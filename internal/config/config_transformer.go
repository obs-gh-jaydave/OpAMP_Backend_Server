package config

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

// CollectorConfig represents a simplified structure of the OpenTelemetry collector configuration.
type CollectorConfig struct {
	Logging struct {
		Level string `yaml:"level"`
	} `yaml:"logging"`
}

// GetCurrentCollectorConfig simulates retrieving the current collector configuration.
// In a real implementation, you might read this from a file or a database.
func GetCurrentCollectorConfig() (string, error) {
	// For demonstration, we return a sample YAML.
	sampleConfig := `
logging:
  level: info
`
	return sampleConfig, nil
}

// UpdateLogLevelInConfig updates the log level in the provided YAML configuration.
func UpdateLogLevelInConfig(originalConfig string, newLogLevel string) (string, error) {
	var cfg CollectorConfig
	if err := yaml.Unmarshal([]byte(originalConfig), &cfg); err != nil {
		return "", fmt.Errorf("failed to unmarshal collector config: %v", err)
	}
	cfg.Logging.Level = newLogLevel

	updated, err := yaml.Marshal(&cfg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal updated config: %v", err)
	}
	return string(updated), nil
}

func PropagateConfigToAgent(agentID string, updatedConfig string) error {
	// TODO: Lookup agent from agent manager and push the configuration update.
	// For the MVP, we simulate success.
	fmt.Printf("Propagating updated configuration to agent %s:\n%s\n", agentID, updatedConfig)
	return nil
}
