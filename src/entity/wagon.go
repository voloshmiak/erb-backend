package entity

import "github.com/google/uuid"

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
	Type   WagonType
	Status WagonStatus
}

type WagonStatusCount struct {
	Type   WagonType
	Status WagonStatus
	Count  int
}

type LoadedWagon struct {
	ID          uuid.UUID
	WagonNumber string
}
