package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"opamp-backend/internal/api"
	"os"
	"testing"
	"time"
)

const testConfig = `
opamp:
  listen_address: ":34321"
  tls:
    cert_file: ""
    key_file: ""
api:
  listen_address: ":38081"
`

func TestServerHTTPAPI(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "backend.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(testConfig)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	s, err := NewServer(tmpfile.Name())
	if err != nil {
		t.Fatalf("NewServer error: %v", err)
	}

	// Ensure server is properly stopped after test
	defer func() {
		s.Stop()
		time.Sleep(100 * time.Millisecond) // Give time for ports to be released
	}()

	// Start the server in a goroutine.
	go s.Start()

	// Give the server time to start.
	time.Sleep(500 * time.Millisecond)

	// Create a request directly to the handler
	req, err := http.NewRequest("POST", "/api/config", bytes.NewBufferString(`{"test": "value"}`))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Skip the authentication for this test by calling the handler directly
	w := httptest.NewRecorder()
	handler := api.HandleConfigUpdate()
	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	expected := "Configuration received successfully"
	if buf.String() != expected {
		t.Errorf("Expected response body %q, got %q", expected, buf.String())
	}
}
