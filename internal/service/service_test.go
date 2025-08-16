package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/meeting-scheduler/internal/domain"
)

// MockRepository implements the Repository interface for testing
type MockRepository struct {
	users  map[string]*domain.User
	events map[string][]domain.CalendarEvent
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		users:  make(map[string]*domain.User),
		events: make(map[string][]domain.CalendarEvent),
	}
}

func (m *MockRepository) GetUser(ctx context.Context, id string) (*domain.User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (m *MockRepository) GetUserEvents(ctx context.Context, userID string, start, end time.Time) ([]domain.CalendarEvent, error) {
	events := m.events[userID]
	var filtered []domain.CalendarEvent
	for _, event := range events {
		if (event.StartTime.Equal(start) || event.StartTime.After(start)) &&
			(event.EndTime.Equal(end) || event.EndTime.Before(end)) {
			filtered = append(filtered, event)
		}
	}
	return filtered, nil
}

func (m *MockRepository) CreateEvent(ctx context.Context, event *domain.CalendarEvent) error {
	m.events[event.UserID] = append(m.events[event.UserID], *event)
	return nil
}

func TestSchedule(t *testing.T) {
	// Helper function to create time
	parseTime := func(s string) time.Time {
		t, _ := time.Parse(time.RFC3339, s)
		return t
	}

	// Create mock repository with test data
	repo := NewMockRepository()

	// Add test users
	users := []*domain.User{
		{ID: "user1", Name: "Alice"},
		{ID: "user2", Name: "Bob"},
	}
	for _, user := range users {
		repo.users[user.ID] = user
	}

	// Create service
	svc := NewService(repo)

	tests := []struct {
		name         string
		request      domain.ScheduleRequest
		setupEvents  func()
		expectError  bool
		errorType    error
		validateResp func(*testing.T, *domain.ScheduleResponse)
	}{
		{
			name: "Successful scheduling",
			request: domain.ScheduleRequest{
				ParticipantIDs:  []string{"user1", "user2"},
				DurationMinutes: 60,
				TimeRange: domain.TimeRange{
					Start: parseTime("2025-09-01T09:00:00Z"),
					End:   parseTime("2025-09-01T17:00:00Z"),
				},
			},
			setupEvents: func() {
				repo.events = make(map[string][]domain.CalendarEvent)
			},
			expectError: false,
			validateResp: func(t *testing.T, resp *domain.ScheduleResponse) {
				if resp == nil {
					t.Fatal("Expected response but got nil")
				}
				if resp.StartTime.Hour() != 9 {
					t.Errorf("Expected meeting to start at 9 AM, got %v", resp.StartTime.Hour())
				}
			},
		},
		{
			name: "No available slots",
			request: domain.ScheduleRequest{
				ParticipantIDs:  []string{"user1", "user2"},
				DurationMinutes: 60,
				TimeRange: domain.TimeRange{
					Start: parseTime("2025-09-01T09:00:00Z"),
					End:   parseTime("2025-09-01T11:00:00Z"),
				},
			},
			setupEvents: func() {
				repo.events = map[string][]domain.CalendarEvent{
					"user1": {
						{
							StartTime: parseTime("2025-09-01T09:00:00Z"),
							EndTime:   parseTime("2025-09-01T10:00:00Z"),
						},
					},
					"user2": {
						{
							StartTime: parseTime("2025-09-01T10:00:00Z"),
							EndTime:   parseTime("2025-09-01T11:00:00Z"),
						},
					},
				}
			},
			expectError: true,
			errorType:   ErrNoAvailableSlot,
		},
		{
			name: "Invalid user",
			request: domain.ScheduleRequest{
				ParticipantIDs:  []string{"user1", "nonexistent"},
				DurationMinutes: 60,
				TimeRange: domain.TimeRange{
					Start: parseTime("2025-09-01T09:00:00Z"),
					End:   parseTime("2025-09-01T17:00:00Z"),
				},
			},
			setupEvents: func() {
				repo.events = make(map[string][]domain.CalendarEvent)
			},
			expectError: true,
			errorType:   ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test events
			tt.setupEvents()

			// Execute test
			resp, err := svc.Schedule(context.Background(), tt.request)

			// Validate results
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if err != tt.errorType {
					t.Errorf("Expected error %v but got %v", tt.errorType, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			tt.validateResp(t, resp)
		})
	}
}

func TestGetUserCalendar(t *testing.T) {
	// Create mock repository with test data
	repo := NewMockRepository()

	// Add test user
	user := &domain.User{ID: "user1", Name: "Alice"}
	repo.users[user.ID] = user

	// Add test events
	startTime := time.Now()
	events := []domain.CalendarEvent{
		{
			ID:        "event1",
			Title:     "Meeting 1",
			StartTime: startTime,
			EndTime:   startTime.Add(time.Hour),
			UserID:    user.ID,
		},
		{
			ID:        "event2",
			Title:     "Meeting 2",
			StartTime: startTime.Add(2 * time.Hour),
			EndTime:   startTime.Add(3 * time.Hour),
			UserID:    user.ID,
		},
	}
	repo.events[user.ID] = events

	// Create service
	svc := NewService(repo)

	tests := []struct {
		name        string
		userID      string
		start       time.Time
		end         time.Time
		expectError bool
		errorType   error
		eventCount  int
	}{
		{
			name:        "Get all events",
			userID:      user.ID,
			start:       startTime,
			end:         startTime.Add(4 * time.Hour),
			expectError: false,
			eventCount:  2,
		},
		{
			name:        "Get partial events",
			userID:      user.ID,
			start:       startTime,
			end:         startTime.Add(time.Hour),
			expectError: false,
			eventCount:  1,
		},
		{
			name:        "Invalid user",
			userID:      "nonexistent",
			start:       startTime,
			end:         startTime.Add(time.Hour),
			expectError: true,
			errorType:   ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events, err := svc.GetUserCalendar(context.Background(), tt.userID, tt.start, tt.end)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if err != tt.errorType {
					t.Errorf("Expected error %v but got %v", tt.errorType, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(events) != tt.eventCount {
				t.Errorf("Expected %d events but got %d", tt.eventCount, len(events))
			}
		})
	}
}

func TestValidateScheduleRequest(t *testing.T) {
	tests := []struct {
		name    string
		request domain.ScheduleRequest
		wantErr bool
	}{
		{
			name: "Valid request",
			request: domain.ScheduleRequest{
				ParticipantIDs:  []string{"user1", "user2"},
				DurationMinutes: 60,
				TimeRange: domain.TimeRange{
					Start: time.Now().Add(time.Hour),
					End:   time.Now().Add(2 * time.Hour),
				},
			},
			wantErr: false,
		},
		{
			name: "Empty participant IDs",
			request: domain.ScheduleRequest{
				ParticipantIDs:  []string{},
				DurationMinutes: 60,
				TimeRange: domain.TimeRange{
					Start: time.Now().Add(time.Hour),
					End:   time.Now().Add(2 * time.Hour),
				},
			},
			wantErr: true,
		},
		{
			name: "Duplicate participant IDs",
			request: domain.ScheduleRequest{
				ParticipantIDs:  []string{"user1", "user1"},
				DurationMinutes: 60,
				TimeRange: domain.TimeRange{
					Start: time.Now().Add(time.Hour),
					End:   time.Now().Add(2 * time.Hour),
				},
			},
			wantErr: true,
		},
		{
			name: "Empty participant ID",
			request: domain.ScheduleRequest{
				ParticipantIDs:  []string{"user1", ""},
				DurationMinutes: 60,
				TimeRange: domain.TimeRange{
					Start: time.Now().Add(time.Hour),
					End:   time.Now().Add(2 * time.Hour),
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid duration - zero",
			request: domain.ScheduleRequest{
				ParticipantIDs:  []string{"user1", "user2"},
				DurationMinutes: 0,
				TimeRange: domain.TimeRange{
					Start: time.Now().Add(time.Hour),
					End:   time.Now().Add(2 * time.Hour),
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid duration - negative",
			request: domain.ScheduleRequest{
				ParticipantIDs:  []string{"user1", "user2"},
				DurationMinutes: -10,
				TimeRange: domain.TimeRange{
					Start: time.Now().Add(time.Hour),
					End:   time.Now().Add(2 * time.Hour),
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid duration - too long",
			request: domain.ScheduleRequest{
				ParticipantIDs:  []string{"user1", "user2"},
				DurationMinutes: 500, // More than 8 hours
				TimeRange: domain.TimeRange{
					Start: time.Now().Add(time.Hour),
					End:   time.Now().Add(2 * time.Hour),
				},
			},
			wantErr: true,
		},
		{
			name: "Start time in the past",
			request: domain.ScheduleRequest{
				ParticipantIDs:  []string{"user1", "user2"},
				DurationMinutes: 60,
				TimeRange: domain.TimeRange{
					Start: time.Now().Add(-time.Hour),
					End:   time.Now().Add(time.Hour),
				},
			},
			wantErr: true,
		},
		{
			name: "End time before start time",
			request: domain.ScheduleRequest{
				ParticipantIDs:  []string{"user1", "user2"},
				DurationMinutes: 60,
				TimeRange: domain.TimeRange{
					Start: time.Now().Add(2 * time.Hour),
					End:   time.Now().Add(time.Hour),
				},
			},
			wantErr: true,
		},
		{
			name: "Duration doesn't fit in time range",
			request: domain.ScheduleRequest{
				ParticipantIDs:  []string{"user1", "user2"},
				DurationMinutes: 120, // 2 hours
				TimeRange: domain.TimeRange{
					Start: time.Now().Add(time.Hour),
					End:   time.Now().Add(time.Hour + 30*time.Minute), // Only 30 minutes
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateScheduleRequest(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateScheduleRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateMeetingID(t *testing.T) {
	// Test that generated IDs are unique
	id1 := generateMeetingID()
	id2 := generateMeetingID()

	if id1 == "" {
		t.Error("Generated meeting ID should not be empty")
	}

	if id2 == "" {
		t.Error("Generated meeting ID should not be empty")
	}

	if id1 == id2 {
		t.Error("Generated meeting IDs should be unique")
	}

	// Test that IDs are valid UUIDs
	_, err1 := uuid.Parse(id1)
	if err1 != nil {
		t.Errorf("Generated meeting ID should be a valid UUID: %v", err1)
	}

	_, err2 := uuid.Parse(id2)
	if err2 != nil {
		t.Errorf("Generated meeting ID should be a valid UUID: %v", err2)
	}
}
