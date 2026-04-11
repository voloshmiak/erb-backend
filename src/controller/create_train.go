package controller

import (
	"encoding/json"
	"erb-backend/src/usecase"
	"net/http"

	"github.com/google/uuid"
)

type CreateTrainController struct {
	usecase *usecase.CreateTrainUseCase
}

func NewCreateTrainController(usecase *usecase.CreateTrainUseCase) *CreateTrainController {
	return &CreateTrainController{usecase: usecase}
}

type CreateTrainInput struct {
	WagonIDs []uuid.UUID `json:"wagonIds"`
	Route    []uuid.UUID `json:"route"`
}

// ServeHTTP godoc
// @Summary     Create Train
// @Description Creates a new train from a list of wagons and a route
// @Tags        Train
// @Accept      json
// @Produce     json
// @Param       body  body      CreateTrainInput  true  "Train payload"
// @Success     200   {object}  entity.Train
// @Failure     400   {object}  string  "Invalid request"
// @Failure     500   {object}  string  "Failed to create train"
// @Router      /trains [post]
func (ctrl *CreateTrainController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body CreateTrainInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	train, err := ctrl.usecase.Execute(r.Context(), body.WagonIDs, body.Route)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(train)
}
