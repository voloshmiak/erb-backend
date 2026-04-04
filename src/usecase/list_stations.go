package usecase

import (
	"context"
	"erb-backend/src/entity"

	"github.com/pkg/errors"
)

type ListStationsUseCase struct {
	repository StationRepository
}

func NewListStationsUseCase(repository StationRepository) *ListStationsUseCase {
	return &ListStationsUseCase{
		repository: repository,
	}
}

func (u *ListStationsUseCase) Execute(ctx context.Context) ([]*entity.Station,
	[]*entity.Edge, error) {
	stations, edges, err := u.repository.List(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to list stations")
	}
	return stations, edges, nil
}
