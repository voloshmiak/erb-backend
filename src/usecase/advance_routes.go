package usecase

import (
	"context"
	"erb-backend/src/broadcaster"
	"erb-backend/src/entity"
	"log"
	"time"
)

type AdvanceRoutesUseCase struct {
	routeSteps  RouteStepRepository
	wagons      WagonRepository
	assignments AssignmentRepository
	orders      OrderRepository
	broadcaster Broadcaster
}

func NewAdvanceRoutesUseCase(
	routeSteps RouteStepRepository,
	wagons WagonRepository,
	assignments AssignmentRepository,
	orders OrderRepository,
	b Broadcaster,
) *AdvanceRoutesUseCase {
	return &AdvanceRoutesUseCase{
		routeSteps:  routeSteps,
		wagons:      wagons,
		assignments: assignments,
		orders:      orders,
		broadcaster: b,
	}
}

func (uc *AdvanceRoutesUseCase) Execute(ctx context.Context) {
	activeSteps, err := uc.routeSteps.GetActiveRouteSteps(ctx)
	if err != nil {
		log.Printf("advance_routes: failed to get active steps: %v", err)
		return
	}

	now := time.Now()

	for _, step := range activeSteps {
		uc.processStep(ctx, step, now)
	}
}

func (uc *AdvanceRoutesUseCase) processStep(ctx context.Context,
	step *entity.ActiveRouteStep, now time.Time) {
	if err := uc.routeSteps.CompleteRouteStep(ctx, step.ID); err != nil {
		log.Printf("advance_routes: failed to complete step %s: %v", step.ID, err)
		return
	}

	nextStep, err := uc.routeSteps.GetNextRouteStep(ctx, step.AssignmentID, step.StepIndex+1)
	if err != nil {
		log.Printf("advance_routes: failed to get next step for assignment %s: %v",
			step.AssignmentID, err)
		return
	}

	if nextStep != nil {
		uc.advanceToNextStep(ctx, step, nextStep, now)
	} else {
		uc.completeAssignment(ctx, step)
	}
}

func (uc *AdvanceRoutesUseCase) advanceToNextStep(ctx context.Context, step *entity.ActiveRouteStep,
	nextStep *entity.RouteStep, now time.Time) {
	if err := uc.routeSteps.ActivateRouteStep(ctx, nextStep.ID); err != nil {
		log.Printf("advance_routes: failed to activate step %s: %v", nextStep.ID, err)
		return
	}
	if err := uc.wagons.UpdateStation(ctx, step.WagonID, nextStep.StationID); err != nil {
		log.Printf("advance_routes: failed to update wagon station %s: %v", step.WagonID, err)
		return
	}
	uc.broadcaster.Publish(broadcaster.NewEvent(broadcaster.WagonMoved, entity.WagonMovedPayload{
		WagonID:     step.WagonID,
		WagonNumber: step.WagonNumber,
		StationID:   nextStep.StationID,
		StationName: nextStep.StationName,
		Lat:         nextStep.Lat,
		Lng:         nextStep.Lng,
		StepIndex:   nextStep.StepIndex,
		TotalSteps:  step.TotalSteps,
		ArrivedAt:   now,
	}))
}

func (uc *AdvanceRoutesUseCase) completeAssignment(ctx context.Context,
	step *entity.ActiveRouteStep) {
	if err := uc.wagons.UpdateStatus(ctx, step.WagonID, entity.Loaded); err != nil {
		log.Printf("advance_routes: failed to update wagon status %s: %v", step.WagonID, err)
		return
	}
	if err := uc.assignments.UpdateStatus(ctx, step.AssignmentID, entity.AssignmentDelivered); err != nil {
		log.Printf("advance_routes: failed to update assignment status %s: %v", step.AssignmentID, err)
		return
	}
	fulfilled, err := uc.orders.UpdateIfFulfilled(ctx, step.OrderID)
	if err != nil {
		log.Printf("advance_routes: failed to update order %s: %v", step.OrderID, err)
		return
	}
	if fulfilled {
		uc.broadcaster.Publish(broadcaster.NewEvent(broadcaster.OrderFulfilled, step.OrderID))
	}
	uc.broadcaster.Publish(broadcaster.NewEvent(broadcaster.WagonArrived, entity.WagonArrivedPayload{
		WagonID:      step.WagonID,
		OrderID:      step.OrderID,
		AssignmentID: step.AssignmentID,
	}))
}
