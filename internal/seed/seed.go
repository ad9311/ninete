package seed

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ad9311/ninete/internal/db"
	"github.com/ad9311/ninete/internal/prog"
	"golang.org/x/crypto/bcrypt"
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

	seeds := []struct {
		name     string
		seedFunc func(context.Context, *sql.DB) error
	}{
		{"users", seedUsers},
		{"categories", seedCategories},
	}

	ctx := context.Background()
	for _, s := range seeds {
		err := s.seedFunc(ctx, sqlDB)
		if err != nil {
			return fmt.Errorf("failed to run seed '%s': %w", s.name, err)
		}
	}

	return nil
}

func seedUsers(ctx context.Context, sqlDB *sql.DB) error {
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

	if _, err := sqlDB.ExecContext(ctx, query, passHash); err != nil {
		return err
	}

	return nil
}

func seedCategories(ctx context.Context, sqlDB *sql.DB) error {
	query := `
		BEGIN;
		INSERT INTO "categories" ("name", "uid")
		VALUES
				('Housing',  'housing'),
				('Transportation',  'transportation'),
				('Groceries',  'groceries'),
				('Food Delivery', 'foodDelivery'),
				('Healthcare',  'healthcare'),
				('Personal Care',  'personalCare'),
				('Entertainment',  'entertainment'),
				('Shopping',  'shopping'),
				('Online Shopping',  'onlineShopping'),
				('Travel',  'travel'),
				('Financial',  'financial'),
				('Pets', 'pets'),
				('Taxes', 'taxes'),
				('Other', 'other')
		ON CONFLICT DO NOTHING;
		COMMIT;
	`

	if _, err := sqlDB.ExecContext(ctx, query); err != nil {
		return err
	}

	return nil
}
