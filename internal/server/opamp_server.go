package server

import (
	"context"
	"log"
	"net"
	"net/http"
	"opamp-backend/internal/api"
	"opamp-backend/internal/config"
	"opamp-backend/internal/middleware"

	"github.com/open-telemetry/opamp-go/server"
)

// SimpleLogger implements the types.Logger interface needed by OpAMP.
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

type Server struct {
	opampServer server.OpAMPServer
	config      config.Config
	httpServer  *http.Server
}

func NewServer(configPath string) (*Server, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	// Create a logger for the OpAMP server.
	logger := &SimpleLogger{}

	// Create a new OpAMP server implementation.
	opampSrv := server.New(logger)

	return &Server{
		opampServer: opampSrv,
		config:      cfg,
	}, nil
}

func (s *Server) Start() {
	// Start OpAMP server in a goroutine.
	go func() {
		startSettings := server.StartSettings{
			Settings: server.Settings{
				// TODO: Implement protocol callbacks for Hello, AgentDescription, RemoteConfig, etc.
			},
			ListenEndpoint: s.config.OpAMP.ListenAddress,
			ListenPath:     "/v1/opamp",
		}

		// Add TLS if configured.
		if s.config.OpAMP.TLS.CertFile != "" && s.config.OpAMP.TLS.KeyFile != "" {
			tlsConfig, err := s.config.TLSConfig()
			if err != nil {
				log.Fatalf("Error in TLS configuration: %v", err)
			}
			startSettings.TLSConfig = tlsConfig
		}

		if err := s.opampServer.Start(startSettings); err != nil {
			log.Fatalf("OpAMP Server failed: %v", err)
		}
	}()

	// Setup HTTP API endpoints.
	mux := http.NewServeMux()
	mux.Handle("/api/config", middleware.AuthMiddleware(http.HandlerFunc(api.HandleConfigUpdate())))
	mux.Handle("/api/loglevel", middleware.AuthMiddleware(http.HandlerFunc(api.HandleLogLevelUpdate())))
	mux.Handle("/api/agent/loglevel", middleware.AuthMiddleware(http.HandlerFunc(api.HandleAgentLogLevelUpdate())))
	mux.Handle("/api/agents", middleware.AuthMiddleware(http.HandlerFunc(api.HandleListAgents())))

	// Use net.Listen to capture the actual API address.
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
