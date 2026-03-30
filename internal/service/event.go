package service

import (
	"context"
	"fmt"
	"time"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	apperr "github.com/georgifotev1/nuvelaone-api/internal/errors"
	"github.com/georgifotev1/nuvelaone-api/internal/repository"
	"github.com/georgifotev1/nuvelaone-api/internal/txmanager"
	"github.com/georgifotev1/nuvelaone-api/pkg/timeutil"
	"github.com/segmentio/ksuid"
)

type EventService interface {
	Create(ctx context.Context, tenantID string, req domain.EventRequest) (*domain.Event, error)
	CreateForCustomer(ctx context.Context, tenantID, customerID string, req domain.EventRequest) (*domain.Event, error)
	Update(ctx context.Context, tenantID string, eventID string, req domain.EventUpdateRequest) (*domain.Event, error)
	List(ctx context.Context, tenantID string, filter domain.EventListFilter) ([]domain.Event, error)
	ListByCustomer(ctx context.Context, tenantID, customerID string) ([]domain.Event, error)
	GetTimeslots(ctx context.Context, tenantID string, req domain.TimeslotRequest) ([]string, error)
}

type eventService struct {
	eventRepo    repository.EventRepository
	serviceRepo  repository.ServiceRepository
	customerRepo repository.CustomerRepository
	userRepo     repository.UserRepository
	tenantRepo   repository.TenantRepository
	tx           txmanager.TxManager
}

func NewEventService(eventRepo repository.EventRepository, serviceRepo repository.ServiceRepository, customerRepo repository.CustomerRepository, userRepo repository.UserRepository, tenantRepo repository.TenantRepository, tx txmanager.TxManager) EventService {
	return &eventService{eventRepo: eventRepo, serviceRepo: serviceRepo, customerRepo: customerRepo, userRepo: userRepo, tenantRepo: tenantRepo, tx: tx}
}

func (s *eventService) Create(ctx context.Context, tenantID string, req domain.EventRequest) (*domain.Event, error) {
	if req.EndTime.Before(req.StartTime) || req.EndTime.Equal(req.StartTime) {
		return nil, apperr.Validation("end_time must be after start_time")
	}

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

		_, err = s.userRepo.GetByID(ctx, req.UserID)
		if err != nil {
			return fmt.Errorf("eventService.Create get user: %w", err)
		}

		available, err := s.eventRepo.CheckUserAvailability(ctx, tenantID, req.UserID, req.StartTime, req.EndTime, "")
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

func (s *eventService) CreateForCustomer(ctx context.Context, tenantID, customerID string, req domain.EventRequest) (*domain.Event, error) {
	if req.EndTime.Before(req.StartTime) || req.EndTime.Equal(req.StartTime) {
		return nil, apperr.Validation("end_time must be after start_time")
	}

	req.CustomerID = customerID

	var event *domain.Event

	err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		_, err := s.serviceRepo.GetByID(ctx, tenantID, req.ServiceID)
		if err != nil {
			return fmt.Errorf("eventService.CreateForCustomer get service: %w", err)
		}

		_, err = s.customerRepo.GetByID(ctx, tenantID, customerID)
		if err != nil {
			return fmt.Errorf("eventService.CreateForCustomer get customer: %w", err)
		}

		_, err = s.userRepo.GetByID(ctx, req.UserID)
		if err != nil {
			return fmt.Errorf("eventService.CreateForCustomer get user: %w", err)
		}

		available, err := s.eventRepo.CheckUserAvailability(ctx, tenantID, req.UserID, req.StartTime, req.EndTime, "")
		if err != nil {
			return fmt.Errorf("eventService.CreateForCustomer check availability: %w", err)
		}
		if !available {
			return apperr.Conflict("user is not available at this time")
		}

		now := time.Now()
		event = &domain.Event{
			ID:         ksuid.New().String(),
			CustomerID: customerID,
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
			return fmt.Errorf("eventService.CreateForCustomer repo: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (s *eventService) Update(ctx context.Context, tenantID string, eventID string, req domain.EventUpdateRequest) (*domain.Event, error) {
	newStartTime := req.StartTime
	newEndTime := req.EndTime

	if !newStartTime.IsZero() && !newEndTime.IsZero() {
		if newEndTime.Before(newStartTime) || newEndTime.Equal(newStartTime) {
			return nil, apperr.Validation("end_time must be after start_time")
		}
	}

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
		currentUserID := event.UserID
		currentStartTime := event.StartTime
		currentEndTime := event.EndTime

		if req.UserID != "" && req.UserID != event.UserID {
			_, err := s.userRepo.GetByID(ctx, req.UserID)
			if err != nil {
				return fmt.Errorf("eventService.Update get user: %w", err)
			}
			currentUserID = req.UserID
			checkAvailability = true
		}

		if !req.StartTime.IsZero() && req.StartTime != event.StartTime {
			currentStartTime = req.StartTime
			checkAvailability = true
		}

		if !req.EndTime.IsZero() && req.EndTime != event.EndTime {
			currentEndTime = req.EndTime
			checkAvailability = true
		}

		if checkAvailability {
			available, err := s.eventRepo.CheckUserAvailability(ctx, tenantID, currentUserID, currentStartTime, currentEndTime, event.ID)
			if err != nil {
				return fmt.Errorf("eventService.Update check availability: %w", err)
			}
			if !available {
				return apperr.Conflict("user is not available at this time")
			}
			event.UserID = currentUserID
			event.StartTime = currentStartTime
			event.EndTime = currentEndTime
		}

		if req.Status != "" {
			if !isValidStatus(req.Status) {
				return apperr.Validation("invalid status, must be one of: pending, confirmed, cancelled, completed")
			}
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

func (s *eventService) List(ctx context.Context, tenantID string, filter domain.EventListFilter) ([]domain.Event, error) {
	startDate, err := time.Parse(timeutil.DateOnly, filter.StartDate)
	if err != nil {
		return nil, apperr.Validation("invalid startDate format, use YYYY-MM-DD")
	}

	endDate, err := time.Parse(timeutil.DateOnly, filter.EndDate)
	if err != nil {
		return nil, apperr.Validation("invalid endDate format, use YYYY-MM-DD")
	}

	startTime := startDate
	endTime := endDate.Add(24 * time.Hour)

	events, err := s.eventRepo.List(ctx, tenantID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("eventService.List: %w", err)
	}

	return events, nil
}

func (s *eventService) ListByCustomer(ctx context.Context, tenantID, customerID string) ([]domain.Event, error) {
	events, err := s.eventRepo.ListByCustomer(ctx, tenantID, customerID)
	if err != nil {
		return nil, fmt.Errorf("eventService.ListByCustomer: %w", err)
	}

	return events, nil
}

func isValidStatus(status string) bool {
	switch status {
	case "pending", "confirmed", "cancelled", "completed":
		return true
	}
	return false
}

func (s *eventService) GetTimeslots(ctx context.Context, tenantID string, req domain.TimeslotRequest) ([]string, error) {
	svc, err := s.serviceRepo.GetByID(ctx, tenantID, req.ServiceID)
	if err != nil {
		return nil, fmt.Errorf("eventService.GetTimeslots get service: %w", err)
	}

	_, err = s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("eventService.GetTimeslots get user: %w", err)
	}

	workingHours, err := s.tenantRepo.GetWorkingHours(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("eventService.GetTimeslots get working hours: %w", err)
	}

	date, err := time.Parse(timeutil.DateOnly, req.Date)
	if err != nil {
		return nil, apperr.Validation("invalid date format, use YYYY-MM-DD")
	}

	dayOfWeek := int(date.Weekday())
	var dayHours *domain.WorkingHours
	for _, wh := range workingHours {
		if wh.DayOfWeek == dayOfWeek {
			dayHours = &wh
			break
		}
	}

	if dayHours == nil || dayHours.IsClosed {
		return []string{}, nil
	}

	tenant := domain.TenantFromContext(ctx)
	loc, err := time.LoadLocation(tenant.Timezone)
	if err != nil {
		return nil, fmt.Errorf("eventService.GetTimeslots load location: %w", err)
	}
	date = date.In(loc)

	opensAt, err := time.Parse(timeutil.TimeOnly, dayHours.OpensAt)
	if err != nil {
		return nil, fmt.Errorf("eventService.GetTimeslots parse opens_at: %w", err)
	}
	closesAt, err := time.Parse(timeutil.TimeOnly, dayHours.ClosesAt)
	if err != nil {
		return nil, fmt.Errorf("eventService.GetTimeslots parse closes_at: %w", err)
	}

	dayStart := time.Date(date.Year(), date.Month(), date.Day(), opensAt.Hour(), opensAt.Minute(), 0, 0, loc)
	dayEnd := time.Date(date.Year(), date.Month(), date.Day(), closesAt.Hour(), closesAt.Minute(), 0, 0, loc)

	bookings, err := s.eventRepo.ListByUserAndDateRange(ctx, tenantID, req.UserID, dayStart, dayEnd)
	if err != nil {
		return nil, fmt.Errorf("eventService.GetTimeslots get bookings: %w", err)
	}

	serviceDuration := time.Duration(svc.Duration) * time.Minute
	bufferDuration := time.Duration(svc.Buffer) * time.Minute
	totalSlotDuration := serviceDuration + bufferDuration

	type blockedPeriod struct {
		start time.Time
		end   time.Time
	}
	var blockedPeriods []blockedPeriod
	for _, booking := range bookings {
		eventEnd := booking.EndTime
		if svc.Buffer > 0 {
			eventEnd = eventEnd.Add(time.Duration(svc.Buffer) * time.Minute)
		}
		blockedPeriods = append(blockedPeriods, blockedPeriod{
			start: booking.StartTime,
			end:   eventEnd,
		})
	}

	slotInterval := 15 * time.Minute
	var timeslots []string

	for currentSlotStart := dayStart; ; currentSlotStart = currentSlotStart.Add(slotInterval) {
		currentSlotEnd := currentSlotStart.Add(totalSlotDuration)

		if currentSlotEnd.After(dayEnd) {
			break
		}

		isAvailable := true
		for _, blocked := range blockedPeriods {
			if currentSlotStart.Before(blocked.end) && currentSlotEnd.After(blocked.start) {
				isAvailable = false
				break
			}
		}

		if isAvailable {
			timeslots = append(timeslots, currentSlotStart.Format(timeutil.TimeOnly))
		}
	}

	return timeslots, nil
}
