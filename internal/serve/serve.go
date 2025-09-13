package serve

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	Router chi.Router

	app            *prog.App
	store          *logic.Store
	port           string
	allowedOrigins []string
}

func New(app *prog.App, store *logic.Store) (*Server, error) {
	allowedOrigins, err := prog.LoadList("ALLOWED_ORIGINS")
	if err != nil {
		return nil, err
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	s := &Server{
		Router:         chi.NewRouter(),
		app:            app,
		store:          store,
		port:           port,
		allowedOrigins: allowedOrigins,
	}

	s.setUpMiddlewares()
	s.setUpRoutes()

	return s, nil
}

func (s *Server) Start() error {
	server := &http.Server{
		Addr:              ":" + s.port,
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
		s.app.Logger.Log("Server starting on port %s\n", s.port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.app.Logger.Error("ListenAndServe error: %v", err)
		}
	}()

	<-ctx.Done()
	s.app.Logger.Log("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		s.app.Logger.Error("Graceful shutdown failed: %v", err)

		return err
	}
	s.app.Logger.Log("Server stopped cleanly.")

	return nil
}
