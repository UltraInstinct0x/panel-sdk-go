# panel-sdk-go

thin client for [panel](https://github.com/UltraInstinct0x/panel). go 1.22+. stdlib only.

```
go get github.com/UltraInstinct0x/panel-sdk-go
```

```go
import panelsdk "github.com/UltraInstinct0x/panel-sdk-go"

c := panelsdk.New(panelsdk.Options{
    BaseURL:        "https://panel.example.com",
    SiteKey:        os.Getenv("PANEL_SITE_KEY"),
    SiteSecret:     os.Getenv("PANEL_SITE_SECRET"),
    ScrubberSecret: os.Getenv("SCRUBBER_JWT_SECRET"), // omit for first-party keys
})

v, err := c.VerifyToken(token)
if err != nil || !v.OK || (v.Trust == nil || *v.Trust < 0.5) {
    http.Error(w, "forbidden", 403); return
}

_, err = c.IngestTrace(panelsdk.IngestTraceInput{
    TraceID: "tr_" + uuid.NewString(), SourceAgent: "myapp",
    Blob: map[string]interface{}{"messages": msgs},
})
```

methods: `IngestUnit`, `IngestTrace`, `VerifyToken`, `FetchUnit`, `FetchTrace`.
auth: HMAC-SHA256 (`x-panel-ingest-sig`) + optional scrubber JWT.
zero external deps.
