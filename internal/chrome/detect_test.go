package chrome

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDetectExecutableUsesExecutableOverride(t *testing.T) {
	path := filepath.Join(t.TempDir(), "chrome")
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CHROMEPROBE_CHROME", path)
	got, err := DetectExecutable()
	if err != nil {
		t.Fatal(err)
	}
	if got != path {
		t.Fatalf("DetectExecutable() = %q, want %q", got, path)
	}
}

func TestDetectExecutableRejectsInvalidOverride(t *testing.T) {
	t.Setenv("CHROMEPROBE_CHROME", filepath.Join(t.TempDir(), "missing"))
	if _, err := DetectExecutable(); err == nil {
		t.Fatal("DetectExecutable accepted a missing override")
	}
}

func TestIsExecutableRejectsDirectoriesAndMissingPaths(t *testing.T) {
	dir := t.TempDir()
	if isExecutable(dir) {
		t.Fatal("isExecutable accepted a directory")
	}
	if isExecutable(filepath.Join(dir, "missing")) {
		t.Fatal("isExecutable accepted a missing path")
	}
}
func TestDetectExecutableFindsLinuxPathCandidate(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux PATH candidate branch")
	}
	bin := t.TempDir()
	path := filepath.Join(bin, "google-chrome")
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CHROMEPROBE_CHROME", "")
	t.Setenv("PATH", bin)
	got, err := DetectExecutable()
	if err != nil {
		t.Fatal(err)
	}
	if got != path {
		t.Fatalf("DetectExecutable() = %q, want %q", got, path)
	}
}
