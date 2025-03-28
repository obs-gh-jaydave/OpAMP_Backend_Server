package api

import (
	"encoding/json"
	"net/http"

	"opamp-backend/internal/config"
)

// AgentLogLevelUpdateRequest represents the request payload to update an agent's log level.
// It now includes additional metadata (e.g. IP address and location) for selecting the target agent.
type AgentLogLevelUpdateRequest struct {
	AgentID   string `json:"agent_id"`   // Unique identifier for the agent.
	IPAddress string `json:"ip_address"` // Optional: IP address of the agent.
	Location  string `json:"location"`   // Optional: Physical location or region.
	LogLevel  string `json:"log_level"`
}

// HandleAgentLogLevelUpdate updates the log level for a specific agent.
// It retrieves the current collector configuration, applies the log level change, and propagates the update.
func HandleAgentLogLevelUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AgentLogLevelUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate log level.
		switch req.LogLevel {
		case "debug", "info", "warn", "error":
			// Valid log level.
		default:
			http.Error(w, "Invalid log level", http.StatusBadRequest)
			return
		}

		// Retrieve current collector configuration.
		currentConfig, err := config.GetCurrentCollectorConfig()
		if err != nil {
			http.Error(w, "Failed to retrieve current configuration", http.StatusInternalServerError)
			return
		}

		// Update configuration YAML with the new log level.
		updatedConfig, err := config.UpdateLogLevelInConfig(currentConfig, req.LogLevel)
		if err != nil {
			http.Error(w, "Failed to update configuration", http.StatusInternalServerError)
			return
		}

		// Propagate the updated configuration to the specified agent.
		if err := config.PropagateConfigToAgent(req.AgentID, updatedConfig); err != nil {
			http.Error(w, "Failed to propagate configuration", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Agent log level updated successfully"))
	}
}
