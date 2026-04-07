package usecase

import (
	"context"
	"erb-backend/src/entity"

	"github.com/pkg/errors"
)

type ListWagonsUseCase struct {
	repository WagonRepository
}

func NewListWagonsUseCase(repository WagonRepository) *ListWagonsUseCase {
	return &ListWagonsUseCase{
		repository: repository,
	}
}

func (u *ListWagonsUseCase) Execute(ctx context.Context) ([]*entity.Wagon, error) {
	wagons, err := u.repository.List(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list stations")
	}
	return wagons, nil
}
