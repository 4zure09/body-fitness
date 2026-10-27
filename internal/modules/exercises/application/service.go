package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/4zure09/body-fitness/internal/modules/exercises/domain"
	"github.com/4zure09/body-fitness/internal/modules/exercises/port"
	"github.com/google/uuid"
)

type Service struct {
	repository port.Repository
	now        func() time.Time
	newID      func() string
}

type MuscleGroupInput struct {
	Name        string
	BodyRegion  domain.BodyRegion
	Description string
}

type EquipmentInput struct {
	Name        string
	Description string
}

type ExerciseInput struct {
	Name               string
	Description        string
	Instructions       string
	Difficulty         domain.Difficulty
	EquipmentID        *string
	PrimaryMuscleIDs   []string
	SecondaryMuscleIDs []string
}

func NewService(repository port.Repository) *Service {
	return &Service{repository: repository, now: time.Now, newID: uuid.NewString}
}

func (s *Service) CreateMuscleGroup(ctx context.Context, input MuscleGroupInput) (domain.MuscleGroup, error) {
	muscle, err := s.buildMuscleGroup(s.newID(), input)
	if err != nil {
		return domain.MuscleGroup{}, err
	}
	if err := s.repository.CreateMuscleGroup(ctx, muscle); err != nil {
		return domain.MuscleGroup{}, err
	}
	return muscle, nil
}

func (s *Service) ListMuscleGroups(ctx context.Context, options port.ListOptions) (port.MuscleGroupPage, error) {
	return s.repository.ListMuscleGroups(ctx, options)
}

func (s *Service) GetMuscleGroup(ctx context.Context, id string) (domain.MuscleGroup, error) {
	if err := validateID(id); err != nil {
		return domain.MuscleGroup{}, err
	}
	return s.repository.GetMuscleGroup(ctx, id)
}

func (s *Service) UpdateMuscleGroup(
	ctx context.Context,
	id string,
	input MuscleGroupInput,
) (domain.MuscleGroup, error) {
	if err := validateID(id); err != nil {
		return domain.MuscleGroup{}, err
	}
	muscle, err := s.buildMuscleGroup(id, input)
	if err != nil {
		return domain.MuscleGroup{}, err
	}
	if err := s.repository.UpdateMuscleGroup(ctx, muscle); err != nil {
		return domain.MuscleGroup{}, err
	}
	return s.repository.GetMuscleGroup(ctx, id)
}

func (s *Service) DeleteMuscleGroup(ctx context.Context, id string) error {
	if err := validateID(id); err != nil {
		return err
	}
	return s.repository.DeleteMuscleGroup(ctx, id)
}

func (s *Service) CreateEquipment(ctx context.Context, input EquipmentInput) (domain.Equipment, error) {
	equipment, err := s.buildEquipment(s.newID(), input)
	if err != nil {
		return domain.Equipment{}, err
	}
	if err := s.repository.CreateEquipment(ctx, equipment); err != nil {
		return domain.Equipment{}, err
	}
	return equipment, nil
}

func (s *Service) ListEquipment(ctx context.Context, options port.ListOptions) (port.EquipmentPage, error) {
	return s.repository.ListEquipment(ctx, options)
}

func (s *Service) GetEquipment(ctx context.Context, id string) (domain.Equipment, error) {
	if err := validateID(id); err != nil {
		return domain.Equipment{}, err
	}
	return s.repository.GetEquipment(ctx, id)
}

func (s *Service) UpdateEquipment(
	ctx context.Context,
	id string,
	input EquipmentInput,
) (domain.Equipment, error) {
	if err := validateID(id); err != nil {
		return domain.Equipment{}, err
	}
	equipment, err := s.buildEquipment(id, input)
	if err != nil {
		return domain.Equipment{}, err
	}
	if err := s.repository.UpdateEquipment(ctx, equipment); err != nil {
		return domain.Equipment{}, err
	}
	return s.repository.GetEquipment(ctx, id)
}

func (s *Service) DeleteEquipment(ctx context.Context, id string) error {
	if err := validateID(id); err != nil {
		return err
	}
	return s.repository.DeleteEquipment(ctx, id)
}

func (s *Service) CreateExercise(ctx context.Context, input ExerciseInput) (domain.Exercise, error) {
	exercise, err := s.buildExercise(s.newID(), input)
	if err != nil {
		return domain.Exercise{}, err
	}
	if err := s.repository.CreateExercise(ctx, exercise); err != nil {
		return domain.Exercise{}, err
	}
	return s.repository.GetExercise(ctx, exercise.ID)
}

func (s *Service) ListExercises(ctx context.Context, options port.ListOptions) (port.ExercisePage, error) {
	if options.Difficulty != "" && !validDifficulty(options.Difficulty) {
		return port.ExercisePage{}, fmt.Errorf("%w: difficulty is invalid", domain.ErrInvalidInput)
	}
	if options.EquipmentID != "" {
		if err := validateID(options.EquipmentID); err != nil {
			return port.ExercisePage{}, err
		}
	}
	if options.MuscleGroupID != "" {
		if err := validateID(options.MuscleGroupID); err != nil {
			return port.ExercisePage{}, err
		}
	}
	return s.repository.ListExercises(ctx, options)
}

func (s *Service) GetExercise(ctx context.Context, id string) (domain.Exercise, error) {
	if err := validateID(id); err != nil {
		return domain.Exercise{}, err
	}
	return s.repository.GetExercise(ctx, id)
}

func (s *Service) UpdateExercise(ctx context.Context, id string, input ExerciseInput) (domain.Exercise, error) {
	if err := validateID(id); err != nil {
		return domain.Exercise{}, err
	}
	exercise, err := s.buildExercise(id, input)
	if err != nil {
		return domain.Exercise{}, err
	}
	if err := s.repository.UpdateExercise(ctx, exercise); err != nil {
		return domain.Exercise{}, err
	}
	return s.repository.GetExercise(ctx, id)
}

func (s *Service) DeleteExercise(ctx context.Context, id string) error {
	if err := validateID(id); err != nil {
		return err
	}
	return s.repository.DeleteExercise(ctx, id)
}

func (s *Service) buildMuscleGroup(id string, input MuscleGroupInput) (domain.MuscleGroup, error) {
	name, err := normalizedRequired(input.Name, "name", 2, 120)
	if err != nil {
		return domain.MuscleGroup{}, err
	}
	if !validBodyRegion(input.BodyRegion) {
		return domain.MuscleGroup{}, fmt.Errorf("%w: body_region is invalid", domain.ErrInvalidInput)
	}
	description, err := normalizedOptional(input.Description, "description", 2000)
	if err != nil {
		return domain.MuscleGroup{}, err
	}
	now := s.now().UTC()
	return domain.MuscleGroup{
		ID: id, Name: name, BodyRegion: input.BodyRegion, Description: description,
		CreatedAt: now, UpdatedAt: now,
	}, nil
}

func (s *Service) buildEquipment(id string, input EquipmentInput) (domain.Equipment, error) {
	name, err := normalizedRequired(input.Name, "name", 2, 120)
	if err != nil {
		return domain.Equipment{}, err
	}
	description, err := normalizedOptional(input.Description, "description", 2000)
	if err != nil {
		return domain.Equipment{}, err
	}
	now := s.now().UTC()
	return domain.Equipment{
		ID: id, Name: name, Description: description, CreatedAt: now, UpdatedAt: now,
	}, nil
}

func (s *Service) buildExercise(id string, input ExerciseInput) (domain.Exercise, error) {
	name, err := normalizedRequired(input.Name, "name", 2, 160)
	if err != nil {
		return domain.Exercise{}, err
	}
	description, err := normalizedOptional(input.Description, "description", 4000)
	if err != nil {
		return domain.Exercise{}, err
	}
	instructions, err := normalizedRequired(input.Instructions, "instructions", 2, 10000)
	if err != nil {
		return domain.Exercise{}, err
	}
	if !validDifficulty(input.Difficulty) {
		return domain.Exercise{}, fmt.Errorf("%w: difficulty is invalid", domain.ErrInvalidInput)
	}
	equipment, err := equipmentReference(input.EquipmentID)
	if err != nil {
		return domain.Exercise{}, err
	}
	muscles, err := exerciseMuscles(input.PrimaryMuscleIDs, input.SecondaryMuscleIDs)
	if err != nil {
		return domain.Exercise{}, err
	}
	now := s.now().UTC()
	return domain.Exercise{
		ID: id, Name: name, Description: description, Instructions: instructions,
		Difficulty: input.Difficulty, Equipment: equipment, Muscles: muscles,
		CreatedAt: now, UpdatedAt: now,
	}, nil
}

func equipmentReference(id *string) (*domain.Equipment, error) {
	if id == nil || strings.TrimSpace(*id) == "" {
		return nil, nil
	}
	if err := validateID(*id); err != nil {
		return nil, err
	}
	return &domain.Equipment{ID: *id}, nil
}

func exerciseMuscles(primaryIDs, secondaryIDs []string) ([]domain.ExerciseMuscle, error) {
	if len(primaryIDs) == 0 {
		return nil, fmt.Errorf("%w: at least one primary_muscle_id is required", domain.ErrInvalidInput)
	}
	seen := make(map[string]struct{}, len(primaryIDs)+len(secondaryIDs))
	muscles := make([]domain.ExerciseMuscle, 0, len(primaryIDs)+len(secondaryIDs))
	appendMuscles := func(ids []string, involvement domain.Involvement) error {
		for _, id := range ids {
			if err := validateID(id); err != nil {
				return err
			}
			if _, exists := seen[id]; exists {
				return fmt.Errorf("%w: muscle group IDs must be unique", domain.ErrInvalidInput)
			}
			seen[id] = struct{}{}
			muscles = append(muscles, domain.ExerciseMuscle{
				Muscle: domain.MuscleGroup{ID: id}, Involvement: involvement,
			})
		}
		return nil
	}
	if err := appendMuscles(primaryIDs, domain.InvolvementPrimary); err != nil {
		return nil, err
	}
	if err := appendMuscles(secondaryIDs, domain.InvolvementSecondary); err != nil {
		return nil, err
	}
	return muscles, nil
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

func validateID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("%w: ID must be a valid UUID", domain.ErrInvalidInput)
	}
	return nil
}

func validBodyRegion(value domain.BodyRegion) bool {
	switch value {
	case domain.BodyRegionUpper, domain.BodyRegionLower, domain.BodyRegionCore, domain.BodyRegionFull:
		return true
	default:
		return false
	}
}

func validDifficulty(value domain.Difficulty) bool {
	switch value {
	case domain.DifficultyBeginner, domain.DifficultyIntermediate, domain.DifficultyAdvanced:
		return true
	default:
		return false
	}
}
