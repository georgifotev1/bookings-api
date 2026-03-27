package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/georgifotev1/nuvelaone-api/internal/repository"
	"github.com/georgifotev1/nuvelaone-api/pkg/timeutil"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
)

type TenantService interface {
	GetMyTenant(ctx context.Context, tenantID string) (*domain.Tenant, []domain.WorkingHours, error)
	Update(ctx context.Context, tenantID string, req domain.UpdateTenantRequest) (*domain.Tenant, error)
}

type tenantService struct {
	repo   repository.TenantRepository
	logger *zap.SugaredLogger
}

func NewTenantService(repo repository.TenantRepository, logger *zap.SugaredLogger) TenantService {
	return &tenantService{
		repo:   repo,
		logger: logger,
	}
}

func (s *tenantService) GetMyTenant(ctx context.Context, tenantID string) (*domain.Tenant, []domain.WorkingHours, error) {
	tenant, err := s.repo.GetByID(ctx, tenantID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, fmt.Errorf("tenantService.GetMyTenant: %w", err)
	}

	hours, err := s.repo.GetWorkingHours(ctx, tenantID)
	if err != nil {
		return nil, nil, fmt.Errorf("tenantService.GetMyTenant hours: %w", err)
	}

	return tenant, hours, nil
}

func (s *tenantService) Update(ctx context.Context, tenantID string, req domain.UpdateTenantRequest) (*domain.Tenant, error) {
	tenant, err := s.repo.GetByID(ctx, tenantID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("tenantService.Update: %w", err)
	}

	now := time.Now()

	if req.Name != "" {
		tenant.Name = req.Name
		tenant.Slug = domain.NewSlug(req.Name)
	}
	if req.Phone != "" {
		tenant.Phone = req.Phone
	}
	if req.Email != "" {
		tenant.Email = req.Email
	}
	if req.Timezone != "" {
		if _, err := time.LoadLocation(req.Timezone); err != nil {
			return nil, fmt.Errorf("invalid timezone: %w", ErrBadRequest)
		}
		tenant.Timezone = req.Timezone
	}

	tenant.UpdatedAt = now

	if err := s.repo.Update(ctx, tenant); err != nil {
		return nil, fmt.Errorf("tenantService.Update repo: %w", err)
	}

	if len(req.WorkingHours) > 0 {
		loc, err := time.LoadLocation(tenant.Timezone)
		if err != nil {
			return nil, fmt.Errorf("failed to load timezone: %w", err)
		}

		var hours []domain.WorkingHours
		for _, wh := range req.WorkingHours {
			opensAt, err := domain.ParseTimeInLocation(wh.OpensAt, loc)
			if err != nil {
				return nil, fmt.Errorf("invalid opens_at %q: %w", wh.OpensAt, ErrBadRequest)
			}
			closesAt, err := domain.ParseTimeInLocation(wh.ClosesAt, loc)
			if err != nil {
				return nil, fmt.Errorf("invalid closes_at %q: %w", wh.ClosesAt, ErrBadRequest)
			}

			hours = append(hours, domain.WorkingHours{
				ID:        ksuid.New().String(),
				TenantID:  tenantID,
				DayOfWeek: wh.DayOfWeek,
				OpensAt:   opensAt.UTC().Format(timeutil.TimeOnly),
				ClosesAt:  closesAt.UTC().Format(timeutil.TimeOnly),
				IsClosed:  wh.IsClosed,
			})
		}
		if err := s.repo.UpsertWorkingHours(ctx, tenantID, hours); err != nil {
			return nil, fmt.Errorf("tenantService.Update hours: %w", err)
		}
	}

	return tenant, nil
}
