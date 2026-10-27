package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/4zure09/body-fitness/internal/modules/nutrition/domain"
	"github.com/4zure09/body-fitness/internal/modules/nutrition/port"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Service struct {
	repository port.Repository
	now        func() time.Time
	newID      func() string
}

type FoodInput struct {
	Name            string
	Brand           *string
	Description     string
	ReferenceAmount decimal.Decimal
	ReferenceUnit   domain.ReferenceUnit
	Nutrients       domain.Nutrients
}

type ServingInput struct {
	Name   string
	Amount decimal.Decimal
}

func NewService(repository port.Repository) *Service {
	return &Service{repository: repository, now: time.Now, newID: uuid.NewString}
}

func (s *Service) CreateFood(ctx context.Context, input FoodInput) (domain.Food, error) {
	food, err := s.buildFood(s.newID(), input)
	if err != nil {
		return domain.Food{}, err
	}
	if err := s.repository.CreateFood(ctx, food); err != nil {
		return domain.Food{}, err
	}
	return food, nil
}

func (s *Service) ListFoods(ctx context.Context, options port.ListOptions) (port.FoodPage, error) {
	return s.repository.ListFoods(ctx, options)
}

func (s *Service) GetFood(ctx context.Context, id string) (domain.Food, error) {
	if err := validateID(id); err != nil {
		return domain.Food{}, err
	}
	return s.repository.GetFood(ctx, id)
}

func (s *Service) UpdateFood(ctx context.Context, id string, input FoodInput) (domain.Food, error) {
	if err := validateID(id); err != nil {
		return domain.Food{}, err
	}
	food, err := s.buildFood(id, input)
	if err != nil {
		return domain.Food{}, err
	}
	if err := s.repository.UpdateFood(ctx, food); err != nil {
		return domain.Food{}, err
	}
	return s.repository.GetFood(ctx, id)
}

func (s *Service) DeleteFood(ctx context.Context, id string) error {
	if err := validateID(id); err != nil {
		return err
	}
	return s.repository.DeleteFood(ctx, id)
}

func (s *Service) CreateServing(
	ctx context.Context,
	foodID string,
	input ServingInput,
) (domain.Serving, error) {
	if err := validateID(foodID); err != nil {
		return domain.Serving{}, err
	}
	serving, err := s.buildServing(s.newID(), foodID, input)
	if err != nil {
		return domain.Serving{}, err
	}
	if err := s.repository.CreateServing(ctx, serving); err != nil {
		return domain.Serving{}, err
	}
	return serving, nil
}

func (s *Service) ListServings(
	ctx context.Context,
	foodID string,
	options port.ListOptions,
) (port.ServingPage, error) {
	if err := validateID(foodID); err != nil {
		return port.ServingPage{}, err
	}
	if _, err := s.repository.GetFood(ctx, foodID); err != nil {
		return port.ServingPage{}, err
	}
	return s.repository.ListServings(ctx, foodID, options)
}

func (s *Service) GetServing(ctx context.Context, foodID, servingID string) (domain.Serving, error) {
	if err := validateIDs(foodID, servingID); err != nil {
		return domain.Serving{}, err
	}
	return s.repository.GetServing(ctx, foodID, servingID)
}

func (s *Service) UpdateServing(
	ctx context.Context,
	foodID string,
	servingID string,
	input ServingInput,
) (domain.Serving, error) {
	if err := validateIDs(foodID, servingID); err != nil {
		return domain.Serving{}, err
	}
	serving, err := s.buildServing(servingID, foodID, input)
	if err != nil {
		return domain.Serving{}, err
	}
	if err := s.repository.UpdateServing(ctx, serving); err != nil {
		return domain.Serving{}, err
	}
	return s.repository.GetServing(ctx, foodID, servingID)
}

func (s *Service) DeleteServing(ctx context.Context, foodID, servingID string) error {
	if err := validateIDs(foodID, servingID); err != nil {
		return err
	}
	return s.repository.DeleteServing(ctx, foodID, servingID)
}

func (s *Service) buildFood(id string, input FoodInput) (domain.Food, error) {
	name, err := normalizedRequired(input.Name, "name", 2, 160)
	if err != nil {
		return domain.Food{}, err
	}
	description, err := normalizedOptional(input.Description, "description", 4000)
	if err != nil {
		return domain.Food{}, err
	}
	brand, err := normalizedBrand(input.Brand)
	if err != nil {
		return domain.Food{}, err
	}
	if !input.ReferenceAmount.IsPositive() {
		return domain.Food{}, fmt.Errorf("%w: reference_amount must be greater than zero", domain.ErrInvalidInput)
	}
	if !validReferenceUnit(input.ReferenceUnit) {
		return domain.Food{}, fmt.Errorf("%w: reference_unit is invalid", domain.ErrInvalidInput)
	}
	if err := validateNutrients(input.Nutrients); err != nil {
		return domain.Food{}, err
	}
	now := s.now().UTC()
	return domain.Food{
		ID: id, Name: name, Brand: brand, Description: description,
		ReferenceAmount: input.ReferenceAmount, ReferenceUnit: input.ReferenceUnit,
		Nutrients: input.Nutrients, CreatedAt: now, UpdatedAt: now,
	}, nil
}

func (s *Service) buildServing(id, foodID string, input ServingInput) (domain.Serving, error) {
	name, err := normalizedRequired(input.Name, "name", 1, 120)
	if err != nil {
		return domain.Serving{}, err
	}
	if !input.Amount.IsPositive() {
		return domain.Serving{}, fmt.Errorf("%w: amount must be greater than zero", domain.ErrInvalidInput)
	}
	now := s.now().UTC()
	return domain.Serving{
		ID: id, FoodID: foodID, Name: name, Amount: input.Amount,
		CreatedAt: now, UpdatedAt: now,
	}, nil
}

func validateNutrients(nutrients domain.Nutrients) error {
	values := []struct {
		name  string
		value decimal.Decimal
	}{
		{"calories_kcal", nutrients.CaloriesKcal},
		{"protein_g", nutrients.ProteinG},
		{"carbohydrates_g", nutrients.CarbohydratesG},
		{"fat_g", nutrients.FatG},
		{"fiber_g", nutrients.FiberG},
	}
	for _, nutrient := range values {
		if nutrient.value.IsNegative() {
			return fmt.Errorf("%w: %s must not be negative", domain.ErrInvalidInput, nutrient.name)
		}
	}
	return nil
}

func normalizedBrand(value *string) (*string, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	normalized, err := normalizedRequired(*value, "brand", 1, 160)
	if err != nil {
		return nil, err
	}
	return &normalized, nil
}

func normalizedRequired(value, field string, minLength, maxLength int) (string, error) {
	normalized := strings.TrimSpace(value)
	if len(normalized) < minLength || len(normalized) > maxLength {
		return "", fmt.Errorf("%w: %s must contain between %d and %d characters", domain.ErrInvalidInput, field, minLength, maxLength)
	}
	return normalized, nil
}

func normalizedOptional(value, field string, maxLength int) (string, error) {
	normalized := strings.TrimSpace(value)
	if len(normalized) > maxLength {
		return "", fmt.Errorf("%w: %s must not exceed %d characters", domain.ErrInvalidInput, field, maxLength)
	}
	return normalized, nil
}

func validateIDs(ids ...string) error {
	for _, id := range ids {
		if err := validateID(id); err != nil {
			return err
		}
	}
	return nil
}

func validateID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("%w: ID must be a valid UUID", domain.ErrInvalidInput)
	}
	return nil
}

func validReferenceUnit(value domain.ReferenceUnit) bool {
	switch value {
	case domain.ReferenceUnitGram, domain.ReferenceUnitMilliliter, domain.ReferenceUnitItem:
		return true
	default:
		return false
	}
}
