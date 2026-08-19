package netlog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListSessionsRejectsMalformedMetadata(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	t.Setenv("CHROMEPROBE_DATA_DIR", dataDir)
	root := filepath.Join(dataDir, "sessions", "broken")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "session.json"), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ListSessions(); err == nil {
		t.Fatal("ListSessions accepted malformed metadata")
	}
}

func TestListSessionsIgnoresDirectoriesWithoutMetadata(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	t.Setenv("CHROMEPROBE_DATA_DIR", dataDir)
	root := filepath.Join(dataDir, "sessions", "incomplete")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	sessions, err := ListSessions()
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 0 {
		t.Fatalf("ListSessions() = %#v", sessions)
	}
}
func TestStartPropagatesStorageConfigurationErrors(t *testing.T) {
	dataFile := filepath.Join(t.TempDir(), "data-file")
	if err := os.WriteFile(dataFile, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CHROMEPROBE_DATA_DIR", dataFile)
	t.Setenv("CHROMEPROBE_NETLOG_DIR", filepath.Join(t.TempDir(), "netlogs"))
	if _, err := Start(""); err == nil {
		t.Fatal("Start accepted a file as data directory")
	}

	dataDir := filepath.Join(t.TempDir(), "data")
	netlogsFile := filepath.Join(t.TempDir(), "netlogs-file")
	if err := os.WriteFile(netlogsFile, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CHROMEPROBE_DATA_DIR", dataDir)
	t.Setenv("CHROMEPROBE_NETLOG_DIR", netlogsFile)
	if _, err := Start(""); err == nil {
		t.Fatal("Start accepted a file as NetLog directory")
	}
}
