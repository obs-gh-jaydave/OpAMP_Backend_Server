package config

import (
	"testing"
)

func TestTLSConfig_InvalidFiles(t *testing.T) {
	// Create a dummy config with non-existent certificate files.
	var cfg Config
	cfg.OpAMP.TLS.CertFile = "nonexistent.crt"
	cfg.OpAMP.TLS.KeyFile = "nonexistent.key"

	_, err := cfg.TLSConfig()
	if err == nil {
		t.Error("Expected error when certificate files do not exist")
	}
}
