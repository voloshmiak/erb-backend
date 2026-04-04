package usecase

import (
	"erb-backend/src/entity"

	"github.com/pkg/errors"
)

type StationRepository interface {
	List() ([]*entity.Station, []*entity.Edge, error)
}

type ListStationsUseCase struct {
	repository StationRepository
}

func NewListStationsUseCase(repository StationRepository) *ListStationsUseCase {
	return &ListStationsUseCase{
		repository: repository,
	}
}

func (u *ListStationsUseCase) Execute() ([]*entity.Station, []*entity.Edge, error) {
	stations, edges, err := u.repository.List()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to list stations")
	}
	return stations, edges, nil
}
