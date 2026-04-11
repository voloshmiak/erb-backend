package usecase

import (
	"context"
	"erb-backend/src/broadcaster"
	"erb-backend/src/entity"
	"erb-backend/src/simclock"
	"log"
	"math"

	"github.com/pkg/errors"
)

const unloadDurationHours = 24

type AdvanceRoutesUseCase struct {
	routeStepRepository  RouteStepRepository
	wagonRepository      WagonRepository
	assignmentRepository AssignmentRepository
	orderRepository      OrderRepository
	broadcaster          Broadcaster
	simClock             *simclock.SimClock
}

func NewAdvanceRoutesUseCase(
	routeStepRepository RouteStepRepository,
	wagonRepository WagonRepository,
	assignmentRepository AssignmentRepository,
	orderRepository OrderRepository,
	b Broadcaster,
	clock *simclock.SimClock,
) *AdvanceRoutesUseCase {
	return &AdvanceRoutesUseCase{
		routeStepRepository:  routeStepRepository,
		wagonRepository:      wagonRepository,
		assignmentRepository: assignmentRepository,
		orderRepository:      orderRepository,
		broadcaster:          b,
		simClock:             clock,
	}
}

func (uc *AdvanceRoutesUseCase) Execute(ctx context.Context, simHour int64) error {
	activeSteps, err := uc.routeStepRepository.FindActiveRouteSteps(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get active steps")
	}

	for _, step := range activeSteps {
		readyAtHour := step.ActivatedAtHour + int64(math.Ceil(step.DurationHours))
		if readyAtHour > simHour {
			continue
		}

		if err = uc.processStep(ctx, step, simHour); err != nil {
			log.Printf("advance_routes: failed to process step %s: %v", step.ID, err)
		}
	}

	return nil
}

func (uc *AdvanceRoutesUseCase) processStep(ctx context.Context,
	step *entity.ActiveRouteStep, simHour int64) error {
	nextStep, err := uc.routeStepRepository.FindNextRouteStep(ctx, step.AssignmentID, step.StepIndex+1)
	if err != nil {
		return errors.Wrap(err, "advance_routes: failed to get next step for assignment "+
			step.AssignmentID.String())
	}

	if err := uc.routeStepRepository.CompleteRouteStep(ctx, step.ID); err != nil {
		return errors.Wrap(err, "advance_routes: failed to complete step "+step.ID.String())
	}

	if nextStep != nil {
		return uc.advanceToNextStep(ctx, step, nextStep, simHour)
	}
	return uc.completeAssignment(ctx, step, simHour)
}

func (uc *AdvanceRoutesUseCase) advanceToNextStep(ctx context.Context, step *entity.ActiveRouteStep,
	nextStep *entity.RouteStep, simHour int64) error {
	if err := uc.routeStepRepository.ActivateRouteStep(ctx, nextStep.ID, simHour); err != nil {
		return errors.Wrap(err, "advance_routes: failed to activate step "+nextStep.ID.String())
	}
	if err := uc.wagonRepository.UpdateStation(ctx, step.WagonID, nextStep.StationID); err != nil {
		return errors.Wrap(err, "advance_routes: failed to update wagon station "+step.WagonID.String())
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
		ArrivedAt:   uc.simClock.ToDisplayTime(simHour),
	}))

	return nil
}

func (uc *AdvanceRoutesUseCase) completeAssignment(ctx context.Context,
	step *entity.ActiveRouteStep, simHour int64) error {
	unloadUntil := simHour + unloadDurationHours
	if err := uc.wagonRepository.UpdateStatus(ctx, step.WagonID, entity.Loaded, &unloadUntil); err != nil {
		return errors.Wrap(err, "advance_routes: failed to update wagon status "+step.WagonID.String())
	}
	if err := uc.assignmentRepository.UpdateStatus(ctx, step.AssignmentID, entity.AssignmentDelivered); err != nil {
		return errors.Wrap(err, "advance_routes: failed to update assignment status "+step.AssignmentID.String())
	}
	order, err := uc.orderRepository.UpdateIfFulfilled(ctx, step.OrderID)
	if err != nil {
		return errors.Wrap(err, "advance_routes: failed to update order "+step.OrderID.String())
	}
	if order != nil && order.Type == entity.External {
		uc.broadcaster.Publish(broadcaster.NewEvent(broadcaster.OrderFulfilled, order))
	}
	uc.broadcaster.Publish(broadcaster.NewEvent(broadcaster.WagonArrived, entity.WagonArrivedPayload{
		WagonID:      step.WagonID,
		OrderID:      step.OrderID,
		AssignmentID: step.AssignmentID,
	}))

	return nil
}
