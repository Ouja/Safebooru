package safebooru

import (
	"errors"
	"fmt"
	"strings"
)

// ErrSearchDown matches an *APIError whose message says the search is down.
// Check it with errors.Is(err, ErrSearchDown).
var ErrSearchDown = errors.New("safebooru: search is down")

// ErrResponseTooLarge is returned when a response body is bigger than the
// client is willing to read (10 MB).
var ErrResponseTooLarge = errors.New("safebooru: response too large")

// APIError is returned when the API answers with HTTP 200 but reports a
// failure, e.g. {"success": false, "message": "search down"} or the XML
// <response success="false" reason="..."/>. Use errors.As to read Message.
type APIError struct {
	Message string
}

func (e *APIError) Error() string {
	if e.Message == "" {
		return "safebooru: api error"
	}
	return "safebooru: api error: " + e.Message
}

// Is makes errors.Is(err, ErrSearchDown) true when the API says its
// search is down.
func (e *APIError) Is(target error) bool {
	return target == ErrSearchDown && strings.Contains(strings.ToLower(e.Message), "search down")
}

// StatusError is returned when the API responds with a non-200 status code.
// Use errors.As to inspect StatusCode (e.g. 429 = rate limited).
type StatusError struct {
	StatusCode int
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("safebooru: unexpected status %d", e.StatusCode)
}
