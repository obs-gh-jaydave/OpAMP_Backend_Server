// internal/server/opamp_server.go
package server

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"opamp-backend/internal/agents"
	"opamp-backend/internal/api"
	"opamp-backend/internal/common"
	"opamp-backend/internal/config"
	"opamp-backend/internal/middleware"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/open-telemetry/opamp-go/protobufs"
	"github.com/open-telemetry/opamp-go/server"
	opampTypes "github.com/open-telemetry/opamp-go/server/types"
	"gopkg.in/yaml.v2"
)

// SimpleLogger implements the Logger interface required by opamp-go.
type SimpleLogger struct{}

func (l *SimpleLogger) Debug(msg string, args ...interface{}) {
	log.Printf("DEBUG: "+msg, args...)
}

func (l *SimpleLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	log.Printf("DEBUG: "+format, args...)
}

func (l *SimpleLogger) Info(msg string, args ...interface{}) {
	log.Printf("INFO: "+msg, args...)
}

func (l *SimpleLogger) Error(msg string, args ...interface{}) {
	log.Printf("ERROR: "+msg, args...)
}

func (l *SimpleLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	log.Printf("ERROR: "+format, args...)
}

// Helper function to get keys from a map
func getMapKeys(m map[string]*protobufs.AgentConfigFile) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Server wraps both the OpAMP server and the HTTP API server.
type Server struct {
	opampServer    server.OpAMPServer
	config         config.Config
	httpServer     *http.Server
	agentManager   *agents.Manager
	restartOpampMu sync.Mutex
	stopping       bool
}

// NewServer initializes a new Server instance using the configuration file.
func NewServer(configPath string) (*Server, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	logger := &SimpleLogger{}
	opampSrv := server.New(logger)

	return &Server{
		opampServer:  opampSrv,
		config:       cfg,
		agentManager: agents.NewManager(),
	}, nil
}

// GetAgentIDs returns just the IDs of all registered agents
func (s *Server) GetAgentIDs() []string {
	agents := s.agentManager.GetAllAgents()
	ids := make([]string, 0, len(agents))
	for _, agent := range agents {
		ids = append(ids, agent.ID)
	}
	return ids
}

// GetAgent returns an agent by ID.
func (s *Server) GetAgent(agentID string) (*agents.Agent, bool) {
	return s.agentManager.GetAgent(agentID)
}

// RequestAgentConfig requests the current configuration from an agent.
func (s *Server) RequestAgentConfig(agentID string) error {
	agent, exists := s.agentManager.GetAgent(agentID)
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	conn, ok := agent.Conn.(opampTypes.Connection)
	if !ok {
		return fmt.Errorf("agent connection is not valid")
	}

	// Create a message to request the agent's configuration
	message := &protobufs.ServerToAgent{
		InstanceUid:  []byte(agentID),
		Capabilities: uint64(protobufs.ServerCapabilities_ServerCapabilities_AcceptsEffectiveConfig),
	}

	ctx := context.Background()
	if err := conn.Send(ctx, message); err != nil {
		log.Printf("Failed to send configuration request: %v", err)
		return err
	}

	log.Printf("Configuration request sent to agent %s", agentID)
	return nil
}

// UpdateAgentLogLevel updates the log level for a specific agent
func (s *Server) UpdateAgentLogLevel(agentID string, logLevel string) error {
	log.Printf("UpdateAgentLogLevel called for agent %s with level %s", agentID, logLevel)

	agent, exists := s.agentManager.GetAgent(agentID)
	if !exists {
		log.Printf("Agent %s not found in manager", agentID)
		return fmt.Errorf("agent %s not found", agentID)
	}

	if agent.Conn == nil {
		log.Printf("Agent %s has nil connection", agentID)
		return fmt.Errorf("agent %s connection is nil", agentID)
	}

	log.Printf("Retrieving current collector config for agent %s", agentID)
	currentConfig, err := config.GetCurrentCollectorConfig(agentID)
	if err != nil {
		log.Printf("Failed to get current collector config: %v", err)
		return err
	}

	// Print a preview of the current config
	configLines := strings.Split(currentConfig, "\n")
	previewLines := 10
	if len(configLines) < previewLines {
		previewLines = len(configLines)
	}
	configPreview := strings.Join(configLines[:previewLines], "\n")
	log.Printf("Current config preview for agent %s: \n%s...", agentID, configPreview)

	log.Printf("Updating config with log level: %s", logLevel)
	updatedConfig, err := config.UpdateLogLevelInConfig(currentConfig, logLevel)
	if err != nil {
		log.Printf("Failed to update log level in config: %v", err)
		return err
	}

	// Print a preview of the updated config
	updatedLines := strings.Split(updatedConfig, "\n")
	previewLines = 10
	if len(updatedLines) < previewLines {
		previewLines = len(updatedLines)
	}
	updatedPreview := strings.Join(updatedLines[:previewLines], "\n")
	log.Printf("Updated config preview for agent %s: \n%s...", agentID, updatedPreview)

	// Update the agent's stored configuration.
	log.Printf("Updating stored configuration for agent %s", agentID)
	s.agentManager.UpdateAgentConfig(agentID, updatedConfig)

	// Assert that the stored connection implements opampTypes.Connection.
	conn, ok := agent.Conn.(opampTypes.Connection)
	if !ok {
		log.Printf("Agent connection not valid (type assertion failed)")
		return fmt.Errorf("agent connection is not valid")
	}

	// Create a context and a ServerToAgent message.
	ctx := context.Background()

	// Calculate hash of the config for tracking changes
	configHash := sha256.Sum256([]byte(updatedConfig))

	log.Printf("Creating ServerToAgent message with updated config for agent %s", agentID)
	// Construct the ServerToAgent message with the config update
	message := &protobufs.ServerToAgent{
		InstanceUid: []byte(agentID), // Use the agent's ID as the instance UID
		RemoteConfig: &protobufs.AgentRemoteConfig{
			Config: &protobufs.AgentConfigMap{
				ConfigMap: map[string]*protobufs.AgentConfigFile{
					"collector": {
						Body:        []byte(updatedConfig),
						ContentType: "text/yaml",
					},
				},
			},
			ConfigHash: configHash[:], // Use the hash we calculated
		},
		// Set the appropriate capability flag
		Capabilities: uint64(protobufs.ServerCapabilities_ServerCapabilities_OffersRemoteConfig),
	}

	// Send the message.
	log.Printf("Sending configuration update to agent %s", agentID)
	if err := conn.Send(ctx, message); err != nil {
		log.Printf("Failed to send configuration update: %v", err)
		return fmt.Errorf("failed to send configuration update: %v", err)
	}

	log.Printf("Configuration update successfully sent to agent %s", agentID)
	return nil
}

// GetAllAgents returns a list of currently registered agents.
func (s *Server) GetAllAgents() []*agents.Agent {
	return s.agentManager.GetAllAgents()
}

// RestartOpampServer restarts the OpAMP server after a crash
func (s *Server) RestartOpampServer() {
	s.restartOpampMu.Lock()
	defer s.restartOpampMu.Unlock()

	if s.stopping {
		return
	}

	log.Println("Restarting OpAMP server after crash...")

	// Small delay to allow resources to be freed
	time.Sleep(2 * time.Second)

	// Create a new OpAMP server instance
	s.opampServer = server.New(&SimpleLogger{})

	// Start it again
	go s.startOpampServer()
}

// startOpampServer starts the OpAMP server in a safe way
func (s *Server) startOpampServer() {
	// Wrap the entire function in a recovery handler
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			log.Printf("PANIC RECOVERED from OpAMP server: %v\n%s", r, stack)

			// Restart the server if it wasn't intentionally stopped
			if !s.stopping {
				go s.RestartOpampServer()
			}
		}
	}()

	var tlsConfig *tls.Config
	if s.config.OpAMP.TLS.CertFile != "" && s.config.OpAMP.TLS.KeyFile != "" {
		var err error
		tlsConfig, err = s.config.TLSConfig()
		if err != nil {
			log.Printf("Error in TLS configuration: %v", err)
			return
		}
	}

	startSettings := server.StartSettings{
		ListenEndpoint: s.config.OpAMP.ListenAddress,
		ListenPath:     "/v1/opamp",
		TLSConfig:      tlsConfig,
		Settings: server.Settings{
			Callbacks: opampTypes.Callbacks{
				// Update the OnConnecting callback in the Start method
				OnConnecting: func(request *http.Request) opampTypes.ConnectionResponse {
					// Wrap this entire callback in panic recovery
					var result opampTypes.ConnectionResponse

					func() {
						defer func() {
							if r := recover(); r != nil {
								stack := debug.Stack()
								log.Printf("PANIC RECOVERED in OnConnecting: %v\n%s", r, stack)
								// Default to rejecting the connection if we panic
								result = opampTypes.ConnectionResponse{Accept: false}
							}
						}()

						log.Printf("Agent connecting from: %s", request.RemoteAddr)

						// Extract the agent ID
						// Instead of using X-Agent-ID header, we'll use the instance_uid from
						// the first message. For now, use remoteAddr as temporary ID
						agentID := request.RemoteAddr

						// Create connection callbacks
						callbacks := opampTypes.ConnectionCallbacks{
							OnConnected: func(ctx context.Context, conn opampTypes.Connection) {
								// Wrap this callback in panic recovery
								defer func() {
									if r := recover(); r != nil {
										log.Printf("PANIC RECOVERED in OnConnected: %v", r)
									}
								}()

								if conn == nil {
									log.Printf("OnConnected called with nil connection object for agent %s", agentID)
									return
								}

								// Register the agent with the temporary ID
								s.agentManager.RegisterAgent(&agents.Agent{
									ID:   agentID,
									Conn: conn,
								})
								log.Printf("Agent connected: %s", agentID)
							},
							// Replace the OnMessage callback in the server.Start() method with this enhanced version:

							OnMessage: func(ctx context.Context, conn opampTypes.Connection, message *protobufs.AgentToServer) *protobufs.ServerToAgent {
								// Wrap this callback in panic recovery and ensure we always return something
								var response *protobufs.ServerToAgent = &protobufs.ServerToAgent{}

								func() {
									defer func() {
										if r := recover(); r != nil {
											log.Printf("PANIC RECOVERED in OnMessage: %v", r)
										}
									}()

									if conn == nil {
										log.Printf("OnMessage called with nil connection object for agent %s", agentID)
										return
									}

									// Message can be nil in some cases
									if message == nil {
										return
									}

									// Check if this is the first status message and contains instance_uid
									if len(message.InstanceUid) > 0 {
										// Convert bytes to hex representation for safer handling
										instanceID := fmt.Sprintf("%x", message.InstanceUid)

										if instanceID != agentID && instanceID != "" {
											// Found the real agent ID! Update our registry
											log.Printf("Updating agent ID from %s to %s", agentID, instanceID)

											// First deregister the temporary ID
											s.agentManager.DeregisterAgent(agentID)

											// Then register with the real ID
											s.agentManager.RegisterAgent(&agents.Agent{
												ID:   instanceID,
												Conn: conn,
											})

											// Update our local variable
											agentID = instanceID
										}
									}

									// Check if the message contains effective configuration
									if message.GetEffectiveConfig() != nil && message.GetEffectiveConfig().GetConfigMap() != nil {
										effectiveConfig := message.GetEffectiveConfig().GetConfigMap()

										// Log the available keys in the effective config
										availableKeys := getMapKeys(effectiveConfig.ConfigMap)
										log.Printf("Received effective config with keys: %v", availableKeys)

										// Try to process the effective configuration with a more flexible approach
										foundEffectiveConfig := false
										for key, configFile := range effectiveConfig.ConfigMap {
											log.Printf("Processing config key: %s", key)

											if configFile != nil {
												// Store the effective configuration
												effectiveConfigContent := string(configFile.Body)

												// Log a preview of the effective config
												configLines := strings.Split(effectiveConfigContent, "\n")
												previewLines := 10
												if len(configLines) < previewLines {
													previewLines = len(configLines)
												}
												configPreview := strings.Join(configLines[:previewLines], "\n")

												log.Printf("Received effective configuration from agent %s with key '%s': \n%s...",
													agentID, key, configPreview)
												s.agentManager.UpdateAgentEffectiveConfig(agentID, effectiveConfigContent)
												log.Printf("Updated stored effective configuration for agent %s", agentID)
												foundEffectiveConfig = true
												break // Use the first valid config we find
											}
										}

										if !foundEffectiveConfig {
											log.Printf("Received effective config but no valid config content found for agent %s", agentID)
										}
									}

									// Set instance ID in response
									response.InstanceUid = message.InstanceUid

									// Check for disconnect message but DO NOT deregister here
									if message.GetAgentDisconnect() != nil {
										log.Printf("Agent signaled disconnect: %s", agentID)
									}
								}()

								return response
							},
							OnConnectionClose: func(conn opampTypes.Connection) {
								// Wrap this callback in panic recovery
								defer func() {
									if r := recover(); r != nil {
										log.Printf("PANIC RECOVERED in OnConnectionClose: %v", r)
									}
								}()

								// Handle deregistration only here, not in OnMessage
								s.agentManager.DeregisterAgent(agentID)
								log.Printf("Agent connection closed: %s", agentID)
							},
						}

						// Return a connection response that accepts the connection
						// and sets up our callbacks
						result = opampTypes.ConnectionResponse{
							Accept:              true,
							ConnectionCallbacks: callbacks,
						}
					}()

					return result
				},
			},
		},
	}

	// Start the OpAMP server
	if err := s.opampServer.Start(startSettings); err != nil {
		log.Printf("OpAMP Server failed: %v", err)
	}
}

func (s *Server) Start() {
	s.stopping = false
	common.SetServerInstance(s)

	// Start the OpAMP server in a goroutine
	go s.startOpampServer()

	// Set up the HTTP API.
	mux := http.NewServeMux()
	mux.Handle("/api/config", middleware.AuthMiddleware(http.HandlerFunc(api.HandleConfigUpdate())))
	mux.Handle("/api/loglevel", middleware.AuthMiddleware(http.HandlerFunc(api.HandleLogLevelUpdate())))
	mux.Handle("/api/agent/loglevel", middleware.AuthMiddleware(http.HandlerFunc(api.HandleAgentLogLevelUpdate())))
	mux.Handle("/api/agents", middleware.AuthMiddleware(http.HandlerFunc(api.HandleListAgents())))

	mux.Handle("/api/debug/trigger-logs", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get log level to generate
		level := r.URL.Query().Get("level")
		if level == "" {
			level = "debug" // Default to debug
		}

		// Get count parameter (how many log messages to generate)
		countStr := r.URL.Query().Get("count")
		count := 10 // Default
		if countStr != "" {
			if c, err := strconv.Atoi(countStr); err == nil && c > 0 {
				count = c
			}
		}

		// Get all agents
		agents := s.agentManager.GetAllAgents()
		if len(agents) == 0 {
			http.Error(w, "No agents connected", http.StatusBadRequest)
			return
		}

		// Create a test message to send to all agents
		for _, agent := range agents {
			// Skip agents with nil connection
			if agent.Conn == nil {
				continue
			}

			conn, ok := agent.Conn.(opampTypes.Connection)
			if !ok {
				continue
			}

			// Create a configuration that will generate a lot of log messages
			// This config sets log level and adds a batch processor that will generate log info
			testConfig := fmt.Sprintf(`
	logging:
	  level: %s
	  development: true
	  format: json
	  verbose: true
	
	processors:
	  batch:
		timeout: 100ms
		send_batch_size: %d
		# This will cause the batch processor to log at the configured level
		send_batch_max_size: %d
			`, level, count, count*2)

			configHash := sha256.Sum256([]byte(testConfig))

			// Send as a configuration update to trigger logs
			message := &protobufs.ServerToAgent{
				InstanceUid: []byte(agent.ID),
				RemoteConfig: &protobufs.AgentRemoteConfig{
					Config: &protobufs.AgentConfigMap{
						ConfigMap: map[string]*protobufs.AgentConfigFile{
							"debug_logs": {
								Body:        []byte(testConfig),
								ContentType: "text/yaml",
							},
						},
					},
					ConfigHash: configHash[:],
				},
				Capabilities: uint64(protobufs.ServerCapabilities_ServerCapabilities_OffersRemoteConfig),
			}

			// Send the message
			ctx := context.Background()
			if err := conn.Send(ctx, message); err != nil {
				log.Printf("Failed to send test message to agent %s: %v", agent.ID, err)
			} else {
				log.Printf("Sent test config to agent %s to generate %d %s logs",
					agent.ID, count, level)
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Requested %d %s-level log messages from all agents", count, level)))
	})))

	// Add this endpoint to create synthetic logs by manipulating configurations
	mux.Handle("/api/debug/synthetic-logs", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// First update log level to ensure logs will be visible
		level := r.URL.Query().Get("level")
		if level == "" {
			level = "debug" // Default to debug
		}

		// Get all agents
		agents := s.agentManager.GetAllAgents()
		if len(agents) == 0 {
			http.Error(w, "No agents connected", http.StatusBadRequest)
			return
		}

		// Update log level for all agents
		for _, agentID := range s.GetAgentIDs() {
			err := s.UpdateAgentLogLevel(agentID, level)
			if err != nil {
				log.Printf("Failed to update log level for agent %s: %v", agentID, err)
			} else {
				log.Printf("Updated log level to %s for agent %s", level, agentID)
			}
		}

		// Now send some configuration modifications that will cause log entries
		// For instance, try to send an invalid config that will be rejected and logged
		for _, agent := range agents {
			if agent.Conn == nil {
				continue
			}

			conn, ok := agent.Conn.(opampTypes.Connection)
			if !ok {
				continue
			}

			// Create a slightly invalid configuration that will generate debug logs but not crash
			// For example, try setting a non-existent but harmless property
			invalidConfig := `
logging:
  level: ` + level + `
processors:
  debug_generator:
    enabled: true
    message: "This is a synthetic debug message"
    `

			configHash := sha256.Sum256([]byte(invalidConfig))

			// Send as a configuration update
			message := &protobufs.ServerToAgent{
				InstanceUid: []byte(agent.ID),
				RemoteConfig: &protobufs.AgentRemoteConfig{
					Config: &protobufs.AgentConfigMap{
						ConfigMap: map[string]*protobufs.AgentConfigFile{
							"synthetic": {
								Body:        []byte(invalidConfig),
								ContentType: "text/yaml",
							},
						},
					},
					ConfigHash: configHash[:],
				},
				Capabilities: uint64(protobufs.ServerCapabilities_ServerCapabilities_OffersRemoteConfig),
			}

			ctx := context.Background()
			if err := conn.Send(ctx, message); err != nil {
				log.Printf("Failed to send test config to agent %s: %v", agent.ID, err)
			} else {
				log.Printf("Sent test config to agent %s to generate %s logs", agent.ID, level)
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Updated log level to %s and sent synthetic configuration to generate logs", level)))
	})))
	// Add a debug endpoint to test agent connectivity
	mux.Handle("/api/debug/agents", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agents := s.agentManager.GetAllAgents()
		result := make([]map[string]interface{}, 0, len(agents))

		for _, agent := range agents {
			agentData := map[string]interface{}{
				"id":       agent.ID,
				"id_len":   len(agent.ID),
				"id_bytes": []byte(agent.ID),
			}
			result = append(result, agentData)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})))

	// Add a debug endpoint to inspect agent configuration
	mux.Handle("/api/debug/agent-config", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agentID := r.URL.Query().Get("agent_id")
		if agentID == "" {
			// If no specific agent ID, list all agents with their config status
			agents := s.agentManager.GetAllAgents()
			result := make([]map[string]interface{}, 0, len(agents))

			for _, agent := range agents {
				configInfo := map[string]interface{}{
					"agent_id":             agent.ID,
					"has_effective_config": agent.EffectiveConfig != "",
					"has_config":           agent.Config != "",
					"config_source":        "none",
				}

				// Determine which config source would be used
				if agent.EffectiveConfig != "" {
					configInfo["config_source"] = "effective"
				} else if agent.Config != "" {
					configInfo["config_source"] = "sent"
				} else {
					configInfo["config_source"] = "default"
				}

				result = append(result, configInfo)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(result)
			return
		}

		// If specific agent ID is provided, show detailed config info
		agent, exists := s.agentManager.GetAgent(agentID)
		if !exists {
			http.Error(w, "Agent not found", http.StatusNotFound)
			return
		}

		// Get the current configuration that would be used
		currentConfig, err := config.GetCurrentCollectorConfig(agentID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting current config: %v", err), http.StatusInternalServerError)
			return
		}

		// Create preview of the configs
		effectivePreview := ""
		if agent.EffectiveConfig != "" {
			lines := strings.Split(agent.EffectiveConfig, "\n")
			maxLines := 20
			if len(lines) > maxLines {
				effectivePreview = strings.Join(lines[:maxLines], "\n") + "\n... (truncated)"
			} else {
				effectivePreview = agent.EffectiveConfig
			}
		}

		configPreview := ""
		if agent.Config != "" {
			lines := strings.Split(agent.Config, "\n")
			maxLines := 20
			if len(lines) > maxLines {
				configPreview = strings.Join(lines[:maxLines], "\n") + "\n... (truncated)"
			} else {
				configPreview = agent.Config
			}
		}

		currentConfigPreview := ""
		if currentConfig != "" {
			lines := strings.Split(currentConfig, "\n")
			maxLines := 20
			if len(lines) > maxLines {
				currentConfigPreview = strings.Join(lines[:maxLines], "\n") + "\n... (truncated)"
			} else {
				currentConfigPreview = currentConfig
			}
		}

		// Extract log level from current config
		logLevel := "unknown"
		var configMap map[string]interface{}
		if err := yaml.Unmarshal([]byte(currentConfig), &configMap); err == nil {
			if service, ok := configMap["service"].(map[string]interface{}); ok {
				if telemetry, ok := service["telemetry"].(map[string]interface{}); ok {
					if logs, ok := telemetry["logs"].(map[string]interface{}); ok {
						if level, ok := logs["level"].(string); ok {
							logLevel = level
						}
					}
				}
			}
		}

		// Create result
		result := map[string]interface{}{
			"agent_id":                 agentID,
			"has_effective_config":     agent.EffectiveConfig != "",
			"has_config":               agent.Config != "",
			"config_source":            "none",
			"current_log_level":        logLevel,
			"effective_config_preview": effectivePreview,
			"config_preview":           configPreview,
			"current_config_preview":   currentConfigPreview,
		}

		// Determine which config source is being used
		if agent.EffectiveConfig != "" {
			result["config_source"] = "effective"
		} else if agent.Config != "" {
			result["config_source"] = "sent"
		} else {
			result["config_source"] = "default"
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})))

	l, err := net.Listen("tcp", s.config.API.ListenAddress)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", s.config.API.ListenAddress, err)
	}
	s.httpServer = &http.Server{
		Handler: mux,
	}

	log.Printf("API server listening on %s", l.Addr().String())
	if err := s.httpServer.Serve(l); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server failed: %v", err)
	}
}

// Stop cleanly stops the server
func (s *Server) Stop() {
	s.stopping = true

	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.httpServer.Shutdown(ctx)
	}
}
