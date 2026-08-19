package netlog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRotateRemovesExpiredSessionAndNetLog(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	netlogsDir := filepath.Join(t.TempDir(), "netlogs")
	if err := os.MkdirAll(netlogsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CHROMEPROBE_NETLOG_DIR", netlogsDir)
	t.Setenv("CHROMEPROBE_DATA_DIR", dataDir)
	root := filepath.Join(dataDir, "sessions")
	oldDir := filepath.Join(root, "old")
	recentDir := filepath.Join(root, "recent")
	if err := os.MkdirAll(oldDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(recentDir, 0o755); err != nil {
		t.Fatal(err)
	}
	oldNetLog := filepath.Join(netlogsDir, "old.json")
	if err := os.WriteFile(oldNetLog, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	metadata, _ := json.Marshal(&Session{ID: "old", NetLogPath: oldNetLog})
	if err := os.WriteFile(filepath.Join(oldDir, "session.json"), metadata, 0o600); err != nil {
		t.Fatal(err)
	}
	oldTime := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(oldDir, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	removed, err := Rotate(24 * time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if removed != 1 {
		t.Fatalf("Rotate() removed %d sessions, want 1", removed)
	}
	if _, err := os.Stat(oldDir); !os.IsNotExist(err) {
		t.Fatalf("old session still exists: %v", err)
	}
	if _, err := os.Stat(oldNetLog); !os.IsNotExist(err) {
		t.Fatalf("old NetLog still exists: %v", err)
	}
	if _, err := os.Stat(recentDir); err != nil {
		t.Fatalf("recent session was removed: %v", err)
	}
}
func TestRotateRejectsNegativeRetention(t *testing.T) {
	if _, err := Rotate(-time.Second); err == nil {
		t.Fatal("Rotate accepted negative retention")
	}
}

func TestIsWithinDirRejectsEmptyPath(t *testing.T) {
	if isWithinDir("", t.TempDir()) {
		t.Fatal("empty path accepted")
	}
}
func TestRotatePropagatesStorageErrors(t *testing.T) {
	dataFile := filepath.Join(t.TempDir(), "data-file")
	if err := os.WriteFile(dataFile, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CHROMEPROBE_DATA_DIR", dataFile)
	if _, err := Rotate(0); err == nil {
		t.Fatal("Rotate accepted a file as data directory")
	}
}

func TestIsWithinDirRejectsOutsideAndSamePath(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "netlogs")
	inside := filepath.Join(dir, "capture.json")
	outside := filepath.Join(t.TempDir(), "capture.json")
	if !isWithinDir(inside, dir) {
		t.Fatal("inside path rejected")
	}
	if isWithinDir(outside, dir) {
		t.Fatal("outside path accepted")
	}
	if isWithinDir(dir, dir) {
		t.Fatal("directory itself accepted as a file")
	}
}
