// Package panelsdk is a thin client for the panel HTTP api.
// See https://github.com/UltraInstinct0x/panel for the server.
package panelsdk

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const DefaultEngineVersion = "0.2.0"

// Options configures a Client.
type Options struct {
	BaseURL         string
	SiteKey         string
	SiteSecret      string
	ScrubberSecret  string // optional; required for third-party site keys
	EngineVersion   string // default "0.2.0"
	HTTPClient      *http.Client
}

// Client is the panel HTTP client.
type Client struct {
	o    Options
	http *http.Client
}

// New returns a configured Client.
func New(o Options) *Client {
	if o.EngineVersion == "" {
		o.EngineVersion = DefaultEngineVersion
	}
	hc := o.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: 10 * time.Second}
	}
	o.BaseURL = strings.TrimRight(o.BaseURL, "/")
	return &Client{o: o, http: hc}
}

// Error is returned for non-2xx responses.
type Error struct {
	Status int
	Body   []byte
}

func (e *Error) Error() string {
	return fmt.Sprintf("panel %d: %s", e.Status, truncate(string(e.Body), 300))
}

// VerifyResult is the response shape from /v1/verify.
type VerifyResult struct {
	OK       bool     `json:"ok"`
	Trust    *float64 `json:"trust,omitempty"`
	TierUsed string   `json:"tier_used,omitempty"`
	UnitIDs  []string `json:"unit_ids,omitempty"`
	Reason   string   `json:"reason,omitempty"`
}

// IngestUnitInput is the request body for ingestUnit.
type IngestUnitInput struct {
	Type    string                 `json:"type"`
	Pool    string                 `json:"pool,omitempty"`
	Payload map[string]interface{} `json:"payload"`
}

// IngestTraceInput is the request body for ingestTrace.
type IngestTraceInput struct {
	TraceID     string                 `json:"trace_id"`
	SourceAgent string                 `json:"source_agent"`
	Blob        map[string]interface{} `json:"blob"`
}

// IngestUnit posts a single unit.
func (c *Client) IngestUnit(in IngestUnitInput) (map[string]interface{}, error) {
	body, _ := json.Marshal(in)
	return c.postSigned("/api/units/ingest", body)
}

// IngestTrace posts a trace for splitting.
func (c *Client) IngestTrace(in IngestTraceInput) (map[string]interface{}, error) {
	body, _ := json.Marshal(in)
	return c.postSigned("/api/v1/traces", body)
}

// VerifyToken verifies a widget token server-side.
func (c *Client) VerifyToken(token string) (*VerifyResult, error) {
	body, _ := json.Marshal(map[string]string{"token": token, "site_key": c.o.SiteKey})
	req, _ := http.NewRequest("POST", c.o.BaseURL+"/v1/verify", bytes.NewReader(body))
	req.Header.Set("content-type", "application/json")
	var out VerifyResult
	if err := c.do(req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FetchUnit gets a unit by id.
func (c *Client) FetchUnit(id string) (map[string]interface{}, error) {
	return c.get("/api/units/" + id)
}

// FetchTrace gets a trace by id.
func (c *Client) FetchTrace(id string) (map[string]interface{}, error) {
	return c.get("/api/v1/traces/" + id)
}

// --- internals -----------------------------------------------------

func (c *Client) postSigned(path string, body []byte) (map[string]interface{}, error) {
	req, _ := http.NewRequest("POST", c.o.BaseURL+path, bytes.NewReader(body))
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-panel-site-key", c.o.SiteKey)
	req.Header.Set("x-panel-ingest-sig", hmacHex(c.o.SiteSecret, body))
	if c.o.ScrubberSecret != "" {
		req.Header.Set("x-scrubber-attestation", c.attest(body))
	}
	var out map[string]interface{}
	if err := c.do(req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) get(path string) (map[string]interface{}, error) {
	req, _ := http.NewRequest("GET", c.o.BaseURL+path, nil)
	var out map[string]interface{}
	if err := c.do(req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) do(req *http.Request, out interface{}) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &Error{Status: resp.StatusCode, Body: b}
	}
	if out == nil || len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, out)
}

func (c *Client) attest(body []byte) string {
	now := time.Now().Unix()
	jti := make([]byte, 16)
	_, _ = rand.Read(jti)
	payload := map[string]interface{}{
		"jti":            hex.EncodeToString(jti),
		"iat":            now,
		"exp":            now + 300,
		"input_hash":     "x",
		"output_hash":    sha256Hex(body),
		"mode":           "text",
		"engine_version": c.o.EngineVersion,
	}
	return jwtHS256(c.o.ScrubberSecret, payload)
}

func hmacHex(secret string, body []byte) string {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write(body)
	return hex.EncodeToString(m.Sum(nil))
}

func sha256Hex(body []byte) string {
	s := sha256.Sum256(body)
	return hex.EncodeToString(s[:])
}

func b64u(b []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}

func jwtHS256(secret string, payload map[string]interface{}) string {
	headerB, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payloadB, _ := json.Marshal(payload)
	si := b64u(headerB) + "." + b64u(payloadB)
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(si))
	return si + "." + b64u(m.Sum(nil))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
