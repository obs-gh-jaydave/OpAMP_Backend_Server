package api

import (
	"encoding/json"
	"log"
	"net/http"
	"opamp-backend/internal/common"
)

// AgentLogLevelUpdateRequest represents the request payload to update an agent's log level.
type AgentLogLevelUpdateRequest struct {
	AgentID   string `json:"agent_id"`
	IPAddress string `json:"ip_address"`
	Location  string `json:"location"`
	LogLevel  string `json:"log_level"`
}

func HandleAgentLogLevelUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received agent log level update request")

		var req AgentLogLevelUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Failed to parse request body: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		log.Printf("Agent log level update request for agent %s to level %s", req.AgentID, req.LogLevel)

		// Validate log level.
		switch req.LogLevel {
		case "debug", "info", "warn", "error":
			// Valid log level.
		default:
			log.Printf("Invalid log level: %s", req.LogLevel)
			http.Error(w, "Invalid log level", http.StatusBadRequest)
			return
		}

		// Get the server instance
		srv := common.GetServerInstance()
		if srv == nil {
			log.Printf("Server not initialized")
			http.Error(w, "Server not initialized", http.StatusInternalServerError)
			return
		}

		// Update IP address and location information if available
		// This requires modifying your Agent struct to include these fields
		if req.IPAddress != "" || req.Location != "" {
			agent, exists := srv.GetAgent(req.AgentID)
			if exists {
				if req.IPAddress != "" {
					// If your Agent struct has an IP field:
					agent.IP = req.IPAddress
					log.Printf("Updated IP address for agent %s to %s", req.AgentID, req.IPAddress)
				}
				if req.Location != "" {
					// If your Agent struct has a Location field:
					agent.Location = req.Location
					log.Printf("Updated location for agent %s to %s", req.AgentID, req.Location)
				}
			}
		}

		// Before calling UpdateAgentLogLevel:
		log.Printf("Attempting to update log level for agent %s to %s", req.AgentID, req.LogLevel)
		err := srv.UpdateAgentLogLevel(req.AgentID, req.LogLevel)
		if err != nil {
			log.Printf("Failed to update agent log level: %v", err)
			http.Error(w, "Failed to update agent log level: "+err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("Agent log level updated successfully for %s", req.AgentID)
		w.WriteHeader(http.StatusOK)
		response := map[string]string{
			"status":    "success",
			"agent_id":  req.AgentID,
			"log_level": req.LogLevel,
			"message":   "Agent log level updated successfully",
		}
		json.NewEncoder(w).Encode(response)
	}
}
