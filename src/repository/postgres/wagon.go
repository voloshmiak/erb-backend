package postgres

import (
	"database/sql"
	"erb-backend/src/entity"
)

type WagonRepository struct {
	conn *sql.DB
}

func NewWagonRepository(conn *sql.DB) *WagonRepository {
	return &WagonRepository{conn: conn}
}

func (r *WagonRepository) ListStatusCounts() ([]entity.WagonStatusCount, error) {
	rows, err := r.conn.Query(`
		SELECT wagon_type, status, COUNT(*) AS count
		FROM wagons
		GROUP BY wagon_type, status
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var counts []entity.WagonStatusCount
	for rows.Next() {
		var c entity.WagonStatusCount
		if err := rows.Scan(&c.Type, &c.Status, &c.Count); err != nil {
			return nil, err
		}
		counts = append(counts, c)
	}
	return counts, rows.Err()
}
