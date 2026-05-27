package panelsdk

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (c *PanelClient) scrubberMode(ctx context.Context, body []byte) ([]byte, string, error) {
	switch c.o.ScrubberMode {
	case "", "off":
		return body, "", nil
	case "self-sign":
		if c.o.ScrubberSecret == "" {
			return nil, "", fmt.Errorf("scrubber_mode self-sign requires scrubber_secret")
		}
		att, err := makeSelfSignedAttestation(c.o.ScrubberSecret, body, c.o.EngineVersion)
		if err != nil {
			return nil, "", fmt.Errorf("self-sign attestation: %w", err)
		}
		return body, att, nil
	case "proxy":
		if c.o.ScrubberURL == "" {
			return nil, "", fmt.Errorf("scrubber_mode proxy requires scrubber_url")
		}
		return c.proxyScrub(ctx, body)
	default:
		return nil, "", fmt.Errorf("invalid scrubber_mode: %s", c.o.ScrubberMode)
	}
}

func (c *PanelClient) proxyScrub(ctx context.Context, body []byte) ([]byte, string, error) {
	url := strings.TrimRight(c.o.ScrubberURL, "/") + "/scrub"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, "", fmt.Errorf("build scrubber request: %w", err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-panel-site-key", c.o.SiteKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("call scrubber proxy: %w", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read scrubber response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", &ScrubberError{Status: resp.StatusCode, Body: b}
	}
	att := resp.Header.Get("x-scrubber-attestation")
	if att == "" {
		return nil, "", fmt.Errorf("scrubber proxy missing x-scrubber-attestation")
	}
	if len(b) == 0 {
		return body, att, nil
	}
	return b, att, nil
}
