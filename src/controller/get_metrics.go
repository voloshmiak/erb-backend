package controller

import (
	"encoding/json"
	"erb-backend/src/usecase"
	"net/http"
)

type GetMetricsController struct {
	usecase *usecase.GetMetricsUseCase
}

func NewGetMetricsController(uc *usecase.GetMetricsUseCase) *GetMetricsController {
	return &GetMetricsController{usecase: uc}
}

// ServeHTTP godoc
// @Summary     Financial Metrics
// @Description Returns financial and operational metrics: empty run costs, revenue, cost saved vs naive
// @Tags        Metrics
// @Produce     json
// @Success     200  {object}  usecase.MetricsOutput
// @Failure     500  {object}  string  "Failed to retrieve metrics"
// @Router      /metrics [get]
func (c *GetMetricsController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	output, err := c.usecase.Execute(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err = json.NewEncoder(w).Encode(output); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
