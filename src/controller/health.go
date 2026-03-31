package controller

import "net/http"

type HealthController struct {
}

func NewHealthController() *HealthController {
	return &HealthController{}
}

func (h *HealthController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
