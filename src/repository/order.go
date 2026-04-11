package repository

import (
	"context"
	"database/sql"
	"erb-backend/src/entity"
	"log"

	"github.com/google/uuid"
)

type OrderRepository struct {
	conn *sql.DB
}

func NewOrderRepository(conn *sql.DB) *OrderRepository {
	return &OrderRepository{conn: conn}
}

func (r *OrderRepository) FindPending(ctx context.Context) ([]*entity.Order, error) {
	rows, err := r.conn.QueryContext(ctx, `
		SELECT id::text, client_name, station_to_id::text, wagon_type, 
		       quantity, desired_date, status, type, created_at
		FROM orders
		WHERE status = $1 AND type = $2
	`, entity.Pending, entity.External)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		if err = rows.Close(); err != nil {
			log.Println("failed to close rows:", err)
		}
	}(rows)

	var orders []*entity.Order
	for rows.Next() {
		var idStr, stationToIDStr string
		o := &entity.Order{}
		if err = rows.Scan(&idStr, &o.ClientName, &stationToIDStr, &o.WagonType,
			&o.Quantity, &o.DesiredDate, &o.Status, &o.Type, &o.CreatedAt); err != nil {
			return nil, err
		}
		if o.ID, err = uuid.Parse(idStr); err != nil {
			return nil, err
		}
		if o.StationToID, err = uuid.Parse(stationToIDStr); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

func (r *OrderRepository) UpdateIfFulfilled(ctx context.Context, orderID uuid.UUID) (*entity.Order, error) {
	var idStr, stationToIDStr string
	o := &entity.Order{}
	err := r.conn.QueryRowContext(ctx, `
		UPDATE orders SET status = $2
		WHERE id = $1
		AND NOT EXISTS (
			SELECT 1 FROM assignments
			WHERE order_id = $1 AND status != $3
		)
		RETURNING id::text, client_name, station_to_id::text, wagon_type,
		          quantity, desired_date, status, type, created_at
	`, orderID, entity.Fulfilled, entity.AssignmentDelivered).
		Scan(&idStr, &o.ClientName, &stationToIDStr, &o.WagonType,
			&o.Quantity, &o.DesiredDate, &o.Status, &o.Type, &o.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if o.ID, err = uuid.Parse(idStr); err != nil {
		return nil, err
	}
	if o.StationToID, err = uuid.Parse(stationToIDStr); err != nil {
		return nil, err
	}
	return o, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context,
	orderID uuid.UUID, status entity.OrderStatus) error {
	_, err := r.conn.ExecContext(ctx,
		"UPDATE orders SET status = $1 WHERE id = $2",
		status, orderID,
	)
	return err
}

func (r *OrderRepository) Create(ctx context.Context, order *entity.Order) error {
	_, err := r.conn.ExecContext(ctx, `
		INSERT INTO orders (id, client_name, station_to_id, wagon_type, quantity, desired_date, status, type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, order.ID, order.ClientName, order.StationToID, order.WagonType,
		order.Quantity, order.DesiredDate, order.Status, order.Type, order.CreatedAt)
	return err
}

func (r *OrderRepository) List(ctx context.Context) ([]*entity.Order, error) {
	rows, err := r.conn.QueryContext(ctx, `
		SELECT id::text, client_name, station_to_id::text, wagon_type, 
		       quantity, desired_date, status, type, created_at
		FROM orders
		WHERE type = $1
		ORDER BY created_at
	`, entity.External)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		if err = rows.Close(); err != nil {
			log.Println("failed to close rows:", err)
		}
	}(rows)

	var orders []*entity.Order
	for rows.Next() {
		var idStr, stationToIDStr string
		o := &entity.Order{}
		if err = rows.Scan(&idStr, &o.ClientName, &stationToIDStr, &o.WagonType,
			&o.Quantity, &o.DesiredDate, &o.Status, &o.Type, &o.CreatedAt); err != nil {
			return nil, err
		}
		if o.ID, err = uuid.Parse(idStr); err != nil {
			return nil, err
		}
		if o.StationToID, err = uuid.Parse(stationToIDStr); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}
