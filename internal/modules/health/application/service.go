package application

import (
	"context"
	"time"
)

type DatabaseChecker interface {
	PingContext(ctx context.Context) error
}

type Service struct {
	database DatabaseChecker
	now      func() time.Time
}

type Status struct {
	Status     string            `json:"status"`
	CheckedAt  time.Time         `json:"checked_at"`
	Components map[string]string `json:"components,omitempty"`
}

func NewService(database DatabaseChecker) Service {
	return Service{
		database: database,
		now:      time.Now,
	}
}

func (s Service) Liveness() Status {
	return Status{
		Status:    "ok",
		CheckedAt: s.now().UTC(),
	}
}

func (s Service) Readiness(ctx context.Context) (Status, error) {
	if err := s.database.PingContext(ctx); err != nil {
		return Status{
			Status:     "unavailable",
			CheckedAt:  s.now().UTC(),
			Components: map[string]string{"database": "unavailable"},
		}, err
	}

	return Status{
		Status:     "ok",
		CheckedAt:  s.now().UTC(),
		Components: map[string]string{"database": "ok"},
	}, nil
}
