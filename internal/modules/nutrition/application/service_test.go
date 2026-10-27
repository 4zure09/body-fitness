package application

import (
	"errors"
	"testing"
	"time"

	"github.com/4zure09/body-fitness/internal/modules/nutrition/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestBuildFoodRejectsNegativeNutrients(t *testing.T) {
	service := testService()
	_, err := service.buildFood(uuid.NewString(), FoodInput{
		Name: "Chicken breast", ReferenceAmount: decimal.NewFromInt(100),
		ReferenceUnit: domain.ReferenceUnitGram,
		Nutrients:     domain.Nutrients{CaloriesKcal: decimal.NewFromInt(-1)},
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("buildFood() error = %v, want invalid input", err)
	}
}

func TestBuildFoodNormalizesOptionalBrand(t *testing.T) {
	service := testService()
	brand := "  Local Farm  "
	food, err := service.buildFood(uuid.NewString(), FoodInput{
		Name: " Chicken breast ", Brand: &brand, ReferenceAmount: decimal.NewFromInt(100),
		ReferenceUnit: domain.ReferenceUnitGram,
		Nutrients:     domain.Nutrients{CaloriesKcal: decimal.NewFromInt(165)},
	})
	if err != nil {
		t.Fatalf("buildFood() error = %v", err)
	}
	if food.Name != "Chicken breast" || food.Brand == nil || *food.Brand != "Local Farm" {
		t.Fatalf("unexpected food: %+v", food)
	}
}

func TestBuildServingRejectsZeroAmount(t *testing.T) {
	service := testService()
	_, err := service.buildServing(uuid.NewString(), uuid.NewString(), ServingInput{
		Name: "One portion", Amount: decimal.Zero,
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("buildServing() error = %v, want invalid input", err)
	}
}

func testService() *Service {
	fixedNow := time.Date(2026, 10, 27, 10, 0, 0, 0, time.UTC)
	return &Service{now: func() time.Time { return fixedNow }, newID: uuid.NewString}
}
