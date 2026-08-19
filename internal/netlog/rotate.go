package netlog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jin-doh/chrome-debug-toolkit/internal/chrome"
)

func Rotate(retention time.Duration) (int, error) {
	if retention < 0 {
		return 0, fmt.Errorf("retention must be non-negative")
	}
	root, err := chrome.SessionsDir()
	if err != nil {
		return 0, err
	}
	netlogs, err := chrome.NetLogsDir()
	if err != nil {
		return 0, err
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return 0, fmt.Errorf("read sessions directory: %w", err)
	}
	cutoff := time.Now().Add(-retention)
	removed := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(root, entry.Name())
		info, err := entry.Info()
		if err != nil {
			return removed, fmt.Errorf("stat session %s: %w", entry.Name(), err)
		}
		if info.ModTime().After(cutoff) {
			continue
		}
		var metadata Session
		if data, readErr := os.ReadFile(filepath.Join(path, "session.json")); readErr == nil {
			_ = json.Unmarshal(data, &metadata)
		}
		if err := os.RemoveAll(path); err != nil {
			return removed, fmt.Errorf("remove session %s: %w", entry.Name(), err)
		}
		if isWithinDir(metadata.NetLogPath, netlogs) {
			if err := os.Remove(metadata.NetLogPath); err != nil && !os.IsNotExist(err) {
				return removed, fmt.Errorf("remove NetLog %s: %w", metadata.NetLogPath, err)
			}
		}
		removed++
	}
	return removed, nil
}

func isWithinDir(path, dir string) bool {
	if path == "" {
		return false
	}
	absolutePath, errPath := filepath.Abs(path)
	absoluteDir, errDir := filepath.Abs(dir)
	if errPath != nil || errDir != nil {
		return false
	}
	relative, err := filepath.Rel(absoluteDir, absolutePath)
	return err == nil && relative != "." && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative)
}
