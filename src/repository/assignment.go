package repository

import (
	"context"
	"database/sql"
	"erb-backend/src/entity"
	"errors"

	"github.com/google/uuid"
)

type AssignmentRepository struct {
	conn *sql.DB
}

func NewAssignmentRepository(conn *sql.DB) *AssignmentRepository {
	return &AssignmentRepository{conn: conn}
}

func (r *AssignmentRepository) Create(ctx context.Context, assignment *entity.Assignment) error {
	_, err := r.conn.ExecContext(ctx, `
		INSERT INTO assignments (id, order_id, wagon_id, empty_run_km, loaded_run_km, cost_empty_run, estimated_arrival, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, assignment.ID, assignment.OrderID, assignment.WagonID,
		assignment.EmptyRunKM, assignment.LoadedRunKM, assignment.CostEmptyRun,
		assignment.EstimatedArrival, assignment.Status)
	return err
}

func (r *AssignmentRepository) GetStats(ctx context.Context) (*entity.AssignmentStats, error) {
	s := &entity.AssignmentStats{}
	err := r.conn.QueryRowContext(ctx, `
		SELECT
			COUNT(*),
			COALESCE(SUM(empty_run_km), 0),
			COALESCE(SUM(cost_empty_run), 0),
			COALESCE(AVG(empty_run_km), 0),
			COALESCE(SUM(loaded_run_km), 0)
		FROM assignments
		WHERE status = $1
	`, entity.AssignmentDelivered).Scan(
		&s.TotalDelivered,
		&s.TotalEmptyRunKM,
		&s.TotalCostEmptyRun,
		&s.AvgEmptyRunKM,
		&s.TotalLoadedRunKM,
	)
	return s, err
}

func (r *AssignmentRepository) FindOldestPlanned(ctx context.Context) (*entity.PlannedAssignment, error) {
	var id, orderID, wagonID, firstStepID string
	err := r.conn.QueryRowContext(ctx, `
		SELECT a.id::text, a.order_id::text, a.wagon_id::text, rs.id::text
		FROM assignments a
		JOIN route_steps rs ON rs.assignment_id = a.id AND rs.step_index = 0
		WHERE a.status = $1
		LIMIT 1
	`, entity.AssignmentPlanned).Scan(&id, &orderID, &wagonID, &firstStepID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	a := &entity.PlannedAssignment{}
	for _, pair := range []struct {
		dst *uuid.UUID
		src string
	}{
		{&a.ID, id},
		{&a.OrderID, orderID},
		{&a.WagonID, wagonID},
		{&a.FirstStepID, firstStepID},
	} {
		parsed, err := uuid.Parse(pair.src)
		if err != nil {
			return nil, err
		}
		*pair.dst = parsed
	}
	return a, nil
}

func (r *AssignmentRepository) UpdateStatus(ctx context.Context, assignmentID uuid.UUID,
	status entity.AssignmentStatus) error {
	_, err := r.conn.ExecContext(ctx,
		"UPDATE assignments SET status = $1 WHERE id = $2",
		status, assignmentID,
	)
	return err
}
