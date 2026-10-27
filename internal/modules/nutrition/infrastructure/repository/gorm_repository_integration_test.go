package repository

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/4zure09/body-fitness/internal/modules/nutrition/domain"
	"github.com/4zure09/body-fitness/internal/modules/nutrition/port"
	"github.com/4zure09/body-fitness/internal/platform/config"
	"github.com/4zure09/body-fitness/internal/platform/database"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestGORMRepositoryNutritionCatalogLifecycle(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN is not configured")
	}
	db, err := database.Open(context.Background(), config.Database{
		DSN: dsn, MaxOpenConns: 5, MaxIdleConns: 2, ConnMaxLifetime: time.Minute,
	})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { _ = database.Close(db) })

	repository := NewGORMRepository(db)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	brand := "Test Brand"
	food := domain.Food{
		ID: uuid.NewString(), Name: "Food " + uuid.NewString(), Brand: &brand,
		Description: "Integration test food", ReferenceAmount: decimal.NewFromInt(100),
		ReferenceUnit: domain.ReferenceUnitGram,
		Nutrients: domain.Nutrients{
			CaloriesKcal: decimal.NewFromInt(165), ProteinG: decimal.NewFromFloat(31.2),
			CarbohydratesG: decimal.Zero, FatG: decimal.NewFromFloat(3.6), FiberG: decimal.Zero,
		},
		CreatedAt: now, UpdatedAt: now,
	}
	serving := domain.Serving{
		ID: uuid.NewString(), FoodID: food.ID, Name: "One portion",
		Amount: decimal.NewFromInt(150), CreatedAt: now, UpdatedAt: now,
	}
	t.Cleanup(func() {
		_ = db.Delete(&servingModel{}, "food_id = ?", food.ID).Error
		_ = db.Unscoped().Delete(&foodModel{}, "id = ?", food.ID).Error
	})

	if err := repository.CreateFood(ctx, food); err != nil {
		t.Fatalf("CreateFood() error = %v", err)
	}
	duplicate := food
	duplicate.ID = uuid.NewString()
	if err := repository.CreateFood(ctx, duplicate); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("duplicate food error = %v, want conflict", err)
	}
	stored, err := repository.GetFood(ctx, food.ID)
	if err != nil {
		t.Fatalf("GetFood() error = %v", err)
	}
	if !stored.Nutrients.ProteinG.Equal(food.Nutrients.ProteinG) {
		t.Fatalf("stored protein = %s, want %s", stored.Nutrients.ProteinG, food.Nutrients.ProteinG)
	}
	page, err := repository.ListFoods(ctx, port.ListOptions{Query: food.Name, Brand: brand, Limit: 20})
	if err != nil || page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("unexpected food page: %+v, error=%v", page, err)
	}

	if err := repository.CreateServing(ctx, serving); err != nil {
		t.Fatalf("CreateServing() error = %v", err)
	}
	duplicateServing := serving
	duplicateServing.ID = uuid.NewString()
	duplicateServing.Name = "ONE PORTION"
	if err := repository.CreateServing(ctx, duplicateServing); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("duplicate serving error = %v, want conflict", err)
	}
	servingPage, err := repository.ListServings(ctx, food.ID, port.ListOptions{Limit: 20})
	if err != nil || servingPage.Total != 1 || len(servingPage.Items) != 1 {
		t.Fatalf("unexpected serving page: %+v, error=%v", servingPage, err)
	}

	serving.Name = "Large portion"
	serving.Amount = decimal.NewFromInt(200)
	serving.UpdatedAt = now.Add(time.Minute)
	if err := repository.UpdateServing(ctx, serving); err != nil {
		t.Fatalf("UpdateServing() error = %v", err)
	}
	storedServing, err := repository.GetServing(ctx, food.ID, serving.ID)
	if err != nil || storedServing.Name != serving.Name || !storedServing.Amount.Equal(serving.Amount) {
		t.Fatalf("unexpected updated serving: %+v, error=%v", storedServing, err)
	}
	if err := repository.DeleteServing(ctx, food.ID, serving.ID); err != nil {
		t.Fatalf("DeleteServing() error = %v", err)
	}
	if _, err := repository.GetServing(ctx, food.ID, serving.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("deleted serving error = %v, want not found", err)
	}

	food.Nutrients.CaloriesKcal = decimal.NewFromInt(170)
	food.UpdatedAt = now.Add(time.Minute)
	if err := repository.UpdateFood(ctx, food); err != nil {
		t.Fatalf("UpdateFood() error = %v", err)
	}
	if err := repository.DeleteFood(ctx, food.ID); err != nil {
		t.Fatalf("DeleteFood() error = %v", err)
	}
	if _, err := repository.GetFood(ctx, food.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("deleted food error = %v, want not found", err)
	}
	serving.ID = uuid.NewString()
	if err := repository.CreateServing(ctx, serving); !errors.Is(err, domain.ErrInvalidReference) {
		t.Fatalf("CreateServing() for soft-deleted food error = %v, want invalid reference", err)
	}
}
