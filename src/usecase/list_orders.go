package usecase

import (
	"context"
	"erb-backend/src/entity"

	"github.com/pkg/errors"
)

type ListOrdersUseCase struct {
	repository OrderRepository
}

func NewListOrdersUseCase(repository OrderRepository) *ListOrdersUseCase {
	return &ListOrdersUseCase{
		repository: repository,
	}
}

func (u *ListOrdersUseCase) Execute(ctx context.Context) ([]*entity.Order, error) {
	wagons, err := u.repository.List(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list wagons")
	}
	return wagons, nil
}
