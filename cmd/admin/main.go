package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	authapp "github.com/4zure09/body-fitness/internal/modules/auth/application"
	"github.com/4zure09/body-fitness/internal/modules/auth/domain"
	"github.com/4zure09/body-fitness/internal/modules/auth/infrastructure/password"
	authrepo "github.com/4zure09/body-fitness/internal/modules/auth/infrastructure/repository"
	"github.com/4zure09/body-fitness/internal/modules/auth/infrastructure/token"
	"github.com/4zure09/body-fitness/internal/platform/config"
	"github.com/4zure09/body-fitness/internal/platform/database"
	"github.com/4zure09/body-fitness/internal/platform/logging"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	email := flag.String("email", "", "email of the registered user to promote")
	flag.Parse()
	if *email == "" {
		return fmt.Errorf("--email is required")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	logger := logging.New(cfg.Environment)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := database.Open(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer func() {
		if err := database.Close(db); err != nil {
			logger.Error("close database", "error", err)
		}
	}()

	repository := authrepo.NewGORMRepository(db)
	service, err := authapp.NewService(repository, password.New(), token.NewManager(cfg.Auth))
	if err != nil {
		return fmt.Errorf("initialize authentication service: %w", err)
	}
	if err := service.GrantRole(ctx, *email, domain.RoleAdmin); err != nil {
		return fmt.Errorf("grant admin role: %w", err)
	}
	logger.Info("admin role granted", "email", *email)
	return nil
}
