package port

import (
	"context"

	"github.com/4zure09/body-fitness/internal/modules/exercises/domain"
)

type ListOptions struct {
	Query         string
	Difficulty    domain.Difficulty
	EquipmentID   string
	MuscleGroupID string
	Offset        int
	Limit         int
}

type MuscleGroupPage struct {
	Items []domain.MuscleGroup
	Total int64
}

type EquipmentPage struct {
	Items []domain.Equipment
	Total int64
}

type ExercisePage struct {
	Items []domain.Exercise
	Total int64
}

type Repository interface {
	CreateMuscleGroup(ctx context.Context, muscle domain.MuscleGroup) error
	ListMuscleGroups(ctx context.Context, options ListOptions) (MuscleGroupPage, error)
	GetMuscleGroup(ctx context.Context, id string) (domain.MuscleGroup, error)
	UpdateMuscleGroup(ctx context.Context, muscle domain.MuscleGroup) error
	DeleteMuscleGroup(ctx context.Context, id string) error

	CreateEquipment(ctx context.Context, equipment domain.Equipment) error
	ListEquipment(ctx context.Context, options ListOptions) (EquipmentPage, error)
	GetEquipment(ctx context.Context, id string) (domain.Equipment, error)
	UpdateEquipment(ctx context.Context, equipment domain.Equipment) error
	DeleteEquipment(ctx context.Context, id string) error

	CreateExercise(ctx context.Context, exercise domain.Exercise) error
	ListExercises(ctx context.Context, options ListOptions) (ExercisePage, error)
	GetExercise(ctx context.Context, id string) (domain.Exercise, error)
	UpdateExercise(ctx context.Context, exercise domain.Exercise) error
	DeleteExercise(ctx context.Context, id string) error
}
