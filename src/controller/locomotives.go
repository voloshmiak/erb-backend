package controller

import (
	"encoding/json"
	"erb-backend/src/usecase"
	"net/http"

	"github.com/google/uuid"
)

type ListLocomotivesController struct {
	usecase *usecase.ListLocomotivesUseCase
}

func NewListLocomotivesController(uc *usecase.ListLocomotivesUseCase) *ListLocomotivesController {
	return &ListLocomotivesController{usecase: uc}
}

func (c *ListLocomotivesController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	locos, err := c.usecase.Execute(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(locos)
}

type GetLocomotiveController struct {
	usecase *usecase.GetLocomotiveUseCase
}

func NewGetLocomotiveController(uc *usecase.GetLocomotiveUseCase) *GetLocomotiveController {
	return &GetLocomotiveController{usecase: uc}
}

func (c *GetLocomotiveController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	loco, err := c.usecase.Execute(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loco)
}
