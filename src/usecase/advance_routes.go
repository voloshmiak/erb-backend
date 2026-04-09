package usecase

import (
	"context"
	"erb-backend/src/broadcaster"
	"erb-backend/src/entity"
	"time"

	"github.com/pkg/errors"
)

type AdvanceRoutesUseCase struct {
	routeStepRepository  RouteStepRepository
	wagonRepository      WagonRepository
	assignmentRepository AssignmentRepository
	orderRepository      OrderRepository
	broadcaster          Broadcaster
}

func NewAdvanceRoutesUseCase(
	routeStepRepository RouteStepRepository,
	wagonRepository WagonRepository,
	assignmentRepository AssignmentRepository,
	orderRepository OrderRepository,
	b Broadcaster,
) *AdvanceRoutesUseCase {
	return &AdvanceRoutesUseCase{
		routeStepRepository:  routeStepRepository,
		wagonRepository:      wagonRepository,
		assignmentRepository: assignmentRepository,
		orderRepository:      orderRepository,
		broadcaster:          b,
	}
}

func (uc *AdvanceRoutesUseCase) Execute(ctx context.Context) error {
	activeSteps, err := uc.routeStepRepository.FindActiveRouteSteps(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get active steps")
	}

	now := time.Now()

	for _, step := range activeSteps {
		err = uc.processStep(ctx, step, now)
		if err != nil {
			return errors.Wrap(err, "failed to process step")
		}
	}

	return nil
}

func (uc *AdvanceRoutesUseCase) processStep(ctx context.Context,
	step *entity.ActiveRouteStep, now time.Time) error {
	if err := uc.routeStepRepository.CompleteRouteStep(ctx, step.ID); err != nil {
		return errors.Wrap(err, "advance_routes: failed to complete step "+step.ID.String())
	}

	nextStep, err := uc.routeStepRepository.FindNextRouteStep(ctx, step.AssignmentID, step.StepIndex+1)
	if err != nil {
		return errors.Wrap(err, "advance_routes: failed to get next step for assignment "+
			step.AssignmentID.String())
	}

	if nextStep != nil {
		return uc.advanceToNextStep(ctx, step, nextStep, now)
	}
	return uc.completeAssignment(ctx, step)
}

func (uc *AdvanceRoutesUseCase) advanceToNextStep(ctx context.Context, step *entity.ActiveRouteStep,
	nextStep *entity.RouteStep, now time.Time) error {
	if err := uc.routeStepRepository.ActivateRouteStep(ctx, nextStep.ID); err != nil {
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
		ArrivedAt:   now,
	}))

	return nil
}

func (uc *AdvanceRoutesUseCase) completeAssignment(ctx context.Context,
	step *entity.ActiveRouteStep) error {
	if err := uc.wagonRepository.UpdateStatus(ctx, step.WagonID, entity.Loaded); err != nil {
		return errors.Wrap(err, "advance_routes: failed to update wagon status "+step.WagonID.String())
	}
	if err := uc.assignmentRepository.UpdateStatus(ctx, step.AssignmentID, entity.AssignmentDelivered); err != nil {
		return errors.Wrap(err, "advance_routes: failed to update assignment status "+step.AssignmentID.String())
	}
	order, err := uc.orderRepository.UpdateIfFulfilled(ctx, step.OrderID)
	if err != nil {
		return errors.Wrap(err, "advance_routes: failed to update order "+step.OrderID.String())
	}
	if order != nil {
		uc.broadcaster.Publish(broadcaster.NewEvent(broadcaster.OrderFulfilled, order))
	}
	uc.broadcaster.Publish(broadcaster.NewEvent(broadcaster.WagonArrived, entity.WagonArrivedPayload{
		WagonID:      step.WagonID,
		OrderID:      step.OrderID,
		AssignmentID: step.AssignmentID,
	}))

	return nil
}
