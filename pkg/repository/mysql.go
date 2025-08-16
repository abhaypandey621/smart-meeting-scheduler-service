package repository

import (
	"context"
	"time"

	"github.com/meeting-scheduler/internal/domain"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MySQLRepository struct {
	db *gorm.DB
}

// NewMySQLRepository creates a new MySQL repository
func NewMySQLRepository(dsn string) (*MySQLRepository, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto migrate the schema
	err = db.AutoMigrate(&domain.User{}, &domain.CalendarEvent{})
	if err != nil {
		return nil, err
	}

	return &MySQLRepository{
		db: db,
	}, nil
}

// GetUser retrieves a user by ID
func (r *MySQLRepository) GetUser(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	result := r.db.WithContext(ctx).First(&user, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// GetUserEvents retrieves a user's calendar events within a time range
func (r *MySQLRepository) GetUserEvents(ctx context.Context, userID string, start, end time.Time) ([]domain.CalendarEvent, error) {
	var events []domain.CalendarEvent
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND start_time >= ? AND end_time <= ?", userID, start, end).
		Find(&events)
	if result.Error != nil {
		return nil, result.Error
	}
	return events, nil
}

// CreateEvent creates a new calendar event
func (r *MySQLRepository) CreateEvent(ctx context.Context, event *domain.CalendarEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

// CreateUser creates a new user
func (r *MySQLRepository) CreateUser(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// ClearAllData removes all data from the database (useful for testing)
func (r *MySQLRepository) ClearAllData(ctx context.Context) error {
	err := r.db.WithContext(ctx).Exec("DELETE FROM calendar_events").Error
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Exec("DELETE FROM users").Error
}

// SeedTestData populates the database with test data
func (r *MySQLRepository) SeedTestData(ctx context.Context) error {
	// Create test users
	users := []*domain.User{
		domain.NewUser("Alice"),
		domain.NewUser("Bob"),
		domain.NewUser("Charlie"),
	}

	for _, user := range users {
		if err := r.CreateUser(ctx, user); err != nil {
			return err
		}
	}

	// Create some test calendar events
	now := time.Now()
	events := []*domain.CalendarEvent{
		domain.NewCalendarEvent(
			"Team Meeting",
			now.Add(24*time.Hour),
			now.Add(25*time.Hour),
			users[0].ID,
		),
		domain.NewCalendarEvent(
			"Project Review",
			now.Add(26*time.Hour),
			now.Add(27*time.Hour),
			users[1].ID,
		),
		domain.NewCalendarEvent(
			"Client Call",
			now.Add(28*time.Hour),
			now.Add(29*time.Hour),
			users[2].ID,
		),
	}

	for _, event := range events {
		if err := r.CreateEvent(ctx, event); err != nil {
			return err
		}
	}

	return nil
}
