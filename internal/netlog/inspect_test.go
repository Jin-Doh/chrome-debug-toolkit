package netlog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInspectCountsEventsAndErrorsWithoutWholeFileModel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "netlog.json")
	data := `{
  "constants": {"logEventTypes": {"REQUEST_ALIVE": 7, "URL_REQUEST_START_JOB": 8}},
  "events": [
    {"type": 7, "params": {"net_error": -105}},
    {"type": 7, "params": {"net_error": "-105"}},
    {"type": 8, "params": {}},
    {"type": 999, "params": {"net_error": -101}}
  ]
}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}

	summary, err := Inspect(path)
	if err != nil {
		t.Fatal(err)
	}
	if summary.TotalEvents != 4 {
		t.Fatalf("TotalEvents = %d, want 4", summary.TotalEvents)
	}
	if summary.EventCounts["REQUEST_ALIVE"] != 2 || summary.EventCounts["URL_REQUEST_START_JOB"] != 1 || summary.EventCounts["999"] != 1 {
		t.Fatalf("unexpected event counts: %#v", summary.EventCounts)
	}
	if summary.UnknownEvents != 1 {
		t.Fatalf("UnknownEvents = %d, want 1", summary.UnknownEvents)
	}
	if summary.ErrorCounts[-105] != 2 || summary.ErrorCounts[-101] != 1 {
		t.Fatalf("unexpected error counts: %#v", summary.ErrorCounts)
	}
}

func TestInspectRejectsNonObject(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invalid.json")
	if err := os.WriteFile(path, []byte("[]"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Inspect(path); err == nil {
		t.Fatal("Inspect accepted a non-object NetLog")
	}
}
