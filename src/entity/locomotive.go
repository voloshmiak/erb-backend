package entity

import (
	"time"

	"github.com/google/uuid"
)

type LocomotiveStatus string

const (
	LocomotiveIdle      LocomotiveStatus = "idle"
	LocomotiveInTransit LocomotiveStatus = "in_transit"
)

type Locomotive struct {
	ID               uuid.UUID        `json:"id"`
	CurrentStationID uuid.UUID        `json:"current_station_id"`
	Status           LocomotiveStatus `json:"status"`
	AvailableAt      time.Time        `json:"available_at"`
	AvailableAtHour  int64            `json:"available_at_hour"`
	TrainID          *uuid.UUID       `json:"train_id,omitempty"`
}
