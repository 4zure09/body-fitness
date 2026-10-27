package exercisehttp

import (
	"errors"
	"net/http"
	"time"

	authdomain "github.com/4zure09/body-fitness/internal/modules/auth/domain"
	authhttp "github.com/4zure09/body-fitness/internal/modules/auth/transport/http"
	"github.com/4zure09/body-fitness/internal/modules/exercises/application"
	"github.com/4zure09/body-fitness/internal/modules/exercises/domain"
	"github.com/4zure09/body-fitness/internal/modules/exercises/port"
	"github.com/4zure09/body-fitness/internal/shared/httpx"
	"github.com/4zure09/body-fitness/internal/shared/pagination"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *application.Service
	auth    authhttp.Middleware
}

type muscleGroupRequest struct {
	Name        string            `json:"name"`
	BodyRegion  domain.BodyRegion `json:"body_region"`
	Description string            `json:"description"`
}

type equipmentRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type exerciseRequest struct {
	Name               string            `json:"name"`
	Description        string            `json:"description"`
	Instructions       string            `json:"instructions"`
	Difficulty         domain.Difficulty `json:"difficulty"`
	EquipmentID        *string           `json:"equipment_id"`
	PrimaryMuscleIDs   []string          `json:"primary_muscle_ids"`
	SecondaryMuscleIDs []string          `json:"secondary_muscle_ids"`
}

type muscleGroupResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	BodyRegion  domain.BodyRegion `json:"body_region"`
	Description string            `json:"description"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type equipmentResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type exerciseMuscleResponse struct {
	Muscle      muscleGroupResponse `json:"muscle_group"`
	Involvement domain.Involvement  `json:"involvement"`
}

type exerciseResponse struct {
	ID           string                   `json:"id"`
	Name         string                   `json:"name"`
	Description  string                   `json:"description"`
	Instructions string                   `json:"instructions"`
	Difficulty   domain.Difficulty        `json:"difficulty"`
	Equipment    *equipmentResponse       `json:"equipment,omitempty"`
	Muscles      []exerciseMuscleResponse `json:"muscles"`
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at"`
}

func NewHandler(service *application.Service, auth authhttp.Middleware) Handler {
	return Handler{service: service, auth: auth}
}

func (h Handler) RegisterRoutes(router chi.Router) {
	h.registerMuscleGroupRoutes(router)
	h.registerEquipmentRoutes(router)
	h.registerExerciseRoutes(router)
}

func (h Handler) registerMuscleGroupRoutes(router chi.Router) {
	router.Get("/muscle-groups", h.ListMuscleGroups)
	router.With(h.adminOnly()).Post("/muscle-groups", h.CreateMuscleGroup)
	router.Get("/muscle-groups/{id}", h.GetMuscleGroup)
	router.With(h.adminOnly()).Put("/muscle-groups/{id}", h.UpdateMuscleGroup)
	router.With(h.adminOnly()).Delete("/muscle-groups/{id}", h.DeleteMuscleGroup)
}

func (h Handler) registerEquipmentRoutes(router chi.Router) {
	router.Get("/equipment", h.ListEquipment)
	router.With(h.adminOnly()).Post("/equipment", h.CreateEquipment)
	router.Get("/equipment/{id}", h.GetEquipment)
	router.With(h.adminOnly()).Put("/equipment/{id}", h.UpdateEquipment)
	router.With(h.adminOnly()).Delete("/equipment/{id}", h.DeleteEquipment)
}

func (h Handler) registerExerciseRoutes(router chi.Router) {
	router.Get("/exercises", h.ListExercises)
	router.With(h.adminOnly()).Post("/exercises", h.CreateExercise)
	router.Get("/exercises/{id}", h.GetExercise)
	router.With(h.adminOnly()).Put("/exercises/{id}", h.UpdateExercise)
	router.With(h.adminOnly()).Delete("/exercises/{id}", h.DeleteExercise)
}

func (h Handler) adminOnly() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return h.auth.Authenticate(h.auth.RequireAnyRole(authdomain.RoleAdmin)(next))
	}
}

func (h Handler) CreateMuscleGroup(w http.ResponseWriter, r *http.Request) {
	var request muscleGroupRequest
	if !decodeRequest(w, r, &request) {
		return
	}
	item, err := h.service.CreateMuscleGroup(r.Context(), application.MuscleGroupInput(request))
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toMuscleGroupResponse(item))
}

func (h Handler) ListMuscleGroups(w http.ResponseWriter, r *http.Request) {
	pageRequest, ok := parsePage(w, r)
	if !ok {
		return
	}
	result, err := h.service.ListMuscleGroups(r.Context(), port.ListOptions{
		Query: r.URL.Query().Get("q"), Offset: pageRequest.Offset(), Limit: pageRequest.PageSize,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	items := make([]muscleGroupResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toMuscleGroupResponse(item))
	}
	httpx.WriteJSON(w, http.StatusOK, pagination.Result[muscleGroupResponse]{
		Items: items, Page: pageRequest.Page, PageSize: pageRequest.PageSize, Total: result.Total,
	})
}

func (h Handler) GetMuscleGroup(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetMuscleGroup(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toMuscleGroupResponse(item))
}

func (h Handler) UpdateMuscleGroup(w http.ResponseWriter, r *http.Request) {
	var request muscleGroupRequest
	if !decodeRequest(w, r, &request) {
		return
	}
	item, err := h.service.UpdateMuscleGroup(
		r.Context(), chi.URLParam(r, "id"), application.MuscleGroupInput(request),
	)
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toMuscleGroupResponse(item))
}

func (h Handler) DeleteMuscleGroup(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteMuscleGroup(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) CreateEquipment(w http.ResponseWriter, r *http.Request) {
	var request equipmentRequest
	if !decodeRequest(w, r, &request) {
		return
	}
	item, err := h.service.CreateEquipment(r.Context(), application.EquipmentInput(request))
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toEquipmentResponse(item))
}

func (h Handler) ListEquipment(w http.ResponseWriter, r *http.Request) {
	pageRequest, ok := parsePage(w, r)
	if !ok {
		return
	}
	result, err := h.service.ListEquipment(r.Context(), port.ListOptions{
		Query: r.URL.Query().Get("q"), Offset: pageRequest.Offset(), Limit: pageRequest.PageSize,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	items := make([]equipmentResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toEquipmentResponse(item))
	}
	httpx.WriteJSON(w, http.StatusOK, pagination.Result[equipmentResponse]{
		Items: items, Page: pageRequest.Page, PageSize: pageRequest.PageSize, Total: result.Total,
	})
}

func (h Handler) GetEquipment(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetEquipment(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toEquipmentResponse(item))
}

func (h Handler) UpdateEquipment(w http.ResponseWriter, r *http.Request) {
	var request equipmentRequest
	if !decodeRequest(w, r, &request) {
		return
	}
	item, err := h.service.UpdateEquipment(
		r.Context(), chi.URLParam(r, "id"), application.EquipmentInput(request),
	)
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toEquipmentResponse(item))
}

func (h Handler) DeleteEquipment(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteEquipment(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) CreateExercise(w http.ResponseWriter, r *http.Request) {
	var request exerciseRequest
	if !decodeRequest(w, r, &request) {
		return
	}
	item, err := h.service.CreateExercise(r.Context(), toExerciseInput(request))
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toExerciseResponse(item))
}

func (h Handler) ListExercises(w http.ResponseWriter, r *http.Request) {
	pageRequest, ok := parsePage(w, r)
	if !ok {
		return
	}
	query := r.URL.Query()
	result, err := h.service.ListExercises(r.Context(), port.ListOptions{
		Query: query.Get("q"), Difficulty: domain.Difficulty(query.Get("difficulty")),
		EquipmentID: query.Get("equipment_id"), MuscleGroupID: query.Get("muscle_group_id"),
		Offset: pageRequest.Offset(), Limit: pageRequest.PageSize,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	items := make([]exerciseResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toExerciseResponse(item))
	}
	httpx.WriteJSON(w, http.StatusOK, pagination.Result[exerciseResponse]{
		Items: items, Page: pageRequest.Page, PageSize: pageRequest.PageSize, Total: result.Total,
	})
}

func (h Handler) GetExercise(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetExercise(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toExerciseResponse(item))
}

func (h Handler) UpdateExercise(w http.ResponseWriter, r *http.Request) {
	var request exerciseRequest
	if !decodeRequest(w, r, &request) {
		return
	}
	item, err := h.service.UpdateExercise(r.Context(), chi.URLParam(r, "id"), toExerciseInput(request))
	if err != nil {
		writeError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toExerciseResponse(item))
}

func (h Handler) DeleteExercise(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteExercise(r.Context(), chi.URLParam(r, "id")); err != nil {
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
		httpx.WriteError(w, http.StatusNotFound, "not_found", "The catalog item was not found.")
	case errors.Is(err, domain.ErrConflict):
		httpx.WriteError(w, http.StatusConflict, "catalog_conflict", "A catalog item with these values already exists.")
	case errors.Is(err, domain.ErrInvalidReference):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "invalid_reference", "A referenced catalog item does not exist.")
	case errors.Is(err, domain.ErrInUse):
		httpx.WriteError(w, http.StatusConflict, "catalog_item_in_use", "This catalog item is used by an active exercise.")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred.")
	}
}

func toExerciseInput(request exerciseRequest) application.ExerciseInput {
	return application.ExerciseInput{
		Name: request.Name, Description: request.Description, Instructions: request.Instructions,
		Difficulty: request.Difficulty, EquipmentID: request.EquipmentID,
		PrimaryMuscleIDs: request.PrimaryMuscleIDs, SecondaryMuscleIDs: request.SecondaryMuscleIDs,
	}
}

func toMuscleGroupResponse(item domain.MuscleGroup) muscleGroupResponse {
	return muscleGroupResponse{
		ID: item.ID, Name: item.Name, BodyRegion: item.BodyRegion, Description: item.Description,
		CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}

func toEquipmentResponse(item domain.Equipment) equipmentResponse {
	return equipmentResponse{
		ID: item.ID, Name: item.Name, Description: item.Description,
		CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}

func toExerciseResponse(item domain.Exercise) exerciseResponse {
	var equipment *equipmentResponse
	if item.Equipment != nil {
		mapped := toEquipmentResponse(*item.Equipment)
		equipment = &mapped
	}
	muscles := make([]exerciseMuscleResponse, 0, len(item.Muscles))
	for _, muscle := range item.Muscles {
		muscles = append(muscles, exerciseMuscleResponse{
			Muscle: toMuscleGroupResponse(muscle.Muscle), Involvement: muscle.Involvement,
		})
	}
	return exerciseResponse{
		ID: item.ID, Name: item.Name, Description: item.Description, Instructions: item.Instructions,
		Difficulty: item.Difficulty, Equipment: equipment, Muscles: muscles,
		CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}
