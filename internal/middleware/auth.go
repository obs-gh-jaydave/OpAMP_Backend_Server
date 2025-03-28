package middleware

import (
	"net/http"
	"os"
)

var authToken = os.Getenv("AUTH_TOKEN") // Retrieve the token from environment variables

// AuthMiddleware checks for a valid Authorization header.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		// If authToken is not set or does not match the header, reject the request.
		if authToken == "" || token != authToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
