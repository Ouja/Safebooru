package safebooru

import (
	"errors"
	"fmt"
)

// ErrSearchDown is returned when the API answers with HTTP 200 but
// {"success": false} (for example when its search is down).
var ErrSearchDown = errors.New("safebooru: search is down")

// StatusError is returned when the API responds with a non-200 status code.
// Use errors.As to inspect StatusCode (e.g. 429 = rate limited).
type StatusError struct {
	StatusCode int
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("safebooru: unexpected status %d", e.StatusCode)
}
