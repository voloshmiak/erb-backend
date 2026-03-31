package controller

import "net/http"

type HealthController struct {
}

func NewHealthController() *HealthController {
	return &HealthController{}
}

// HealthСheck godoc
// @Summary     Health Check
// @Description Checks the health of the API
// @Tags        System
// @Accept      json
// @Produce     json
// @Router      /health [get]
func (h *HealthController) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
