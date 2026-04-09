package entity

import (
	"time"

	"github.com/google/uuid"
)

type WagonType string

const (
	Gondola      WagonType = "gondola"
	GrainHopper  WagonType = "grain_hopper"
	CementHopper WagonType = "cement_hopper"
)

type WagonStatus string

const (
	Loaded      WagonStatus = "loaded"
	EmptyMoving WagonStatus = "empty_moving"
	Idle        WagonStatus = "idle"
	Maintenance WagonStatus = "maintenance"
)

type Wagon struct {
	ID               uuid.UUID   `json:"id"`
	Number           string      `json:"number"`
	Type             WagonType   `json:"type"`
	Status           WagonStatus `json:"status"`
	CurrentStationID uuid.UUID   `json:"currentStationId"`
	LastUnloadTime   time.Time   `json:"lastUnloadTime"`
}

type WagonStatusCount struct {
	Type   WagonType
	Status WagonStatus
	Count  int
}

type LoadedWagon struct {
	ID          uuid.UUID `json:"id"`
	WagonNumber string    `json:"wagonNumber"`
}
