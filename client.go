package panelsdk

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type PanelClient struct {
	o    Options
	http *http.Client
}

type ingestCallOptions struct {
	scrubberText string
	externalRef  string
}

type IngestOpt func(*ingestCallOptions)

func WithScrubberText(text string) IngestOpt {
	return func(o *ingestCallOptions) { o.scrubberText = text }
}

func WithExternalRef(ref string) IngestOpt {
	return func(o *ingestCallOptions) { o.externalRef = ref }
}

func (c *Client) IngestTrace(ctx context.Context, in IngestTraceInput) (*TraceResult, error) {
	return c.panel.IngestTrace(ctx, in)
}

func (c *Client) IngestTraceAndWait(ctx context.Context, in IngestTraceInput, maxWaitSeconds float64, pollIntervalSeconds float64) (*FetchTraceResult, error) {
	return c.panel.IngestTraceAndWait(ctx, in, maxWaitSeconds, pollIntervalSeconds)
}

func (c *Client) FetchTrace(ctx context.Context, id string) (*FetchTraceResult, error) {
	return c.panel.FetchTrace(ctx, id)
}

func (c *Client) IngestUnits(ctx context.Context, payload interface{}) (map[string]interface{}, error) {
	return c.panel.IngestUnits(ctx, payload)
}

func (c *Client) ScoreUnit(ctx context.Context, ref string, id string) (*ScoreResult, error) {
	return c.panel.ScoreUnit(ctx, ref, id)
}

func (c *Client) SkillReview(ctx context.Context, in SkillReviewInput) (*SkillReviewResult, error) {
	return c.panel.SkillReview(ctx, in)
}

func (c *Client) VerifyToken(ctx context.Context, token string) (*VerifyResult, error) {
	return c.panel.VerifyToken(ctx, token)
}

func (c *Client) IngestUnit(ctx context.Context, in IngestUnitInput, opts ...IngestOpt) (*IngestUnitResult, error) {
	callOpts := ingestCallOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&callOpts)
		}
	}

	bodyMap := map[string]interface{}{
		"type": in.Type,
	}
	if in.Pool != "" {
		bodyMap["pool"] = in.Pool
	}
	if in.Payload != nil {
		bodyMap["payload"] = in.Payload
	}
	if in.ExternalRef != "" {
		bodyMap["external_ref"] = in.ExternalRef
	}
	if callOpts.externalRef != "" {
		bodyMap["external_ref"] = callOpts.externalRef
	}
	if in.SourceAgent != "" {
		bodyMap["source_agent"] = in.SourceAgent
	}
	if in.Question != "" {
		bodyMap["question"] = in.Question
	}

	body, err := json.Marshal(bodyMap)
	if err != nil {
		return nil, fmt.Errorf("marshal ingest unit body: %w", err)
	}
	att := ""
	if callOpts.scrubberText != "" {
		att, err = c.scrubberAttestation(sha256Hex(body))
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.o.BaseURL+"/api/units/ingest", io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		return nil, fmt.Errorf("build ingest unit request: %w", err)
	}
	req.ContentLength = int64(len(body))
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-panel-site-key", c.o.SiteKey)
	req.Header.Set("x-panel-ingest-sig", c.signBody(body))
	if att != "" {
		req.Header.Set("x-scrubber-attestation", att)
	}

	b, _, err := c.panel.do(req)
	if err != nil {
		return nil, err
	}
	var out IngestUnitResult
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("parse ingest unit response: %w", err)
	}
	return &out, nil
}

func (c *Client) GetTrace(ctx context.Context, traceID string) (*TraceStatus, error) {
	path := c.o.BaseURL + "/api/v1/traces/" + url.PathEscape(traceID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("build get trace request: %w", err)
	}
	req.Header.Set("x-panel-site-key", c.o.SiteKey)
	req.Header.Set("x-panel-ingest-sig", c.signBody([]byte{}))
	b, _, err := c.panel.do(req)
	if err != nil {
		return nil, err
	}
	var out TraceStatus
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("parse trace status response: %w", err)
	}
	return &out, nil
}

func (c *Client) SubmitJudgment(ctx context.Context, in JudgmentInput) (*JudgmentResult, error) {
	body, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("marshal judgment body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.o.BaseURL+"/api/judgments", io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		return nil, fmt.Errorf("build submit judgment request: %w", err)
	}
	req.ContentLength = int64(len(body))
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-panel-site-key", c.o.SiteKey)
	req.Header.Set("x-panel-ingest-sig", c.signBody(body))

	b, _, err := c.panel.do(req)
	if err != nil {
		return nil, err
	}
	var out JudgmentResult
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("parse judgment response: %w", err)
	}
	return &out, nil
}

func (c *Client) Scrub(ctx context.Context, text string) (*ScrubResult, error) {
	body, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return nil, fmt.Errorf("marshal scrub request: %w", err)
	}
	url := strings.TrimRight(c.o.ScrubberURL, "/")
	if url == "" {
		url = strings.TrimRight(c.o.BaseURL, "/")
	}
	url += "/v1/scrub"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		return nil, fmt.Errorf("build scrub request: %w", err)
	}
	req.ContentLength = int64(len(body))
	req.Header.Set("content-type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("scrub request failed: %w", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read scrub response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &ScrubberError{Status: resp.StatusCode, Body: b}
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, fmt.Errorf("parse scrub response: %w", err)
	}
	out := &ScrubResult{Extra: map[string]interface{}{}}
	if v, ok := raw["scrubbed"].(string); ok {
		out.Scrubbed = v
	}
	if v, ok := raw["mapping_id"].(string); ok {
		out.MappingID = v
	}
	for k, v := range raw {
		if k == "scrubbed" || k == "mapping_id" {
			continue
		}
		out.Extra[k] = v
	}
	return out, nil
}

func (c *Client) signBody(body []byte) string {
	m := hmac.New(sha256.New, []byte(c.o.SiteSecret))
	_, _ = m.Write(body)
	return hex.EncodeToString(m.Sum(nil))
}

func (c *Client) scrubberAttestation(outputHashHex string) (string, error) {
	if c.o.ScrubberSecret == "" {
		return "", fmt.Errorf("scrubber_secret is required for scrubber attestation")
	}
	now := time.Now().Unix()
	jti := make([]byte, 16)
	if _, err := rand.Read(jti); err != nil {
		return "", fmt.Errorf("random jti: %w", err)
	}
	headerJSON, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", fmt.Errorf("marshal jwt header: %w", err)
	}
	payloadJSON, err := json.Marshal(map[string]interface{}{
		"jti":            hex.EncodeToString(jti),
		"iat":            now,
		"exp":            now + 300,
		"input_hash":     outputHashHex,
		"output_hash":    outputHashHex,
		"mode":           "text",
		"engine_version": c.o.EngineVersion,
	})
	if err != nil {
		return "", fmt.Errorf("marshal jwt payload: %w", err)
	}
	signingInput := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(payloadJSON)
	sig := hmac.New(sha256.New, []byte(c.o.ScrubberSecret))
	_, _ = sig.Write([]byte(signingInput))
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig.Sum(nil)), nil
}

func (c *PanelClient) IngestTrace(ctx context.Context, in IngestTraceInput) (*TraceResult, error) {
	body, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("marshal ingest trace body: %w", err)
	}
	res, status, err := c.signedJSON(ctx, http.MethodPost, "/api/v1/traces", nil, body)
	if err != nil {
		return nil, err
	}
	if status == http.StatusAccepted {
		var pending struct {
			TraceID string `json:"trace_id"`
			Status  string `json:"status"`
			Poll    string `json:"poll"`
		}
		if err := json.Unmarshal(res, &pending); err != nil {
			return nil, fmt.Errorf("parse pending ingest trace response: %w", err)
		}
		pollURL := pending.Poll
		if strings.HasPrefix(pollURL, "/") {
			pollURL = c.o.BaseURL + "/api" + strings.TrimPrefix(pollURL, "/v1")
		}
		return &TraceResult{Status: "pending", TraceID: pending.TraceID, PollURL: &pollURL}, nil
	}

	var out struct {
		TraceID         string   `json:"trace_id"`
		UnitIDs         []string `json:"unit_ids"`
		StructuralCount int      `json:"structural_count"`
		LLMCount        int      `json:"llm_count"`
		SkippedCount    int      `json:"skipped_count"`
	}
	if err := json.Unmarshal(res, &out); err != nil {
		return nil, fmt.Errorf("parse ingest trace response: %w", err)
	}
	return &TraceResult{
		Status:          "done",
		TraceID:         out.TraceID,
		UnitIDs:         out.UnitIDs,
		StructuralCount: &out.StructuralCount,
		LLMCount:        &out.LLMCount,
		SkippedCount:    &out.SkippedCount,
	}, nil
}

func (c *PanelClient) IngestTraceAndWait(ctx context.Context, in IngestTraceInput, maxWaitSeconds float64, pollIntervalSeconds float64) (*FetchTraceResult, error) {
	if maxWaitSeconds <= 0 {
		maxWaitSeconds = 60
	}
	if pollIntervalSeconds <= 0 {
		pollIntervalSeconds = 1.5
	}
	res, err := c.IngestTrace(ctx, in)
	if err != nil {
		return nil, err
	}
	if res.Status == "done" {
		return &FetchTraceResult{
			TraceID:         res.TraceID,
			Status:          "done",
			UnitIDs:         res.UnitIDs,
			StructuralCount: res.StructuralCount,
			LLMCount:        res.LLMCount,
			SkippedCount:    res.SkippedCount,
		}, nil
	}
	deadline := time.Now().Add(durationFromSeconds(maxWaitSeconds))
	for time.Now().Before(deadline) {
		fetched, err := c.FetchTrace(ctx, res.TraceID)
		if err != nil {
			return nil, err
		}
		if fetched.Status != "pending" {
			return fetched, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(durationFromSeconds(pollIntervalSeconds)):
		}
	}
	return nil, fmt.Errorf("trace wait timed out after %.2fs", maxWaitSeconds)
}

func (c *PanelClient) FetchTrace(ctx context.Context, id string) (*FetchTraceResult, error) {
	path := "/api/v1/traces/" + url.PathEscape(id)
	b, _, err := c.signedJSON(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	var out FetchTraceResult
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("parse fetch trace response: %w", err)
	}
	return &out, nil
}

func (c *PanelClient) IngestUnits(ctx context.Context, payload interface{}) (map[string]interface{}, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal ingest units body: %w", err)
	}
	b, _, err := c.signedJSON(ctx, http.MethodPost, "/api/units/ingest", nil, body)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("parse ingest units response: %w", err)
	}
	return out, nil
}

func (c *PanelClient) ScoreUnit(ctx context.Context, ref string, id string) (*ScoreResult, error) {
	query := url.Values{}
	if ref != "" {
		query.Set("ref", ref)
	}
	if id != "" {
		query.Set("id", id)
	}
	canonical := scoreCanonicalString(c.o.SiteKey, ref, id)
	headers := map[string]string{"x-panel-ingest-sig": hmacHex(c.o.SiteSecret, []byte(canonical))}
	b, _, err := c.signedJSON(ctx, http.MethodGet, "/api/units/score", query, nil, headers)
	if err != nil {
		return nil, err
	}
	var out ScoreResult
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("parse score response: %w", err)
	}
	return &out, nil
}

func (c *PanelClient) SkillReview(ctx context.Context, in SkillReviewInput) (*SkillReviewResult, error) {
	body, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("marshal skill review body: %w", err)
	}
	b, _, err := c.signedJSON(ctx, http.MethodPost, "/api/v1/skill-review", nil, body)
	if err != nil {
		return nil, err
	}
	var out SkillReviewResult
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("parse skill review response: %w", err)
	}
	return &out, nil
}

func (c *PanelClient) VerifyToken(ctx context.Context, token string) (*VerifyResult, error) {
	body, err := json.Marshal(map[string]string{"token": token, "site_key": c.o.SiteKey})
	if err != nil {
		return nil, fmt.Errorf("marshal verify token body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.o.BaseURL+"/v1/verify", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build verify request: %w", err)
	}
	req.Header.Set("content-type", "application/json")
	b, _, err := c.do(req)
	if err != nil {
		return nil, err
	}
	var out VerifyResult
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("parse verify response: %w", err)
	}
	return &out, nil
}

func (c *PanelClient) signedJSON(ctx context.Context, method string, path string, query url.Values, body []byte, extraHeaders ...map[string]string) ([]byte, int, error) {
	bodyToSend := body
	attestation := ""
	var err error
	if method == http.MethodPost && (path == "/api/v1/traces" || path == "/api/units/ingest") {
		bodyToSend, attestation, err = c.scrubberMode(ctx, body)
		if err != nil {
			return nil, 0, err
		}
	}

	u := c.o.BaseURL + path
	if query != nil {
		q := query.Encode()
		if q != "" {
			u += "?" + q
		}
	}

	var reader io.Reader
	if bodyToSend != nil {
		reader = bytes.NewReader(bodyToSend)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, reader)
	if err != nil {
		return nil, 0, fmt.Errorf("build request: %w", err)
	}
	if bodyToSend != nil {
		req.Header.Set("content-type", "application/json")
	}
	req.Header.Set("x-panel-site-key", c.o.SiteKey)
	if c.o.SiteSecret != "" {
		sigPayload := bodyToSend
		if sigPayload == nil {
			sigPayload = []byte{}
		}
		req.Header.Set("x-panel-ingest-sig", hmacHex(c.o.SiteSecret, sigPayload))
	}
	if c.o.SiteSecretSource == "raw" {
		req.Header.Set("x-panel-ingest-secret", c.o.SiteSecret)
	}
	if attestation != "" {
		req.Header.Set("x-scrubber-attestation", attestation)
	}
	if len(extraHeaders) > 0 {
		for k, v := range extraHeaders[0] {
			req.Header.Set(k, v)
		}
	}

	return c.do(req)
}

func (c *PanelClient) do(req *http.Request) ([]byte, int, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		var rl struct {
			Error       string `json:"error"`
			Scope       string `json:"scope"`
			RetryAfterS int    `json:"retry_after_s"`
		}
		_ = json.Unmarshal(b, &rl)
		if rl.RetryAfterS == 0 {
			ra := resp.Header.Get("Retry-After")
			if ra != "" {
				rl.RetryAfterS, _ = strconv.Atoi(ra)
			}
		}
		return nil, resp.StatusCode, &RateLimitError{Scope: rl.Scope, RetryAfter: rl.RetryAfterS, Status: resp.StatusCode, Body: b}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.StatusCode, &Error{Status: resp.StatusCode, Body: b}
	}
	return b, resp.StatusCode, nil
}
