package application

import (
	"context"
	"errors"
	"testing"
)

type databaseStub struct {
	err error
}

func (s databaseStub) PingContext(context.Context) error {
	return s.err
}

func TestReadinessReportsDatabaseFailure(t *testing.T) {
	service := NewService(databaseStub{err: errors.New("database offline")})

	status, err := service.Readiness(context.Background())
	if err == nil {
		t.Fatal("expected database error")
	}
	if status.Status != "unavailable" {
		t.Fatalf("status = %q, want %q", status.Status, "unavailable")
	}
}

func TestReadinessReportsHealthyDatabase(t *testing.T) {
	service := NewService(databaseStub{})

	status, err := service.Readiness(context.Background())
	if err != nil {
		t.Fatalf("Readiness() error = %v", err)
	}
	if status.Components["database"] != "ok" {
		t.Fatalf("database component = %q, want %q", status.Components["database"], "ok")
	}
}
