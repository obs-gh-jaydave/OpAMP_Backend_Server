package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleConfigUpdate_ValidJSON(t *testing.T) {
	handler := HandleConfigUpdate()

	payload := `{"key": "value"}`
	req := httptest.NewRequest("POST", "/api/config", bytes.NewBufferString(payload))
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	expectedResponse := "Configuration received successfully"
	if w.Body.String() != expectedResponse {
		t.Errorf("expected response %q, got %q", expectedResponse, w.Body.String())
	}
}

func TestHandleConfigUpdate_InvalidJSON(t *testing.T) {
	handler := HandleConfigUpdate()

	payload := `{"key": "value"` // malformed JSON
	req := httptest.NewRequest("POST", "/api/config", bytes.NewBufferString(payload))
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid JSON, got %d", w.Code)
	}
}
