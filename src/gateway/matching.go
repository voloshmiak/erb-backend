package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"erb-backend/src/entity"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	matchEndpoint = "/api/match"
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

type orderDTO struct {
	OrderID     string `json:"order_id"`
	StationToID string `json:"station_to_id"`
	WagonType   string `json:"wagon_type"`
	Quantity    int    `json:"quantity"`
	DesiredDate string `json:"desired_date"`
}

type wagonDTO struct {
	WagonID          string  `json:"wagon_id"`
	WagonNumber      string  `json:"wagon_number"`
	WagonType        string  `json:"wagon_type"`
	CurrentStationID string  `json:"current_station_id"`
	IdleDays         float64 `json:"idle_days"`
}

type stationDTO struct {
	StationID string  `json:"station_id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
}

type edgeDTO struct {
	FromStationID string  `json:"from_station_id"`
	ToStationID   string  `json:"to_station_id"`
	DistanceKM    float64 `json:"distance_km"`
}

type matchRequest struct {
	Orders   []orderDTO   `json:"orders"`
	Wagons   []wagonDTO   `json:"wagons"`
	Stations []stationDTO `json:"stations"`
	Edges    []edgeDTO    `json:"edges"`
}

type matchedAssignment struct {
	OrderID        string   `json:"order_id"`
	WagonID        string   `json:"wagon_id"`
	WagonNumber    string   `json:"wagon_number"`
	Route          []string `json:"route"`
	EmptyRunKM     float64  `json:"empty_run_km"`
	CostEmptyRun   float64  `json:"cost_empty_run"`
	EstimatedHours float64  `json:"estimated_hours"`
}

type unmatchedOrder struct {
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}

type matchMetrics struct {
	TotalEmptyKM    float64 `json:"total_empty_km"`
	AvgEmptyRunKM   float64 `json:"avg_empty_run_km"`
	TotalCost       float64 `json:"total_cost"`
	NaiveTotalCost  float64 `json:"naive_total_cost"`
	CostSaved       float64 `json:"cost_saved"`
	MatchRate       float64 `json:"match_rate"`
	WagonsMatched   int     `json:"wagons_matched"`
	OrdersMatched   int     `json:"orders_matched"`
	OrdersUnmatched int     `json:"orders_unmatched"`
}

type matchResponse struct {
	Assignments     []matchedAssignment `json:"assignments"`
	UnmatchedOrders []unmatchedOrder    `json:"unmatched_orders"`
	Metrics         matchMetrics        `json:"metrics"`
}

// Conversion helpers

func toOrderDTOs(orders []*entity.Order) []orderDTO {
	out := make([]orderDTO, 0, len(orders))
	for _, o := range orders {
		out = append(out, orderDTO{
			OrderID:     o.ID.String(),
			StationToID: o.StationToID.String(),
			WagonType:   string(o.WagonType),
			Quantity:    o.Quantity,
			DesiredDate: o.DesiredDate.Format(time.DateOnly),
		})
	}
	return out
}

func toWagonDTOs(wagons []*entity.Wagon) []wagonDTO {
	out := make([]wagonDTO, 0, len(wagons))
	for _, w := range wagons {
		idleDays := time.Since(w.LastUnloadTime).Hours() / 24
		out = append(out, wagonDTO{
			WagonID:          w.ID.String(),
			WagonNumber:      w.Number,
			WagonType:        string(w.Type),
			CurrentStationID: w.CurrentStationID.String(),
			IdleDays:         idleDays,
		})
	}
	return out
}

func toStationDTOs(stations []*entity.Station) []stationDTO {
	out := make([]stationDTO, 0, len(stations))
	for _, s := range stations {
		out = append(out, stationDTO{
			StationID: s.ID.String(),
			Name:      s.Name,
			Type:      string(s.Type),
			Lat:       s.Location.Latitude,
			Lng:       s.Location.Longitude,
		})
	}
	return out
}

func toEdgeDTOs(edges []*entity.Edge) []edgeDTO {
	out := make([]edgeDTO, 0, len(edges))
	for _, e := range edges {
		if !e.IsActive {
			continue
		}
		out = append(out, edgeDTO{
			FromStationID: e.FromStationID.String(),
			ToStationID:   e.ToStationID.String(),
			DistanceKM:    e.DistanceKM,
		})
	}
	return out
}

func (g *MatchingGateway) Match(ctx context.Context, orders []*entity.Order,
	wagons []*entity.Wagon, stations []*entity.Station, edges []*entity.Edge) ([]*entity.AssignmentResult, error) {

	reqBody := matchRequest{
		Orders:   toOrderDTOs(orders),
		Wagons:   toWagonDTOs(wagons),
		Stations: toStationDTOs(stations),
		Edges:    toEdgeDTOs(edges),
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.URL+matchEndpoint,
		bytes.NewReader(body))
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

	for _, u := range result.UnmatchedOrders {
		log.Printf("matching: unmatched order %s: %s", u.OrderID, u.Reason)
	}

	m := result.Metrics
	log.Printf("matching metrics: matched=%d unmatched=%d match_rate=%.2f total_empty_km=%.1f total_cost=%.1f cost_saved=%.1f",
		m.OrdersMatched, m.OrdersUnmatched, m.MatchRate, m.TotalEmptyKM, m.TotalCost, m.CostSaved)

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
				EmptyRunKM:       int(math.Round(a.EmptyRunKM)),
				CostEmptyRun:     int(math.Round(a.CostEmptyRun)),
				EstimatedArrival: time.Now().Add(time.Duration(a.EstimatedHours * float64(time.Hour))),
			},
			Route: route,
		})
	}

	return out, nil
}
