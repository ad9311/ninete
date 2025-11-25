package seed

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ad9311/ninete/internal/db"
	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"golang.org/x/crypto/bcrypt"
)

const contextTimeout = 10 * time.Second

type Config struct {
	App   *prog.App
	SQLDB *sql.DB
	Store *logic.Store
}

func Run() error {
	app, err := prog.Load()
	if err != nil {
		return fmt.Errorf("failed to load program configuration: %w", err)
	}

	if app.IsProduction() {
		app.Logger.Log("seeds cannot be run in production")

		return nil
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

	sc := Config{
		App:   app,
		SQLDB: sqlDB,
		Store: store,
	}

	seeds := []struct {
		name     string
		seedFunc func() error
	}{
		{"users", sc.SeedUsers},
		{"categories", sc.SeedCategories},
	}

	for _, s := range seeds {
		err := s.seedFunc()
		if err != nil {
			return fmt.Errorf("failed to run seed '%s': %w", s.name, err)
		}
	}

	return nil
}

func CategoryNames() []string {
	return []string{
		"Housing",
		"Transportation",
		"Groceries",
		"Food Delivery",
		"Healthcare",
		"Personal Care",
		"Entertainment",
		"Shopping",
		"Online Shopping",
		"Travel",
		"Financial",
		"Pets",
		"Taxes",
		"Subscriptions",
		"Other",
	}
}

func (sc *Config) SeedUsers() error {
	testPwd := "123456789"
	passHash, err := bcrypt.GenerateFromPassword([]byte(testPwd), bcrypt.MinCost)
	if err != nil {
		return err
	}

	query := `
		BEGIN;
		INSERT INTO "users" ("username", "email", "password_hash")
		VALUES
				('john',  'john@mail.com',   $1),
				('maria', 'maria@mail.com',  $1),
				('peter', 'peter@mail.com',  $1)
		ON CONFLICT DO NOTHING;
		COMMIT;
	`

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	if _, err := sc.SQLDB.ExecContext(ctx, query, passHash); err != nil {
		return err
	}

	return nil
}

func (sc *Config) SeedCategories() error {
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	tx, err := sc.SQLDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO "categories" ("name", "uid")
		VALUES (?, ?)
		ON CONFLICT DO NOTHING;
	`)
	if err != nil {
		_ = tx.Rollback()

		return err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			sc.App.Logger.Error(err)
		}
	}()

	for _, name := range CategoryNames() {
		uid := prog.ToLowerCamel(name)
		if _, err := stmt.ExecContext(ctx, name, uid); err != nil {
			_ = tx.Rollback()

			return err
		}
	}

	return tx.Commit()
}
