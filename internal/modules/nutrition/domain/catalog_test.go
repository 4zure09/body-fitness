package domain

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestNutrientsForServingAmount(t *testing.T) {
	food := Food{
		ReferenceAmount: decimal.NewFromInt(100),
		Nutrients: Nutrients{
			CaloriesKcal:   decimal.NewFromInt(200),
			ProteinG:       decimal.NewFromInt(10),
			CarbohydratesG: decimal.NewFromInt(20),
			FatG:           decimal.NewFromInt(8),
			FiberG:         decimal.NewFromInt(4),
		},
	}
	result := food.NutrientsFor(decimal.NewFromInt(25))
	if !result.CaloriesKcal.Equal(decimal.NewFromInt(50)) ||
		!result.ProteinG.Equal(decimal.NewFromFloat(2.5)) ||
		!result.FatG.Equal(decimal.NewFromInt(2)) {
		t.Fatalf("unexpected serving nutrients: %+v", result)
	}
}
