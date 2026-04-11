package controller

import (
	"encoding/json"
	"erb-backend/src/usecase"
	"net/http"

	"github.com/google/uuid"
)

type GetTrainController struct {
	usecase *usecase.GetTrainUseCase
}

func NewGetTrainController(usecase *usecase.GetTrainUseCase) *GetTrainController {
	return &GetTrainController{usecase: usecase}
}

// ServeHTTP godoc
// @Summary     Get Train Details
// @Description Returns details of a specific train
// @Tags        Train
// @Accept      json
// @Produce     json
// @Param       id    path      string  true  "Train ID"
// @Success     200   {object}  entity.Train
// @Failure     400   {object}  string  "Invalid train id"
// @Failure     404   {object}  string  "Train not found"
// @Router      /trains/{id} [get]
func (ctrl *GetTrainController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid train id", http.StatusBadRequest)
		return
	}

	train, err := ctrl.usecase.Execute(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(train)
}
