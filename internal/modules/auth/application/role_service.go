package application

import (
	"context"
	"fmt"

	"github.com/4zure09/body-fitness/internal/modules/auth/domain"
)

func (s *Service) GrantRole(ctx context.Context, email string, role domain.Role) error {
	normalizedEmail, err := domain.NormalizeEmail(email)
	if err != nil {
		return err
	}
	if role != domain.RoleUser && role != domain.RoleAdmin {
		return fmt.Errorf("%w: unsupported role %q", domain.ErrInvalidInput, role)
	}
	return s.repository.GrantRoleByEmail(ctx, normalizedEmail, role, s.now().UTC())
}
