package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"erb-backend/src/entity"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type MatchingGateway struct {
	URL    string
	Client *http.Client
}

func NewMatchingGateway(url string) *MatchingGateway {
	return &MatchingGateway{
		URL:    url,
		Client: http.DefaultClient,
	}
}

type matchRequest struct {
	Orders []*entity.Order `json:"orders"`
	Wagons []*entity.Wagon `json:"wagons"`
}

type matchedAssignment struct {
	OrderID          string    `json:"order_id"`
	WagonID          string    `json:"wagon_id"`
	Route            []string  `json:"route"`
	EmptyRunKM       int       `json:"empty_run_km"`
	CostEmptyRun     int       `json:"cost_empty_run"`
	EstimatedArrival time.Time `json:"estimated_arrival"`
}

type matchResponse struct {
	Assignments []matchedAssignment `json:"assignments"`
}

func (g *MatchingGateway) Match(ctx context.Context, orders []*entity.Order,
	wagons []*entity.Wagon) ([]*entity.AssignmentResult, error) {

	body, err := json.Marshal(matchRequest{Orders: orders, Wagons: wagons})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.URL+"/match", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			log.Println("failed to close response body: ", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var result matchResponse
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	out := make([]*entity.AssignmentResult, 0, len(result.Assignments))
	for _, a := range result.Assignments {
		orderID, err := uuid.Parse(a.OrderID)
		if err != nil {
			return nil, fmt.Errorf("parse order_id %q: %w", a.OrderID, err)
		}
		wagonID, err := uuid.Parse(a.WagonID)
		if err != nil {
			return nil, fmt.Errorf("parse wagon_id %q: %w", a.WagonID, err)
		}

		route := make([]uuid.UUID, 0, len(a.Route))
		for _, s := range a.Route {
			stationID, err := uuid.Parse(s)
			if err != nil {
				return nil, fmt.Errorf("parse station_id %q: %w", s, err)
			}
			route = append(route, stationID)
		}

		out = append(out, &entity.AssignmentResult{
			Assignment: &entity.Assignment{
				OrderID:          orderID,
				WagonID:          wagonID,
				EmptyRunKM:       a.EmptyRunKM,
				CostEmptyRun:     a.CostEmptyRun,
				EstimatedArrival: a.EstimatedArrival,
			},
			Route: route,
		})
	}

	return out, nil
}
