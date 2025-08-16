package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a participant who can be scheduled for meetings
type User struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CalendarEvent represents a scheduled meeting or event
type CalendarEvent struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	UserID    string    `json:"userId" gorm:"index"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ScheduleRequest represents the input for scheduling a new meeting
type ScheduleRequest struct {
	ParticipantIDs  []string  `json:"participantIds"`
	DurationMinutes int       `json:"durationMinutes"`
	TimeRange       TimeRange `json:"timeRange"`
	Title           string    `json:"title,omitempty"`
}

// TimeRange represents a start and end time window
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ScheduleResponse represents the output of a successful scheduling request
type ScheduleResponse struct {
	MeetingID      string    `json:"meetingId"`
	Title          string    `json:"title"`
	ParticipantIDs []string  `json:"participantIds"`
	StartTime      time.Time `json:"startTime"`
	EndTime        time.Time `json:"endTime"`
}

// NewUser creates a new user with the given name
func NewUser(name string) *User {
	return &User{
		ID:        uuid.New().String(),
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewCalendarEvent creates a new calendar event
func NewCalendarEvent(title string, startTime, endTime time.Time, userID string) *CalendarEvent {
	return &CalendarEvent{
		ID:        uuid.New().String(),
		Title:     title,
		StartTime: startTime,
		EndTime:   endTime,
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
