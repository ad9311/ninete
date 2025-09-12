// Package serve provides the HTTP server setup and graceful shutdown logic for the application.
package serve

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ad9311/ninete/internal/app"
	"github.com/ad9311/ninete/internal/srv"
)

// Server represents the main HTTP server for the application.
type Server struct {
	store          *srv.Store
	port           string
	allowedOrigins []string
}

// New creates and returns a new Server instance using the provided store.
func New(store *srv.Store) (*Server, error) {
	allowedOrigins, err := app.LoadList("ALLOWED_ORIGINS")
	if err != nil {
		return nil, err
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Server{
		store:          store,
		port:           port,
		allowedOrigins: allowedOrigins,
	}, nil
}

// Start launches the HTTP server and handles graceful shutdown on interrupt signals.
func (s *Server) Start() error {
	server := &http.Server{
		Addr: ":" + s.port,
		// Handler:           s.Router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    32 << 10,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		// s.config.Logger.Log("Server starting on port %s\n", s.config.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// s.config.Logger.Error("ListenAndServe error: %v", err)
		}
	}()

	<-ctx.Done()
	// s.config.Logger.Log("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		// s.config.Logger.Error("Graceful shutdown failed: %v", err)

		return err
	}
	// s.config.Logger.Log("Server stopped cleanly.")

	return nil
}
