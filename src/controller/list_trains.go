package controller

import (
	"encoding/json"
	"erb-backend/src/usecase"
	"net/http"
)

type ListTrainsController struct {
	usecase *usecase.ListTrainsUseCase
}

func NewListTrainsController(usecase *usecase.ListTrainsUseCase) *ListTrainsController {
	return &ListTrainsController{usecase: usecase}
}

// ServeHTTP godoc
// @Summary     List Active Trains
// @Description Returns a list of all active trains
// @Tags        Train
// @Accept      json
// @Produce     json
// @Success     200   {array}   entity.Train
// @Failure     500   {object}  string  "Failed to list trains"
// @Router      /trains [get]
func (ctrl *ListTrainsController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	trains, err := ctrl.usecase.Execute(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trains)
}
