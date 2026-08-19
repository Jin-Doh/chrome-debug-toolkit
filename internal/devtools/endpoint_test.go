package devtools

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVersionReadsJSONVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/json/version" {
			t.Fatalf("path = %s, want /json/version", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Browser":"Chrome/140","Protocol-Version":"1.3","webSocketDebuggerUrl":"ws://127.0.0.1/devtools/browser/1"}`))
	}))
	defer server.Close()

	info, err := Version(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	if info.Browser != "Chrome/140" || info.ProtocolVersion != "1.3" || info.WebSocketDebuggerURL == "" {
		t.Fatalf("unexpected version response: %#v", info)
	}
}

func TestVersionRejectsHTTPError(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	if _, err := Version(server.URL); err == nil {
		t.Fatal("Version accepted HTTP 404")
	}
}
