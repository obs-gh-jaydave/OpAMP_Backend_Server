#!/bin/bash

# Configuration
HOST="localhost"
API_PORT="8081"  # From your config, this is the port for API
AUTH_TOKEN="your-secure-token-here"  # Must match AUTH_TOKEN environment variable
BASE_URL="http://${HOST}:${API_PORT}"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to make API calls
call_api() {
    local method=$1
    local endpoint=$2
    local payload=$3
    local description=$4

    echo -e "${BLUE}=== $description ===${NC}"
    echo "API Call: $method $endpoint"
    if [ ! -z "$payload" ]; then
        echo "Payload: $payload"
    fi
    
    # Make the API call
    if [ -z "$payload" ]; then
        response=$(curl -s -X "$method" \
            -H "Authorization: $AUTH_TOKEN" \
            -H "Content-Type: application/json" \
            "${BASE_URL}${endpoint}")
    else
        response=$(curl -s -X "$method" \
            -H "Authorization: $AUTH_TOKEN" \
            -H "Content-Type: application/json" \
            -d "$payload" \
            "${BASE_URL}${endpoint}")
    fi
    
    # Check if response is valid JSON
    if echo "$response" | jq -e . >/dev/null 2>&1; then
        # It's valid JSON, format it nicely
        echo "Response:"
        echo "$response" | jq .
    else
        # Not valid JSON, just print the raw response
        echo "Response: $response"
    fi
    echo -e "${GREEN}--- End of Request ---${NC}\n"
}

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo -e "${RED}Error: jq is not installed. It's required for JSON formatting.${NC}"
    echo "Install with: sudo apt-get install jq (Ubuntu/Debian) or brew install jq (macOS)"
    exit 1
fi

echo -e "${GREEN}=== OpAMP Backend API Test Script ===${NC}"
echo -e "Host: $HOST"
echo -e "API Port: $API_PORT"
echo -e "Base URL: $BASE_URL\n"

# Step 1: List all agents to see what's connected
call_api "GET" "/api/agents" "" "1. List all connected agents"

# Step 2: Debug endpoint to check configuration status
call_api "GET" "/api/debug/agent-config" "" "2. Check configuration status of all agents"

# Step 3: Get first agent ID from the list (if any)
agents_response=$(curl -s -X "GET" \
    -H "Authorization: $AUTH_TOKEN" \
    -H "Content-Type: application/json" \
    "${BASE_URL}/api/agents")

# Try to extract the first agent ID
if echo "$agents_response" | jq -e '. | length > 0' >/dev/null 2>&1; then
    first_agent_id=$(echo "$agents_response" | jq -r '.[0].agent_id')
    echo -e "${GREEN}Found agent with ID: $first_agent_id${NC}"
    
    # Step 3a: Get detailed configuration for the first agent
    call_api "GET" "/api/debug/agent-config?agent_id=${first_agent_id}" "" "3. Check detailed configuration for agent ${first_agent_id}"
    
    # Step 4: Update the log level for this specific agent to DEBUG
    call_api "PUT" "/api/agent/loglevel" "{\"agent_id\": \"${first_agent_id}\", \"log_level\": \"debug\"}" "4. Update log level to DEBUG for agent ${first_agent_id}"
    
    # Step 5: Check the updated configuration
    call_api "GET" "/api/debug/agent-config?agent_id=${first_agent_id}" "" "5. Verify configuration after setting log level to DEBUG"
    
    # Step 6: Update the log level back to INFO
    call_api "PUT" "/api/agent/loglevel" "{\"agent_id\": \"${first_agent_id}\", \"log_level\": \"info\"}" "6. Update log level back to INFO for agent ${first_agent_id}"
    
    # Step 7: Check the updated configuration again
    call_api "GET" "/api/debug/agent-config?agent_id=${first_agent_id}" "" "7. Verify configuration after setting log level back to INFO"
else
    echo -e "${YELLOW}No agents found. Skipping agent-specific tests.${NC}"
fi

# Step 8: Update global log level to DEBUG
call_api "PUT" "/api/loglevel" "{\"log_level\": \"debug\"}" "8. Update global log level to DEBUG"

# Step 9: Check all agents to see updated configurations
call_api "GET" "/api/debug/agent-config" "" "9. Check all agents after global log level update"

# Step 10: Update global log level back to INFO
call_api "PUT" "/api/loglevel" "{\"log_level\": \"info\"}" "10. Update global log level back to INFO"

# Step 11: Final check of all agents
call_api "GET" "/api/debug/agent-config" "" "11. Final check of all agents after tests"

# Step 12: Debug logs trigger (if you've implemented this endpoint)
call_api "GET" "/api/debug/trigger-logs?level=debug&count=5" "" "12. Trigger some debug logs"

echo -e "${GREEN}=== Test Suite Completed ===${NC}"
echo "You should check your server logs to verify that configuration updates are being applied correctly."