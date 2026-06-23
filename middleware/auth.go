package middleware

import (
	"context"
	"corehub/utils"
	"net/http"
)

// AuthMiddleware is a middleware function that intercepts incoming HTTP requests
// to verify JWT tokens, handle CORS preflight requests, and inject the authenticated
// user's username into the request context for downstream handlers.
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers to allow cross-origin requests from the frontend
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight OPTIONS request for browsers
		if r.Method == "OPTIONS" {
			return
		}

		// Extract the JWT token from the URL query parameters
		tokenString := r.URL.Query().Get("token")
		if tokenString == "" {
			http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
			return
		}

		// Parse and validate the token to extract the associated username
		username, err := utils.ParseToken(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		// Inject the authenticated username into the request context
		// This allows handler functions to know exactly who is making the request
		ctx := context.WithValue(r.Context(), "username", username)

		// Pass the request to the next handler in the chain
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
