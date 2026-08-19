package devtools

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVersionRejectsEmptyEndpoint(t *testing.T) {
	if _, err := Version("   "); err == nil {
		t.Fatal("Version accepted an empty endpoint")
	}
}

func TestVersionRejectsMalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{"))
	}))
	defer server.Close()
	if _, err := Version(server.URL); err == nil {
		t.Fatal("Version accepted malformed JSON")
	}
}
