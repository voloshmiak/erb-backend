package usecase

import (
	"erb-backend/src/entity"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type WagonRepository interface {
	ListStatusCounts() ([]entity.WagonStatusCount, error)
	Exists(id uuid.UUID) (bool, error)
}

type FleetStatusUseCase struct {
	repository WagonRepository
}

func NewFleetStatusUseCase(repository WagonRepository) *FleetStatusUseCase {
	return &FleetStatusUseCase{repository: repository}
}

type WagonTypeStatus struct {
	Total       int `json:"total"`
	Loaded      int `json:"loaded"`
	EmptyMoving int `json:"emptyMoving"`
	Idle        int `json:"idle"`
	Maintenance int `json:"maintenance"`
}

type FleetStatus struct {
	TotalWagons        int                                   `json:"totalWagons"`
	ByType             map[entity.WagonType]*WagonTypeStatus `json:"byType"`
	AvgEmptyRunKmToday float64                               `json:"avgEmptyRunKmToday"`
}

func (u *FleetStatusUseCase) Execute() (*FleetStatus, error) {
	counts, err := u.repository.ListStatusCounts()
	if err != nil {
		return nil, errors.Wrap(err, "failed to list wagon status counts")
	}

	status := &FleetStatus{
		ByType: map[entity.WagonType]*WagonTypeStatus{
			entity.Gondola:      {},
			entity.GrainHopper:  {},
			entity.CementHopper: {},
		},
	}

	for _, c := range counts {
		ts := status.ByType[c.Type]
		ts.Total += c.Count
		status.TotalWagons += c.Count
		switch c.Status {
		case entity.Loaded:
			ts.Loaded = c.Count
		case entity.EmptyMoving:
			ts.EmptyMoving = c.Count
		case entity.Idle:
			ts.Idle = c.Count
		case entity.Maintenance:
			ts.Maintenance = c.Count
		}
	}

	return status, nil
}
