package netlog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"syscall"
	"time"

	"github.com/jin-doh/chrome-debug-toolkit/internal/chrome"
)

type Session struct {
	ID         string    `json:"id"`
	PID        int       `json:"pid"`
	URL        string    `json:"url,omitempty"`
	StartedAt  time.Time `json:"started_at"`
	ProfileDir string    `json:"profile_dir"`
	SessionDir string    `json:"session_dir"`
	NetLogPath string    `json:"netlog_path"`
	StdoutPath string    `json:"stdout_path"`
	StderrPath string    `json:"stderr_path"`
	Status     string    `json:"status"`
}

func Start(url string) (*Session, error) {
	sessionsDir, err := chrome.SessionsDir()
	if err != nil {
		return nil, err
	}
	profileDir, err := chrome.ManagedProfileDir()
	if err != nil {
		return nil, err
	}
	netlogsDir, err := chrome.NetLogsDir()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	id := now.Format("20060102-150405.000")
	for n := 1; ; n++ {
		dir := filepath.Join(sessionsDir, id)
		_, statErr := os.Stat(dir)
		if os.IsNotExist(statErr) {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, fmt.Errorf("create session directory: %w", err)
			}
			break
		}
		if statErr != nil {
			return nil, fmt.Errorf("check session directory: %w", statErr)
		}
		id = fmt.Sprintf("%s-%d", now.Format("20060102-150405.000"), n)
	}

	sessionDir := filepath.Join(sessionsDir, id)
	session := &Session{
		ID: id, URL: url, StartedAt: now, ProfileDir: profileDir, SessionDir: sessionDir,
		NetLogPath: filepath.Join(netlogsDir, "chrome-netlog-"+id+".json"),
		StdoutPath: filepath.Join(sessionDir, "chrome.stdout.log"),
		StderrPath: filepath.Join(sessionDir, "chrome.stderr.log"),
		Status:     "running",
	}
	stdout, err := os.OpenFile(session.StdoutPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("create Chrome stdout log: %w", err)
	}
	defer func() {
		_ = stdout.Close()
	}()
	stderr, err := os.OpenFile(session.StderrPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("create Chrome stderr log: %w", err)
	}
	defer func() {
		_ = stderr.Close()
	}()

	result, err := chrome.Launch(chrome.LaunchConfig{
		ProfileDir: profileDir, NetLogPath: session.NetLogPath, URL: url,
		Stdout: stdout, Stderr: stderr,
	})
	if err != nil {
		_ = os.RemoveAll(sessionDir)
		return nil, err
	}
	session.PID = result.PID
	if err := writeSession(session); err != nil {
		if process, findErr := os.FindProcess(session.PID); findErr == nil {
			_ = process.Kill()
		}
		_ = os.RemoveAll(sessionDir)
		return nil, err
	}
	return session, nil
}

func ListSessions() ([]*Session, error) {
	root, err := chrome.SessionsDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read sessions directory: %w", err)
	}

	var sessions []*Session
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(root, entry.Name(), "session.json")
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("read session %s: %w", entry.Name(), err)
		}
		var session Session
		if err := json.Unmarshal(data, &session); err != nil {
			return nil, fmt.Errorf("decode session %s: %w", entry.Name(), err)
		}
		if session.Status == "running" && !processAlive(session.PID) {
			session.Status = "exited"
			_ = writeSession(&session)
		}
		sessions = append(sessions, &session)
	}
	sort.Slice(sessions, func(i, j int) bool { return sessions[i].StartedAt.After(sessions[j].StartedAt) })
	return sessions, nil
}

func writeSession(session *Session) error {
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("encode session metadata: %w", err)
	}
	path := filepath.Join(session.SessionDir, "session.json")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write session metadata: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("commit session metadata: %w", err)
	}
	return nil
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil || err == syscall.EPERM
}
