package controller

import (
	"encoding/json"
	"erb-backend/src/entity"
	"erb-backend/src/usecase"
	"net/http"
)

type ListWagonsController struct {
	usecase *usecase.ListWagonsUseCase
}

func NewListWagonsController(usecase *usecase.ListWagonsUseCase) *ListWagonsController {
	return &ListWagonsController{usecase: usecase}
}

type listWagonsResponse struct {
	Wagons []*entity.Wagon `json:"wagons"`
}

// ServeHTTP godoc
// @Summary     List Wagons
// @Description Retrieves all wagons and their current status
// @Tags        Wagon
// @Accept      json
// @Produce     json
// @Success     200 {object} listWagonsResponse
// @Failure     500 {object} string "Failed to retrieve wagons"
// @Router      /wagons [get]
func (c *ListWagonsController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stations, err := c.usecase.Execute(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := listWagonsResponse{
		Wagons: stations,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
