package nutritionhttp

import (
	"errors"
	"net/http"
	"time"

	authdomain "github.com/4zure09/body-fitness/internal/modules/auth/domain"
	authhttp "github.com/4zure09/body-fitness/internal/modules/auth/transport/http"
	"github.com/4zure09/body-fitness/internal/modules/nutrition/application"
	"github.com/4zure09/body-fitness/internal/modules/nutrition/domain"
	"github.com/4zure09/body-fitness/internal/modules/nutrition/port"
	"github.com/4zure09/body-fitness/internal/shared/httpx"
	"github.com/4zure09/body-fitness/internal/shared/pagination"
	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
)

type Handler struct {
	service *application.Service
	auth    authhttp.Middleware
}

type nutrientsRequest struct {
	CaloriesKcal   float64 `json:"calories_kcal"`
	ProteinG       float64 `json:"protein_g"`
	CarbohydratesG float64 `json:"carbohydrates_g"`
	FatG           float64 `json:"fat_g"`
	FiberG         float64 `json:"fiber_g"`
}

type foodRequest struct {
	Name            string               `json:"name"`
	Brand           *string              `json:"brand"`
	Description     string               `json:"description"`
	ReferenceAmount float64              `json:"reference_amount"`
	ReferenceUnit   domain.ReferenceUnit `json:"reference_unit"`
	Nutrients       nutrientsRequest     `json:"nutrients"`
}

type servingRequest struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

type nutrientsResponse struct {
	CaloriesKcal   float64 `json:"calories_kcal"`
	ProteinG       float64 `json:"protein_g"`
	CarbohydratesG float64 `json:"carbohydrates_g"`
	FatG           float64 `json:"fat_g"`
	FiberG         float64 `json:"fiber_g"`
}

type foodResponse struct {
	ID              string               `json:"id"`
	Name            string               `json:"name"`
	Brand           *string              `json:"brand,omitempty"`
	Description     string               `json:"description"`
	ReferenceAmount float64              `json:"reference_amount"`
	ReferenceUnit   domain.ReferenceUnit `json:"reference_unit"`
	Nutrients       nutrientsResponse    `json:"nutrients"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
}

type servingResponse struct {
	ID        string               `json:"id"`
	FoodID    string               `json:"food_id"`
	Name      string               `json:"name"`
	Amount    float64              `json:"amount"`
	Unit      domain.ReferenceUnit `json:"unit"`
	Nutrition nutrientsResponse    `json:"nutrition"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
}

func NewHandler(service *application.Service, auth authhttp.Middleware) Handler {
	return Handler{service: service, auth: auth}
}

func (h Handler) RegisterRoutes(router chi.Router) {
	router.Get("/foods", h.ListFoods)
	router.With(h.adminOnly()).Post("/foods", h.CreateFood)
	router.Get("/foods/{foodID}", h.GetFood)
	router.With(h.adminOnly()).Put("/foods/{foodID}", h.UpdateFood)
	router.With(h.adminOnly()).Delete("/foods/{foodID}", h.DeleteFood)

	router.Get("/foods/{foodID}/servings", h.ListServings)
	router.With(h.adminOnly()).Post("/foods/{foodID}/servings", h.CreateServing)
	router.Get("/foods/{foodID}/servings/{servingID}", h.GetServing)
	router.With(h.adminOnly()).Put("/foods/{foodID}/servings/{servingID}", h.UpdateServing)
	router.With(h.adminOnly()).Delete("/foods/{foodID}/servings/{servingID}", h.DeleteServing)
}

func (h Handler) adminOnly() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return h.auth.Authenticate(h.auth.RequireAnyRole(authdomain.RoleAdmin)(next))
	}
}

func (h Handler) CreateFood(w http.ResponseWriter, r *http.Request) {
	var request foodRequest
	if !decodeRequest(w, r, &request) {
		return
	}
	food, err := h.service.CreateFood(r.Context(), toFoodInput(request))
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toFoodResponse(food))
}

func (h Handler) ListFoods(w http.ResponseWriter, r *http.Request) {
	pageRequest, ok := parsePage(w, r)
	if !ok {
		return
	}
	result, err := h.service.ListFoods(r.Context(), port.ListOptions{
		Query: r.URL.Query().Get("q"), Brand: r.URL.Query().Get("brand"),
		Offset: pageRequest.Offset(), Limit: pageRequest.PageSize,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	items := make([]foodResponse, 0, len(result.Items))
	for _, food := range result.Items {
		items = append(items, toFoodResponse(food))
	}
	httpx.WriteJSON(w, http.StatusOK, pagination.Result[foodResponse]{
		Items: items, Page: pageRequest.Page, PageSize: pageRequest.PageSize, Total: result.Total,
	})
}

func (h Handler) GetFood(w http.ResponseWriter, r *http.Request) {
	food, err := h.service.GetFood(r.Context(), chi.URLParam(r, "foodID"))
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toFoodResponse(food))
}

func (h Handler) UpdateFood(w http.ResponseWriter, r *http.Request) {
	var request foodRequest
	if !decodeRequest(w, r, &request) {
		return
	}
	food, err := h.service.UpdateFood(r.Context(), chi.URLParam(r, "foodID"), toFoodInput(request))
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toFoodResponse(food))
}

func (h Handler) DeleteFood(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteFood(r.Context(), chi.URLParam(r, "foodID")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) CreateServing(w http.ResponseWriter, r *http.Request) {
	var request servingRequest
	if !decodeRequest(w, r, &request) {
		return
	}
	foodID := chi.URLParam(r, "foodID")
	serving, err := h.service.CreateServing(r.Context(), foodID, toServingInput(request))
	if err != nil {
		writeError(w, err)
		return
	}
	food, err := h.service.GetFood(r.Context(), foodID)
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toServingResponse(serving, food))
}

func (h Handler) ListServings(w http.ResponseWriter, r *http.Request) {
	pageRequest, ok := parsePage(w, r)
	if !ok {
		return
	}
	foodID := chi.URLParam(r, "foodID")
	food, err := h.service.GetFood(r.Context(), foodID)
	if err != nil {
		writeError(w, err)
		return
	}
	result, err := h.service.ListServings(r.Context(), foodID, port.ListOptions{
		Query: r.URL.Query().Get("q"), Offset: pageRequest.Offset(), Limit: pageRequest.PageSize,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	items := make([]servingResponse, 0, len(result.Items))
	for _, serving := range result.Items {
		items = append(items, toServingResponse(serving, food))
	}
	httpx.WriteJSON(w, http.StatusOK, pagination.Result[servingResponse]{
		Items: items, Page: pageRequest.Page, PageSize: pageRequest.PageSize, Total: result.Total,
	})
}

func (h Handler) GetServing(w http.ResponseWriter, r *http.Request) {
	foodID := chi.URLParam(r, "foodID")
	food, err := h.service.GetFood(r.Context(), foodID)
	if err != nil {
		writeError(w, err)
		return
	}
	serving, err := h.service.GetServing(r.Context(), foodID, chi.URLParam(r, "servingID"))
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toServingResponse(serving, food))
}

func (h Handler) UpdateServing(w http.ResponseWriter, r *http.Request) {
	var request servingRequest
	if !decodeRequest(w, r, &request) {
		return
	}
	foodID := chi.URLParam(r, "foodID")
	serving, err := h.service.UpdateServing(
		r.Context(), foodID, chi.URLParam(r, "servingID"), toServingInput(request),
	)
	if err != nil {
		writeError(w, err)
		return
	}
	food, err := h.service.GetFood(r.Context(), foodID)
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toServingResponse(serving, food))
}

func (h Handler) DeleteServing(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteServing(
		r.Context(), chi.URLParam(r, "foodID"), chi.URLParam(r, "servingID"),
	); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decodeRequest(w http.ResponseWriter, r *http.Request, destination any) bool {
	if err := httpx.DecodeJSON(w, r, destination); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_request", "The request body is invalid.")
		return false
	}
	return true
}

func parsePage(w http.ResponseWriter, r *http.Request) (pagination.Request, bool) {
	request, err := pagination.Parse(r.URL.Query())
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_pagination", err.Error())
		return pagination.Request{}, false
	}
	return request, true
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		httpx.WriteError(w, http.StatusBadRequest, "invalid_input", err.Error())
	case errors.Is(err, domain.ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "not_found", "The nutrition catalog item was not found.")
	case errors.Is(err, domain.ErrConflict):
		httpx.WriteError(w, http.StatusConflict, "catalog_conflict", "A nutrition item with these values already exists.")
	case errors.Is(err, domain.ErrInvalidReference):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "invalid_reference", "The referenced food does not exist.")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred.")
	}
}

func toFoodInput(request foodRequest) application.FoodInput {
	return application.FoodInput{
		Name: request.Name, Brand: request.Brand, Description: request.Description,
		ReferenceAmount: decimal.NewFromFloat(request.ReferenceAmount), ReferenceUnit: request.ReferenceUnit,
		Nutrients: domain.Nutrients{
			CaloriesKcal:   decimal.NewFromFloat(request.Nutrients.CaloriesKcal),
			ProteinG:       decimal.NewFromFloat(request.Nutrients.ProteinG),
			CarbohydratesG: decimal.NewFromFloat(request.Nutrients.CarbohydratesG),
			FatG:           decimal.NewFromFloat(request.Nutrients.FatG),
			FiberG:         decimal.NewFromFloat(request.Nutrients.FiberG),
		},
	}
}

func toServingInput(request servingRequest) application.ServingInput {
	return application.ServingInput{Name: request.Name, Amount: decimal.NewFromFloat(request.Amount)}
}

func toFoodResponse(food domain.Food) foodResponse {
	return foodResponse{
		ID: food.ID, Name: food.Name, Brand: food.Brand, Description: food.Description,
		ReferenceAmount: decimalFloat(food.ReferenceAmount, 3), ReferenceUnit: food.ReferenceUnit,
		Nutrients: toNutrientsResponse(food.Nutrients), CreatedAt: food.CreatedAt, UpdatedAt: food.UpdatedAt,
	}
}

func toServingResponse(serving domain.Serving, food domain.Food) servingResponse {
	return servingResponse{
		ID: serving.ID, FoodID: serving.FoodID, Name: serving.Name,
		Amount: decimalFloat(serving.Amount, 3), Unit: food.ReferenceUnit,
		Nutrition: toNutrientsResponse(food.NutrientsFor(serving.Amount)),
		CreatedAt: serving.CreatedAt, UpdatedAt: serving.UpdatedAt,
	}
}

func toNutrientsResponse(nutrients domain.Nutrients) nutrientsResponse {
	return nutrientsResponse{
		CaloriesKcal:   decimalFloat(nutrients.CaloriesKcal, 2),
		ProteinG:       decimalFloat(nutrients.ProteinG, 2),
		CarbohydratesG: decimalFloat(nutrients.CarbohydratesG, 2),
		FatG:           decimalFloat(nutrients.FatG, 2),
		FiberG:         decimalFloat(nutrients.FiberG, 2),
	}
}

func decimalFloat(value decimal.Decimal, places int32) float64 {
	result, _ := value.Round(places).Float64()
	return result
}
