package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/georgifotev1/nuvelaone-api/internal/repository"
	"github.com/georgifotev1/nuvelaone-api/internal/tasks"
	"github.com/georgifotev1/nuvelaone-api/internal/txmanager"
	"github.com/georgifotev1/nuvelaone-api/pkg/auth"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
)

var (
	ErrNotProTier          = errors.New("only pro tier owners can invite users")
	ErrInvitationExists    = errors.New("invitation already exists for this email")
	ErrUserAlreadyInTenant = errors.New("user already exists in this tenant")
	ErrInvitationNotFound  = errors.New("invitation not found")
	ErrInvitationExpired   = errors.New("invitation has expired")
	ErrInvitationAccepted  = errors.New("invitation already accepted")
)

type InvitationConfig struct {
	InvitationExpiry time.Duration
	AppBaseURL       string
}

type InvitationService interface {
	CreateInvitation(ctx context.Context, invitedByUserID, tenantID string, req domain.CreateInvitationRequest) (*domain.UserInvitation, error)
	AcceptInvitation(ctx context.Context, req domain.AcceptInvitationRequest) (*domain.User, error)
	ResendInvitation(ctx context.Context, invitationID, invitedByUserID string) error
	ListInvitations(ctx context.Context, tenantID string) ([]domain.UserInvitation, error)
	RevokeInvitation(ctx context.Context, invitationID, tenantID string) error
}

type invitationService struct {
	invitationRepo repository.InvitationRepository
	userRepo       repository.UserRepository
	tenantRepo     repository.TenantRepository
	txManager      txmanager.TxManager
	taskClient     TaskEnqueuer
	logger         *zap.SugaredLogger
	cfg            InvitationConfig
}

func NewInvitationService(
	invitationRepo repository.InvitationRepository,
	userRepo repository.UserRepository,
	tenantRepo repository.TenantRepository,
	txManager txmanager.TxManager,
	taskClient TaskEnqueuer,
	logger *zap.SugaredLogger,
	cfg InvitationConfig,
) InvitationService {
	return &invitationService{
		invitationRepo: invitationRepo,
		userRepo:       userRepo,
		tenantRepo:     tenantRepo,
		txManager:      txManager,
		taskClient:     taskClient,
		logger:         logger,
		cfg:            cfg,
	}
}

func (s *invitationService) CreateInvitation(ctx context.Context, invitedByUserID, tenantID string, req domain.CreateInvitationRequest) (*domain.UserInvitation, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("invitationService.CreateInvitation get tenant: %w", err)
	}

	if tenant.Tier != domain.TierPro {
		return nil, ErrNotProTier
	}

	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("invitationService.CreateInvitation get user: %w", err)
	}
	if existingUser != nil && existingUser.TenantID == tenantID {
		return nil, ErrUserAlreadyInTenant
	}

	existingInv, err := s.invitationRepo.GetByEmailAndTenant(ctx, req.Email, tenantID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("invitationService.CreateInvitation get existing invitation: %w", err)
	}
	if existingInv != nil {
		return nil, ErrInvitationExists
	}

	token, err := generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("invitationService.CreateInvitation generate token: %w", err)
	}

	role := req.Role
	if role == "" {
		role = string(domain.RoleMember)
	}

	invitation := &domain.UserInvitation{
		ID:        ksuid.New().String(),
		Email:     req.Email,
		Name:      req.Name,
		Phone:     req.Phone,
		Token:     token,
		Role:      role,
		InvitedBy: invitedByUserID,
		ExpiresAt: time.Now().Add(s.cfg.InvitationExpiry),
		TenantID:  tenantID,
		Accepted:  false,
		CreatedAt: time.Now(),
	}

	if err := s.invitationRepo.Create(ctx, invitation); err != nil {
		return nil, fmt.Errorf("invitationService.CreateInvitation create: %w", err)
	}

	if s.taskClient != nil {
		task, err := tasks.NewInvitationEmailTask(tasks.InvitationEmailPayload{
			InvitedByName: invitedByUserID,
			Email:         invitation.Email,
			Name:          invitation.Name,
			Token:         invitation.Token,
			TenantName:    tenant.Name,
			AcceptURL:     fmt.Sprintf("%s/accept-invitation?token=%s", s.cfg.AppBaseURL, invitation.Token),
		})
		if err != nil {
			s.logger.Warnw("failed to create invitation email task", "error", err, "email", invitation.Email)
		} else {
			if _, err := s.taskClient.Enqueue(task); err != nil {
				s.logger.Warnw("failed to enqueue invitation email", "error", err, "email", invitation.Email)
			}
		}
	}

	return invitation, nil
}

func (s *invitationService) AcceptInvitation(ctx context.Context, req domain.AcceptInvitationRequest) (*domain.User, error) {
	invitation, err := s.invitationRepo.GetByToken(ctx, req.Token)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvitationNotFound
		}
		return nil, fmt.Errorf("invitationService.AcceptInvitation get by token: %w", err)
	}

	if invitation.Accepted {
		return nil, ErrInvitationAccepted
	}

	if time.Now().After(invitation.ExpiresAt) {
		return nil, ErrInvitationExpired
	}

	if req.Password != req.PasswordConfirm {
		return nil, fmt.Errorf("passwords do not match")
	}

	hashed, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("invitationService.AcceptInvitation hash: %w", err)
	}

	user := &domain.User{
		ID:        ksuid.New().String(),
		Email:     invitation.Email,
		Password:  string(hashed),
		Name:      invitation.Name,
		Phone:     invitation.Phone,
		TenantID:  invitation.TenantID,
		Role:      domain.Role(invitation.Role),
		Verified:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.txManager.WithTx(ctx, func(ctx context.Context) error {
		if err := s.userRepo.Create(ctx, user); err != nil {
			return fmt.Errorf("invitationService.AcceptInvitation create user: %w", err)
		}

		invitation.Accepted = true
		if err := s.invitationRepo.Update(ctx, invitation); err != nil {
			s.logger.Warnw("failed to mark invitation as accepted", "error", err, "invitationID", invitation.ID)
			return err
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("invitationService.AcceptInvitation tx: %w", err)
	}

	return user, nil
}

func (s *invitationService) ResendInvitation(ctx context.Context, invitationID, invitedByUserID string) error {
	invitation, err := s.invitationRepo.GetByID(ctx, invitationID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrInvitationNotFound
		}
		return fmt.Errorf("invitationService.ResendInvitation get: %w", err)
	}

	if invitation.Accepted {
		return ErrInvitationAccepted
	}

	inviter, err := s.userRepo.GetByID(ctx, invitedByUserID)
	if err != nil {
		return fmt.Errorf("invitationService.ResendInvitation get inviter: %w", err)
	}

	if inviter.TenantID != invitation.TenantID {
		return ErrInvitationNotFound
	}

	tenant, err := s.tenantRepo.GetByID(ctx, invitation.TenantID)
	if err != nil {
		return fmt.Errorf("invitationService.ResendInvitation get tenant: %w", err)
	}

	if tenant.Tier != domain.TierPro {
		return ErrNotProTier
	}

	token, err := generateSecureToken()
	if err != nil {
		return fmt.Errorf("invitationService.ResendInvitation generate token: %w", err)
	}

	invitation.Token = token
	invitation.ExpiresAt = time.Now().Add(s.cfg.InvitationExpiry)

	if err := s.invitationRepo.Update(ctx, invitation); err != nil {
		return fmt.Errorf("invitationService.ResendInvitation update: %w", err)
	}

	if s.taskClient != nil {
		task, err := tasks.NewInvitationEmailTask(tasks.InvitationEmailPayload{
			InvitedByName: invitedByUserID,
			Email:         invitation.Email,
			Name:          invitation.Name,
			Token:         invitation.Token,
			TenantName:    tenant.Name,
			AcceptURL:     fmt.Sprintf("%s/accept-invitation?token=%s", s.cfg.AppBaseURL, invitation.Token),
		})
		if err != nil {
			s.logger.Warnw("failed to create invitation email task", "error", err, "email", invitation.Email)
		} else {
			if _, err := s.taskClient.Enqueue(task); err != nil {
				s.logger.Warnw("failed to enqueue invitation email", "error", err, "email", invitation.Email)
			}
		}
	}

	return nil
}

func (s *invitationService) ListInvitations(ctx context.Context, tenantID string) ([]domain.UserInvitation, error) {
	invitations, err := s.invitationRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("invitationService.ListInvitations: %w", err)
	}
	return invitations, nil
}

func (s *invitationService) RevokeInvitation(ctx context.Context, invitationID, tenantID string) error {
	invitation, err := s.invitationRepo.GetByID(ctx, invitationID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrInvitationNotFound
		}
		return fmt.Errorf("invitationService.RevokeInvitation get: %w", err)
	}

	if invitation.TenantID != tenantID {
		return ErrInvitationNotFound
	}

	if err := s.invitationRepo.Delete(ctx, invitationID); err != nil {
		return fmt.Errorf("invitationService.RevokeInvitation delete: %w", err)
	}

	return nil
}

func generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
