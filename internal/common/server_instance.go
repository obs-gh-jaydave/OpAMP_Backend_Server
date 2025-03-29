package common

import "opamp-backend/internal/agents"

// ServerInterface defines the methods that API handlers need to call on the server
type ServerInterface interface {
	UpdateAgentLogLevel(agentID string, logLevel string) error
	GetAllAgents() []*agents.Agent
	GetAgentIDs() []string
	GetAgent(agentID string) (*agents.Agent, bool) // Added this method
	RequestAgentConfig(agentID string) error       // Added this method
}

var serverInstance ServerInterface

// SetServerInstance sets the global server instance
func SetServerInstance(s ServerInterface) {
	serverInstance = s
}

// GetServerInstance returns the global server instance
func GetServerInstance() ServerInterface {
	return serverInstance
}
