package authhttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/4zure09/body-fitness/internal/modules/auth/domain"
	"github.com/4zure09/body-fitness/internal/modules/auth/port"
)

type tokenParserStub struct {
	claims port.Claims
	err    error
}

func (s tokenParserStub) ParseAccess(string) (port.Claims, error) {
	return s.claims, s.err
}

func TestAuthenticateRejectsMissingBearerToken(t *testing.T) {
	middleware := NewMiddleware(tokenParserStub{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	middleware.Authenticate(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("protected handler must not be called")
	})).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestAuthenticateStoresPrincipal(t *testing.T) {
	middleware := NewMiddleware(tokenParserStub{claims: port.Claims{
		UserID: "user-id",
		Roles:  []domain.Role{domain.RoleUser},
	}})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer valid-token")

	middleware.Authenticate(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		principal, ok := PrincipalFromContext(r.Context())
		if !ok || principal.UserID != "user-id" {
			t.Fatalf("unexpected principal: %+v", principal)
		}
	})).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestAuthenticateRejectsInvalidToken(t *testing.T) {
	middleware := NewMiddleware(tokenParserStub{err: errors.New("invalid")})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer invalid-token")

	middleware.Authenticate(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("protected handler must not be called")
	})).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestRequireAnyRoleAllowsAdminAndRejectsUser(t *testing.T) {
	middleware := NewMiddleware(tokenParserStub{})
	protected := middleware.RequireAnyRole(domain.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	userRecorder := httptest.NewRecorder()
	userRequest := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(
		contextWithPrincipal(Principal{UserID: "user", Roles: []domain.Role{domain.RoleUser}}),
	)
	protected.ServeHTTP(userRecorder, userRequest)
	if userRecorder.Code != http.StatusForbidden {
		t.Fatalf("user status = %d, want %d", userRecorder.Code, http.StatusForbidden)
	}

	adminRecorder := httptest.NewRecorder()
	adminRequest := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(
		contextWithPrincipal(Principal{UserID: "admin", Roles: []domain.Role{domain.RoleAdmin}}),
	)
	protected.ServeHTTP(adminRecorder, adminRequest)
	if adminRecorder.Code != http.StatusNoContent {
		t.Fatalf("admin status = %d, want %d", adminRecorder.Code, http.StatusNoContent)
	}
}

func contextWithPrincipal(principal Principal) context.Context {
	return context.WithValue(context.Background(), principalContextKey{}, principal)
}
