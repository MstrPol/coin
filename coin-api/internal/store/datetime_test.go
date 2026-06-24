package store

import (
	"testing"
	"time"
)

func TestParseQueryDateStart(t *testing.T) {
	start, err := ParseQueryDateStart("2026-01-15")
	if err != nil {
		t.Fatal(err)
	}
	if start == nil || !start.Equal(time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("start=%v", start)
	}
	if _, err := ParseQueryDateStart(""); err != nil {
		t.Fatal(err)
	}
}

func TestParseQueryDateEnd(t *testing.T) {
	end, err := ParseQueryDateEnd("2026-01-15")
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 1, 15, 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)
	if end == nil || !end.Equal(want) {
		t.Fatalf("end=%v want %v", end, want)
	}
}

func TestParseQueryDateRFC3339(t *testing.T) {
	start, err := ParseQueryDateStart("2026-03-10T14:30:00Z")
	if err != nil || start == nil {
		t.Fatalf("rfc3339 start: %v err=%v", start, err)
	}
}
