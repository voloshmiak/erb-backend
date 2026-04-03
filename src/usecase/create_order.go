package usecase

import (
	"erb-backend/src/broadcaster"
	"erb-backend/src/entity"
	"time"

	"github.com/pkg/errors"

	"github.com/google/uuid"
)

var (
	ErrDesiredDateInPast = errors.New("desired date cannot be in the past")
	ErrInvalidInput      = errors.New("invalid input")
	ErrFailedToParseDate = errors.New("failed to parse desired date")
)

type OrderRepository interface {
	Create(order *entity.Order) error
}

type Broadcaster interface {
	Subscribe() chan string
	Unsubscribe(chan string)
	Publish(broadcaster.Event)
}

type CreateOrderUseCase struct {
	repository  OrderRepository
	broadcaster Broadcaster
}

func NewCreateOrderUseCase(repository OrderRepository,
	broadcaster Broadcaster) *CreateOrderUseCase {
	return &CreateOrderUseCase{
		repository:  repository,
		broadcaster: broadcaster,
	}
}

type CreateOrderInput struct {
	ClientName  string           `json:"clientName"`
	StationToID uuid.UUID        `json:"stationToId"`
	WagonType   entity.WagonType `json:"wagonType"`
	Quantity    int              `json:"quantity"`
	DesiredDate string           `json:"desiredDate"`
}

func (u *CreateOrderUseCase) Execute(input CreateOrderInput) (*entity.Order, error) {
	if input.ClientName == "" ||
		input.StationToID.String() == "00000000-0000-0000-0000-000000000000" ||
		input.Quantity <= 0 {
		return nil, ErrInvalidInput
	}

	parsedDesiredDate, err := time.Parse(time.DateOnly, input.DesiredDate)
	if err != nil {
		return nil, ErrFailedToParseDate
	}

	if parsedDesiredDate.Before(time.Now()) {
		return nil, ErrDesiredDateInPast
	}

	order := entity.NewOrder(input.ClientName, input.StationToID, input.WagonType,
		input.Quantity, parsedDesiredDate)

	if err := u.repository.Create(order); err != nil {
		return nil, err
	}

	u.broadcaster.Publish(broadcaster.NewEvent(broadcaster.OrderCreated, order))

	return order, nil
}
