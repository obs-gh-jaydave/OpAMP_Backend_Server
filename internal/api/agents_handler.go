package api

import (
	"encoding/json"
	"net/http"
	"opamp-backend/internal/common"
)

// AgentInfo represents information about a connected agent.
type AgentInfo struct {
	AgentID   string `json:"agent_id"`
	IPAddress string `json:"ip_address"`
	Status    string `json:"status"`
}

// HandleListAgents returns a list of connected agents.
func HandleListAgents() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the server instance
		srv := common.GetServerInstance()
		if srv == nil {
			http.Error(w, "Server not initialized", http.StatusInternalServerError)
			return
		}

		// Get actual agents from the agent manager
		allAgents := srv.GetAllAgents()

		// Convert to AgentInfo objects for the response
		agents := make([]AgentInfo, 0, len(allAgents))
		for _, agent := range allAgents {
			agents = append(agents, AgentInfo{
				AgentID:   agent.ID, // Use the exact ID as stored
				IPAddress: agent.IP,
				Status:    "active",
			})
		}

		// If no agents found, return an empty array
		if len(agents) == 0 {
			agents = []AgentInfo{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(agents)
	}
}
