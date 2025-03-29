// in internal/api/agent_loglevel_handler_test.go
package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"opamp-backend/internal/agents"
	"opamp-backend/internal/common"
	"testing"
)

// Mock server implementation
type mockServerImpl struct{}

func (m *mockServerImpl) UpdateAgentLogLevel(agentID string, logLevel string) error {
	return nil // Just return success for tests
}

func (m *mockServerImpl) GetAllAgents() []*agents.Agent {
	return []*agents.Agent{}
}

func (m *mockServerImpl) GetAgentIDs() []string {
	return []string{}
}

func (m *mockServerImpl) GetAgent(agentID string) (*agents.Agent, bool) {
	// Return a properly initialized agent
	return &agents.Agent{
		ID:       "agent-123",
		IP:       "",
		Location: "",
	}, true
}

func (m *mockServerImpl) RequestAgentConfig(agentID string) error {
	return nil
}

func TestHandleAgentLogLevelUpdate_WithMetadata(t *testing.T) {
	// Create and set mock server
	mockServer := &mockServerImpl{}
	common.SetServerInstance(mockServer)

	handler := HandleAgentLogLevelUpdate()

	// Include additional metadata: ip_address and location.
	payload := `{"agent_id": "agent-123", "ip_address": "192.168.1.100", "location": "datacenter-1", "log_level": "warn"}`
	req := httptest.NewRequest("PUT", "/api/agent/loglevel", bytes.NewBufferString(payload))
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Check for the string in the JSON response
	if !bytes.Contains(w.Body.Bytes(), []byte(`"message":"Agent log level updated successfully"`)) {
		t.Errorf("expected response to contain 'Agent log level updated successfully', got %s", w.Body.String())
	}
}
