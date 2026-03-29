package service

import (
	"context"
	"fmt"
	"time"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	apperr "github.com/georgifotev1/nuvelaone-api/internal/errors"
	"github.com/georgifotev1/nuvelaone-api/internal/repository"
	"github.com/georgifotev1/nuvelaone-api/internal/txmanager"
	"github.com/segmentio/ksuid"
)

type EventService interface {
	Create(ctx context.Context, tenantID string, req domain.EventRequest) (*domain.Event, error)
	Update(ctx context.Context, tenantID string, eventID string, req domain.EventUpdateRequest) (*domain.Event, error)
}

type eventService struct {
	eventRepo    repository.EventRepository
	serviceRepo  repository.ServiceRepository
	customerRepo repository.CustomerRepository
	tx           txmanager.TxManager
}

func NewEventService(eventRepo repository.EventRepository, serviceRepo repository.ServiceRepository, customerRepo repository.CustomerRepository, tx txmanager.TxManager) EventService {
	return &eventService{eventRepo: eventRepo, serviceRepo: serviceRepo, customerRepo: customerRepo, tx: tx}
}

func (s *eventService) Create(ctx context.Context, tenantID string, req domain.EventRequest) (*domain.Event, error) {
	var event *domain.Event

	err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		_, err := s.serviceRepo.GetByID(ctx, tenantID, req.ServiceID)
		if err != nil {
			return fmt.Errorf("eventService.Create get service: %w", err)
		}

		_, err = s.customerRepo.GetByID(ctx, tenantID, req.CustomerID)
		if err != nil {
			return fmt.Errorf("eventService.Create get customer: %w", err)
		}

		available, err := s.eventRepo.CheckUserAvailability(ctx, req.UserID, req.StartTime, req.EndTime, "")
		if err != nil {
			return fmt.Errorf("eventService.Create check availability: %w", err)
		}
		if !available {
			return apperr.Conflict("user is not available at this time")
		}

		now := time.Now()
		event = &domain.Event{
			ID:         ksuid.New().String(),
			CustomerID: req.CustomerID,
			ServiceID:  req.ServiceID,
			UserID:     req.UserID,
			TenantID:   tenantID,
			StartTime:  req.StartTime,
			EndTime:    req.EndTime,
			Status:     "pending",
			Notes:      req.Notes,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		if err := s.eventRepo.Create(ctx, event); err != nil {
			return fmt.Errorf("eventService.Create repo: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (s *eventService) Update(ctx context.Context, tenantID string, eventID string, req domain.EventUpdateRequest) (*domain.Event, error) {
	var updated *domain.Event

	err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		event, err := s.eventRepo.GetByID(ctx, tenantID, eventID)
		if err != nil {
			return fmt.Errorf("eventService.Update get event: %w", err)
		}

		if req.CustomerID != "" {
			_, err := s.customerRepo.GetByID(ctx, tenantID, req.CustomerID)
			if err != nil {
				return fmt.Errorf("eventService.Update get customer: %w", err)
			}
			event.CustomerID = req.CustomerID
		}

		if req.ServiceID != "" {
			_, err := s.serviceRepo.GetByID(ctx, tenantID, req.ServiceID)
			if err != nil {
				return fmt.Errorf("eventService.Update get service: %w", err)
			}
			event.ServiceID = req.ServiceID
		}

		checkAvailability := false
		newUserID := event.UserID
		newStartTime := event.StartTime
		newEndTime := event.EndTime

		if req.UserID != "" && req.UserID != event.UserID {
			newUserID = req.UserID
			checkAvailability = true
		}

		if !req.StartTime.IsZero() && req.StartTime != event.StartTime {
			newStartTime = req.StartTime
			checkAvailability = true
		}

		if !req.EndTime.IsZero() && req.EndTime != event.EndTime {
			newEndTime = req.EndTime
			checkAvailability = true
		}

		if checkAvailability {
			available, err := s.eventRepo.CheckUserAvailability(ctx, newUserID, newStartTime, newEndTime, event.ID)
			if err != nil {
				return fmt.Errorf("eventService.Update check availability: %w", err)
			}
			if !available {
				return apperr.Conflict("user is not available at this time")
			}
			event.UserID = newUserID
			event.StartTime = newStartTime
			event.EndTime = newEndTime
		}

		if req.Status != "" {
			event.Status = req.Status
		}

		if req.Notes != "" {
			event.Notes = req.Notes
		}

		event.UpdatedAt = time.Now()

		if err := s.eventRepo.Update(ctx, event); err != nil {
			return fmt.Errorf("eventService.Update repo: %w", err)
		}

		updated = event
		return nil
	})
	if err != nil {
		return nil, err
	}

	return updated, nil
}
