package logging

import (
	"log/slog"
	"os"
)

func New(environment string) *slog.Logger {
	level := slog.LevelInfo
	if environment == "local" {
		level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(handler).With("service", "body-fitness-api", "environment", environment)
}
