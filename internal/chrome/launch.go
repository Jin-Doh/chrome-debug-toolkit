package chrome

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

type LaunchConfig struct {
	ProfileDir string
	NetLogPath string
	URL        string
	Stdout     io.Writer
	Stderr     io.Writer
	ExtraArgs  []string
}

type LaunchResult struct {
	PID int
}

// Launch starts a detached, profile-isolated Chrome process. It intentionally
// does not wait: the CLI must return while the capture browser remains open.
func Launch(config LaunchConfig) (*LaunchResult, error) {
	executable, err := DetectExecutable()
	if err != nil {
		return nil, err
	}
	if config.ProfileDir == "" {
		return nil, fmt.Errorf("profile directory is required")
	}
	if config.NetLogPath == "" {
		return nil, fmt.Errorf("NetLog path is required")
	}
	if err := os.MkdirAll(config.ProfileDir, 0o755); err != nil {
		return nil, fmt.Errorf("create Chrome profile: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(config.NetLogPath), 0o755); err != nil {
		return nil, fmt.Errorf("create NetLog parent directory: %w", err)
	}

	args := []string{
		"--user-data-dir=" + config.ProfileDir,
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-sync",
		"--log-net-log=" + config.NetLogPath,
		"--net-log-capture-mode=Everything",
	}
	args = append(args, config.ExtraArgs...)
	if config.URL != "" {
		args = append(args, config.URL)
	}

	command := exec.Command(executable, args...)
	command.Stdout = config.Stdout
	command.Stderr = config.Stderr
	if err := command.Start(); err != nil {
		return nil, fmt.Errorf("start Chrome: %w", err)
	}
	pid := command.Process.Pid
	// Release the Go process handle. Chrome remains alive, while the parent CLI
	// does not retain a waitable child that could become a zombie.
	_ = command.Process.Release()
	return &LaunchResult{PID: pid}, nil
}
