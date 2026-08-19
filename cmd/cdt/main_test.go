package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/jin-doh/chrome-debug-toolkit/internal/chrome"
)

func TestCommandArgumentValidation(t *testing.T) {
	if err := runNetLog([]string{"one", "two"}); err == nil {
		t.Fatal("runNetLog accepted two URLs")
	}
	if err := runInspect([]string{"one", "two"}); err == nil {
		t.Fatal("runInspect accepted two targets")
	}
	if err := runInspect([]string{"missing"}); err == nil {
		t.Fatal("runInspect accepted a missing target")
	}
	if err := runKill([]string{"--unknown"}); err == nil {
		t.Fatal("runKill accepted an unknown option")
	}
	if err := runClean([]string{"not-a-number"}); err == nil {
		t.Fatal("runClean accepted invalid days")
	}
	if err := runClean([]string{"7", "8"}); err == nil {
		t.Fatal("runClean accepted multiple retention values")
	}
	if err := runClean([]string{"-1"}); err == nil {
		t.Fatal("runClean accepted negative days")
	}
}

func TestExecuteDispatchesMetaCommands(t *testing.T) {
	if err := execute([]string{}); err == nil {
		t.Fatal("execute accepted no command")
	}
	for _, command := range []string{"version", "--version", "-v", "help", "--help", "-h"} {
		if err := execute([]string{command}); err != nil {
			t.Fatalf("execute(%q): %v", command, err)
		}
	}
	if err := execute([]string{"unknown"}); err == nil {
		t.Fatal("execute accepted an unknown command")
	}
}

func TestExecuteDispatchesRuntimeCommands(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	t.Setenv("CHROMEPROBE_DATA_DIR", dataDir)
	t.Setenv("CHROMEPROBE_NETLOG_DIR", filepath.Join(t.TempDir(), "netlogs"))
	t.Setenv("CHROMEPROBE_CHROME", writeFakeChrome(t))
	profile, err := chrome.ManagedProfileDir()
	if err != nil {
		t.Fatal(err)
	}
	fakeChrome := filepath.Join(t.TempDir(), "Google Chrome")
	if err := os.Symlink("/bin/sleep", fakeChrome); err != nil {
		t.Fatal(err)
	}
	child := exec.Command(fakeChrome, "5", "--user-data-dir="+profile)
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = child.Process.Kill() }()

	direct := filepath.Join(t.TempDir(), "netlog.json")
	if err := os.WriteFile(direct, []byte(`{"constants":{"logEventTypes":{"ONE":1}},"events":[{"type":1,"params":{}}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	commands := [][]string{
		{"ps"}, {"kill", "--force"}, {"sessions"}, {"netlog", "about:blank"},
		{"inspect", direct}, {"doctor"}, {"clean", "7"},
	}
	for _, command := range commands {
		if err := execute(command); err != nil {
			t.Fatalf("execute(%v): %v", command, err)
		}
	}
	_ = child.Wait()
}

func TestCommandsRunAgainstIsolatedStorage(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	netlogsDir := filepath.Join(t.TempDir(), "netlogs")
	t.Setenv("CHROMEPROBE_DATA_DIR", dataDir)
	t.Setenv("CHROMEPROBE_NETLOG_DIR", netlogsDir)
	t.Setenv("CHROMEPROBE_CHROME", writeFakeChrome(t))

	if err := runDoctor(); err != nil {
		t.Fatal(err)
	}
	if err := runPS(); err != nil {
		t.Fatal(err)
	}
	if err := runSessions(); err != nil {
		t.Fatal(err)
	}
	if err := runKill(nil); err != nil {
		t.Fatal(err)
	}
	if err := runNetLog([]string{"about:blank"}); err != nil {
		t.Fatal(err)
	}
	if err := runSessions(); err != nil {
		t.Fatal(err)
	}
	if err := runClean([]string{"7"}); err != nil {
		t.Fatal(err)
	}
}

func TestRunInspectReadsDirectFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "netlog.json")
	if err := os.WriteFile(path, []byte(`{"constants":{"logEventTypes":{"ONE":1}},"events":[{"type":1,"params":{}}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := runInspect([]string{path}); err != nil {
		t.Fatal(err)
	}
}

func TestUsageAndVersionRemainAvailable(t *testing.T) {
	printUsage()
	if version == "" {
		t.Fatalf("invalid version: %q", version)
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
func TestRunPSReportsNoChromeProcesses(t *testing.T) {
	t.Setenv("CHROMEPROBE_DATA_DIR", filepath.Join(t.TempDir(), "data"))
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "ps"), []byte("#!/bin/sh\nprintf '1 0 launchd\\n'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	if err := runPS(); err != nil {
		t.Fatal(err)
	}
}
func TestRunKillReportsKilledProcess(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")
	t.Setenv("CHROMEPROBE_DATA_DIR", dataDir)
	profile, err := chrome.ManagedProfileDir()
	if err != nil {
		t.Fatal(err)
	}
	child := exec.Command("sleep", "5")
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = child.Process.Kill() }()
	bin := t.TempDir()
	ps := filepath.Join(bin, "ps")
	script := "#!/bin/sh\nprintf '%s 1 Google Chrome --user-data-dir=%s\\\\n' \"$CHILD_PID\" \"$CHROME_PROFILE\"\n"
	if err := os.WriteFile(ps, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CHILD_PID", strconv.Itoa(child.Process.Pid))
	t.Setenv("CHROME_PROFILE", profile)
	t.Setenv("PATH", bin)
	if err := runKill([]string{"--force"}); err != nil {
		t.Fatal(err)
	}
	_ = child.Wait()
}
func TestMainVersionPath(t *testing.T) {
	original := os.Args
	t.Cleanup(func() { os.Args = original })
	os.Args = []string{"cdt", "version"}
	main()
}
func TestDoctorReportsUnavailableChrome(t *testing.T) {
	t.Setenv("CHROMEPROBE_CHROME", filepath.Join(t.TempDir(), "missing"))
	t.Setenv("CHROMEPROBE_DATA_DIR", filepath.Join(t.TempDir(), "data"))
	t.Setenv("CHROMEPROBE_NETLOG_DIR", filepath.Join(t.TempDir(), "netlogs"))
	if err := runDoctor(); err != nil {
		t.Fatal(err)
	}
}
