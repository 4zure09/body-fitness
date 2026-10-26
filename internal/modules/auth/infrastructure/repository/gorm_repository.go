package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/4zure09/body-fitness/internal/modules/auth/domain"
	"github.com/4zure09/body-fitness/internal/modules/auth/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GORMRepository struct {
	db *gorm.DB
}

type userModel struct {
	ID           string       `gorm:"type:uuid;primaryKey"`
	Email        string       `gorm:"not null"`
	PasswordHash string       `gorm:"not null"`
	Status       string       `gorm:"not null"`
	CreatedAt    time.Time    `gorm:"not null"`
	UpdatedAt    time.Time    `gorm:"not null"`
	Profile      profileModel `gorm:"foreignKey:UserID"`
	Roles        []roleModel  `gorm:"many2many:user_roles;joinForeignKey:UserID;joinReferences:RoleID"`
}

func (userModel) TableName() string { return "users" }

type profileModel struct {
	UserID      string    `gorm:"type:uuid;primaryKey"`
	DisplayName string    `gorm:"not null"`
	Timezone    string    `gorm:"not null"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

func (profileModel) TableName() string { return "user_profiles" }

type roleModel struct {
	ID   int16  `gorm:"primaryKey"`
	Name string `gorm:"not null;unique"`
}

func (roleModel) TableName() string { return "roles" }

type userRoleModel struct {
	UserID    string    `gorm:"type:uuid;primaryKey"`
	RoleID    int16     `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"not null"`
}

func (userRoleModel) TableName() string { return "user_roles" }

type refreshTokenModel struct {
	ID                string    `gorm:"type:uuid;primaryKey"`
	UserID            string    `gorm:"type:uuid;not null;index"`
	TokenHash         string    `gorm:"not null;unique"`
	ExpiresAt         time.Time `gorm:"not null;index"`
	RevokedAt         *time.Time
	ReplacedByTokenID *string   `gorm:"column:replaced_by_token_id;type:uuid"`
	CreatedAt         time.Time `gorm:"not null"`
}

func (refreshTokenModel) TableName() string { return "refresh_tokens" }

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db}
}

func (r *GORMRepository) CreateUser(
	ctx context.Context,
	user domain.User,
	session port.RefreshSession,
) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := user.CreatedAt
		model := userModel{
			ID:           user.ID,
			Email:        user.Email,
			PasswordHash: user.PasswordHash,
			Status:       string(user.Status),
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := tx.Create(&model).Error; err != nil {
			return err
		}
		profile := profileModel{
			UserID:      user.ID,
			DisplayName: user.DisplayName,
			Timezone:    "UTC",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := tx.Create(&profile).Error; err != nil {
			return err
		}
		if err := assignRoles(ctx, tx, user.ID, user.Roles, now); err != nil {
			return err
		}
		return tx.Create(newRefreshTokenModel(session, now)).Error
	})
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return domain.ErrEmailAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *GORMRepository) FindUserByEmail(ctx context.Context, email string) (domain.User, error) {
	return r.findUser(ctx, r.db.Where("LOWER(email) = ?", email))
}

func (r *GORMRepository) FindUserByID(ctx context.Context, userID string) (domain.User, error) {
	return r.findUser(ctx, r.db.Where("id = ?", userID))
}

func (r *GORMRepository) FindUserByActiveRefreshToken(
	ctx context.Context,
	tokenHash string,
	now time.Time,
) (domain.User, error) {
	query := r.db.
		Joins("JOIN refresh_tokens ON refresh_tokens.user_id = users.id").
		Where("refresh_tokens.token_hash = ?", tokenHash).
		Where("refresh_tokens.revoked_at IS NULL").
		Where("refresh_tokens.expires_at > ?", now)
	return r.findUser(ctx, query)
}

func (r *GORMRepository) StoreRefreshToken(ctx context.Context, session port.RefreshSession) error {
	model := newRefreshTokenModel(session, time.Now().UTC())
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("store refresh token: %w", err)
	}
	return nil
}

func (r *GORMRepository) GrantRoleByEmail(
	ctx context.Context,
	email string,
	role domain.Role,
	grantedAt time.Time,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user userModel
		if err := tx.Select("id").Where("LOWER(email) = ?", email).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
			return fmt.Errorf("find user for role assignment: %w", err)
		}
		var storedRole roleModel
		if err := tx.Where("name = ?", role).First(&storedRole).Error; err != nil {
			return fmt.Errorf("find role %q: %w", role, err)
		}
		mapping := userRoleModel{UserID: user.ID, RoleID: storedRole.ID, CreatedAt: grantedAt}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&mapping).Error; err != nil {
			return fmt.Errorf("grant role %q: %w", role, err)
		}
		return nil
	})
}

func (r *GORMRepository) RotateRefreshToken(
	ctx context.Context,
	currentHash string,
	replacement port.RefreshSession,
	revokedAt time.Time,
) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		replacementModel := newRefreshTokenModel(replacement, revokedAt)
		if err := tx.Create(&replacementModel).Error; err != nil {
			return err
		}
		result := tx.Model(&refreshTokenModel{}).
			Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", currentHash, revokedAt).
			Updates(map[string]any{
				"revoked_at":           revokedAt,
				"replaced_by_token_id": replacement.ID,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return domain.ErrNotFound
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("rotate refresh token: %w", err)
	}
	return nil
}

func (r *GORMRepository) RevokeRefreshToken(
	ctx context.Context,
	tokenHash string,
	revokedAt time.Time,
) error {
	result := r.db.WithContext(ctx).
		Model(&refreshTokenModel{}).
		Where("token_hash = ? AND revoked_at IS NULL", tokenHash).
		Update("revoked_at", revokedAt)
	if result.Error != nil {
		return fmt.Errorf("revoke refresh token: %w", result.Error)
	}
	return nil
}

func (r *GORMRepository) findUser(ctx context.Context, query *gorm.DB) (domain.User, error) {
	var model userModel
	err := query.WithContext(ctx).
		Preload("Profile").
		Preload("Roles").
		First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("find user: %w", err)
	}
	return toDomainUser(model), nil
}

func assignRoles(
	ctx context.Context,
	tx *gorm.DB,
	userID string,
	roles []domain.Role,
	createdAt time.Time,
) error {
	for _, role := range roles {
		var storedRole roleModel
		if err := tx.WithContext(ctx).Where("name = ?", role).First(&storedRole).Error; err != nil {
			return fmt.Errorf("find role %q: %w", role, err)
		}
		mapping := userRoleModel{UserID: userID, RoleID: storedRole.ID, CreatedAt: createdAt}
		if err := tx.WithContext(ctx).Create(&mapping).Error; err != nil {
			return fmt.Errorf("assign role %q: %w", role, err)
		}
	}
	return nil
}

func newRefreshTokenModel(session port.RefreshSession, createdAt time.Time) refreshTokenModel {
	return refreshTokenModel{
		ID:                session.ID,
		UserID:            session.UserID,
		TokenHash:         session.TokenHash,
		ExpiresAt:         session.ExpiresAt,
		RevokedAt:         session.RevokedAt,
		ReplacedByTokenID: session.ReplacementID,
		CreatedAt:         createdAt,
	}
}

func toDomainUser(model userModel) domain.User {
	roles := make([]domain.Role, 0, len(model.Roles))
	for _, role := range model.Roles {
		roles = append(roles, domain.Role(role.Name))
	}
	sort.Slice(roles, func(i, j int) bool { return roles[i] < roles[j] })
	return domain.User{
		ID:           model.ID,
		Email:        model.Email,
		PasswordHash: model.PasswordHash,
		DisplayName:  model.Profile.DisplayName,
		Status:       domain.UserStatus(model.Status),
		Roles:        roles,
		CreatedAt:    model.CreatedAt,
	}
}
