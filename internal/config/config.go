package config

import (
	"crypto/tls"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	OpAMP struct {
		ListenAddress string `yaml:"listen_address"`
		TLS           struct {
			CertFile string `yaml:"cert_file"`
			KeyFile  string `yaml:"key_file"`
		} `yaml:"tls"`
	} `yaml:"opamp"`
	API struct {
		ListenAddress string `yaml:"listen_address"`
	} `yaml:"api"`
}

func LoadConfig(path string) (Config, error) {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	err = yaml.Unmarshal(data, &cfg)
	return cfg, err
}

func (c *Config) TLSConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(c.OpAMP.TLS.CertFile, c.OpAMP.TLS.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificates: %v", err)
	}
	return &tls.Config{Certificates: []tls.Certificate{cert}}, nil
}
