package chrome

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestListProcessesFiltersToChromeCommands(t *testing.T) {
	t.Setenv("CHROMEPROBE_DATA_DIR", filepath.Join(t.TempDir(), "data"))
	processes, err := ListProcesses()
	if err != nil {
		t.Fatal(err)
	}
	for _, process := range processes {
		if process.PID <= 0 || process.Command == "" {
			t.Fatalf("invalid process: %#v", process)
		}
	}
}

func TestListProcessesHandlesMalformedRows(t *testing.T) {
	t.Setenv("CHROMEPROBE_DATA_DIR", filepath.Join(t.TempDir(), "data"))
	profile, err := ManagedProfileDir()
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	ps := filepath.Join(bin, "ps")
	script := "#!/bin/sh\nprintf 'bad\\n'\nprintf 'abc 1 Google Chrome\\n'\nprintf '123 bad Google Chrome\\n'\nprintf '0 1 Google Chrome\\n'\nprintf '123 1 Google Chrome --user-data-dir=" + profile + "\\n'\n"
	if err := os.WriteFile(ps, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	processes, err := ListProcesses()
	if err != nil {
		t.Fatal(err)
	}
	if len(processes) != 1 || !processes[0].Managed {
		t.Fatalf("ListProcesses() = %#v", processes)
	}
}

func TestKillManagedSignalsManagedRoots(t *testing.T) {
	child := exec.Command("sleep", "5")
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = child.Process.Kill() }()
	original := processList
	t.Cleanup(func() { processList = original })
	processList = func() ([]Process, error) {
		return []Process{{PID: child.Process.Pid, PPID: os.Getpid(), Managed: true}}, nil
	}
	killed, err := KillManaged(false)
	if err != nil {
		t.Fatal(err)
	}
	if killed != 1 {
		t.Fatalf("KillManaged(false) = %d, want 1", killed)
	}
	_ = child.Wait()
}

func TestKillManagedSupportsForce(t *testing.T) {
	child := exec.Command("sleep", "5")
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = child.Process.Kill() }()
	original := processList
	t.Cleanup(func() { processList = original })
	processList = func() ([]Process, error) {
		return []Process{{PID: child.Process.Pid, PPID: os.Getpid(), Managed: true}}, nil
	}
	killed, err := KillManaged(true)
	if err != nil {
		t.Fatal(err)
	}
	if killed != 1 {
		t.Fatalf("KillManaged(true) = %d, want 1", killed)
	}
	_ = child.Wait()
}

func TestKillManagedWithIsolatedProfileHasNoTargets(t *testing.T) {
	t.Setenv("CHROMEPROBE_DATA_DIR", filepath.Join(t.TempDir(), "data"))
	killed, err := KillManaged(false)
	if err != nil {
		t.Fatal(err)
	}
	if killed != 0 {
		t.Fatalf("KillManaged() killed %d unrelated processes", killed)
	}
}
