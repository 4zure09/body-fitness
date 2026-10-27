package application

import (
	"errors"
	"testing"
	"time"

	"github.com/4zure09/body-fitness/internal/modules/exercises/domain"
	"github.com/google/uuid"
)

func TestBuildExerciseRequiresPrimaryMuscle(t *testing.T) {
	service := testService()
	_, err := service.buildExercise(uuid.NewString(), ExerciseInput{
		Name: "Push-up", Instructions: "Lower and raise the body.", Difficulty: domain.DifficultyBeginner,
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("buildExercise() error = %v, want invalid input", err)
	}
}

func TestBuildExerciseRejectsDuplicateMuscleAcrossInvolvement(t *testing.T) {
	service := testService()
	muscleID := uuid.NewString()
	_, err := service.buildExercise(uuid.NewString(), ExerciseInput{
		Name: "Push-up", Instructions: "Lower and raise the body.", Difficulty: domain.DifficultyBeginner,
		PrimaryMuscleIDs: []string{muscleID}, SecondaryMuscleIDs: []string{muscleID},
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("buildExercise() error = %v, want invalid input", err)
	}
}

func TestBuildExerciseCreatesPrimaryAndSecondaryMuscles(t *testing.T) {
	service := testService()
	primaryID := uuid.NewString()
	secondaryID := uuid.NewString()
	exercise, err := service.buildExercise(uuid.NewString(), ExerciseInput{
		Name: " Push-up ", Instructions: " Lower and raise the body. ", Difficulty: domain.DifficultyBeginner,
		PrimaryMuscleIDs: []string{primaryID}, SecondaryMuscleIDs: []string{secondaryID},
	})
	if err != nil {
		t.Fatalf("buildExercise() error = %v", err)
	}
	if exercise.Name != "Push-up" || len(exercise.Muscles) != 2 {
		t.Fatalf("unexpected exercise: %+v", exercise)
	}
	if exercise.Muscles[0].Involvement != domain.InvolvementPrimary ||
		exercise.Muscles[1].Involvement != domain.InvolvementSecondary {
		t.Fatalf("unexpected muscle involvement: %+v", exercise.Muscles)
	}
}

func TestBuildMuscleGroupRejectsUnknownRegion(t *testing.T) {
	service := testService()
	_, err := service.buildMuscleGroup(uuid.NewString(), MuscleGroupInput{
		Name: "Chest", BodyRegion: domain.BodyRegion("unknown"),
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("buildMuscleGroup() error = %v, want invalid input", err)
	}
}

func testService() *Service {
	fixedNow := time.Date(2026, 10, 27, 10, 0, 0, 0, time.UTC)
	return &Service{now: func() time.Time { return fixedNow }, newID: uuid.NewString}
}
