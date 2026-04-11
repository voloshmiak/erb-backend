package controller

import (
	"erb-backend/src/simclock"
	"erb-backend/src/usecase"
	"net/http"

	"github.com/google/uuid"
)

type DispatchTrainController struct {
	usecase *usecase.DispatchTrainUseCase
	clock   *simclock.SimClock
}

func NewDispatchTrainController(usecase *usecase.DispatchTrainUseCase, clock *simclock.SimClock) *DispatchTrainController {
	return &DispatchTrainController{usecase: usecase, clock: clock}
}

// ServeHTTP godoc
// @Summary     Dispatch Train
// @Description Dispatches a forming train to start its journey
// @Tags        Train
// @Accept      json
// @Produce     json
// @Param       id    path      string  true  "Train ID"
// @Success     204   "No Content"
// @Failure     400   {object}  string  "Invalid train id"
// @Failure     500   {object}  string  "Failed to dispatch train"
// @Router      /trains/{id}/dispatch [post]
func (ctrl *DispatchTrainController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid train id", http.StatusBadRequest)
		return
	}

	currentHour := ctrl.clock.Tick()
	if err := ctrl.usecase.Execute(r.Context(), id, currentHour); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
