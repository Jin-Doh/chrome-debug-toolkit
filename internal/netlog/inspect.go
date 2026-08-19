package netlog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/jin-doh/chrome-debug-toolkit/internal/chrome"
)

// Summary contains bounded-memory counts extracted from one NetLog file.
type Summary struct {
	Path          string
	TotalEvents   int
	EventCounts   map[string]int
	ErrorCounts   map[int]int
	UnknownEvents int
}

type netLogEvent struct {
	Type   int                        `json:"type"`
	Params map[string]json.RawMessage `json:"params"`
}

func Inspect(target string) (*Summary, error) {
	path, err := resolveTarget(target)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open NetLog %s: %w", path, err)
	}
	defer func() {
		_ = file.Close()
	}()

	decoder := json.NewDecoder(file)
	token, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("read NetLog root: %w", err)
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '{' {
		return nil, fmt.Errorf("NetLog root must be a JSON object")
	}

	typeCounts := make(map[int]int)
	errorCounts := make(map[int]int)
	nameByType := make(map[int]string)
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return nil, fmt.Errorf("read NetLog field: %w", err)
		}
		key := keyToken.(string)
		switch key {
		case "constants":
			if err := decodeConstants(decoder, nameByType); err != nil {
				return nil, err
			}
		case "events":
			if err := decodeEvents(decoder, typeCounts, errorCounts); err != nil {
				return nil, err
			}
		default:
			var ignored json.RawMessage
			if err := decoder.Decode(&ignored); err != nil {
				return nil, fmt.Errorf("skip NetLog field %q: %w", key, err)
			}
		}
	}
	if _, err := decoder.Token(); err != nil {
		return nil, fmt.Errorf("close NetLog root: %w", err)
	}

	summary := &Summary{Path: path, EventCounts: make(map[string]int), ErrorCounts: errorCounts}
	for eventType, count := range typeCounts {
		name, ok := nameByType[eventType]
		if !ok || name == "" {
			summary.UnknownEvents += count
			name = strconv.Itoa(eventType)
		}
		summary.EventCounts[name] += count
		summary.TotalEvents += count
	}
	return summary, nil
}

func decodeConstants(decoder *json.Decoder, names map[int]string) error {
	var constants struct {
		LogEventTypes map[string]json.RawMessage `json:"logEventTypes"`
	}
	if err := decoder.Decode(&constants); err != nil {
		return fmt.Errorf("decode NetLog constants: %w", err)
	}
	for name, raw := range constants.LogEventTypes {
		var value int
		if err := json.Unmarshal(raw, &value); err == nil {
			names[value] = name
		}
	}
	return nil
}

func decodeEvents(decoder *json.Decoder, typeCounts, errorCounts map[int]int) error {
	token, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("read NetLog events: %w", err)
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '[' {
		return fmt.Errorf("NetLog events must be a JSON array")
	}
	for decoder.More() {
		var event netLogEvent
		if err := decoder.Decode(&event); err != nil {
			return fmt.Errorf("decode NetLog event: %w", err)
		}
		typeCounts[event.Type]++
		if raw, ok := event.Params["net_error"]; ok {
			if value, ok := decodeInt(raw); ok {
				errorCounts[value]++
			}
		}
	}
	if _, err := decoder.Token(); err != nil {
		return fmt.Errorf("close NetLog events: %w", err)
	}
	return nil
}

func decodeInt(raw json.RawMessage) (int, bool) {
	var value int
	if err := json.Unmarshal(raw, &value); err == nil {
		return value, true
	}
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return 0, false
	}
	value, err := strconv.Atoi(text)
	return value, err == nil
}

func resolveTarget(target string) (string, error) {
	if target == "" {
		target = "latest"
	}
	if target != "latest" {
		if info, err := os.Stat(target); err == nil && !info.IsDir() {
			path, absErr := filepath.Abs(target)
			if absErr != nil {
				return "", fmt.Errorf("resolve NetLog path: %w", absErr)
			}
			return path, nil
		}
		root, err := chrome.SessionsDir()
		if err != nil {
			return "", err
		}
		path := filepath.Join(root, target, "session.json")
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				return "", fmt.Errorf("NetLog target not found: %s", target)
			}
			return "", fmt.Errorf("read session %s: %w", target, err)
		}
		var session Session
		if err := json.Unmarshal(data, &session); err != nil {
			return "", fmt.Errorf("decode session %s: %w", target, err)
		}
		return session.NetLogPath, nil
	}

	sessions, err := ListSessions()
	if err != nil {
		return "", err
	}
	if len(sessions) == 0 {
		return "", fmt.Errorf("no NetLog sessions found")
	}
	return sessions[0].NetLogPath, nil
}
