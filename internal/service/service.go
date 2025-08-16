package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/meeting-scheduler/internal/domain"
	"github.com/meeting-scheduler/pkg/algorithm"
)

var (
	ErrInvalidRequest  = errors.New("invalid request parameters")
	ErrNoAvailableSlot = errors.New("no available time slot found for all participants")
	ErrUserNotFound    = errors.New("user not found")
	ErrInternalError   = errors.New("internal server error")
)

// SchedulerService defines the interface for our meeting scheduler
type SchedulerService interface {
	Schedule(ctx context.Context, req domain.ScheduleRequest) (*domain.ScheduleResponse, error)

	GetUserCalendar(ctx context.Context, userID string, start, end time.Time) ([]domain.CalendarEvent, error)
}

// Repository defines the interface for data persistence
type Repository interface {
	GetUser(ctx context.Context, id string) (*domain.User, error)
	GetUserEvents(ctx context.Context, userID string, start, end time.Time) ([]domain.CalendarEvent, error)
	CreateEvent(ctx context.Context, event *domain.CalendarEvent) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) SchedulerService {
	return &service{
		repo: repo,
	}
}

// Schedule implements the core scheduling logic
func (s *service) Schedule(ctx context.Context, req domain.ScheduleRequest) (*domain.ScheduleResponse, error) {
	if err := validateScheduleRequest(req); err != nil {
		return nil, err
	}

	for _, userID := range req.ParticipantIDs {
		if _, err := s.repo.GetUser(ctx, userID); err != nil {
			return nil, ErrUserNotFound
		}
	}

	allEvents := make(map[string][]domain.CalendarEvent)
	for _, userID := range req.ParticipantIDs {
		events, err := s.repo.GetUserEvents(ctx, userID, req.TimeRange.Start, req.TimeRange.End)
		if err != nil {
			return nil, ErrInternalError
		}
		allEvents[userID] = events
	}

	slot, err := algorithm.FindOptimalSlot(req, allEvents)
	if err != nil {
		return nil, ErrInternalError
	}
	if slot == nil {
		return nil, ErrNoAvailableSlot
	}

	meetingID := generateMeetingID()
	meetingTitle := req.Title
	if meetingTitle == "" {
		meetingTitle = "New Meeting"
	}
	for _, userID := range req.ParticipantIDs {
		event := domain.NewCalendarEvent(
			meetingTitle,
			slot.Start,
			slot.End,
			userID,
		)
		if err := s.repo.CreateEvent(ctx, event); err != nil {
			return nil, ErrInternalError
		}
	}

	return &domain.ScheduleResponse{
		MeetingID:      meetingID,
		Title:          meetingTitle,
		ParticipantIDs: req.ParticipantIDs,
		StartTime:      slot.Start,
		EndTime:        slot.End,
	}, nil
}

func (s *service) GetUserCalendar(ctx context.Context, userID string, start, end time.Time) ([]domain.CalendarEvent, error) {
	if _, err := s.repo.GetUser(ctx, userID); err != nil {
		return nil, ErrUserNotFound
	}

	events, err := s.repo.GetUserEvents(ctx, userID, start, end)
	if err != nil {
		return nil, ErrInternalError
	}
	if len(events) == 0 {
		return nil, errors.New("no meetings found for the specified user and time window")
	}
	return events, nil
}

func validateScheduleRequest(req domain.ScheduleRequest) error {
	if len(req.ParticipantIDs) == 0 {
		return errors.New("at least one participant is required")
	}

	// Check for duplicate participant IDs
	participantMap := make(map[string]bool)
	for _, id := range req.ParticipantIDs {
		if id == "" {
			return errors.New("participant ID cannot be empty")
		}
		if participantMap[id] {
			return errors.New("duplicate participant IDs are not allowed")
		}
		participantMap[id] = true
	}

	if req.DurationMinutes <= 0 {
		return errors.New("duration must be greater than 0 minutes")
	}
	if req.DurationMinutes > 480 { // 8 hours max
		return errors.New("duration cannot exceed 8 hours (480 minutes)")
	}

	if req.TimeRange.Start.IsZero() {
		return errors.New("start time is required")
	}
	if req.TimeRange.End.IsZero() {
		return errors.New("end time is required")
	}
	if req.TimeRange.Start.After(req.TimeRange.End) {
		return errors.New("start time must be before end time")
	}

	now := time.Now()
	if req.TimeRange.Start.Before(now) {
		return errors.New("start time cannot be in the past. Please enter a valid future start date.")
	}

	maxFuture := now.AddDate(1, 0, 0)
	if req.TimeRange.End.After(maxFuture) {
		return errors.New("end time cannot be more than 1 year in the future")
	}

	duration := time.Duration(req.DurationMinutes) * time.Minute
	if req.TimeRange.Start.Add(duration).After(req.TimeRange.End) {
		return errors.New("duration does not fit within the specified time range")
	}

	return nil
}

func generateMeetingID() string {
	return uuid.New().String()
}
