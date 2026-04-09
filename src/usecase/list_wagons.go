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

type ListWagonsResult struct {
	Wagons       []*entity.Wagon
	StatusCounts map[entity.WagonStatus]int
}

func (u *ListWagonsUseCase) Execute(ctx context.Context) (*ListWagonsResult, error) {
	wagons, err := u.repository.List(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list wagons")
	}

	counts, err := u.repository.ListStatusCounts(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list status counts")
	}

	statusCounts := make(map[entity.WagonStatus]int)
	for _, c := range counts {
		statusCounts[c.Status] += c.Count
	}

	return &ListWagonsResult{
		Wagons:       wagons,
		StatusCounts: statusCounts,
	}, nil
}
