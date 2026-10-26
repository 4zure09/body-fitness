package authhttp

import (
	"context"
	"errors"
	"net/http"

	"github.com/4zure09/body-fitness/internal/modules/auth/application"
	"github.com/4zure09/body-fitness/internal/modules/auth/domain"
	"github.com/4zure09/body-fitness/internal/shared/httpx"
	"github.com/go-chi/chi/v5"
)

type authService interface {
	Register(ctx context.Context, input application.RegisterInput) (application.AuthResult, error)
	Login(ctx context.Context, input application.LoginInput) (application.AuthResult, error)
	Refresh(ctx context.Context, rawRefreshToken string) (application.AuthResult, error)
	Logout(ctx context.Context, rawRefreshToken string) error
	CurrentUser(ctx context.Context, userID string) (application.UserView, error)
}

type Handler struct {
	service    authService
	middleware Middleware
}

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func NewHandler(service authService, middleware Middleware) Handler {
	return Handler{service: service, middleware: middleware}
}

func (h Handler) RegisterRoutes(router chi.Router) {
	router.Route("/auth", func(auth chi.Router) {
		auth.Post("/register", h.Register)
		auth.Post("/login", h.Login)
		auth.Post("/refresh", h.Refresh)
		auth.Post("/logout", h.Logout)
		auth.With(h.middleware.Authenticate).Get("/me", h.Me)
	})

	router.Route("/admin", func(admin chi.Router) {
		admin.Use(h.middleware.Authenticate)
		admin.Use(h.middleware.RequireAnyRole(domain.RoleAdmin))
		admin.Get("/access-check", h.AdminAccessCheck)
	})
}

func (h Handler) Register(w http.ResponseWriter, r *http.Request) {
	var request registerRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_request", "The request body is invalid.")
		return
	}
	result, err := h.service.Register(r.Context(), application.RegisterInput{
		Email:       request.Email,
		Password:    request.Password,
		DisplayName: request.DisplayName,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	setTokenResponseHeaders(w)
	httpx.WriteJSON(w, http.StatusCreated, result)
}

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_request", "The request body is invalid.")
		return
	}
	result, err := h.service.Login(r.Context(), application.LoginInput{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	setTokenResponseHeaders(w)
	httpx.WriteJSON(w, http.StatusOK, result)
}

func (h Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	request, ok := decodeRefreshTokenRequest(w, r)
	if !ok {
		return
	}
	result, err := h.service.Refresh(r.Context(), request.RefreshToken)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	setTokenResponseHeaders(w)
	httpx.WriteJSON(w, http.StatusOK, result)
}

func (h Handler) Logout(w http.ResponseWriter, r *http.Request) {
	request, ok := decodeRefreshTokenRequest(w, r)
	if !ok {
		return
	}
	if err := h.service.Logout(r.Context(), request.RefreshToken); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) Me(w http.ResponseWriter, r *http.Request) {
	principal, ok := PrincipalFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}
	user, err := h.service.CurrentUser(r.Context(), principal.UserID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, user)
}

func (h Handler) AdminAccessCheck(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func decodeRefreshTokenRequest(w http.ResponseWriter, r *http.Request) (refreshTokenRequest, bool) {
	var request refreshTokenRequest
	if err := httpx.DecodeJSON(w, r, &request); err != nil || request.RefreshToken == "" {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_request", "A refresh_token is required.")
		return refreshTokenRequest{}, false
	}
	return request, true
}

func setTokenResponseHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		httpx.WriteError(w, http.StatusBadRequest, "invalid_input", err.Error())
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		httpx.WriteError(w, http.StatusConflict, "email_already_exists", "An account with this email already exists.")
	case errors.Is(err, domain.ErrInvalidCredentials), errors.Is(err, domain.ErrInvalidToken):
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "The supplied credentials or token are invalid.")
	case errors.Is(err, domain.ErrAccountDisabled):
		httpx.WriteError(w, http.StatusForbidden, "account_disabled", "This account is not active.")
	case errors.Is(err, domain.ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "not_found", "The requested resource was not found.")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred.")
	}
}
