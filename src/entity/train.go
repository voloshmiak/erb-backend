package entity

import (
	"time"

	"github.com/google/uuid"
)

type TrainStatus string

const (
	TrainForming   TrainStatus = "forming"
	TrainInTransit TrainStatus = "in_transit"
	TrainArrived   TrainStatus = "arrived"
)

type Train struct {
	ID              uuid.UUID   `json:"id"`
	WagonIDs        []uuid.UUID `json:"wagonIds"`
	Route           []uuid.UUID `json:"route"`
	StepIndex       int         `json:"stepIndex"`
	SourceStationID uuid.UUID   `json:"sourceStationId"`
	NextStationID   uuid.UUID   `json:"nextStationId"`
	Status          TrainStatus `json:"status"`
	CreatedAt       time.Time   `json:"createdAt"`
	DepartedAt      *time.Time  `json:"departedAt,omitempty"`
	ArrivedAt       *time.Time  `json:"arrivedAt,omitempty"`

	// Ticker internal fields for movement tracking
	ActivatedAtHour int64   `json:"-"`
	DurationHours   float64 `json:"-"`
}
