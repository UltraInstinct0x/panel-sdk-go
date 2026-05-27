package panelsdk

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"time"
)

type retryTransport struct {
	base       http.RoundTripper
	maxRetries int
}

func newRetryTransport(base http.RoundTripper, maxRetries int) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &retryTransport{base: base, maxRetries: maxRetries}
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.maxRetries <= 0 {
		return t.base.RoundTrip(req)
	}

	var bodyBytes []byte
	if req.Body != nil && req.GetBody == nil {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		_ = req.Body.Close()
		bodyBytes = b
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(bodyBytes)), nil
		}
	}

	for attempt := 0; ; attempt++ {
		resp, err := t.base.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		if !shouldRetry(resp.StatusCode) || attempt >= t.maxRetries {
			return resp, nil
		}

		delay := retryDelay(resp, attempt)
		_ = resp.Body.Close()
		time.Sleep(delay)

		if req.GetBody != nil {
			rc, err := req.GetBody()
			if err != nil {
				return nil, err
			}
			req.Body = rc
		}
	}
}

func shouldRetry(status int) bool {
	if status == http.StatusTooManyRequests {
		return true
	}
	return status >= 500 && status <= 599
}

func retryDelay(resp *http.Response, attempt int) time.Duration {
	if resp.StatusCode == http.StatusTooManyRequests {
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if s, err := strconv.Atoi(ra); err == nil {
				if s > 30 {
					s = 30
				}
				if s < 0 {
					s = 0
				}
				return time.Duration(s) * time.Second
			}
		}
		return 500 * time.Millisecond
	}

	backoff := []time.Duration{500 * time.Millisecond, 1500 * time.Millisecond, 3500 * time.Millisecond}
	if attempt < len(backoff) {
		return backoff[attempt]
	}
	return backoff[len(backoff)-1]
}
