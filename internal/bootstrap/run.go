package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	healthapp "github.com/4zure09/body-fitness/internal/modules/health/application"
	healthhttp "github.com/4zure09/body-fitness/internal/modules/health/transport/http"
	"github.com/4zure09/body-fitness/internal/platform/config"
	"github.com/4zure09/body-fitness/internal/platform/database"
	"github.com/4zure09/body-fitness/internal/platform/httpapi"
	"github.com/4zure09/body-fitness/internal/platform/server"
)

// Run wires application dependencies and owns their lifecycle.
func Run(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	db, err := database.Open(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer func() {
		if err := database.Close(db); err != nil {
			logger.Error("close database", "error", err)
		}
	}()

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get database pool: %w", err)
	}

	healthService := healthapp.NewService(sqlDB)
	healthHandler := healthhttp.NewHandler(healthService)
	router := httpapi.NewRouter(logger, healthHandler)
	httpServer := server.New(cfg.HTTP, router)

	return serve(ctx, httpServer, cfg.HTTP.ShutdownTimeout, logger)
}

func serve(
	ctx context.Context,
	httpServer *http.Server,
	shutdownTimeout time.Duration,
	logger *slog.Logger,
) error {
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("HTTP server started", "address", httpServer.Addr)
		serverErr <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}
	return nil
}
