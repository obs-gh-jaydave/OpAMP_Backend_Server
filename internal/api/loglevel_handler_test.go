package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleLogLevelUpdate_Valid(t *testing.T) {
	// Reset global log level before test.
	GlobalLogLevel = "info"
	handler := HandleLogLevelUpdate()
	payload := `{"log_level": "debug"}`
	req := httptest.NewRequest("PUT", "/api/loglevel", bytes.NewBufferString(payload))
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	expected := "Log level updated successfully"
	if w.Body.String() != expected {
		t.Errorf("expected response %q, got %q", expected, w.Body.String())
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
