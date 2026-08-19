package netlog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestStartAndListSessionsPersistMetadata(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	netlogsDir := filepath.Join(t.TempDir(), "netlogs")
	t.Setenv("CHROMEPROBE_DATA_DIR", dataDir)
	t.Setenv("CHROMEPROBE_NETLOG_DIR", netlogsDir)
	t.Setenv("CHROMEPROBE_CHROME", writeFakeChrome(t))

	session, err := Start("https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	if session.ID == "" || session.PID <= 0 || session.Status != "running" {
		t.Fatalf("unexpected started session: %#v", session)
	}
	if _, err := os.Stat(filepath.Join(session.SessionDir, "session.json")); err != nil {
		t.Fatal(err)
	}

	var sessions []*Session
	sessions, err = ListSessions()
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].ID != session.ID {
		t.Fatalf("ListSessions() = %#v", sessions)
	}
	if sessions[0].Status != "running" && sessions[0].Status != "exited" {
		t.Fatalf("unexpected session status = %q", sessions[0].Status)
	}
}
func TestListSessionsMarksDeadSessionExited(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	t.Setenv("CHROMEPROBE_DATA_DIR", dataDir)
	root := filepath.Join(dataDir, "sessions")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	session := &Session{
		ID: "dead", PID: 999999999, Status: "running",
		SessionDir: filepath.Join(root, "dead"),
	}
	if err := os.MkdirAll(session.SessionDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeSession(session); err != nil {
		t.Fatal(err)
	}
	sessions, err := ListSessions()
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].Status != "exited" {
		t.Fatalf("ListSessions() = %#v", sessions)
	}
}

func TestStartRemovesSessionOnLaunchFailure(t *testing.T) {
	t.Setenv("CHROMEPROBE_DATA_DIR", filepath.Join(t.TempDir(), "data"))
	t.Setenv("CHROMEPROBE_NETLOG_DIR", filepath.Join(t.TempDir(), "netlogs"))
	t.Setenv("CHROMEPROBE_CHROME", filepath.Join(t.TempDir(), "missing"))
	if _, err := Start(""); err == nil {
		t.Fatal("Start succeeded with missing Chrome")
	}
	root, err := os.ReadDir(filepath.Join(os.Getenv("CHROMEPROBE_DATA_DIR"), "sessions"))
	if err != nil {
		t.Fatal(err)
	}
	if len(root) != 0 {
		t.Fatalf("failed session directories remain: %d", len(root))
	}
}

func TestWriteSessionRoundTrips(t *testing.T) {
	dir := t.TempDir()
	session := &Session{ID: "id", SessionDir: dir, PID: 42, Status: "running"}
	if err := writeSession(session); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "session.json"))
	if err != nil {
		t.Fatal(err)
	}
	var decoded Session
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID != session.ID || decoded.PID != session.PID || decoded.Status != session.Status {
		t.Fatalf("decoded session = %#v", decoded)
	}
}

func writeFakeChrome(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fake-chrome")
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}
