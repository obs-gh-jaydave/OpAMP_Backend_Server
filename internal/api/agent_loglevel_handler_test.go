package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleAgentLogLevelUpdate_WithMetadata(t *testing.T) {
	handler := HandleAgentLogLevelUpdate()

	// Include additional metadata: ip_address and location.
	payload := `{"agent_id": "agent-123", "ip_address": "192.168.1.100", "location": "datacenter-1", "log_level": "warn"}`
	req := httptest.NewRequest("PUT", "/api/agent/loglevel", bytes.NewBufferString(payload))
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	expected := "Agent log level updated successfully"
	if w.Body.String() != expected {
		t.Errorf("expected response %q, got %q", expected, w.Body.String())
	}
}
