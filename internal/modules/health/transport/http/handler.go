package healthhttp

import (
	"net/http"

	healthapp "github.com/4zure09/body-fitness/internal/modules/health/application"
	"github.com/4zure09/body-fitness/internal/shared/httpx"
)

type Handler struct {
	service healthapp.Service
}

func NewHandler(service healthapp.Service) Handler {
	return Handler{service: service}
}

func (h Handler) Liveness(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, h.service.Liveness())
}

func (h Handler) Readiness(w http.ResponseWriter, r *http.Request) {
	status, err := h.service.Readiness(r.Context())
	if err != nil {
		httpx.WriteJSON(w, http.StatusServiceUnavailable, status)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, status)
}
