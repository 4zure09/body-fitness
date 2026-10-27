-- +goose Up
CREATE TABLE foods (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    brand TEXT,
    description TEXT NOT NULL DEFAULT '',
    reference_amount NUMERIC(10, 3) NOT NULL,
    reference_unit TEXT NOT NULL,
    calories_kcal NUMERIC(10, 2) NOT NULL,
    protein_g NUMERIC(10, 2) NOT NULL,
    carbohydrates_g NUMERIC(10, 2) NOT NULL,
    fat_g NUMERIC(10, 2) NOT NULL,
    fiber_g NUMERIC(10, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT foods_name_check CHECK (char_length(name) BETWEEN 2 AND 160),
    CONSTRAINT foods_reference_amount_check CHECK (reference_amount > 0),
    CONSTRAINT foods_reference_unit_check CHECK (reference_unit IN ('g', 'ml', 'unit')),
    CONSTRAINT foods_nutrition_nonnegative_check CHECK (
        calories_kcal >= 0 AND protein_g >= 0 AND carbohydrates_g >= 0 AND fat_g >= 0 AND fiber_g >= 0
    )
);

CREATE UNIQUE INDEX foods_name_brand_unique
    ON foods (LOWER(name), LOWER(COALESCE(brand, ''))) WHERE deleted_at IS NULL;
CREATE INDEX foods_active_idx ON foods (id) WHERE deleted_at IS NULL;

CREATE TABLE food_servings (
    id UUID PRIMARY KEY,
    food_id UUID NOT NULL REFERENCES foods (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    amount NUMERIC(10, 3) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT food_servings_name_check CHECK (char_length(name) BETWEEN 1 AND 120),
    CONSTRAINT food_servings_amount_check CHECK (amount > 0)
);

CREATE INDEX food_servings_food_id_idx ON food_servings (food_id);
CREATE UNIQUE INDEX food_servings_name_unique ON food_servings (food_id, LOWER(name));

-- +goose Down
DROP TABLE IF EXISTS food_servings;
DROP TABLE IF EXISTS foods;
