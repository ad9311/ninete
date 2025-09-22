package seed

import (
	"context"
	"fmt"

	"github.com/ad9311/ninete/internal/db"
	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
)

func Run() error {
	app, err := prog.Load()
	if err != nil {
		return fmt.Errorf("failed to load program configuration: %w", err)
	}

	sqlDB, err := db.Open()
	if err != nil {
		return fmt.Errorf("failed to open the database: %w", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			app.Logger.Errorf("failed to close database, %v", err)
		}
	}()

	queries := repo.New(app, sqlDB)

	store, err := logic.New(app, queries)
	if err != nil {
		return fmt.Errorf("failed to set up store: %w", err)
	}

	seeds := []struct {
		name     string
		seedFunc func(*prog.App, *logic.Store) error
	}{
		{"users", seedUsers},
	}

	for _, s := range seeds {
		err := s.seedFunc(app, store)
		if err != nil {
			return fmt.Errorf("failed to run seed '%s': %w", s.name, err)
		}
	}

	return nil
}

func seedUsers(app *prog.App, store *logic.Store) error {
	if app.IsProduction() {
		app.Logger.Log("skipping user seeds")

		return nil
	}

	seedPassword := "123456789"

	params := []logic.SignUpParams{
		{
			Username:             "mary",
			Email:                "mary@mail.com",
			Password:             seedPassword,
			PasswordConfirmation: seedPassword,
		},
		{
			Username:             "joseph",
			Email:                "joseph@mail.com",
			Password:             seedPassword,
			PasswordConfirmation: seedPassword,
		},
	}

	ctx := context.Background()
	for _, p := range params {
		if _, err := store.SignUpUser(ctx, p); err != nil {
			return err
		}
	}

	return nil
}
