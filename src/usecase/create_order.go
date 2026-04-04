package usecase

import (
	"erb-backend/src/broadcaster"
	"erb-backend/src/entity"
	"time"

	"github.com/pkg/errors"

	"github.com/google/uuid"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("not found")
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
	wagonRepository WagonRepository
	orderRepository OrderRepository
	broadcaster     Broadcaster
}

func NewCreateOrderUseCase(repository OrderRepository,
	broadcaster Broadcaster) *CreateOrderUseCase {
	return &CreateOrderUseCase{
		orderRepository: repository,
		broadcaster:     broadcaster,
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
		return nil, errors.Wrap(ErrInvalidInput, "missing or invalid fields")
	}

	parsedDesiredDate, err := time.Parse(time.DateOnly, input.DesiredDate)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse desired date")
	}

	if parsedDesiredDate.Before(time.Now()) {
		return nil, errors.Wrap(ErrInvalidInput, "desired date cannot be in the past")
	}

	exists, err := u.wagonRepository.Exists(input.StationToID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check station existence")
	}

	if !exists {
		return nil, errors.Wrap(ErrNotFound, "station")
	}

	order := entity.NewOrder(input.ClientName, input.StationToID, input.WagonType,
		input.Quantity, parsedDesiredDate)

	if err = u.orderRepository.Create(order); err != nil {
		return nil, errors.Wrap(err, "failed to create order")
	}

	u.broadcaster.Publish(broadcaster.NewEvent(broadcaster.OrderCreated, order))

	return order, nil
}
