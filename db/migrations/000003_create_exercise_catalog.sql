-- +goose Up
CREATE TABLE muscle_groups (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    body_region TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT muscle_groups_name_check CHECK (char_length(name) BETWEEN 2 AND 120),
    CONSTRAINT muscle_groups_body_region_check CHECK (body_region IN ('upper_body', 'lower_body', 'core', 'full_body'))
);

CREATE UNIQUE INDEX muscle_groups_name_unique
    ON muscle_groups (LOWER(name)) WHERE deleted_at IS NULL;
CREATE INDEX muscle_groups_active_idx ON muscle_groups (id) WHERE deleted_at IS NULL;

CREATE TABLE equipment (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT equipment_name_check CHECK (char_length(name) BETWEEN 2 AND 120)
);

CREATE UNIQUE INDEX equipment_name_unique
    ON equipment (LOWER(name)) WHERE deleted_at IS NULL;
CREATE INDEX equipment_active_idx ON equipment (id) WHERE deleted_at IS NULL;

CREATE TABLE exercises (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    instructions TEXT NOT NULL,
    difficulty TEXT NOT NULL,
    equipment_id UUID REFERENCES equipment (id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT exercises_name_check CHECK (char_length(name) BETWEEN 2 AND 160),
    CONSTRAINT exercises_difficulty_check CHECK (difficulty IN ('beginner', 'intermediate', 'advanced'))
);

CREATE UNIQUE INDEX exercises_name_unique
    ON exercises (LOWER(name)) WHERE deleted_at IS NULL;
CREATE INDEX exercises_equipment_id_idx ON exercises (equipment_id) WHERE deleted_at IS NULL;
CREATE INDEX exercises_difficulty_idx ON exercises (difficulty) WHERE deleted_at IS NULL;

CREATE TABLE exercise_muscles (
    exercise_id UUID NOT NULL REFERENCES exercises (id) ON DELETE CASCADE,
    muscle_group_id UUID NOT NULL REFERENCES muscle_groups (id) ON DELETE RESTRICT,
    involvement TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (exercise_id, muscle_group_id),
    CONSTRAINT exercise_muscles_involvement_check CHECK (involvement IN ('primary', 'secondary'))
);

CREATE INDEX exercise_muscles_group_idx ON exercise_muscles (muscle_group_id, involvement);

-- +goose Down
DROP TABLE IF EXISTS exercise_muscles;
DROP TABLE IF EXISTS exercises;
DROP TABLE IF EXISTS equipment;
DROP TABLE IF EXISTS muscle_groups;
