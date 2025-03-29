package agents

import (
	"fmt"
	"log"
	"sync"
)

// Agent represents an agent and its configuration, along with its connection.
type Agent struct {
	ID              string
	IP              string
	Location        string
	Config          string      // Stores complete configuration
	EffectiveConfig string      // Stores what the agent reports as its active config
	Conn            interface{} // Stores the agent's connection
}

// Manager handles agent registration and information.
type Manager struct {
	mu     sync.RWMutex
	agents map[string]*Agent
}

// NewManager creates a new agent manager.
func NewManager() *Manager {
	return &Manager{
		agents: make(map[string]*Agent),
	}
}

// RegisterAgent registers a new agent.
func (m *Manager) RegisterAgent(agent *Agent) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.agents[agent.ID]; exists {
		log.Printf("Replacing existing agent with ID: %s", agent.ID)
		// We could implement additional cleanup for the existing agent if needed
	}

	m.agents[agent.ID] = agent
}

// DeregisterAgent removes an agent.
// Returns true if the agent was found and removed, false otherwise.
func (m *Manager) DeregisterAgent(agentID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.agents[agentID]; !exists {
		// Log if agent doesn't exist
		log.Printf("Attempted to deregister non-existent agent: %s", agentID)
		return false
	}

	delete(m.agents, agentID)
	log.Printf("Agent deregistered: %s", agentID)
	return true
}

// GetAgent returns an agent by ID.
func (m *Manager) GetAgent(agentID string) (*Agent, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	agent, exists := m.agents[agentID]
	return agent, exists
}

// UpdateAgentConfig updates the configuration of an agent.
func (m *Manager) UpdateAgentConfig(agentID string, config string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	agent, exists := m.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	agent.Config = config
	return nil
}

// UpdateAgentEffectiveConfig updates the effective configuration of an agent.
func (m *Manager) UpdateAgentEffectiveConfig(agentID string, config string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	agent, exists := m.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	agent.EffectiveConfig = config
	return nil
}

// GetAllAgents returns a slice of all registered agents.
func (m *Manager) GetAllAgents() []*Agent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agents := make([]*Agent, 0, len(m.agents))
	for _, agent := range m.agents {
		agents = append(agents, agent)
	}
	return agents
}
