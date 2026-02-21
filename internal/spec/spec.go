package spec

import (
	"testing"

	"github.com/ad9311/ninete/internal/db"
	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/serve"
)

type Spec struct {
	Store   *logic.Store
	Server  *serve.Server
	Queries repo.Queries
}

func New(t *testing.T) Spec {
	t.Helper()

	app, err := prog.Load()
	if err != nil {
		t.Fatalf("failed to load app configuration: %v", err)
	}

	sqlDB, err := db.Open()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if closeErr := sqlDB.Close(); closeErr != nil {
			t.Fatalf("failed to close test database: %v", closeErr)
		}
	})

	queries := repo.New(app, sqlDB)

	store := logic.New(app, queries)

	server := serve.New(app, store)
	if err := server.LoadTemplates(); err != nil {
		t.Fatalf("failed to load templates, %v", err)
	}

	return Spec{
		Store:   store,
		Server:  server,
		Queries: queries,
	}
}
