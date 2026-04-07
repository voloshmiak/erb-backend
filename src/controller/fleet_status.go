package controller

import (
	"encoding/json"
	"erb-backend/src/entity"
	"erb-backend/src/usecase"
	"net/http"
)

type FleetStatusController struct {
	usecase *usecase.FleetStatusUseCase
}

func NewFleetStatusController(usecase *usecase.FleetStatusUseCase) *FleetStatusController {
	return &FleetStatusController{
		usecase: usecase,
	}
}

type fleetStatusResponse struct {
	TotalWagons        int                                           `json:"totalWagons"`
	ByType             map[entity.WagonType]*usecase.WagonTypeStatus `json:"byType"`
	AvgEmptyRunKmToday float64                                       `json:"avgEmptyRunKmToday"`
}

// ServeHTTP godoc
// @Summary     Fleet Status
// @Description Returns wagon counts by type and status, plus average empty-run km for today
// @Tags        Wagon
// @Produce     json
// @Success     200  {object}  fleetStatusResponse
// @Failure     500  {object}  string  "Failed to retrieve fleet status"
// @Router      /fleet/status [get]
func (h *FleetStatusController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status, err := h.usecase.Execute(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := fleetStatusResponse{
		TotalWagons: status.TotalWagons,
		ByType: map[entity.WagonType]*usecase.WagonTypeStatus{
			entity.Gondola:      status.ByType[entity.Gondola],
			entity.GrainHopper:  status.ByType[entity.GrainHopper],
			entity.CementHopper: status.ByType[entity.CementHopper],
		},
		AvgEmptyRunKmToday: status.AvgEmptyRunKmToday,
	}

	w.Header().Set("Content-Type", "application/json")

	if err = json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
