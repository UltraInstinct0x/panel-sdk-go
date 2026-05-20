package panelsdk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIngestUnitSignsBody(t *testing.T) {
	var sig, key, attest string
	var bodyBytes []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ = io.ReadAll(r.Body)
		sig = r.Header.Get("x-panel-ingest-sig")
		key = r.Header.Get("x-panel-site-key")
		attest = r.Header.Get("x-scrubber-attestation")
		_, _ = w.Write([]byte(`{"id":"u_1"}`))
	}))
	defer ts.Close()
	c := New(Options{BaseURL: ts.URL, SiteKey: "pk_test", SiteSecret: "sek"})
	out, err := c.IngestUnit(IngestUnitInput{Type: "step_validity", Payload: map[string]interface{}{"foo": 1}})
	if err != nil {
		t.Fatal(err)
	}
	if out["id"] != "u_1" {
		t.Fatalf("unexpected: %#v", out)
	}
	if key != "pk_test" {
		t.Fatalf("site key: %q", key)
	}
	if attest != "" {
		t.Fatalf("attestation should be empty when no scrubber secret: %q", attest)
	}
	m := hmac.New(sha256.New, []byte("sek"))
	m.Write(bodyBytes)
	if want := hex.EncodeToString(m.Sum(nil)); sig != want {
		t.Fatalf("sig %q != want %q", sig, want)
	}
}

func TestIngestTraceAttachesAttestation(t *testing.T) {
	var attest string
	var bodyBytes []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ = io.ReadAll(r.Body)
		attest = r.Header.Get("x-scrubber-attestation")
		_, _ = w.Write([]byte(`{"trace_id":"tr_1","units_emitted":3}`))
	}))
	defer ts.Close()
	c := New(Options{BaseURL: ts.URL, SiteKey: "pk_test", SiteSecret: "sek", ScrubberSecret: "scrub"})
	_, err := c.IngestTrace(IngestTraceInput{TraceID: "tr_1", SourceAgent: "hermes", Blob: map[string]interface{}{"messages": []interface{}{}}})
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(attest, ".")
	if len(parts) != 3 {
		t.Fatalf("attest parts: %d", len(parts))
	}
	pb, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]interface{}
	_ = json.Unmarshal(pb, &payload)
	want := sha256Hex(bodyBytes)
	if payload["output_hash"] != want {
		t.Fatalf("output_hash %v != %s", payload["output_hash"], want)
	}
}

func TestVerifyTokenParses(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true,"trust":0.8,"tier_used":"C1","unit_ids":["u_a"]}`))
	}))
	defer ts.Close()
	c := New(Options{BaseURL: ts.URL, SiteKey: "pk_test", SiteSecret: "sek"})
	v, err := c.VerifyToken("t.t.t")
	if err != nil {
		t.Fatal(err)
	}
	if !v.OK || v.TierUsed != "C1" || v.Trust == nil || *v.Trust != 0.8 {
		t.Fatalf("unexpected: %#v", v)
	}
}

func TestNon2xxReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		_, _ = w.Write([]byte(`{"error":"scrubber_attestation_required"}`))
	}))
	defer ts.Close()
	c := New(Options{BaseURL: ts.URL, SiteKey: "pk_test", SiteSecret: "sek"})
	_, err := c.IngestUnit(IngestUnitInput{Type: "x", Payload: map[string]interface{}{}})
	if err == nil {
		t.Fatal("expected error")
	}
	pe, ok := err.(*Error)
	if !ok || pe.Status != 422 {
		t.Fatalf("not panel Error 422: %#v", err)
	}
}

func TestFetchTraceUsesGET(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("method %s", r.Method)
		}
		if r.URL.Path != "/api/v1/traces/tr_y" {
			t.Fatalf("path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"trace_id":"tr_y"}`))
	}))
	defer ts.Close()
	c := New(Options{BaseURL: ts.URL + "/", SiteKey: "pk_test", SiteSecret: "sek"})
	out, err := c.FetchTrace("tr_y")
	if err != nil {
		t.Fatal(err)
	}
	if out["trace_id"] != "tr_y" {
		t.Fatalf("unexpected: %#v", out)
	}
}
