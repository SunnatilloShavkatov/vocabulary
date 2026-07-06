package notificationservice

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidWordAddedEvent = errors.New("invalid WordAdded event")
)

type WordAddedEvent struct {
	WordID    string    `json:"word_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type ScheduleItem struct {
	WordID       string    `json:"word_id"`
	UserID       string    `json:"user_id"`
	RemindAt     time.Time `json:"remind_at"`
	IntervalDays int       `json:"interval_days"`
}

type NotificationRepository interface {
	SaveSchedules(ctx context.Context, schedules []ScheduleItem) error
}

type NotificationService struct {
	repo NotificationRepository
}

func NewNotificationService(repo NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) BuildSRSPlan(event WordAddedEvent) ([]ScheduleItem, error) {
	wordID := strings.TrimSpace(event.WordID)
	userID := strings.TrimSpace(event.UserID)
	if wordID == "" || userID == "" {
		return nil, ErrInvalidWordAddedEvent
	}

	base := event.CreatedAt.UTC()
	if base.IsZero() {
		base = time.Now().UTC()
	}

	days := []int{1, 3, 7, 30}
	schedules := make([]ScheduleItem, 0, len(days))
	for _, d := range days {
		schedules = append(schedules, ScheduleItem{
			WordID:       wordID,
			UserID:       userID,
			RemindAt:     base.AddDate(0, 0, d),
			IntervalDays: d,
		})
	}

	return schedules, nil
}

func (s *NotificationService) OnWordAdded(ctx context.Context, event WordAddedEvent) ([]ScheduleItem, error) {
	schedules, err := s.BuildSRSPlan(event)
	if err != nil {
		return nil, err
	}
	if s.repo != nil {
		if err := s.repo.SaveSchedules(ctx, schedules); err != nil {
			return nil, err
		}
	}
	return schedules, nil
}
