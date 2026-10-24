package server

import (
	"net/http"

	"github.com/4zure09/body-fitness/internal/platform/config"
)

func New(cfg config.HTTP, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         cfg.Address,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}
