package usecase

import (
	"context"
	"erb-backend/src/broadcaster"
	"erb-backend/src/entity"

	"github.com/pkg/errors"
)

type DispatchPlannedUseCase struct {
	assignments AssignmentRepository
	routeSteps  RouteStepRepository
	wagons      WagonRepository
	broadcaster Broadcaster
}

func NewDispatchPlannedUseCase(
	assignments AssignmentRepository,
	routeSteps RouteStepRepository,
	wagons WagonRepository,
	b Broadcaster,
) *DispatchPlannedUseCase {
	return &DispatchPlannedUseCase{
		assignments: assignments,
		routeSteps:  routeSteps,
		wagons:      wagons,
		broadcaster: b,
	}
}

func (uc *DispatchPlannedUseCase) Execute(ctx context.Context) error {
	a, err := uc.assignments.FindOldestPlanned(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get oldest planned")
	}
	if a == nil {
		return nil
	}

	if err = uc.assignments.UpdateStatus(ctx, a.ID, entity.AssignmentInTransit); err != nil {
		return errors.Wrap(err, "failed to update assignment status")
	}
	if err = uc.routeSteps.ActivateRouteStep(ctx, a.FirstStepID); err != nil {
		return errors.Wrap(err, "failed to activate first step")
	}
	if err = uc.wagons.UpdateStatus(ctx, a.WagonID, entity.EmptyMoving); err != nil {
		return errors.Wrap(err, "failed to update wagon status")
	}

	uc.broadcaster.Publish(broadcaster.NewEvent(broadcaster.WagonDispatched, a))

	return nil
}
