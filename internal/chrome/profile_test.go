package chrome

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoragePathsUseOverridesAndCreateDirectories(t *testing.T) {
	root := filepath.Join(t.TempDir(), "data")
	netlogs := filepath.Join(t.TempDir(), "netlogs")
	t.Setenv("CHROMEPROBE_DATA_DIR", root)
	t.Setenv("CHROMEPROBE_NETLOG_DIR", netlogs)

	data, err := DataDir()
	if err != nil {
		t.Fatal(err)
	}
	if data != root {
		t.Fatalf("DataDir() = %q, want %q", data, root)
	}
	profile, err := ManagedProfileDir()
	if err != nil {
		t.Fatal(err)
	}
	if profile != filepath.Join(root, "profiles", "netlog") {
		t.Fatalf("ManagedProfileDir() = %q", profile)
	}
	sessions, err := SessionsDir()
	if err != nil {
		t.Fatal(err)
	}
	if sessions != filepath.Join(root, "sessions") {
		t.Fatalf("SessionsDir() = %q", sessions)
	}
	gotNetlogs, err := NetLogsDir()
	if err != nil {
		t.Fatal(err)
	}
	if gotNetlogs != netlogs {
		t.Fatalf("NetLogsDir() = %q, want %q", gotNetlogs, netlogs)
	}
	for _, path := range []string{data, profile, sessions, gotNetlogs} {
		if info, err := os.Stat(path); err != nil || !info.IsDir() {
			t.Fatalf("storage path %q was not created", path)
		}
	}
}
