package panelsdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type RaterClient struct {
	o    Options
	http *http.Client
}

func (r *RaterClient) NextUnit(ctx context.Context, pool string, raterID string) (*NextUnitResult, error) {
	query := url.Values{}
	if pool != "" {
		query.Set("pool", pool)
	}
	if raterID != "" {
		query.Set("rater_id", raterID)
	}
	u := r.o.BaseURL + "/api/units/next"
	if q := query.Encode(); q != "" {
		u += "?" + q
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("build next unit request: %w", err)
	}
	req.Header.Set("x-panel-site-key", r.o.SiteKey)
	b, _, err := r.do(req)
	if err != nil {
		return nil, err
	}
	var out NextUnitResult
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("parse next unit response: %w", err)
	}
	return &out, nil
}

func (r *RaterClient) SubmitJudgment(ctx context.Context, in SubmitJudgmentInput) (map[string]interface{}, error) {
	body, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("marshal submit judgment body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.o.BaseURL+"/api/units/judgment", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build submit judgment request: %w", err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-panel-site-key", r.o.SiteKey)
	b, _, err := r.do(req)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("parse submit judgment response: %w", err)
	}
	return out, nil
}

func (r *RaterClient) do(req *http.Request) ([]byte, int, error) {
	panel := PanelClient{o: r.o, http: r.http}
	return panel.do(req)
}
