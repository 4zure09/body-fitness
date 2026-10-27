package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/4zure09/body-fitness/internal/modules/exercises/domain"
	"github.com/4zure09/body-fitness/internal/modules/exercises/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GORMRepository struct {
	db *gorm.DB
}

type muscleGroupModel struct {
	ID          string         `gorm:"type:uuid;primaryKey"`
	Name        string         `gorm:"not null"`
	BodyRegion  string         `gorm:"not null"`
	Description string         `gorm:"not null"`
	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (muscleGroupModel) TableName() string { return "muscle_groups" }

type equipmentModel struct {
	ID          string         `gorm:"type:uuid;primaryKey"`
	Name        string         `gorm:"not null"`
	Description string         `gorm:"not null"`
	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (equipmentModel) TableName() string { return "equipment" }

type exerciseModel struct {
	ID           string                `gorm:"type:uuid;primaryKey"`
	Name         string                `gorm:"not null"`
	Description  string                `gorm:"not null"`
	Instructions string                `gorm:"not null"`
	Difficulty   string                `gorm:"not null"`
	EquipmentID  *string               `gorm:"type:uuid"`
	Equipment    *equipmentModel       `gorm:"foreignKey:EquipmentID"`
	Muscles      []exerciseMuscleModel `gorm:"foreignKey:ExerciseID"`
	CreatedAt    time.Time             `gorm:"not null"`
	UpdatedAt    time.Time             `gorm:"not null"`
	DeletedAt    gorm.DeletedAt        `gorm:"index"`
}

func (exerciseModel) TableName() string { return "exercises" }

type exerciseMuscleModel struct {
	ExerciseID    string           `gorm:"type:uuid;primaryKey"`
	MuscleGroupID string           `gorm:"type:uuid;primaryKey"`
	Involvement   string           `gorm:"not null"`
	MuscleGroup   muscleGroupModel `gorm:"foreignKey:MuscleGroupID"`
	CreatedAt     time.Time        `gorm:"not null"`
}

func (exerciseMuscleModel) TableName() string { return "exercise_muscles" }

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db}
}

func (r *GORMRepository) CreateMuscleGroup(ctx context.Context, muscle domain.MuscleGroup) error {
	model := muscleGroupModel{
		ID: muscle.ID, Name: muscle.Name, BodyRegion: string(muscle.BodyRegion),
		Description: muscle.Description, CreatedAt: muscle.CreatedAt, UpdatedAt: muscle.UpdatedAt,
	}
	return mapWriteError(r.db.WithContext(ctx).Create(&model).Error)
}

func (r *GORMRepository) ListMuscleGroups(
	ctx context.Context,
	options port.ListOptions,
) (port.MuscleGroupPage, error) {
	query := applyTextQuery(r.db.WithContext(ctx).Model(&muscleGroupModel{}), options.Query)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return port.MuscleGroupPage{}, fmt.Errorf("count muscle groups: %w", err)
	}
	var models []muscleGroupModel
	if err := query.Order("name ASC").Offset(options.Offset).Limit(options.Limit).Find(&models).Error; err != nil {
		return port.MuscleGroupPage{}, fmt.Errorf("list muscle groups: %w", err)
	}
	items := make([]domain.MuscleGroup, 0, len(models))
	for _, model := range models {
		items = append(items, toMuscleGroup(model))
	}
	return port.MuscleGroupPage{Items: items, Total: total}, nil
}

func (r *GORMRepository) GetMuscleGroup(ctx context.Context, id string) (domain.MuscleGroup, error) {
	var model muscleGroupModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.MuscleGroup{}, mapReadError("get muscle group", err)
	}
	return toMuscleGroup(model), nil
}

func (r *GORMRepository) UpdateMuscleGroup(ctx context.Context, muscle domain.MuscleGroup) error {
	result := r.db.WithContext(ctx).Model(&muscleGroupModel{}).
		Where("id = ?", muscle.ID).
		Updates(map[string]any{
			"name": muscle.Name, "body_region": muscle.BodyRegion,
			"description": muscle.Description, "updated_at": muscle.UpdatedAt,
		})
	return mapUpdateResult(result, "update muscle group")
}

func (r *GORMRepository) DeleteMuscleGroup(ctx context.Context, id string) error {
	var references int64
	err := r.db.WithContext(ctx).Table("exercise_muscles").
		Joins("JOIN exercises ON exercises.id = exercise_muscles.exercise_id AND exercises.deleted_at IS NULL").
		Where("exercise_muscles.muscle_group_id = ?", id).Count(&references).Error
	if err != nil {
		return fmt.Errorf("check muscle group references: %w", err)
	}
	if references > 0 {
		return domain.ErrInUse
	}
	result := r.db.WithContext(ctx).Delete(&muscleGroupModel{}, "id = ?", id)
	return mapDeleteResult(result, "delete muscle group")
}

func (r *GORMRepository) CreateEquipment(ctx context.Context, equipment domain.Equipment) error {
	model := equipmentModel{
		ID: equipment.ID, Name: equipment.Name, Description: equipment.Description,
		CreatedAt: equipment.CreatedAt, UpdatedAt: equipment.UpdatedAt,
	}
	return mapWriteError(r.db.WithContext(ctx).Create(&model).Error)
}

func (r *GORMRepository) ListEquipment(
	ctx context.Context,
	options port.ListOptions,
) (port.EquipmentPage, error) {
	query := applyTextQuery(r.db.WithContext(ctx).Model(&equipmentModel{}), options.Query)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return port.EquipmentPage{}, fmt.Errorf("count equipment: %w", err)
	}
	var models []equipmentModel
	if err := query.Order("name ASC").Offset(options.Offset).Limit(options.Limit).Find(&models).Error; err != nil {
		return port.EquipmentPage{}, fmt.Errorf("list equipment: %w", err)
	}
	items := make([]domain.Equipment, 0, len(models))
	for _, model := range models {
		items = append(items, toEquipment(model))
	}
	return port.EquipmentPage{Items: items, Total: total}, nil
}

func (r *GORMRepository) GetEquipment(ctx context.Context, id string) (domain.Equipment, error) {
	var model equipmentModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return domain.Equipment{}, mapReadError("get equipment", err)
	}
	return toEquipment(model), nil
}

func (r *GORMRepository) UpdateEquipment(ctx context.Context, equipment domain.Equipment) error {
	result := r.db.WithContext(ctx).Model(&equipmentModel{}).
		Where("id = ?", equipment.ID).
		Updates(map[string]any{
			"name": equipment.Name, "description": equipment.Description,
			"updated_at": equipment.UpdatedAt,
		})
	return mapUpdateResult(result, "update equipment")
}

func (r *GORMRepository) DeleteEquipment(ctx context.Context, id string) error {
	var references int64
	if err := r.db.WithContext(ctx).Model(&exerciseModel{}).Where("equipment_id = ?", id).Count(&references).Error; err != nil {
		return fmt.Errorf("check equipment references: %w", err)
	}
	if references > 0 {
		return domain.ErrInUse
	}
	result := r.db.WithContext(ctx).Delete(&equipmentModel{}, "id = ?", id)
	return mapDeleteResult(result, "delete equipment")
}

func (r *GORMRepository) CreateExercise(ctx context.Context, exercise domain.Exercise) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := validateActiveExerciseReferences(tx, exercise); err != nil {
			return err
		}
		model := toExerciseModel(exercise)
		if err := tx.Omit("Equipment", "Muscles").Create(&model).Error; err != nil {
			return err
		}
		return createExerciseMuscles(tx, exercise)
	})
	return mapWriteError(err)
}

func (r *GORMRepository) ListExercises(
	ctx context.Context,
	options port.ListOptions,
) (port.ExercisePage, error) {
	query := r.db.WithContext(ctx).Model(&exerciseModel{})
	query = applyTextQuery(query, options.Query)
	if options.Difficulty != "" {
		query = query.Where("difficulty = ?", options.Difficulty)
	}
	if options.EquipmentID != "" {
		query = query.Where("equipment_id = ?", options.EquipmentID)
	}
	if options.MuscleGroupID != "" {
		query = query.Where(
			"EXISTS (SELECT 1 FROM exercise_muscles em WHERE em.exercise_id = exercises.id AND em.muscle_group_id = ?)",
			options.MuscleGroupID,
		)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return port.ExercisePage{}, fmt.Errorf("count exercises: %w", err)
	}
	var models []exerciseModel
	err := withExerciseAssociations(query).
		Order("name ASC").Offset(options.Offset).Limit(options.Limit).Find(&models).Error
	if err != nil {
		return port.ExercisePage{}, fmt.Errorf("list exercises: %w", err)
	}
	items := make([]domain.Exercise, 0, len(models))
	for _, model := range models {
		items = append(items, toExercise(model))
	}
	return port.ExercisePage{Items: items, Total: total}, nil
}

func (r *GORMRepository) GetExercise(ctx context.Context, id string) (domain.Exercise, error) {
	var model exerciseModel
	err := withExerciseAssociations(r.db.WithContext(ctx)).First(&model, "id = ?", id).Error
	if err != nil {
		return domain.Exercise{}, mapReadError("get exercise", err)
	}
	return toExercise(model), nil
}

func (r *GORMRepository) UpdateExercise(ctx context.Context, exercise domain.Exercise) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := validateActiveExerciseReferences(tx, exercise); err != nil {
			return err
		}
		model := toExerciseModel(exercise)
		result := tx.Model(&exerciseModel{}).Where("id = ?", exercise.ID).Updates(map[string]any{
			"name": model.Name, "description": model.Description, "instructions": model.Instructions,
			"difficulty": model.Difficulty, "equipment_id": model.EquipmentID, "updated_at": model.UpdatedAt,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrNotFound
		}
		if err := tx.Where("exercise_id = ?", exercise.ID).Delete(&exerciseMuscleModel{}).Error; err != nil {
			return err
		}
		return createExerciseMuscles(tx, exercise)
	})
	if errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrInvalidReference) {
		return err
	}
	return mapWriteError(err)
}

func (r *GORMRepository) DeleteExercise(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&exerciseModel{}, "id = ?", id)
	return mapDeleteResult(result, "delete exercise")
}

func withExerciseAssociations(query *gorm.DB) *gorm.DB {
	return query.Preload("Equipment").Preload("Muscles").Preload("Muscles.MuscleGroup")
}

func createExerciseMuscles(tx *gorm.DB, exercise domain.Exercise) error {
	models := make([]exerciseMuscleModel, 0, len(exercise.Muscles))
	for _, muscle := range exercise.Muscles {
		models = append(models, exerciseMuscleModel{
			ExerciseID: exercise.ID, MuscleGroupID: muscle.Muscle.ID,
			Involvement: string(muscle.Involvement), CreatedAt: exercise.UpdatedAt,
		})
	}
	if len(models) == 0 {
		return nil
	}
	return tx.Create(&models).Error
}

func validateActiveExerciseReferences(tx *gorm.DB, exercise domain.Exercise) error {
	if exercise.Equipment != nil {
		var equipment equipmentModel
		err := tx.Clauses(clause.Locking{Strength: "SHARE"}).
			Select("id").First(&equipment, "id = ?", exercise.Equipment.ID).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrInvalidReference
		}
		if err != nil {
			return fmt.Errorf("validate exercise equipment: %w", err)
		}
	}

	ids := make([]string, 0, len(exercise.Muscles))
	for _, muscle := range exercise.Muscles {
		ids = append(ids, muscle.Muscle.ID)
	}
	if len(ids) == 0 {
		return nil
	}
	var muscles []muscleGroupModel
	err := tx.Clauses(clause.Locking{Strength: "SHARE"}).
		Select("id").Where("id IN ?", ids).Find(&muscles).Error
	if err != nil {
		return fmt.Errorf("validate exercise muscle groups: %w", err)
	}
	if len(muscles) != len(ids) {
		return domain.ErrInvalidReference
	}
	return nil
}

func applyTextQuery(query *gorm.DB, value string) *gorm.DB {
	value = strings.TrimSpace(value)
	if value == "" {
		return query
	}
	return query.Where("name ILIKE ?", "%"+value+"%")
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
		return fmt.Errorf("write exercise catalog: %w", err)
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

func toMuscleGroup(model muscleGroupModel) domain.MuscleGroup {
	return domain.MuscleGroup{
		ID: model.ID, Name: model.Name, BodyRegion: domain.BodyRegion(model.BodyRegion),
		Description: model.Description, CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt,
	}
}

func toEquipment(model equipmentModel) domain.Equipment {
	return domain.Equipment{
		ID: model.ID, Name: model.Name, Description: model.Description,
		CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt,
	}
}

func toExerciseModel(exercise domain.Exercise) exerciseModel {
	var equipmentID *string
	if exercise.Equipment != nil {
		id := exercise.Equipment.ID
		equipmentID = &id
	}
	return exerciseModel{
		ID: exercise.ID, Name: exercise.Name, Description: exercise.Description,
		Instructions: exercise.Instructions, Difficulty: string(exercise.Difficulty),
		EquipmentID: equipmentID, CreatedAt: exercise.CreatedAt, UpdatedAt: exercise.UpdatedAt,
	}
}

func toExercise(model exerciseModel) domain.Exercise {
	var equipment *domain.Equipment
	if model.Equipment != nil {
		mapped := toEquipment(*model.Equipment)
		equipment = &mapped
	}
	muscles := make([]domain.ExerciseMuscle, 0, len(model.Muscles))
	for _, muscle := range model.Muscles {
		muscles = append(muscles, domain.ExerciseMuscle{
			Muscle: toMuscleGroup(muscle.MuscleGroup), Involvement: domain.Involvement(muscle.Involvement),
		})
	}
	sort.Slice(muscles, func(i, j int) bool {
		if muscles[i].Involvement == muscles[j].Involvement {
			return muscles[i].Muscle.Name < muscles[j].Muscle.Name
		}
		return muscles[i].Involvement < muscles[j].Involvement
	})
	return domain.Exercise{
		ID: model.ID, Name: model.Name, Description: model.Description,
		Instructions: model.Instructions, Difficulty: domain.Difficulty(model.Difficulty),
		Equipment: equipment, Muscles: muscles, CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt,
	}
}
