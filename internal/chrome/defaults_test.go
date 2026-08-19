package chrome

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultDetectionAndStoragePaths(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CHROMEPROBE_CHROME", "")
	if _, err := DetectExecutable(); err != nil {
		// CI may not have Chrome installed; the call still exercises platform
		// candidate discovery and returns a useful diagnostic.
		t.Logf("Chrome unavailable in test environment: %v", err)
	}
	t.Setenv("CHROMEPROBE_DATA_DIR", "")
	if _, err := DataDir(); err != nil {
		t.Fatal(err)
	}
	if _, err := ManagedProfileDir(); err != nil {
		t.Fatal(err)
	}
	if _, err := SessionsDir(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CHROMEPROBE_NETLOG_DIR", "")
	if _, err := NetLogsDir(); err != nil {
		t.Fatal(err)
	}
}
func TestStorageOverridesRejectFiles(t *testing.T) {
	dataFile := filepath.Join(t.TempDir(), "data-file")
	if err := os.WriteFile(dataFile, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CHROMEPROBE_DATA_DIR", dataFile)
	if _, err := DataDir(); err == nil {
		t.Fatal("DataDir accepted a file")
	}
	if _, err := ManagedProfileDir(); err == nil {
		t.Fatal("ManagedProfileDir accepted a file")
	}
	if _, err := SessionsDir(); err == nil {
		t.Fatal("SessionsDir accepted a file")
	}
	netlogsFile := filepath.Join(t.TempDir(), "netlogs-file")
	if err := os.WriteFile(netlogsFile, []byte("netlogs"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CHROMEPROBE_NETLOG_DIR", netlogsFile)
	if _, err := NetLogsDir(); err == nil {
		t.Fatal("NetLogsDir accepted a file")
	}
}
