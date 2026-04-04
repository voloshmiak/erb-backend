package usecase

import (
	"context"
	"erb-backend/src/broadcaster"
	"erb-backend/src/entity"
	"log"
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

func (uc *DispatchPlannedUseCase) Execute(ctx context.Context) {
	a, err := uc.assignments.GetOldestPlanned(ctx)
	if err != nil {
		log.Printf("dispatch_planned: failed to get oldest planned: %v", err)
		return
	}
	if a == nil {
		return
	}

	if err = uc.assignments.UpdateStatus(ctx, a.ID, entity.AssignmentInTransit); err != nil {
		log.Printf("dispatch_planned: failed to update assignment status %s: %v", a.ID, err)
		return
	}
	if err = uc.routeSteps.ActivateRouteStep(ctx, a.FirstStepID); err != nil {
		log.Printf("dispatch_planned: failed to activate first step %s: %v", a.FirstStepID, err)
		return
	}
	if err = uc.wagons.UpdateStatus(ctx, a.WagonID, entity.EmptyMoving); err != nil {
		log.Printf("dispatch_planned: failed to update wagon status %s: %v", a.WagonID, err)
		return
	}

	uc.broadcaster.Publish(broadcaster.NewEvent(broadcaster.WagonDispatched, a))
}
