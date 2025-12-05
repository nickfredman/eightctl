package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type TravelActions struct{ c *Client }

func (c *Client) Travel() *TravelActions { return &TravelActions{c: c} }

func (t *TravelActions) Trips(ctx context.Context) (any, error) {
	if err := t.c.requireUser(ctx); err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/users/%s/travel/trips", t.c.UserID)
	var res any
	err := t.c.do(ctx, http.MethodGet, path, nil, nil, &res)
	return res, err
}

func (t *TravelActions) CreateTrip(ctx context.Context, body map[string]any) error {
	if err := t.c.requireUser(ctx); err != nil {
		return err
	}
	path := fmt.Sprintf("/users/%s/travel/trips", t.c.UserID)
	return t.c.do(ctx, http.MethodPost, path, nil, body, nil)
}

func (t *TravelActions) CreatePlan(ctx context.Context, tripID string, body map[string]any) error {
	if err := t.c.requireUser(ctx); err != nil {
		return err
	}
	path := fmt.Sprintf("/users/%s/travel/trips/%s/plans", t.c.UserID, tripID)
	return t.c.do(ctx, http.MethodPost, path, nil, body, nil)
}

func (t *TravelActions) UpdatePlan(ctx context.Context, planID string, body map[string]any) error {
	if err := t.c.requireUser(ctx); err != nil {
		return err
	}
	path := fmt.Sprintf("/users/%s/travel/plans/%s", t.c.UserID, planID)
	return t.c.do(ctx, http.MethodPatch, path, nil, body, nil)
}

func (t *TravelActions) DeleteTrip(ctx context.Context, tripID string) error {
	if err := t.c.requireUser(ctx); err != nil {
		return err
	}
	path := fmt.Sprintf("/users/%s/travel/trips/%s", t.c.UserID, tripID)
	return t.c.do(ctx, http.MethodDelete, path, nil, nil, nil)
}

func (t *TravelActions) Plans(ctx context.Context, tripID string) (any, error) {
	if err := t.c.requireUser(ctx); err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/users/%s/travel/trips/%s/plans", t.c.UserID, tripID)
	var res any
	err := t.c.do(ctx, http.MethodGet, path, nil, nil, &res)
	return res, err
}

func (t *TravelActions) PlanTasks(ctx context.Context, planID string) (any, error) {
	if err := t.c.requireUser(ctx); err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/users/%s/travel/plans/%s/tasks", t.c.UserID, planID)
	var res any
	err := t.c.do(ctx, http.MethodGet, path, nil, nil, &res)
	return res, err
}

func (t *TravelActions) AirportSearch(ctx context.Context, query string) (any, error) {
	q := url.Values{"query": []string{query}}
	var res any
	err := t.c.do(ctx, http.MethodGet, "/travel/airport-search", q, nil, &res)
	return res, err
}

func (t *TravelActions) FlightStatus(ctx context.Context, flight string) (any, error) {
	q := url.Values{"flightNumber": []string{flight}}
	var res any
	err := t.c.do(ctx, http.MethodGet, "/travel/flight-status", q, nil, &res)
	return res, err
}
