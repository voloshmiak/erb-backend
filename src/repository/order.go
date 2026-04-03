package repository

import (
	"database/sql"
	"erb-backend/src/entity"
)

type OrderRepository struct {
	conn *sql.DB
}

func NewOrderRepository(conn *sql.DB) *OrderRepository {
	return &OrderRepository{conn: conn}
}

func (r *OrderRepository) Create(order *entity.Order) error {
	_, err := r.conn.Exec(`
		INSERT INTO orders (id, client_name, station_to_id, wagon_type, quantity, desired_date, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, order.ID, order.ClientName, order.StationToID, order.WagonType,
		order.Quantity, order.DesiredDate, order.Status, order.CreatedAt)
	return err
}
