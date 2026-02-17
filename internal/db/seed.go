package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
)

func RunSeeds() error {
	app, err := prog.Load()
	if err != nil {
		return err
	}

	sqlDB, err := Open()
	if err != nil {
		return err
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			app.Logger.Errorf("failed to close database: %v", err)
		}
	}()

	queries := repo.New(app, sqlDB)

	store := logic.New(app, queries)

	seeds := []struct {
		name       string
		f          func(*logic.Store) error
		skipOnProd bool
	}{
		{
			"user",
			seedUsers,
			true,
		},
		{
			"category",
			seedCategories,
			false,
		},
	}

	for _, s := range seeds {
		if s.skipOnProd && app.IsProduction() {
			continue
		}

		if err := s.f(store); err != nil {
			return fmt.Errorf("failed to run %s seeds, %w", s.name, err)
		}
	}

	return nil
}

func newContext() (context.Context, context.CancelFunc) {
	ctx := context.Background()

	return context.WithTimeout(ctx, 30*time.Second)
}

func seedUsers(s *logic.Store) error {
	ctx, cancel := newContext()
	defer cancel()

	commonPassword := "123456789"

	userNames := []string{
		"john",
		"mario",
		"claudia",
	}

	for _, u := range userNames {
		hashedPassword, err := logic.HashPassword(commonPassword)
		if err != nil {
			return err
		}

		if _, err := s.CreateUser(ctx, repo.InsertUserParams{
			Username:     u,
			Email:        u + "@email.com",
			PasswordHash: hashedPassword,
		}); err != nil && err.Error() != "UNIQUE constraint failed: users.email" {
			return err
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
		"Utilities",
		"Restaurants",
	}
}

func seedCategories(s *logic.Store) error {
	ctx, cancel := newContext()
	defer cancel()

	for _, c := range CategoryNames() {
		_, err := s.CreateCategory(ctx, c)
		if err != nil && !strings.Contains(err.Error(), "UNIQUE constraint failed: categories.") {
			return err
		}
	}

	return nil
}
