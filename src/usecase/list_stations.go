package usecase

import "erb-backend/src/entity"

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
	return u.repository.List()
}
