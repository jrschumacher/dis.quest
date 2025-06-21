// Package server provides HTTP server initialization and configuration
package server

import (
	"net/http"
	"time"

	"github.com/jrschumacher/dis.quest/internal/config"
	"github.com/jrschumacher/dis.quest/internal/db"
	"github.com/jrschumacher/dis.quest/internal/logger"
	"github.com/jrschumacher/dis.quest/internal/pds"
	apphandlers "github.com/jrschumacher/dis.quest/server/app"
	authhandlers "github.com/jrschumacher/dis.quest/server/auth-handlers"
	wellknownhandlers "github.com/jrschumacher/dis.quest/server/dot-well-known-handlers"
	healthhandlers "github.com/jrschumacher/dis.quest/server/health-handlers"
)

const (
	readTimeout  = 10 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 60 * time.Second

	// Headers
	contentTypeOptions    = "nosniff"
	frameOptions          = "DENY"
	xssProtection         = "1; mode=block"
	contentSecurityPolicy = "default-src 'self'; script-src 'self' 'unsafe-eval' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline'"
	referrerPolicy        = "strict-origin-when-cross-origin"
)

// Start initializes and starts the HTTP server with the given configuration
func Start(cfg *config.Config) {
	if err := config.Validate(cfg); err != nil {
		logger.Error("invalid config", "error", err)
		panic("invalid config")
	}

	// Initialize database service
	dbService, err := db.NewService(cfg)
	if err != nil {
		logger.Error("failed to initialize database service", "error", err)
		panic("failed to initialize database service")
	}
	defer func() {
		if err := dbService.Close(); err != nil {
			logger.Error("failed to close database service", "error", err)
		}
	}()

	// Initialize PDS service (real ATProtocol implementation)
	pdsService := pds.NewATProtoService()

	mux := http.NewServeMux()

	// Serve static assets with existence check
	mux.HandleFunc("/assets/", func(w http.ResponseWriter, r *http.Request) {
		assetPath := "." + r.URL.Path
		if fi, err := http.Dir(".").Open(assetPath); err == nil {
			if closeErr := fi.Close(); closeErr != nil {
				logger.Error("failed to close file", "error", closeErr)
			}
			http.ServeFile(w, r, assetPath)
		} else {
			http.NotFound(w, r)
		}
	})

	wellknownhandlers.RegisterRoutes(mux, "/.well-known", cfg)
	authhandlers.RegisterRoutes(mux, "/auth", cfg)
	healthhandlers.RegisterRoutes(mux, "/health", cfg)
	apphandlers.RegisterRoutes(mux, "/", cfg, dbService, pdsService)

	// Secure headers middleware
	handler := secureHeaders(mux)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	logger.Info("Listening on " + srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server error", "error", err)
	}
}

// secureHeaders adds common security headers to all responses
func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", contentTypeOptions)
		w.Header().Set("X-Frame-Options", frameOptions)
		w.Header().Set("X-XSS-Protection", xssProtection)
		w.Header().Set("Content-Security-Policy", contentSecurityPolicy)
		w.Header().Set("Referrer-Policy", referrerPolicy)
		next.ServeHTTP(w, r)
	})
}
