package api

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gopkg.in/yaml.v2"
)

// HandleConfigUpdate creates a handler function for updating agent configurations
func HandleConfigUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cfg map[string]interface{}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(body, &cfg); err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}

		yamlConfig, err := yaml.Marshal(cfg)
		if err != nil {
			http.Error(w, "Failed to convert JSON to YAML", http.StatusInternalServerError)
			return
		}

		// Manually calculate SHA256 hash for configuration
		configHash := sha256.Sum256(yamlConfig)

		// Log the config that would be sent
		fmt.Printf("Received configuration update with hash: %x\n", configHash[:])
		fmt.Printf("Configuration content (YAML):\n%s\n", string(yamlConfig))

		fmt.Printf("In a real implementation, this would be sent to OpAMP agents\n")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Configuration received successfully"))
	}
}
