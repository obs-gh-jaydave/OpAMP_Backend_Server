package server

import (
	"testing"
)

func TestNewServer_InvalidConfig(t *testing.T) {
	_, err := NewServer("nonexistent.yaml")
	if err == nil {
		t.Error("expected error when config file does not exist")
	}
}
