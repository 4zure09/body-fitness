package repository

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/4zure09/body-fitness/internal/modules/exercises/domain"
	"github.com/4zure09/body-fitness/internal/modules/exercises/port"
	"github.com/4zure09/body-fitness/internal/platform/config"
	"github.com/4zure09/body-fitness/internal/platform/database"
	"github.com/google/uuid"
)

func TestGORMRepositoryExerciseCatalogLifecycle(t *testing.T) {
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
	primary := domain.MuscleGroup{
		ID: uuid.NewString(), Name: "Primary " + uuid.NewString(), BodyRegion: domain.BodyRegionUpper,
		Description: "Primary muscle", CreatedAt: now, UpdatedAt: now,
	}
	secondary := domain.MuscleGroup{
		ID: uuid.NewString(), Name: "Secondary " + uuid.NewString(), BodyRegion: domain.BodyRegionUpper,
		Description: "Secondary muscle", CreatedAt: now, UpdatedAt: now,
	}
	equipment := domain.Equipment{
		ID: uuid.NewString(), Name: "Equipment " + uuid.NewString(),
		Description: "Test equipment", CreatedAt: now, UpdatedAt: now,
	}
	exerciseID := uuid.NewString()
	t.Cleanup(func() {
		_ = db.Unscoped().Delete(&exerciseModel{}, "id = ?", exerciseID).Error
		_ = db.Unscoped().Delete(&equipmentModel{}, "id = ?", equipment.ID).Error
		_ = db.Unscoped().Delete(&muscleGroupModel{}, "id IN ?", []string{primary.ID, secondary.ID}).Error
	})

	if err := repository.CreateMuscleGroup(ctx, primary); err != nil {
		t.Fatalf("CreateMuscleGroup(primary) error = %v", err)
	}
	if err := repository.CreateMuscleGroup(ctx, secondary); err != nil {
		t.Fatalf("CreateMuscleGroup(secondary) error = %v", err)
	}
	duplicate := primary
	duplicate.ID = uuid.NewString()
	if err := repository.CreateMuscleGroup(ctx, duplicate); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("duplicate muscle error = %v, want conflict", err)
	}
	if err := repository.CreateEquipment(ctx, equipment); err != nil {
		t.Fatalf("CreateEquipment() error = %v", err)
	}

	exercise := domain.Exercise{
		ID: exerciseID, Name: "Exercise " + uuid.NewString(), Description: "Description",
		Instructions: "Perform the exercise safely.", Difficulty: domain.DifficultyBeginner,
		Equipment: &domain.Equipment{ID: equipment.ID},
		Muscles: []domain.ExerciseMuscle{
			{Muscle: domain.MuscleGroup{ID: primary.ID}, Involvement: domain.InvolvementPrimary},
			{Muscle: domain.MuscleGroup{ID: secondary.ID}, Involvement: domain.InvolvementSecondary},
		},
		CreatedAt: now, UpdatedAt: now,
	}
	if err := repository.CreateExercise(ctx, exercise); err != nil {
		t.Fatalf("CreateExercise() error = %v", err)
	}
	stored, err := repository.GetExercise(ctx, exercise.ID)
	if err != nil {
		t.Fatalf("GetExercise() error = %v", err)
	}
	if stored.Equipment == nil || stored.Equipment.ID != equipment.ID || len(stored.Muscles) != 2 {
		t.Fatalf("unexpected stored exercise: %+v", stored)
	}

	filtered, err := repository.ListExercises(ctx, port.ListOptions{
		MuscleGroupID: primary.ID, Difficulty: domain.DifficultyBeginner, Limit: 20,
	})
	if err != nil {
		t.Fatalf("ListExercises() error = %v", err)
	}
	if filtered.Total != 1 || len(filtered.Items) != 1 || filtered.Items[0].ID != exercise.ID {
		t.Fatalf("unexpected filtered exercises: %+v", filtered)
	}
	if err := repository.DeleteMuscleGroup(ctx, primary.ID); !errors.Is(err, domain.ErrInUse) {
		t.Fatalf("DeleteMuscleGroup() error = %v, want in use", err)
	}
	if err := repository.DeleteEquipment(ctx, equipment.ID); !errors.Is(err, domain.ErrInUse) {
		t.Fatalf("DeleteEquipment() error = %v, want in use", err)
	}

	exercise.Name = "Updated " + exercise.Name
	exercise.Difficulty = domain.DifficultyIntermediate
	exercise.Muscles = exercise.Muscles[:1]
	exercise.UpdatedAt = now.Add(time.Minute)
	if err := repository.UpdateExercise(ctx, exercise); err != nil {
		t.Fatalf("UpdateExercise() error = %v", err)
	}
	stored, err = repository.GetExercise(ctx, exercise.ID)
	if err != nil || stored.Name != exercise.Name || len(stored.Muscles) != 1 {
		t.Fatalf("unexpected updated exercise: %+v, error=%v", stored, err)
	}

	if err := repository.DeleteExercise(ctx, exercise.ID); err != nil {
		t.Fatalf("DeleteExercise() error = %v", err)
	}
	if _, err := repository.GetExercise(ctx, exercise.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("deleted exercise error = %v, want not found", err)
	}
	if err := repository.DeleteMuscleGroup(ctx, primary.ID); err != nil {
		t.Fatalf("DeleteMuscleGroup() after exercise delete error = %v", err)
	}
	if err := repository.DeleteEquipment(ctx, equipment.ID); err != nil {
		t.Fatalf("DeleteEquipment() after exercise delete error = %v", err)
	}
	invalidExercise := exercise
	invalidExercise.ID = uuid.NewString()
	invalidExercise.Name = "Invalid reference " + uuid.NewString()
	if err := repository.CreateExercise(ctx, invalidExercise); !errors.Is(err, domain.ErrInvalidReference) {
		t.Fatalf("CreateExercise() with soft-deleted references error = %v, want invalid reference", err)
	}
}
