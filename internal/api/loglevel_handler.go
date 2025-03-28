package api

import (
	"encoding/json"
	"net/http"
)

type LogLevelUpdateRequest struct {
	LogLevel string `json:"log_level"`
}

// GlobalLogLevel holds the global log level setting.
// The default is "info".
var GlobalLogLevel = "info"

// HandleLogLevelUpdate updates the global log level for the Observe Agent.
func HandleLogLevelUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LogLevelUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Only allow a fixed set of valid log levels.
		switch req.LogLevel {
		case "debug", "info", "warn", "error":
			GlobalLogLevel = req.LogLevel
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Log level updated successfully"))
		default:
			http.Error(w, "Invalid log level", http.StatusBadRequest)
		}
	}
}
