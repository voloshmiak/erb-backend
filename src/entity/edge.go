package entity

import (
	"time"

	"github.com/google/uuid"
)

type Edge struct {
	ID            uuid.UUID `json:"id"`
	FromStationID uuid.UUID `json:"fromStationId"`
	ToStationID   uuid.UUID `json:"toStationId"`
	DistanceKM    float64   `json:"distanceKm"`
	IsActive      bool      `json:"isActive"`
	CreatedAt     time.Time `json:"createdAt"`
}
