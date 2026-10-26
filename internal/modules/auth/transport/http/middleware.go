package authhttp

import (
	"context"
	"net/http"
	"strings"

	"github.com/4zure09/body-fitness/internal/modules/auth/domain"
	"github.com/4zure09/body-fitness/internal/modules/auth/port"
	"github.com/4zure09/body-fitness/internal/shared/httpx"
)

type accessTokenParser interface {
	ParseAccess(rawToken string) (port.Claims, error)
}

type principalContextKey struct{}

type Principal struct {
	UserID string
	Roles  []domain.Role
}

type Middleware struct {
	tokens accessTokenParser
}

func NewMiddleware(tokens accessTokenParser) Middleware {
	return Middleware{tokens: tokens}
}

func (m Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawToken, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "A valid Bearer access token is required.")
			return
		}
		claims, err := m.tokens.ParseAccess(rawToken)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "A valid Bearer access token is required.")
			return
		}
		principal := Principal{UserID: claims.UserID, Roles: append([]domain.Role(nil), claims.Roles...)}
		ctx := context.WithValue(r.Context(), principalContextKey{}, principal)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m Middleware) RequireAnyRole(allowedRoles ...domain.Role) func(http.Handler) http.Handler {
	allowed := make(map[domain.Role]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		allowed[role] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := PrincipalFromContext(r.Context())
			if !ok {
				httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
				return
			}
			if !hasAnyRole(principal.Roles, allowed) {
				httpx.WriteError(w, http.StatusForbidden, "forbidden", "You do not have permission to access this resource.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(Principal)
	return principal, ok
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func hasAnyRole(userRoles []domain.Role, allowed map[domain.Role]struct{}) bool {
	for _, role := range userRoles {
		if _, ok := allowed[role]; ok {
			return true
		}
	}
	return false
}
