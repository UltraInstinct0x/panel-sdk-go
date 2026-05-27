package panelsdk

import "fmt"

type Error struct {
	Status int
	Body   []byte
}

func (e *Error) Error() string {
	return fmt.Sprintf("panel %d: %s", e.Status, truncate(string(e.Body), 300))
}

type RateLimitError struct {
	Scope      string
	RetryAfter int
	Status     int
	Body       []byte
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("panel rate limited (scope=%s retry_after_s=%d)", e.Scope, e.RetryAfter)
}

type ScrubberError struct {
	Status int
	Body   []byte
}

func (e *ScrubberError) Error() string {
	return fmt.Sprintf("scrubber %d: %s", e.Status, truncate(string(e.Body), 300))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
