package config

import (
	"fmt"
	"log"
	"opamp-backend/internal/common"
	"strings"

	"gopkg.in/yaml.v2"
)

// CollectorConfig represents a simplified structure of the OpenTelemetry collector configuration.
type CollectorConfig struct {
	Logging struct {
		Level string `yaml:"level"`
	} `yaml:"logging"`
}

// GetCurrentCollectorConfig retrieves the current configuration for a specific agent.
func GetCurrentCollectorConfig(agentID string) (string, error) {
	srv := common.GetServerInstance()
	if srv == nil {
		return "", fmt.Errorf("server not initialized")
	}

	// Get the agent
	agent, exists := srv.GetAgent(agentID)
	if !exists {
		return "", fmt.Errorf("agent %s not found", agentID)
	}

	// If we have effective configuration reported by the agent, use it
	if agent.EffectiveConfig != "" {
		log.Printf("Using agent-reported effective configuration for agent %s", agentID)
		return agent.EffectiveConfig, nil
	}

	// If not, use the last known configuration we sent
	if agent.Config != "" {
		log.Printf("Using last sent configuration for agent %s", agentID)
		return agent.Config, nil
	}

	// If we have nothing stored, return a default configuration
	log.Printf("Using default configuration for agent %s (no stored config found)", agentID)
	defaultConfig := getDefaultConfig()
	return defaultConfig, nil
}

// UpdateLogLevelInConfig updates the log level in the given configuration.
func UpdateLogLevelInConfig(originalConfig string, newLogLevel string) (string, error) {
	// Log the first few lines of the original config for debugging
	lines := strings.Split(originalConfig, "\n")
	preview := strings.Join(lines[:min(10, len(lines))], "\n")
	log.Printf("Updating log level to %s in config: \n%s...", newLogLevel, preview)

	// Parse the original YAML
	var configMap map[string]interface{}
	if err := yaml.Unmarshal([]byte(originalConfig), &configMap); err != nil {
		return "", fmt.Errorf("failed to unmarshal collector config: %v", err)
	}

	// Check if service section exists
	service, ok := configMap["service"].(map[string]interface{})
	if !ok {
		service = make(map[string]interface{})
		configMap["service"] = service
	}

	// Check if telemetry section exists
	telemetry, ok := service["telemetry"].(map[string]interface{})
	if !ok {
		telemetry = make(map[string]interface{})
		service["telemetry"] = telemetry
	}

	// Check if logs section exists
	logs, ok := telemetry["logs"].(map[string]interface{})
	if !ok {
		logs = make(map[string]interface{})
		telemetry["logs"] = logs
	}

	// Update the log level
	logs["level"] = newLogLevel

	// Marshal back to YAML
	updated, err := yaml.Marshal(configMap)
	if err != nil {
		return "", fmt.Errorf("failed to marshal updated config: %v", err)
	}

	// Log a preview of the updated config
	updatedLines := strings.Split(string(updated), "\n")
	updatedPreview := strings.Join(updatedLines[:min(10, len(updatedLines))], "\n")
	log.Printf("Updated config preview: \n%s...", updatedPreview)

	return string(updated), nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getDefaultConfig returns the default collector configuration
func getDefaultConfig() string {
	defaultConfig := `receivers:
  otlp:
    protocols:
      grpc:
        endpoint: "0.0.0.0:4317"
      http:
        endpoint: "0.0.0.0:4318"
  filelog:
    include: ["/logs/*.log"]
    start_at: beginning
    operators:
      - type: regex_parser
        regex: '^(?P<timestamp>\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}) (?P<message>.*)$'

exporters:
  otlphttp/observe:
    endpoint: "https://134414420961.collect.observeinc.com/v2/otel"
    headers:
      authorization: "Bearer ${OBSERVE_TOKEN}"

processors:
  batch:

extensions:
  opamp:
    instance_uid: "01HFPQ7Z5G8YPJM6QRNCT9KD8N"
    capabilities:
      reports_effective_config: true
    server:
      ws:
        endpoint: "wss://opamp-backend:4320/v1/opamp"
        headers:
          Authorization: "Secret-Key your-secure-token-here"
        tls:
          insecure: true

service:
  telemetry:
    logs:
      level: "info"
      development: true
      output_paths: ["stdout"]
  extensions: [opamp]
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlphttp/observe]
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlphttp/observe]
    logs: 
      receivers: [otlp, filelog]
      processors: [batch]
      exporters: [otlphttp/observe]`
	return defaultConfig
}
