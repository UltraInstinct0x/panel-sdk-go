## panel-sdk-go

Go 1.22+ SDK for panel operator endpoints.

```bash
go get github.com/UltraInstinct0x/panel-sdk-go
```

SDK version constant:

```go
panelsdk.Version // "0.2.0"
```

## Quick start

```go
ctx := context.Background()

c := panelsdk.New(panelsdk.Options{
	BaseURL:        "https://panel.example.com",
	SiteKey:        os.Getenv("PANEL_SITE_KEY"),
	SiteSecret:     os.Getenv("PANEL_INGEST_SECRET"),
	ScrubberSecret: os.Getenv("SCRUBBER_JWT_SECRET"),
})
```

## Verify

```go
vr, err := c.VerifyToken(ctx, token)
if err != nil {
	log.Fatal(err)
}
fmt.Println(vr.OK)
```

## Ingest a unit

```go
u, err := c.IngestUnit(ctx, panelsdk.IngestUnitInput{
	Type:    "process_output_rating",
	Payload: map[string]any{"passage": "model output"},
}, panelsdk.WithExternalRef("ext-123"))
if err != nil {
	log.Fatal(err)
}
fmt.Println(u.Accepted, u.IDs)
```

## Ingest a trace and fetch status

```go
tr, err := c.IngestTrace(ctx, panelsdk.IngestTraceInput{
	TraceID:     "tr_demo_1",
	SourceAgent: "sdk-demo",
	Blob:        map[string]any{"messages": []string{"hello"}},
})
if err != nil {
	log.Fatal(err)
}

status, err := c.GetTrace(ctx, tr.TraceID)
if err != nil {
	log.Fatal(err)
}
fmt.Println(status.Status)
```

## Submit judgment

```go
j, err := c.SubmitJudgment(ctx, panelsdk.JudgmentInput{
	UnitID:    "u_123",
	RaterID:   "operator-1",
	Choice:    "yes",
	LatencyMS: 3000,
})
if err != nil {
	log.Fatal(err)
}
fmt.Println(j.OK, j.Token)
```

## Scrub text

```go
c = panelsdk.New(panelsdk.Options{
	BaseURL:     "https://panel.example.com",
	ScrubberURL: "https://scrubber.example.com",
})

sr, err := c.Scrub(ctx, "email me at test@example.com")
if err != nil {
	log.Fatal(err)
}
fmt.Println(sr.Scrubbed)
```
