package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/4zure09/body-fitness/internal/modules/nutrition/domain"
	"github.com/4zure09/body-fitness/internal/modules/nutrition/port"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GORMRepository struct {
	db *gorm.DB
}

type foodModel struct {
	ID              string `gorm:"type:uuid;primaryKey"`
	Name            string `gorm:"not null"`
	Brand           *string
	Description     string          `gorm:"not null"`
	ReferenceAmount decimal.Decimal `gorm:"type:numeric(10,3);not null"`
	ReferenceUnit   string          `gorm:"not null"`
	CaloriesKcal    decimal.Decimal `gorm:"type:numeric(10,2);not null"`
	ProteinG        decimal.Decimal `gorm:"type:numeric(10,2);not null"`
	CarbohydratesG  decimal.Decimal `gorm:"type:numeric(10,2);not null"`
	FatG            decimal.Decimal `gorm:"type:numeric(10,2);not null"`
	FiberG          decimal.Decimal `gorm:"type:numeric(10,2);not null"`
	CreatedAt       time.Time       `gorm:"not null"`
	UpdatedAt       time.Time       `gorm:"not null"`
	DeletedAt       gorm.DeletedAt  `gorm:"index"`
}

func (foodModel) TableName() string { return "foods" }

type servingModel struct {
	ID        string          `gorm:"type:uuid;primaryKey"`
	FoodID    string          `gorm:"type:uuid;not null;index"`
	Name      string          `gorm:"not null"`
	Amount    decimal.Decimal `gorm:"type:numeric(10,3);not null"`
	CreatedAt time.Time       `gorm:"not null"`
	UpdatedAt time.Time       `gorm:"not null"`
}

func (servingModel) TableName() string { return "food_servings" }

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db}
}

func (r *GORMRepository) CreateFood(ctx context.Context, food domain.Food) error {
	model := toFoodModel(food)
	return mapWriteError(r.db.WithContext(ctx).Create(&model).Error)
}

func (r *GORMRepository) ListFoods(ctx context.Context, options port.ListOptions) (port.FoodPage, error) {
	query := r.db.WithContext(ctx).Model(&foodModel{})
	if value := strings.TrimSpace(options.Query); value != "" {
		query = query.Where("name ILIKE ? OR COALESCE(brand, '') ILIKE ?", "%"+value+"%", "%"+value+"%")
	}
	if brand := strings.TrimSpace(options.Brand); brand != "" {
		query = query.Where("brand ILIKE ?", brand)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return port.FoodPage{}, fmt.Errorf("count foods: %w", err)
	}
	var models []foodModel
	if err := query.Order("name ASC").Offset(options.Offset).Limit(options.Limit).Find(&models).Error; err != nil {
		return port.FoodPage{}, fmt.Errorf("list foods: %w", err)
	}
	items := make([]domain.Food, 0, len(models))
	for _, model := range models {
		items = append(items, toFood(model))
	}
	return port.FoodPage{Items: items, Total: total}, nil
}

func (r *GORMRepository) GetFood(ctx context.Context, id string) (domain.Food, error) {
	var model foodModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.Food{}, mapReadError("get food", err)
	}
	return toFood(model), nil
}

func (r *GORMRepository) UpdateFood(ctx context.Context, food domain.Food) error {
	model := toFoodModel(food)
	result := r.db.WithContext(ctx).Model(&foodModel{}).Where("id = ?", food.ID).Updates(map[string]any{
		"name": model.Name, "brand": model.Brand, "description": model.Description,
		"reference_amount": model.ReferenceAmount, "reference_unit": model.ReferenceUnit,
		"calories_kcal": model.CaloriesKcal, "protein_g": model.ProteinG,
		"carbohydrates_g": model.CarbohydratesG, "fat_g": model.FatG,
		"fiber_g": model.FiberG, "updated_at": model.UpdatedAt,
	})
	return mapUpdateResult(result, "update food")
}

func (r *GORMRepository) DeleteFood(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&foodModel{}, "id = ?", id)
	return mapDeleteResult(result, "delete food")
}

func (r *GORMRepository) CreateServing(ctx context.Context, serving domain.Serving) error {
	model := toServingModel(serving)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockActiveFood(tx, serving.FoodID); err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return domain.ErrInvalidReference
			}
			return err
		}
		return tx.Create(&model).Error
	})
	if errors.Is(err, domain.ErrInvalidReference) {
		return err
	}
	return mapWriteError(err)
}

func (r *GORMRepository) ListServings(
	ctx context.Context,
	foodID string,
	options port.ListOptions,
) (port.ServingPage, error) {
	query := r.db.WithContext(ctx).Model(&servingModel{}).Where("food_id = ?", foodID)
	if value := strings.TrimSpace(options.Query); value != "" {
		query = query.Where("name ILIKE ?", "%"+value+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return port.ServingPage{}, fmt.Errorf("count food servings: %w", err)
	}
	var models []servingModel
	if err := query.Order("amount ASC, name ASC").Offset(options.Offset).Limit(options.Limit).Find(&models).Error; err != nil {
		return port.ServingPage{}, fmt.Errorf("list food servings: %w", err)
	}
	items := make([]domain.Serving, 0, len(models))
	for _, model := range models {
		items = append(items, toServing(model))
	}
	return port.ServingPage{Items: items, Total: total}, nil
}

func (r *GORMRepository) GetServing(
	ctx context.Context,
	foodID string,
	servingID string,
) (domain.Serving, error) {
	var model servingModel
	if err := r.db.WithContext(ctx).First(&model, "id = ? AND food_id = ?", servingID, foodID).Error; err != nil {
		return domain.Serving{}, mapReadError("get food serving", err)
	}
	return toServing(model), nil
}

func (r *GORMRepository) UpdateServing(ctx context.Context, serving domain.Serving) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockActiveFood(tx, serving.FoodID); err != nil {
			return err
		}
		result := tx.Model(&servingModel{}).
			Where("id = ? AND food_id = ?", serving.ID, serving.FoodID).
			Updates(map[string]any{
				"name": serving.Name, "amount": serving.Amount, "updated_at": serving.UpdatedAt,
			})
		return mapUpdateResult(result, "update food serving")
	})
	return err
}

func (r *GORMRepository) DeleteServing(ctx context.Context, foodID, servingID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockActiveFood(tx, foodID); err != nil {
			return err
		}
		result := tx.Delete(&servingModel{}, "id = ? AND food_id = ?", servingID, foodID)
		return mapDeleteResult(result, "delete food serving")
	})
}

func lockActiveFood(tx *gorm.DB, foodID string) error {
	var food foodModel
	err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Select("id").First(&food, "id = ?", foodID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("validate active food: %w", err)
	}
	return nil
}

func mapWriteError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return domain.ErrConflict
	case errors.Is(err, gorm.ErrForeignKeyViolated):
		return domain.ErrInvalidReference
	default:
		return fmt.Errorf("write nutrition catalog: %w", err)
	}
}

func mapReadError(operation string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ErrNotFound
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func mapUpdateResult(result *gorm.DB, operation string) error {
	if result.Error != nil {
		return mapWriteError(result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func mapDeleteResult(result *gorm.DB, operation string) error {
	if result.Error != nil {
		return fmt.Errorf("%s: %w", operation, result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func toFoodModel(food domain.Food) foodModel {
	return foodModel{
		ID: food.ID, Name: food.Name, Brand: food.Brand, Description: food.Description,
		ReferenceAmount: food.ReferenceAmount, ReferenceUnit: string(food.ReferenceUnit),
		CaloriesKcal: food.Nutrients.CaloriesKcal, ProteinG: food.Nutrients.ProteinG,
		CarbohydratesG: food.Nutrients.CarbohydratesG, FatG: food.Nutrients.FatG,
		FiberG: food.Nutrients.FiberG, CreatedAt: food.CreatedAt, UpdatedAt: food.UpdatedAt,
	}
}

func toFood(model foodModel) domain.Food {
	return domain.Food{
		ID: model.ID, Name: model.Name, Brand: model.Brand, Description: model.Description,
		ReferenceAmount: model.ReferenceAmount, ReferenceUnit: domain.ReferenceUnit(model.ReferenceUnit),
		Nutrients: domain.Nutrients{
			CaloriesKcal: model.CaloriesKcal, ProteinG: model.ProteinG,
			CarbohydratesG: model.CarbohydratesG, FatG: model.FatG, FiberG: model.FiberG,
		},
		CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt,
	}
}

func toServingModel(serving domain.Serving) servingModel {
	return servingModel{
		ID: serving.ID, FoodID: serving.FoodID, Name: serving.Name, Amount: serving.Amount,
		CreatedAt: serving.CreatedAt, UpdatedAt: serving.UpdatedAt,
	}
}

func toServing(model servingModel) domain.Serving {
	return domain.Serving{
		ID: model.ID, FoodID: model.FoodID, Name: model.Name, Amount: model.Amount,
		CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt,
	}
}
