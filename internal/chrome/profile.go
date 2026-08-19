package chrome

import (
	"fmt"
	"os"
	"path/filepath"
)

const appName = "chrome-debug-toolkit"

// DataDir returns and creates the toolkit's persistent data directory.
// CHROMEPROBE_DATA_DIR is intended for isolated test and CI runs.
func DataDir() (string, error) {
	if override := os.Getenv("CHROMEPROBE_DATA_DIR"); override != "" {
		path, err := filepath.Abs(override)
		if err != nil {
			return "", fmt.Errorf("resolve CHROMEPROBE_DATA_DIR: %w", err)
		}
		if err := os.MkdirAll(path, 0o755); err != nil {
			return "", fmt.Errorf("create data directory: %w", err)
		}
		return path, nil
	}

	root, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config directory: %w", err)
	}
	path := filepath.Join(root, appName)
	if err := os.MkdirAll(path, 0o755); err != nil {
		return "", fmt.Errorf("create data directory: %w", err)
	}
	return path, nil
}

func ManagedProfileDir() (string, error) {
	root, err := DataDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(root, "profiles", "netlog")
	if err := os.MkdirAll(path, 0o755); err != nil {
		return "", fmt.Errorf("create managed profile directory: %w", err)
	}
	return path, nil
}

func SessionsDir() (string, error) {
	root, err := DataDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(root, "sessions")
	if err := os.MkdirAll(path, 0o755); err != nil {
		return "", fmt.Errorf("create sessions directory: %w", err)
	}
	return path, nil
}

func NetLogsDir() (string, error) {
	if override := os.Getenv("CHROMEPROBE_NETLOG_DIR"); override != "" {
		path, err := filepath.Abs(override)
		if err != nil {
			return "", fmt.Errorf("resolve CHROMEPROBE_NETLOG_DIR: %w", err)
		}
		if err := os.MkdirAll(path, 0o755); err != nil {
			return "", fmt.Errorf("create NetLog directory: %w", err)
		}
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	path := filepath.Join(home, "Downloads", "chrome-netlogs")
	if err := os.MkdirAll(path, 0o755); err != nil {
		return "", fmt.Errorf("create NetLog directory: %w", err)
	}
	return path, nil
}
