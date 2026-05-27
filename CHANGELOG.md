# CHANGELOG

## v0.2.0 — server sync

- Refactored SDK into focused files:
  - `panelsdk.go` (options and constructor)
  - `client.go` (`PanelClient` methods)
  - `rater.go` (`RaterClient` methods)
  - `signing.go` (HMAC/JWT/canonical helpers)
  - `scrubber.go` (scrubber mode dispatch)
  - `retry.go` (429/5xx retry round-tripper)
  - `errors.go` (typed errors)
  - `types.go` (request/response structs)
- Added full server-sync panel surface:
  - `IngestTrace`, `FetchTrace`, `IngestUnits`, `ScoreUnit`, `SkillReview`, `VerifyToken`
  - `IngestTraceAndWait` convenience poller
- Added `RaterClient` methods:
  - `NextUnit(pool, raterID)`
  - `SubmitJudgment(unitID, choice)`
- Added option parity with server sync spec:
  - `SiteSecretSource`, `ScrubberMode`, `ScrubberURL`, `TimeoutSeconds`, `MaxRetries`, `EngineVersion`
- Added dual-secret auth header support (`x-panel-ingest-secret`) when `SiteSecretSource="raw"`.
- Added scrubber proxy mode (`scrubber_mode="proxy"`) and self-sign mode (`"self-sign"`).
- Added typed rate limit error with exported fields:
  - `RateLimitError.Scope`
  - `RateLimitError.RetryAfter`
- Added `httptest` unit tests for:
  1. body HMAC signatures
  2. self-signed scrubber JWT payload shape
  3. async trace polling to done
  4. 429 retry handling
  5. canonical score signature

### Compatibility note

- `New(o Options) *Client` remains the constructor entrypoint.
