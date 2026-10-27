package domain

import "time"

type BodyRegion string

const (
	BodyRegionUpper BodyRegion = "upper_body"
	BodyRegionLower BodyRegion = "lower_body"
	BodyRegionCore  BodyRegion = "core"
	BodyRegionFull  BodyRegion = "full_body"
)

type Difficulty string

const (
	DifficultyBeginner     Difficulty = "beginner"
	DifficultyIntermediate Difficulty = "intermediate"
	DifficultyAdvanced     Difficulty = "advanced"
)

type Involvement string

const (
	InvolvementPrimary   Involvement = "primary"
	InvolvementSecondary Involvement = "secondary"
)

type MuscleGroup struct {
	ID          string
	Name        string
	BodyRegion  BodyRegion
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Equipment struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ExerciseMuscle struct {
	Muscle      MuscleGroup
	Involvement Involvement
}

type Exercise struct {
	ID           string
	Name         string
	Description  string
	Instructions string
	Difficulty   Difficulty
	Equipment    *Equipment
	Muscles      []ExerciseMuscle
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
