# panel-sdk-go

Go 1.22+ SDK for [panel](https://github.com/UltraInstinct0x/panel). Standard library only.

```bash
go get github.com/UltraInstinct0x/panel-sdk-go
```

## Options

```go
type Options struct {
    BaseURL          string
    SiteKey          string
    SiteSecret       string
    SiteSecretSource string // "env" (default) or "raw"
    ScrubberMode     string // "off" (default), "self-sign", "proxy"
    ScrubberSecret   string // required for self-sign
    ScrubberURL      string // required for proxy mode
    EngineVersion    string // default: "0.2.0"
    TimeoutSeconds   float64 // default: 10.0
    MaxRetries       int     // default: 0
}
```

`SiteSecretSource="raw"` adds `x-panel-ingest-secret` for dual-secret server mode.

## Clients

`New(o Options) *Client` remains the entry point and exposes:

- `Client` operator methods:
  - `IngestTrace(ctx, in)`
  - `IngestTraceAndWait(ctx, in, maxWaitSeconds, pollIntervalSeconds)`
  - `FetchTrace(ctx, traceID)`
  - `IngestUnits(ctx, payload)`
  - `ScoreUnit(ctx, ref, id)`
  - `SkillReview(ctx, input)`
  - `VerifyToken(ctx, token)`
- `Client.Rater` rater methods:
  - `NextUnit(ctx, pool, raterID)`
  - `SubmitJudgment(ctx, input)`

All public methods take `context.Context` first.

## Example

```go
ctx := context.Background()

c := panelsdk.New(panelsdk.Options{
    BaseURL:      "https://panel.example.com",
    SiteKey:      os.Getenv("PANEL_SITE_KEY"),
    SiteSecret:   os.Getenv("PANEL_SITE_SECRET"),
    ScrubberMode: "self-sign",
    ScrubberSecret: os.Getenv("SCRUBBER_JWT_SECRET"),
    MaxRetries:   3,
})

trace, err := c.IngestTrace(ctx, panelsdk.IngestTraceInput{
    SourceAgent: "myapp",
    Blob: map[string]interface{}{"messages": []string{"hello"}},
})
if err != nil {
    var rl *panelsdk.RateLimitError
    if errors.As(err, &rl) {
        // rl.Scope, rl.RetryAfter
    }
}

if trace.Status == "pending" {
    done, err := c.IngestTraceAndWait(ctx, panelsdk.IngestTraceInput{
        SourceAgent: "myapp",
        Blob: map[string]interface{}{"messages": []string{"hello"}},
    }, 60, 1.5)
    _ = done
    _ = err
}
```
