package entity

import (
	"time"

	"github.com/google/uuid"
)

type AssignmentStatus string

const (
	AssignmentPlanned   AssignmentStatus = "planned"
	AssignmentInTransit AssignmentStatus = "in_transit"
	AssignmentDelivered AssignmentStatus = "delivered"
)

type Assignment struct {
	ID               uuid.UUID        `json:"id"`
	OrderID          uuid.UUID        `json:"orderId"`
	WagonID          uuid.UUID        `json:"wagonId"`
	EmptyRunKM       int              `json:"emptyRunKm"`
	LoadedRunKM      int              `json:"loadedRunKm"`
	CostEmptyRun     int              `json:"costEmptyRun"`
	EstimatedArrival time.Time        `json:"estimatedArrival"`
	Status           AssignmentStatus `json:"status"`
}

// AssignmentStats holds aggregate delivery metrics from the assignments table.
type AssignmentStats struct {
	TotalDelivered    int
	TotalEmptyRunKM   int
	TotalCostEmptyRun int
	AvgEmptyRunKM     float64
	TotalLoadedRunKM  int
}

type AssignmentResult struct {
	Assignment *Assignment `json:"assignment"`
	Route      []uuid.UUID `json:"route"`
}

type PlannedAssignment struct {
	ID          uuid.UUID `json:"id"`
	OrderID     uuid.UUID `json:"orderId"`
	WagonID     uuid.UUID `json:"wagonId"`
	FirstStepID uuid.UUID `json:"firstStepId"`
}

// TrainGroupResult is the Go-side representation of a train_group returned
// by the matching service: a batch of wagons that share a source/destination
// and (optionally) an assigned locomotive.
type TrainGroupResult struct {
	TrainID              uuid.UUID
	SourceStationID      uuid.UUID
	DestinationStationID uuid.UUID
	LocomotiveID         *uuid.UUID
	WagonIDs             []uuid.UUID
	RepositionKM         float64
	DistanceKM           float64
}
