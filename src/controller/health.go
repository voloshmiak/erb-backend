package controller

import "net/http"

type HealthController struct {
}

func NewHealthController() *HealthController {
	return &HealthController{}
}

// ServeHTTP godoc
// @Summary     Health Check
// @Description Returns 200 OK when the service is healthy
// @Tags        System
// @Success     200  {string}  string  "OK"
// @Router      /health [get]
func (h *HealthController) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
