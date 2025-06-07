package middleware

import (
	"net/http"
)

// This file contains examples of how to use the middleware chain system.
// These examples are for documentation purposes and show various patterns.

func exampleRouteRegistration() {
	mux := http.NewServeMux()

	// Pattern 1: Using predefined chains
	mux.Handle("/public", PublicChain.ThenFunc(publicHandler))
	mux.Handle("/auth-required", AuthenticatedChain.ThenFunc(authHandler))
	mux.Handle("/protected", ProtectedChain.ThenFunc(protectedHandler))

	// Pattern 2: Using helper functions
	mux.Handle("/api/user", WithAuthFunc(userAPIHandler))
	mux.Handle("/api/profile", WithProtectionFunc(profileHandler))

	// Pattern 3: Custom middleware chains
	apiChain := NewChain(
		loggingMiddleware,
		corsMiddleware,
		UserContextMiddleware,
	)
	mux.Handle("/api/data", apiChain.ThenFunc(dataHandler))

	// Pattern 4: Composing chains
	baseAPIChain := NewChain(loggingMiddleware, corsMiddleware)
	protectedAPIChain := baseAPIChain.Append(AuthMiddleware, UserContextMiddleware)
	mux.Handle("/api/admin", protectedAPIChain.ThenFunc(adminHandler))

	// Pattern 5: Ad-hoc middleware application
	mux.Handle("/special", Apply(http.HandlerFunc(specialHandler), 
		loggingMiddleware, 
		AuthMiddleware,
		specialMiddleware))

	// Pattern 6: Using WithMiddleware for inline chains
	mux.Handle("/custom", 
		WithMiddleware(
			loggingMiddleware,
			AuthMiddleware,
			UserContextMiddleware,
		).ThenFunc(customHandler))
}

// Example middleware functions (placeholders)
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log request
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

func specialMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Special processing
		next.ServeHTTP(w, r)
	})
}

// Example handler functions (placeholders)
func publicHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Public content"))
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Authenticated content"))
}

func protectedHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Protected content"))
}

func userAPIHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("User API"))
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Profile"))
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Data"))
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Admin"))
}

func specialHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Special"))
}

func customHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Custom"))
}