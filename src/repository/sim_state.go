package repository

import (
	"context"
	"database/sql"
	"time"
)

type SimStateRepository struct {
	conn *sql.DB
}

func NewSimStateRepository(conn *sql.DB) *SimStateRepository {
	return &SimStateRepository{conn: conn}
}

func (r *SimStateRepository) Load(ctx context.Context) (int64, time.Time, error) {
	var currentHour int64
	var startedAt time.Time
	err := r.conn.QueryRowContext(ctx,
		"SELECT current_hour, started_at FROM sim_state WHERE id = 1",
	).Scan(&currentHour, &startedAt)
	return currentHour, startedAt, err
}

func (r *SimStateRepository) Save(ctx context.Context, currentHour int64) error {
	_, err := r.conn.ExecContext(ctx,
		"UPDATE sim_state SET current_hour = $1 WHERE id = 1",
		currentHour,
	)
	return err
}
