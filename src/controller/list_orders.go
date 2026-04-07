package controller

import (
	"encoding/json"
	"erb-backend/src/entity"
	"erb-backend/src/usecase"
	"net/http"
)

type ListOrdersController struct {
	usecase *usecase.ListOrdersUseCase
}

func NewListOrdersController(usecase *usecase.ListOrdersUseCase) *ListOrdersController {
	return &ListOrdersController{usecase: usecase}
}

type listOrdersResponse struct {
	Orders []*entity.Order `json:"orders"`
}

// ServeHTTP godoc
// @Summary     List Orders
// @Description Retrieves all orders and their current status
// @Tags        Order
// @Accept      json
// @Produce     json
// @Success     200 {object} listOrdersResponse
// @Failure     500 {object} string "Failed to retrieve orders"
// @Router      /orders [get]
func (c *ListOrdersController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stations, err := c.usecase.Execute(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := listOrdersResponse{
		Orders: stations,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
