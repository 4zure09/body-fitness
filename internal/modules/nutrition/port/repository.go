package port

import (
	"context"

	"github.com/4zure09/body-fitness/internal/modules/nutrition/domain"
)

type ListOptions struct {
	Query  string
	Brand  string
	Offset int
	Limit  int
}

type FoodPage struct {
	Items []domain.Food
	Total int64
}

type ServingPage struct {
	Items []domain.Serving
	Total int64
}

type Repository interface {
	CreateFood(ctx context.Context, food domain.Food) error
	ListFoods(ctx context.Context, options ListOptions) (FoodPage, error)
	GetFood(ctx context.Context, id string) (domain.Food, error)
	UpdateFood(ctx context.Context, food domain.Food) error
	DeleteFood(ctx context.Context, id string) error

	CreateServing(ctx context.Context, serving domain.Serving) error
	ListServings(ctx context.Context, foodID string, options ListOptions) (ServingPage, error)
	GetServing(ctx context.Context, foodID, servingID string) (domain.Serving, error)
	UpdateServing(ctx context.Context, serving domain.Serving) error
	DeleteServing(ctx context.Context, foodID, servingID string) error
}
