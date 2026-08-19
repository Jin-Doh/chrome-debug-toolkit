package report

import (
	"errors"
	"strings"
	"testing"

	"github.com/jin-doh/chrome-debug-toolkit/internal/netlog"
)

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("writer failed") }

func TestPrintSummaryFormatsEventsAndErrors(t *testing.T) {
	var output strings.Builder
	err := PrintSummary(&output, &netlog.Summary{
		Path:          "/tmp/netlog.json",
		TotalEvents:   3,
		UnknownEvents: 1,
		EventCounts:   map[string]int{"B": 1, "A": 2},
		ErrorCounts:   map[int]int{-101: 1, -105: 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	text := output.String()
	for _, want := range []string{"NetLog summary", "Total events: 3", "Unknown:      1", "A", "-105"} {
		if !strings.Contains(text, want) {
			t.Fatalf("summary missing %q: %s", want, text)
		}
	}
}

func TestPrintSummaryHandlesNilAndWriterErrors(t *testing.T) {
	if err := PrintSummary(&strings.Builder{}, nil); err != nil {
		t.Fatal(err)
	}
	if err := PrintSummary(failingWriter{}, &netlog.Summary{}); err == nil {
		t.Fatal("PrintSummary accepted a failing writer")
	}
}

type failAfterWriter struct {
	writes int
	limit  int
}

func (w *failAfterWriter) Write(data []byte) (int, error) {
	if w.writes >= w.limit {
		return 0, errors.New("writer failed")
	}
	w.writes++
	return len(data), nil
}

func TestPrintSummaryPropagatesEveryWriteFailurePoint(t *testing.T) {
	summary := &netlog.Summary{
		TotalEvents:   2,
		UnknownEvents: 1,
		EventCounts:   map[string]int{"ONE": 1},
		ErrorCounts:   map[int]int{-1: 1},
	}
	for limit := range 8 {
		if err := PrintSummary(&failAfterWriter{limit: limit}, summary); err == nil {
			t.Fatalf("PrintSummary limit %d returned nil", limit)
		}
	}
}
