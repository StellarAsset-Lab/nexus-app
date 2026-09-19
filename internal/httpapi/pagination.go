package httpapi

import (
	"net/http"
	"strconv"
)

const (
	defaultPageLimit = 50
	maxPageLimit     = 200
)

// parseLimit reads and bounds the `limit` query parameter. Never trusts an
// unbounded caller-supplied limit — spec §30 ("do not accept
// limit=1000000 and hope PostgreSQL handles it").
func parseLimit(r *http.Request) int {
	raw := r.URL.Query().Get("limit")
	if raw == "" {
		return defaultPageLimit
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return defaultPageLimit
	}
	if n > maxPageLimit {
		return maxPageLimit
	}
	return n
}

func parseCursor(r *http.Request) *string {
	raw := r.URL.Query().Get("cursor")
	if raw == "" {
		return nil
	}
	return &raw
}

// parseBoolFilter reads an optional boolean query parameter, returning nil
// when absent (meaning "no filter") rather than defaulting to false.
func parseBoolFilter(r *http.Request, key string) *bool {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return nil
	}
	return &v
}
