package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	authhttp "github.com/4zure09/body-fitness/internal/modules/auth/transport/http"
	healthhttp "github.com/4zure09/body-fitness/internal/modules/health/transport/http"
	"github.com/4zure09/body-fitness/internal/platform/openapi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(logger *slog.Logger, health healthhttp.Handler, auth authhttp.Handler) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(requestLogger(logger))
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(30 * time.Second))

	router.Get("/health/live", health.Liveness)
	router.Get("/health/ready", health.Readiness)
	router.Get("/openapi.yaml", openapi.Spec)
	router.Get("/docs", openapi.Docs)

	router.Route("/api/v1", func(api chi.Router) {
		auth.RegisterRoutes(api)
	})

	return router
}

func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()
			wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(wrapped, r)

			logger.InfoContext(
				r.Context(),
				"HTTP request completed",
				"request_id", middleware.GetReqID(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.Status(),
				"bytes", wrapped.BytesWritten(),
				"duration_ms", time.Since(startedAt).Milliseconds(),
			)
		})
	}
}
