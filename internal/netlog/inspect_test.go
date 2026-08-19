package netlog

import (
	"os"
	"path/filepath"
	"strconv"
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
func TestInspectResolvesSessionAndLatestTargets(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	netlogsDir := filepath.Join(t.TempDir(), "netlogs")
	sessionDir := filepath.Join(dataDir, "sessions", "session-id")
	t.Setenv("CHROMEPROBE_DATA_DIR", dataDir)
	t.Setenv("CHROMEPROBE_NETLOG_DIR", netlogsDir)
	if err := os.MkdirAll(netlogsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatal(err)
	}
	netlogPath := filepath.Join(netlogsDir, "capture.json")
	if err := os.WriteFile(netlogPath, []byte(`{"constants":{"logEventTypes":{"ONE":1}},"events":[{"type":1,"params":{}}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	session := &Session{ID: "session-id", SessionDir: sessionDir, NetLogPath: netlogPath, Status: "exited"}
	if err := writeSession(session); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"session-id", "latest", ""} {
		summary, err := Inspect(target)
		if err != nil {
			t.Fatalf("Inspect(%q): %v", target, err)
		}
		if summary.TotalEvents != 1 {
			t.Fatalf("Inspect(%q) total = %d, want 1", target, summary.TotalEvents)
		}
	}
}

func TestInspectRejectsMissingTargetAndMalformedEvents(t *testing.T) {
	if _, err := Inspect(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("Inspect accepted a missing target")
	}
	path := filepath.Join(t.TempDir(), "malformed.json")
	if err := os.WriteFile(path, []byte(`{"events":[`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Inspect(path); err == nil {
		t.Fatal("Inspect accepted malformed events")
	}
	for index, data := range []string{
		`{"constants":[]}`,
		`{"events":{}}`,
		`{"unknown":`,
		`{"events":[]`,
		`{"events":[{"type":}]}`,
	} {
		path := filepath.Join(t.TempDir(), "malformed-"+strconv.Itoa(index)+".json")
		if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := Inspect(path); err == nil {
			t.Fatalf("Inspect accepted malformed input %q", data)
		}
	}
}
func TestInspectRejectsMalformedSessionMetadata(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "session-data")
	t.Setenv("CHROMEPROBE_DATA_DIR", dataDir)
	brokenDir := filepath.Join(dataDir, "sessions", "broken")
	if err := os.MkdirAll(brokenDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(brokenDir, "session.json"), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Inspect("broken"); err == nil {
		t.Fatal("Inspect accepted malformed session metadata")
	}
}
func TestInspectIgnoresNonNumericNetError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "non-numeric-error.json")
	data := `{"constants":{"logEventTypes":{"ONE":1}},"events":[{"type":1,"params":{"net_error":"not-a-number"}}]}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	summary, err := Inspect(path)
	if err != nil {
		t.Fatal(err)
	}
	if summary.TotalEvents != 1 || len(summary.ErrorCounts) != 0 {
		t.Fatalf("unexpected summary: %#v", summary)
	}
}
