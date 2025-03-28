package agents

import "sync"

// Agent represents an agent and its configuration.
type Agent struct {
	ID       string
	IP       string
	Location string
	Config   string
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
	m.agents[agent.ID] = agent
}

// DeregisterAgent removes an agent.
func (m *Manager) DeregisterAgent(agentID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.agents, agentID)
}

// GetAgent returns an agent by ID.
func (m *Manager) GetAgent(agentID string) (*Agent, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	agent, exists := m.agents[agentID]
	return agent, exists
}
