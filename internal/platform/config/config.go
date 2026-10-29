package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHTTPAddress     = ":8080"
	defaultDatabaseDSN     = "postgres://body_fitness:body_fitness@localhost:5432/body_fitness?sslmode=disable"
	defaultReadTimeout     = 10 * time.Second
	defaultWriteTimeout    = 15 * time.Second
	defaultIdleTimeout     = 60 * time.Second
	defaultShutdownTimeout = 10 * time.Second
	defaultMaxOpenConns    = 20
	defaultMaxIdleConns    = 10
	defaultConnMaxLifetime = 30 * time.Minute
)

type Config struct {
	Environment string
	HTTP        HTTP
	Database    Database
	Auth        Auth
}

type HTTP struct {
	Address         string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type Database struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type Auth struct {
	JWTSecret  string
	Issuer     string
	Audience   string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func Load() (Config, error) {
	httpConfig, err := loadHTTP()
	if err != nil {
		return Config{}, err
	}
	databaseConfig, err := loadDatabase()
	if err != nil {
		return Config{}, err
	}
	authConfig, err := loadAuth()
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Environment: envOrDefault("APP_ENV", "local"),
		HTTP:        httpConfig,
		Database:    databaseConfig,
		Auth:        authConfig,
	}
	if err := validate(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func loadAuth() (Auth, error) {
	accessTTL, err := durationEnv("JWT_ACCESS_TTL", 15*time.Minute)
	if err != nil {
		return Auth{}, err
	}
	refreshTTL, err := durationEnv("JWT_REFRESH_TTL", 30*24*time.Hour)
	if err != nil {
		return Auth{}, err
	}
	return Auth{
		JWTSecret:  envOrDefault("JWT_SECRET", "local-development-secret-change-before-production-2026"),
		Issuer:     envOrDefault("JWT_ISSUER", "body-fitness"),
		Audience:   envOrDefault("JWT_AUDIENCE", "body-fitness-api"),
		AccessTTL:  accessTTL,
		RefreshTTL: refreshTTL,
	}, nil
}

func loadHTTP() (HTTP, error) {
	readTimeout, err := durationEnv("HTTP_READ_TIMEOUT", defaultReadTimeout)
	if err != nil {
		return HTTP{}, err
	}
	writeTimeout, err := durationEnv("HTTP_WRITE_TIMEOUT", defaultWriteTimeout)
	if err != nil {
		return HTTP{}, err
	}
	idleTimeout, err := durationEnv("HTTP_IDLE_TIMEOUT", defaultIdleTimeout)
	if err != nil {
		return HTTP{}, err
	}
	shutdownTimeout, err := durationEnv("HTTP_SHUTDOWN_TIMEOUT", defaultShutdownTimeout)
	if err != nil {
		return HTTP{}, err
	}
	return HTTP{
		Address:         envOrDefault("HTTP_ADDRESS", defaultHTTPAddress),
		ReadTimeout:     readTimeout,
		WriteTimeout:    writeTimeout,
		IdleTimeout:     idleTimeout,
		ShutdownTimeout: shutdownTimeout,
	}, nil
}

func loadDatabase() (Database, error) {
	maxOpenConns, err := intEnv("DATABASE_MAX_OPEN_CONNS", defaultMaxOpenConns)
	if err != nil {
		return Database{}, err
	}
	maxIdleConns, err := intEnv("DATABASE_MAX_IDLE_CONNS", defaultMaxIdleConns)
	if err != nil {
		return Database{}, err
	}
	connMaxLifetime, err := durationEnv("DATABASE_CONN_MAX_LIFETIME", defaultConnMaxLifetime)
	if err != nil {
		return Database{}, err
	}
	return Database{
		DSN:             DatabaseDSN(),
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,
	}, nil
}

func validate(cfg Config) error {
	if strings.TrimSpace(cfg.Database.DSN) == "" {
		return errors.New("DATABASE_DSN or DATABASE_URL must not be empty")
	}
	if cfg.Database.MaxOpenConns < 1 {
		return errors.New("DATABASE_MAX_OPEN_CONNS must be greater than zero")
	}
	if cfg.Database.MaxIdleConns < 0 || cfg.Database.MaxIdleConns > cfg.Database.MaxOpenConns {
		return errors.New("DATABASE_MAX_IDLE_CONNS must be between zero and DATABASE_MAX_OPEN_CONNS")
	}
	if len(cfg.Auth.JWTSecret) < 32 {
		return errors.New("JWT_SECRET must contain at least 32 characters")
	}
	if strings.TrimSpace(cfg.Auth.Issuer) == "" || strings.TrimSpace(cfg.Auth.Audience) == "" {
		return errors.New("JWT_ISSUER and JWT_AUDIENCE must not be empty")
	}
	if cfg.Auth.AccessTTL >= cfg.Auth.RefreshTTL {
		return errors.New("JWT_ACCESS_TTL must be shorter than JWT_REFRESH_TTL")
	}
	if cfg.Environment != "local" && cfg.Auth.JWTSecret == "local-development-secret-change-before-production-2026" {
		return errors.New("JWT_SECRET must be changed outside the local environment")
	}
	return nil
}

func DatabaseDSN() string {
	if value := strings.TrimSpace(os.Getenv("DATABASE_DSN")); value != "" {
		return value
	}
	if value := strings.TrimSpace(os.Getenv("DATABASE_URL")); value != "" {
		return value
	}
	return defaultDatabaseDSN
}

func envOrDefault(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	raw := envOrDefault(key, fallback.String())
	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}
	return value, nil
}

func intEnv(key string, fallback int) (int, error) {
	raw := envOrDefault(key, strconv.Itoa(fallback))
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return value, nil
}
