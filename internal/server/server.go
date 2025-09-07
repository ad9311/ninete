// Package server provides functionality to create and run an HTTP server instance.
// Routes and middleware are configured in separate files.
package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ad9311/go-api-base/internal/app"
	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/ad9311/go-api-base/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Server represents an HTTP server instance with configuration, service store, and router.
type Server struct {
	config       *app.Config    // Application configuration
	serviceStore *service.Store // Service layer store
	Router       chi.Router     // HTTP router
}

// New initializes and returns a new Server instance with the provided configuration and service store.
func New(config *app.Config, store *service.Store) *Server {
	s := &Server{
		config:       config,
		serviceStore: store,
		Router:       chi.NewRouter(),
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// setupMiddleware configures middleware for the server's router.
func (s *Server) setupMiddleware() {
	if s.config.Env != app.EnvTest {
		s.Router.Use(middleware.Logger)
	}

	s.Router.Use(middleware.Recoverer)
	s.Router.Use(middleware.RequestID)
	s.Router.Use(WithTimeout(3 * time.Second)) // TODO
	s.Router.Use(JSONMiddleware)
	s.Router.Use(s.CORS)

	s.Router.NotFound(NotFoundHandler)
	s.Router.MethodNotAllowed(MethodNotAllowedHandler)
}

// Start runs the HTTP server and listens for termination signals to gracefully shut down.
func (s *Server) Start() error {
	if !s.config.IsSafeEnv() {
		return errs.ErrServiceFuncNotAvailable
	}

	srv := &http.Server{
		Addr:              ":" + s.config.Port,
		Handler:           s.Router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    32 << 10,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		s.config.Logger.Log("Server starting on port %s\n", s.config.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.config.Logger.Error("ListenAndServe error: %v", err)
		}
	}()

	<-ctx.Done()
	s.config.Logger.Log("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		s.config.Logger.Error("Graceful shutdown failed: %v", err)

		return err
	}
	s.config.Logger.Log("Server stopped cleanly.")

	return nil
}
