package store

import (
	"fmt"
	"strings"
	"time"
)

// ParseQueryDateStart parses YYYY-MM-DD or RFC3339 as start of day UTC.
func ParseQueryDateStart(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		utc := t.UTC()
		return &utc, nil
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.UTC)
	if err != nil {
		return nil, fmt.Errorf("invalid date %q: use YYYY-MM-DD or RFC3339", s)
	}
	return &t, nil
}

// ParseQueryDateEnd parses YYYY-MM-DD or RFC3339 as end of day UTC.
func ParseQueryDateEnd(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		utc := t.UTC()
		return &utc, nil
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.UTC)
	if err != nil {
		return nil, fmt.Errorf("invalid date %q: use YYYY-MM-DD or RFC3339", s)
	}
	end := t.Add(24*time.Hour - time.Nanosecond)
	return &end, nil
}
