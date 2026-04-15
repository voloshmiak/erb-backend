package usecase

import (
	"context"
	"erb-backend/src/entity"

	"github.com/pkg/errors"
)

type FleetStatusUseCase struct {
	repository           WagonRepository
	assignmentRepository AssignmentRepository
}

func NewFleetStatusUseCase(repository WagonRepository, assignmentRepository AssignmentRepository) *FleetStatusUseCase {
	return &FleetStatusUseCase{repository: repository, assignmentRepository: assignmentRepository}
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

func (u *FleetStatusUseCase) Execute(ctx context.Context) (*FleetStatus, error) {
	counts, err := u.repository.ListStatusCounts(ctx)
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

	stats, err := u.assignmentRepository.GetStats(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get assignment stats")
	}
	status.AvgEmptyRunKmToday = stats.AvgEmptyRunKM

	return status, nil
}
