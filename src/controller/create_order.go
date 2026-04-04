package controller

import (
	"encoding/json"
	"erb-backend/src/usecase"
	"net/http"

	"github.com/pkg/errors"
)

type CreateOrderController struct {
	usecase *usecase.CreateOrderUseCase
}

func NewCreateOrderController(usecase *usecase.CreateOrderUseCase) *CreateOrderController {
	return &CreateOrderController{usecase: usecase}
}

// ServeHTTP godoc
// @Summary     Create Order
// @Description Places a new wagon order for a client
// @Tags        Order
// @Accept      json
// @Produce     json
// @Param       body  body      usecase.CreateOrderInput  true  "Order payload"
// @Success     201   {object}  entity.Order
// @Failure     400   {object}  string  "Invalid request"
// @Failure     500   {object}  string  "Failed to create order"
// @Router      /orders [post]
func (h *CreateOrderController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var input usecase.CreateOrderInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	order, err := h.usecase.Execute(r.Context(), input)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidInput) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err = json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
