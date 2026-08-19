package devtools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type VersionInfo struct {
	Browser              string `json:"Browser"`
	ProtocolVersion      string `json:"Protocol-Version"`
	UserAgent            string `json:"User-Agent"`
	V8Version            string `json:"V8-Version"`
	WebKitVersion        string `json:"WebKit-Version"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

func Version(endpoint string) (*VersionInfo, error) {
	endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	if endpoint == "" {
		return nil, fmt.Errorf("DevTools endpoint is required")
	}
	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Get(endpoint + "/json/version")
	if err != nil {
		return nil, fmt.Errorf("connect to DevTools endpoint: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DevTools endpoint returned %s", response.Status)
	}
	var info VersionInfo
	if err := json.NewDecoder(response.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode DevTools version: %w", err)
	}
	return &info, nil
}
