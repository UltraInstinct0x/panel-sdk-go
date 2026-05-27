package panelsdk

import (
	"net/http"
	"strings"
	"time"
)

const DefaultEngineVersion = "0.2.0"

type Options struct {
	BaseURL          string
	SiteKey          string
	SiteSecret       string
	SiteSecretSource string
	ScrubberMode     string
	ScrubberSecret   string
	ScrubberURL      string
	EngineVersion    string
	TimeoutSeconds   float64
	MaxRetries       int
	HTTPClient       *http.Client
}

type Client struct {
	panel *PanelClient
	Rater *RaterClient
}

func New(o Options) *Client {
	o = normalizeOptions(o)
	hc := o.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: durationFromSeconds(o.TimeoutSeconds)}
	}
	hc.Transport = newRetryTransport(hc.Transport, o.MaxRetries)

	panel := &PanelClient{o: o, http: hc}
	rater := &RaterClient{o: o, http: hc}
	return &Client{panel: panel, Rater: rater}
}

func normalizeOptions(o Options) Options {
	o.BaseURL = strings.TrimRight(o.BaseURL, "/")
	if o.EngineVersion == "" {
		o.EngineVersion = DefaultEngineVersion
	}
	if o.TimeoutSeconds <= 0 {
		o.TimeoutSeconds = 10.0
	}
	if o.MaxRetries < 0 {
		o.MaxRetries = 0
	}
	if o.SiteSecretSource == "" {
		o.SiteSecretSource = "env"
	}
	if o.ScrubberMode == "" {
		o.ScrubberMode = "off"
	}
	return o
}

func durationFromSeconds(v float64) time.Duration {
	return time.Duration(v * float64(time.Second))
}
