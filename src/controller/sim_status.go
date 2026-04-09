package controller

import (
	"encoding/json"
	"erb-backend/src/simclock"
	"net/http"
	"time"
)

type SimStatusController struct {
	simClock *simclock.SimClock
}

func NewSimStatusController(clock *simclock.SimClock) *SimStatusController {
	return &SimStatusController{simClock: clock}
}

type simStatusResponse struct {
	CurrentHour int64     `json:"currentHour"`
	DisplayTime time.Time `json:"displayTime"`
	Speed       int       `json:"speed"`
}

// ServeHTTP godoc
// @Summary     Simulation Status
// @Description Returns current simulation clock state
// @Tags        Simulation
// @Produce     json
// @Success     200  {object}  simStatusResponse
// @Router      /simulation [get]
func (c *SimStatusController) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	hour := c.simClock.Now()
	resp := simStatusResponse{
		CurrentHour: hour,
		DisplayTime: c.simClock.ToDisplayTime(hour),
		Speed:       c.simClock.Speed(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
