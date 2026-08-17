package serve

import (
	"context"
	"database/sql"
	"errors"
	"html/template"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ad9311/ninete/internal/handlers"
	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	Router  chi.Router
	Session *scs.SessionManager

	templates map[handlers.TemplateName]*template.Template
	handlers  *handlers.Handler
	app       *prog.App
	store     *logic.Store
	host      string
	port      string
}

// defaultHost keeps the listener on loopback, which is what Caddy proxies to and
// what docs/deployment.md describes. It is also what makes middleware.RealIP
// safe: no client can reach the port directly to forge a forwarded address and
// slip past the auth rate limiter. Override with HOST only if something other
// than a local reverse proxy needs to connect.
const defaultHost = "127.0.0.1"

func New(app *prog.App, store *logic.Store, db *sql.DB) *Server {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	host, ok := os.LookupEnv("HOST")
	if !ok {
		host = defaultHost
	}

	s := &Server{
		Router:  chi.NewRouter(),
		Session: scs.New(),
		app:     app,
		store:   store,
		host:    host,
		port:    port,
	}

	s.Session.Store = sqlite3store.New(db)

	s.handlers = handlers.New(handlers.Deps{
		App:            app,
		Store:          store,
		Session:        s.Session,
		TemplateByName: s.templateByName,
		ReloadTemplates: func() error {
			return s.LoadTemplates()
		},
	})

	s.setUpMiddlewares()
	s.setUpRoutes()
	s.setUpSession()

	return s
}

func (s *Server) Start() error {
	server := &http.Server{
		Addr:              net.JoinHostPort(s.host, s.port),
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
		s.app.Logger.Logf("Server starting on %s\n", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.app.Logger.Errorf("ListenAndServe error: %v", err)
		}
	}()

	<-ctx.Done()
	s.app.Logger.Log("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		s.app.Logger.Errorf("Graceful shutdown failed: %v", err)

		return err
	}
	s.app.Logger.Log("Server stopped cleanly.")

	return nil
}
