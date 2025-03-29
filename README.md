# OpAMP Backend Server

## Overview
This project implements an OpAMP backend server for managing OpenTelemetry agents. It supports:
- WebSocket connections for the OpAMP protocol.
- REST API endpoints to update configurations, log levels, and list agents.
- TLS for secure communication.
- Basic authentication for API endpoints.

## Setup

1. **Install Go:**  
   Ensure you have Go installed (version 1.22.1 or higher).

2. **Set the Authentication Token:**  
   The backend uses an environment variable (`AUTH_TOKEN`) for API authentication. Set this variable to a secure, randomly generated value before running the server. For example, in your shell:
   ```bash
   export AUTH_TOKEN="your-secure-token-here"
   ```
   In production, use a secure method to set this (for example, using a secrets manager or Docker environment variables).

3. **Create the Configuration File:**  
   Create a configuration file at `config/backend.yaml` (or use the provided sample):
   ```yaml
   opamp:
     listen_address: ":4320"
     tls:
       cert_file: "config/certs/server.crt"
       key_file: "config/certs/server.key"

   api:
     listen_address: ":8080"
   ```

4. **Build the Server:**  
   Run:
   ```bash
   make build
   ```

5. **Run the Server:**  
   Run:
   ```bash 
   ./opamp-server
   ```

## API Endpoints

### Update Global Log Level
* Endpoint: `/api/loglevel`
* Method: PUT
* Headers:
  * `Authorization: <your-auth-token>`
  * `Content-Type: application/json`
* Payload:
  ```json
  { "log_level": "debug" }
  ```

### Update Agent-specific Log Level
* Endpoint: `/api/agent/loglevel`
* Method: PUT
* Headers:
  * `Authorization: <your-auth-token>`
  * `Content-Type: application/json`
* Payload:
  ```json
  {
    "agent_id": "agent-123",
    "ip_address": "192.168.1.100",
    "location": "datacenter-1",
    "log_level": "warn"
  }
  ```

### List Agents
* Endpoint: `/api/agents`
* Method: GET
* Headers:
  * `Authorization: <your-auth-token>`

### Update Configuration
* Endpoint: `/api/config`
* Method: POST
* Headers:
  * `Authorization: <your-auth-token>`
  * `Content-Type: application/json`
* Payload: Any valid configuration JSON

## Testing

Run all tests with:
```bash
go test ./...
```

## Security

Ensure that you set the `AUTH_TOKEN` environment variable to a secure value in your production environment. 

For example, generate a token using a secure random generator and then set it in your deployment environment or Docker container.

## Future Improvements
* Full OpAMP protocol callbacks.
* Advanced agent management and configuration propagation.
* Structured logging and enhanced error handling.
* Advanced authentication (e.g., JWT or OAuth2).

## Architecture and Design

### OpAMP Protocol Implementation

The server implements the Open Agent Management Protocol (OpAMP) to communicate with telemetry agents. Key features include:

- **WebSocket Connections**: Agents connect via secure WebSockets for bidirectional communication
- **Configuration Hierarchy**: The server implements a three-tiered configuration system:
  1. Agent-reported effective configuration (highest priority)
  2. Last sent configuration from server to agent
  3. Default configuration (fallback)

### Configuration Management

The server handles different configuration formats from agents:

- **Flexible Configuration Parsing**: The system now handles various key structures in agent-reported configurations
- **Effective Configuration**: The server stores and uses the agent's reported effective configuration when available
- **Configuration Updates**: Changes to configurations (like log levels) are sent to agents and tracked within the server

### Debug Endpoints

The server includes several debug endpoints to help with troubleshooting:

- `/api/debug/agent-config`: Shows detailed configuration status for all agents or a specific agent
- `/api/debug/trigger-logs`: Generates log messages at specified levels for testing
- `/api/debug/synthetic-logs`: Creates synthetic log entries by sending special configurations

### Configuration Feedback Loop

The system implements a complete configuration feedback loop:
1. Server sends configuration to agent
2. Agent processes and applies configuration
3. Agent reports its effective configuration back to server
4. Server stores this effective configuration for future use

## Demonstration

To demonstrate the system:

1. Start the server and agent using Docker Compose:
   ```bash
   docker compose -f demo/docker-compose.yml up --build
   ```

2. Verify agent connectivity:
   ```bash
   ./test-opamp-api.sh
   ```

3. Update log levels to see configuration changes:
   ```bash
   curl -X PUT "http://localhost:8081/api/loglevel" \
     -H "Authorization: your-secure-token-here" \
     -H "Content-Type: application/json" \
     -d '{"log_level": "debug"}'
   ```

4. Generate debug logs:
   ```bash
   curl "http://localhost:8081/api/debug/trigger-logs?level=debug&count=10" \
     -H "Authorization: your-secure-token-here"
   ```

5. Check agent configuration status:
   ```bash
   curl "http://localhost:8081/api/debug/agent-config" \
     -H "Authorization: your-secure-token-here"
   ```