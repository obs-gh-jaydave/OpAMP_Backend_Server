package server

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func createTempConfig(content string) (string, error) {
	tmpFile, err := ioutil.TempFile("", "backend-*.yaml")
	if err != nil {
		return "", err
	}
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		return "", err
	}
	tmpFile.Close()
	return tmpFile.Name(), nil
}

const testConfigWebSocket = `
opamp:
  listen_address: "127.0.0.1:4322"
  tls:
    cert_file: ""
    key_file: ""
api:
  listen_address: "127.0.0.1:8081"
`

func TestOpAMPWebSocketConnection(t *testing.T) {
	// Create temporary configuration file.
	configPath, err := createTempConfig(testConfigWebSocket)
	if err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}
	defer os.Remove(configPath)

	s, err := NewServer(configPath)
	if err != nil {
		t.Fatalf("NewServer error: %v", err)
	}

	// Start the server in a goroutine.
	go s.Start()
	// Allow server time to start.
	time.Sleep(500 * time.Millisecond)

	// Dial the WebSocket endpoint.
	wsURL := "ws://127.0.0.1:4322/v1/opamp"
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to establish WebSocket connection: %v", err)
	}
	conn.Close()
}
