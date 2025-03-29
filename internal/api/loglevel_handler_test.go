// in internal/api/loglevel_handler_test.go
package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"opamp-backend/internal/agents"
	"opamp-backend/internal/common"
	"testing"
)

// Mock server implementation for tests
type mockLogLevelServer struct{}

func (m *mockLogLevelServer) UpdateAgentLogLevel(agentID string, logLevel string) error {
	return nil
}

func (m *mockLogLevelServer) GetAllAgents() []*agents.Agent {
	return []*agents.Agent{}
}

func (m *mockLogLevelServer) GetAgentIDs() []string {
	return []string{}
}

func (m *mockLogLevelServer) GetAgent(agentID string) (*agents.Agent, bool) {
	return nil, false
}

func (m *mockLogLevelServer) RequestAgentConfig(agentID string) error {
	return nil
}

func TestHandleLogLevelUpdate_Valid(t *testing.T) {
	// Reset global log level before test
	GlobalLogLevel = "info"

	// Setup mock server
	mockServer := &mockLogLevelServer{}
	common.SetServerInstance(mockServer)

	handler := HandleLogLevelUpdate()
	payload := `{"log_level": "debug"}`
	req := httptest.NewRequest("PUT", "/api/loglevel", bytes.NewBufferString(payload))
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Check for the string in the JSON response instead of raw body
	if !bytes.Contains(w.Body.Bytes(), []byte(`"message":"Log level updated successfully for all agents"`)) {
		t.Errorf("expected response to contain 'Log level updated successfully for all agents', got %s", w.Body.String())
	}

	if GlobalLogLevel != "debug" {
		t.Errorf("expected GlobalLogLevel to be 'debug', got %q", GlobalLogLevel)
	}
}

func TestHandleLogLevelUpdate_Invalid(t *testing.T) {
	handler := HandleLogLevelUpdate()
	payload := `{"log_level": "verbose"}` // "verbose" is not allowed.
	req := httptest.NewRequest("PUT", "/api/loglevel", bytes.NewBufferString(payload))
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
