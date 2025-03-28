package api

import (
	"encoding/json"
	"net/http"
)

// AgentInfo represents information about a connected agent.
type AgentInfo struct {
	AgentID   string `json:"agent_id"`
	IPAddress string `json:"ip_address"`
	Status    string `json:"status"`
}

// HandleListAgents returns a list of connected agents.
// In a complete implementation, this would query an agent manager.
func HandleListAgents() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Dummy data for now.
		agents := []AgentInfo{
			{"agent-123", "192.168.1.100", "active"},
			{"agent-456", "192.168.1.101", "inactive"},
		}
		json.NewEncoder(w).Encode(agents)
	}
}
