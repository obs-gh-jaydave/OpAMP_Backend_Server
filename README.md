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

In production, use a secure method to set this (for example, using a secrets manager or Docker environment variables).
	3.	Create the Configuration File:
Create a configuration file at config/backend.yaml (or use the provided sample):
```bash
opamp:
  listen_address: ":4320"
  tls:
    cert_file: "certs/server.crt"
    key_file: "certs/server.key"

api:
  listen_address: ":8080"
```

4.	Build the Server:
Run:

```bash
make build
```


5.	Run the Server:
Run:

```bash 
./opamp-server
```



API Endpoints

Update Global Log Level
* Endpoint: ```/api/loglevel```
* Method: PUT
* Payload:
```bash
{ "log_level": "debug" }
```


Update Agent-specific Log Level
* Endpoint: ```/api/agent/loglevel```
* Method: PUT
* Payload:
```bash
{
  "agent_id": "agent-123",
  "ip_address": "192.168.1.100",
  "location": "datacenter-1",
  "log_level": "warn"
}
```


List Agents
* Endpoint: ```/api/agents```
* Method: GET

Testing

Run all tests with:
```bash
go test ./...
```
Security

Ensure that you set the ```AUTH_TOKEN``` environment variable to a secure value in your production environment. 

For example, generate a token using a secure random generator and then set it in your deployment environment or Docker container.

Improvements Not Added
* Full OpAMP protocol callbacks.
* Advanced agent management and configuration propagation.
* Structured logging and enhanced error handling.
* Advanced authentication (e.g., JWT or OAuth2).