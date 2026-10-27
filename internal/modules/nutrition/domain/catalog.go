package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type ReferenceUnit string

const (
	ReferenceUnitGram       ReferenceUnit = "g"
	ReferenceUnitMilliliter ReferenceUnit = "ml"
	ReferenceUnitItem       ReferenceUnit = "unit"
)

type Nutrients struct {
	CaloriesKcal   decimal.Decimal
	ProteinG       decimal.Decimal
	CarbohydratesG decimal.Decimal
	FatG           decimal.Decimal
	FiberG         decimal.Decimal
}

type Food struct {
	ID              string
	Name            string
	Brand           *string
	Description     string
	ReferenceAmount decimal.Decimal
	ReferenceUnit   ReferenceUnit
	Nutrients       Nutrients
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Serving struct {
	ID        string
	FoodID    string
	Name      string
	Amount    decimal.Decimal
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (f Food) NutrientsFor(amount decimal.Decimal) Nutrients {
	if f.ReferenceAmount.IsZero() {
		return Nutrients{}
	}
	ratio := amount.Div(f.ReferenceAmount)
	return Nutrients{
		CaloriesKcal:   f.Nutrients.CaloriesKcal.Mul(ratio),
		ProteinG:       f.Nutrients.ProteinG.Mul(ratio),
		CarbohydratesG: f.Nutrients.CarbohydratesG.Mul(ratio),
		FatG:           f.Nutrients.FatG.Mul(ratio),
		FiberG:         f.Nutrients.FiberG.Mul(ratio),
	}
}
