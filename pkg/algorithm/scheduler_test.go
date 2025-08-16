package algorithm

import (
	"testing"
	"time"

	"github.com/meeting-scheduler/internal/domain"
)

func TestFindOptimalSlot(t *testing.T) {
	// Helper function to create time
	parseTime := func(s string) time.Time {
		t, _ := time.Parse(time.RFC3339, s)
		return t
	}

	tests := []struct {
		name           string
		request        domain.ScheduleRequest
		events         map[string][]domain.CalendarEvent
		expectSlot     bool
		expectedStart  string
		expectedEnd    string
	}{
		{
			name: "Simple case - one available slot",
			request: domain.ScheduleRequest{
				ParticipantIDs:  []string{"user1", "user2"},
				DurationMinutes: 60,
				TimeRange: domain.TimeRange{
					Start: parseTime("2024-09-01T09:00:00Z"),
					End:   parseTime("2024-09-01T17:00:00Z"),
				},
			},
			events: map[string][]domain.CalendarEvent{
				"user1": {
					{
						StartTime: parseTime("2024-09-01T13:00:00Z"),
						EndTime:   parseTime("2024-09-01T14:00:00Z"),
					},
				},
				"user2": {
					{
						StartTime: parseTime("2024-09-01T15:00:00Z"),
						EndTime:   parseTime("2024-09-01T16:00:00Z"),
					},
				},
			},
			expectSlot:    true,
			expectedStart: "2024-09-01T09:00:00Z",
			expectedEnd:   "2024-09-01T10:00:00Z",
		},
		{
			name: "No available slots",
			request: domain.ScheduleRequest{
				ParticipantIDs:  []string{"user1", "user2"},
				DurationMinutes: 60,
				TimeRange: domain.TimeRange{
					Start: parseTime("2024-09-01T09:00:00Z"),
					End:   parseTime("2024-09-01T11:00:00Z"),
				},
			},
			events: map[string][]domain.CalendarEvent{
				"user1": {
					{
						StartTime: parseTime("2024-09-01T09:00:00Z"),
						EndTime:   parseTime("2024-09-01T10:00:00Z"),
					},
				},
				"user2": {
					{
						StartTime: parseTime("2024-09-01T10:00:00Z"),
						EndTime:   parseTime("2024-09-01T11:00:00Z"),
					},
				},
			},
			expectSlot: false,
		},
		{
			name: "Prefer working hours",
			request: domain.ScheduleRequest{
				ParticipantIDs:  []string{"user1"},
				DurationMinutes: 60,
				TimeRange: domain.TimeRange{
					Start: parseTime("2024-09-01T07:00:00Z"),
					End:   parseTime("2024-09-01T17:00:00Z"),
				},
			},
			events: map[string][]domain.CalendarEvent{},
			expectSlot:    true,
			expectedStart: "2024-09-01T09:00:00Z",
			expectedEnd:   "2024-09-01T10:00:00Z",
		},
		{
			name: "Prefer earlier slots during working hours",
			request: domain.ScheduleRequest{
				ParticipantIDs:  []string{"user1"},
				DurationMinutes: 60,
				TimeRange: domain.TimeRange{
					Start: parseTime("2024-09-01T09:00:00Z"),
					End:   parseTime("2024-09-01T17:00:00Z"),
				},
			},
			events: map[string][]domain.CalendarEvent{},
			expectSlot:    true,
			expectedStart: "2024-09-01T09:00:00Z",
			expectedEnd:   "2024-09-01T10:00:00Z",
		},
		{
			name: "Respect buffer time",
			request: domain.ScheduleRequest{
				ParticipantIDs:  []string{"user1"},
				DurationMinutes: 60,
				TimeRange: domain.TimeRange{
					Start: parseTime("2024-09-01T09:00:00Z"),
					End:   parseTime("2024-09-01T17:00:00Z"),
				},
			},
			events: map[string][]domain.CalendarEvent{
				"user1": {
					{
						StartTime: parseTime("2024-09-01T11:00:00Z"),
						EndTime:   parseTime("2024-09-01T12:00:00Z"),
					},
				},
			},
			expectSlot:    true,
			expectedStart: "2024-09-01T09:00:00Z",
			expectedEnd:   "2024-09-01T10:00:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slot, err := FindOptimalSlot(tt.request, tt.events)
			
			if !tt.expectSlot {
				if slot != nil {
					t.Errorf("Expected no slot, but got one starting at %v", slot.Start)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if slot == nil {
				t.Fatal("Expected a slot but got nil")
			}

			expectedStart := parseTime(tt.expectedStart)
			expectedEnd := parseTime(tt.expectedEnd)

			if !slot.Start.Equal(expectedStart) {
				t.Errorf("Expected start time %v, got %v", expectedStart, slot.Start)
			}

			if !slot.End.Equal(expectedEnd) {
				t.Errorf("Expected end time %v, got %v", expectedEnd, slot.End)
			}
		})
	}
}

func TestWorkingHoursScore(t *testing.T) {
	tests := []struct {
		name     string
		time     string
		expected float64
	}{
		{
			name:     "Middle of working hours",
			time:     "2024-09-01T13:00:00Z",
			expected: 1.0,
		},
		{
			name:     "Early morning",
			time:     "2024-09-01T07:00:00Z",
			expected: 0.0,
		},
		{
			name:     "Just before working hours",
			time:     "2024-09-01T08:00:00Z",
			expected: 0.5,
		},
		{
			name:     "Just after working hours",
			time:     "2024-09-01T17:00:00Z",
			expected: 0.5,
		},
		{
			name:     "Late evening",
			time:     "2024-09-01T20:00:00Z",
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startTime, _ := time.Parse(time.RFC3339, tt.time)
			slot := TimeSlot{
				Start: startTime,
				End:   startTime.Add(time.Hour),
			}

			score := workingHoursScore(slot)
			if score != tt.expected {
				t.Errorf("Expected score %v, got %v", tt.expected, score)
			}
		})
	}
}
