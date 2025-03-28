package server

import (
	"bytes"
	"net/http"
	"os"
	"testing"
	"time"
)

const testConfig = `
opamp:
  listen_address: ":4321"
  tls:
    cert_file: ""
    key_file: ""
api:
  listen_address: ":8081"
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

	// Start the server in a goroutine.
	go s.Start()
	// Give the server time to start.
	time.Sleep(500 * time.Millisecond)

	req, err := http.NewRequest("POST", "http://localhost:8081/api/config", bytes.NewBufferString(`{"test": "value"}`))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	// Add the correct Authorization header.
	req.Header.Set("Authorization", "my-secret-token")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("HTTP POST error: %v", err)
	}
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

	// Shutdown the HTTP server.
	s.httpServer.Close()
}
