package panelsdk

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestIngestTraceSignsBody(t *testing.T) {
	var bodyBytes []byte
	var gotSig string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/traces" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var err error
		bodyBytes, err = io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		gotSig = r.Header.Get("x-panel-ingest-sig")
		_, _ = w.Write([]byte(`{"trace_id":"tr_1","unit_ids":["u_1"],"structural_count":1,"llm_count":2,"skipped_count":0}`))
	}))
	defer ts.Close()

	c := New(Options{BaseURL: ts.URL, SiteKey: "pk_test", SiteSecret: "sek"})
	_, err := c.IngestTrace(context.Background(), IngestTraceInput{SourceAgent: "sdk", Blob: map[string]interface{}{"k": "v"}})
	if err != nil {
		t.Fatal(err)
	}
	m := hmac.New(sha256.New, []byte("sek"))
	_, _ = m.Write(bodyBytes)
	if want := hex.EncodeToString(m.Sum(nil)); gotSig != want {
		t.Fatalf("sig mismatch got=%q want=%q", gotSig, want)
	}
}

func TestSelfSignScrubberJWT(t *testing.T) {
	var attestation string
	var posted []byte

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/units/ingest" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		attestation = r.Header.Get("x-scrubber-attestation")
		posted, _ = io.ReadAll(r.Body)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	c := New(Options{
		BaseURL:        ts.URL,
		SiteKey:        "pk_test",
		SiteSecret:     "sek",
		ScrubberMode:   "self-sign",
		ScrubberSecret: "scrub-secret",
	})
	_, err := c.IngestUnits(context.Background(), map[string]interface{}{"units": []interface{}{map[string]interface{}{"type": "ai_output_rating", "image_url": "https://x"}}})
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(attestation, ".")
	if len(parts) != 3 {
		t.Fatalf("invalid jwt parts: %d", len(parts))
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["output_hash"] != sha256Hex(posted) {
		t.Fatalf("output_hash mismatch: got=%v want=%s", payload["output_hash"], sha256Hex(posted))
	}
	if payload["engine_version"] != DefaultEngineVersion {
		t.Fatalf("engine_version mismatch: %v", payload["engine_version"])
	}
}

func TestIngestTraceAndWaitCompletes(t *testing.T) {
	var polls int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/traces":
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"trace_id":"tr_wait","status":"pending","poll":"/v1/traces/tr_wait"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/traces/tr_wait":
			if atomic.AddInt32(&polls, 1) < 2 {
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write([]byte(`{"trace_id":"tr_wait","status":"pending"}`))
				return
			}
			_, _ = w.Write([]byte(`{"trace_id":"tr_wait","status":"done","unit_ids":["u_1"]}`))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer ts.Close()

	c := New(Options{BaseURL: ts.URL, SiteKey: "pk_test", SiteSecret: "sek"})
	res, err := c.IngestTraceAndWait(context.Background(), IngestTraceInput{SourceAgent: "sdk", Blob: map[string]interface{}{"x": 1}}, 2, 0.01)
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "done" || res.TraceID != "tr_wait" {
		t.Fatalf("unexpected result: %#v", res)
	}
}

func TestRetry429ThenSuccess(t *testing.T) {
	var calls int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"rate_limited","scope":"ingest","retry_after_s":0}`))
			return
		}
		_, _ = w.Write([]byte(`{"trace_id":"tr_2","unit_ids":[],"structural_count":0,"llm_count":0,"skipped_count":0}`))
	}))
	defer ts.Close()

	c := New(Options{BaseURL: ts.URL, SiteKey: "pk_test", SiteSecret: "sek", MaxRetries: 1})
	_, err := c.IngestTrace(context.Background(), IngestTraceInput{SourceAgent: "sdk", Blob: map[string]interface{}{"x": 1}})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}

func TestScoreUnitCanonicalSignature(t *testing.T) {
	var sig string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/units/score" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		sig = r.Header.Get("x-panel-ingest-sig")
		_, _ = w.Write([]byte(`{"counts":{"yes":1},"trust_weighted_score":0.9}`))
	}))
	defer ts.Close()

	c := New(Options{BaseURL: ts.URL, SiteKey: "pk_test", SiteSecret: "sek"})
	_, err := c.ScoreUnit(context.Background(), "ext_1", "")
	if err != nil {
		t.Fatal(err)
	}
	canonical := "GET\n/api/units/score\nref=ext_1\nsite=pk_test"
	m := hmac.New(sha256.New, []byte("sek"))
	_, _ = m.Write([]byte(canonical))
	if want := hex.EncodeToString(m.Sum(nil)); sig != want {
		t.Fatalf("canonical sig mismatch got=%s want=%s", sig, want)
	}
}
