package usecase

import (
	"context"
	"erb-backend/src/entity"

	"github.com/google/uuid"
)

type ListLocomotivesUseCase struct {
	locomotiveRepo LocomotiveRepository
}

func NewListLocomotivesUseCase(locomotiveRepo LocomotiveRepository) *ListLocomotivesUseCase {
	return &ListLocomotivesUseCase{locomotiveRepo: locomotiveRepo}
}

func (uc *ListLocomotivesUseCase) Execute(ctx context.Context) ([]*entity.Locomotive, error) {
	return uc.locomotiveRepo.List(ctx)
}

type GetLocomotiveUseCase struct {
	locomotiveRepo LocomotiveRepository
}

func NewGetLocomotiveUseCase(locomotiveRepo LocomotiveRepository) *GetLocomotiveUseCase {
	return &GetLocomotiveUseCase{locomotiveRepo: locomotiveRepo}
}

func (uc *GetLocomotiveUseCase) Execute(ctx context.Context, id uuid.UUID) (*entity.Locomotive, error) {
	return uc.locomotiveRepo.GetByID(ctx, id)
}
