package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"opamp-backend/internal/common"
)

type LogLevelUpdateRequest struct {
	LogLevel string `json:"log_level"`
}

// GlobalLogLevel holds the global log level setting.
// The default is "info".
var GlobalLogLevel = "info"

// HandleLogLevelUpdate updates the global log level for the Observe Agent.
func HandleLogLevelUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received global log level update request")

		var req LogLevelUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Failed to parse request body: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		log.Printf("Global log level update request to level %s", req.LogLevel)

		// Only allow a fixed set of valid log levels.
		switch req.LogLevel {
		case "debug", "info", "warn", "error":
			GlobalLogLevel = req.LogLevel
		default:
			log.Printf("Invalid log level: %s", req.LogLevel)
			http.Error(w, "Invalid log level", http.StatusBadRequest)
			return
		}

		// Get the server instance
		srv := common.GetServerInstance()
		if srv == nil {
			// If server is not available, just update the global variable
			log.Printf("Server not initialized, only updating global variable")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Global log level updated, but server is not available to update agents"))
			return
		}

		// Update all connected agents
		agentIDs := srv.GetAgentIDs()
		log.Printf("Updating log level to %s for %d agents", req.LogLevel, len(agentIDs))

		updateErrors := 0
		updatedAgents := 0

		for _, agentID := range agentIDs {
			log.Printf("Updating agent %s to log level %s", agentID, req.LogLevel)
			if err := srv.UpdateAgentLogLevel(agentID, req.LogLevel); err != nil {
				// Log the error but continue updating other agents
				log.Printf("Error updating agent %s: %v", agentID, err)
				updateErrors++
			} else {
				updatedAgents++
			}
		}

		log.Printf("Updated %d agents, encountered %d errors", updatedAgents, updateErrors)

		// Prepare the response
		response := map[string]interface{}{
			"global_log_level": GlobalLogLevel,
			"total_agents":     len(agentIDs),
			"updated_agents":   updatedAgents,
			"failed_updates":   updateErrors,
		}

		w.Header().Set("Content-Type", "application/json")
		if updateErrors > 0 {
			w.WriteHeader(http.StatusPartialContent)
			response["message"] = fmt.Sprintf("Global log level updated but failed to update %d agents", updateErrors)
		} else {
			w.WriteHeader(http.StatusOK)
			response["message"] = "Log level updated successfully for all agents"
		}

		json.NewEncoder(w).Encode(response)
	}
}
