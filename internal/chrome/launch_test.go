package chrome

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLaunchValidatesRequiredPaths(t *testing.T) {
	t.Setenv("CHROMEPROBE_CHROME", filepath.Join(t.TempDir(), "missing"))
	if _, err := Launch(LaunchConfig{}); err == nil || !strings.Contains(err.Error(), "not executable") {
		t.Fatalf("Launch() error = %v", err)
	}

	executable := writeFakeChrome(t, "")
	t.Setenv("CHROMEPROBE_CHROME", executable)
	if _, err := Launch(LaunchConfig{NetLogPath: filepath.Join(t.TempDir(), "netlog.json")}); err == nil || !strings.Contains(err.Error(), "profile directory") {
		t.Fatalf("missing profile error = %v", err)
	}
	if _, err := Launch(LaunchConfig{ProfileDir: t.TempDir()}); err == nil || !strings.Contains(err.Error(), "NetLog path") {
		t.Fatalf("missing NetLog error = %v", err)
	}
}

func TestLaunchStartsIsolatedProcessWithExpectedArguments(t *testing.T) {
	argsPath := filepath.Join(t.TempDir(), "args")
	executable := writeFakeChrome(t, fmt.Sprintf("printf '%%s\\\\n' \"$@\" > %s", argsPath))
	t.Setenv("CHROMEPROBE_CHROME", executable)
	profile := filepath.Join(t.TempDir(), "profile")
	netlog := filepath.Join(t.TempDir(), "nested", "netlog.json")

	result, err := Launch(LaunchConfig{
		ProfileDir: profile,
		NetLogPath: netlog,
		URL:        "https://example.com",
		ExtraArgs:  []string{"--headless=new"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.PID <= 0 {
		t.Fatalf("Launch() PID = %d", result.PID)
	}
	var data []byte
	for deadline := time.Now().Add(3 * time.Second); time.Now().Before(deadline); {
		data, _ = os.ReadFile(argsPath)
		if len(data) > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	args := string(data)
	for _, want := range []string{
		"--user-data-dir=" + profile,
		"--log-net-log=" + netlog,
		"--net-log-capture-mode=Everything",
		"--headless=new",
		"https://example.com",
	} {
		if !strings.Contains(args, want) {
			t.Fatalf("launch arguments missing %q: %s", want, args)
		}
	}
}

func writeFakeChrome(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fake-chrome")
	content := "#!/bin/sh\n" + body + "\n"
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}
