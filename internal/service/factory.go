package service

import (
	"context"
	"crypto/rand"
	"math/big"
	"testing"
	"time"

	"github.com/ad9311/go-api-base/internal/app"
	"github.com/ad9311/go-api-base/internal/console"
	"github.com/ad9311/go-api-base/internal/db"
	"github.com/ad9311/go-api-base/internal/repo"
	"github.com/jackc/pgx/v5/pgtype"
)

// JWTTokenFactoryParams contains parameters for creating a new JWT token in tests.
type JWTTokenFactoryParams struct {
	UserID    int32     // User ID for the token
	IssuedAt  time.Time // Token issued-at time
	ExpiresAt time.Time // Token expiration time
}

// SaveRefreshTokenFactoryParams contains parameters for creating a new refresh token in tests.
type SaveRefreshTokenFactoryParams struct {
	UserID    int32     // User ID for the token
	IssuedAt  time.Time // Token issued-at time
	ExpiresAt time.Time // Token expiration time
}

// FactoryStore creates and returns a new Store instance with test configuration and database pool.
func FactoryStore(t *testing.T) *Store {
	t.Helper()

	config := app.FactoryConfig(t)
	pool := db.FactoryDBPool(t, config)

	store, err := New(config, pool)
	if err != nil {
		t.Fatalf("failed to create factory store: %v", err)
	}

	return store
}

// FactoryUser creates a new user for testing. If params are not provided, sets default values.
func (s *Store) FactoryUser(ctx context.Context, t *testing.T, params RegistrationParams) repo.User {
	t.Helper()

	username := FactoryUsername()

	params.Username = app.SetDefaultString(params.Username, username)
	params.Email = app.SetDefaultString(params.Email, username+"@testemail.com")
	params.Password = app.SetDefaultString(params.Password, "123456789")
	params.PasswordConfirmation = app.SetDefaultString(params.PasswordConfirmation, "123456789")

	user, err := s.RegisterUser(ctx, params)
	if err != nil {
		t.Fatalf("failed to create factory user: %v", err)
	}

	return user
}

// FactoryUsername generates a username with the given prefix and a random alphanumeric suffix.
func FactoryUsername() string {
	return randomString()
}

// FactorySavedRefreshToken creates and saves a new refresh token for the given user ID in tests.
// Panics if UserID is zero or saving fails.
func (s *Store) FactorySavedRefreshToken(ctx context.Context, t *testing.T, userID int32) repo.RefreshToken {
	t.Helper()

	token, err := s.SaveRefreshToken(ctx, userID)
	if err != nil {
		t.Fatalf("failed to create factory refresh token: %v", err)
	}

	return token
}

// FactoryExpiredToken creates and inserts an expired refresh token for the specified user into the database.
// The function returns the created repo.RefreshToken and fails the test if an error occurs.
func (s *Store) FactoryExpiredToken(ctx context.Context, t *testing.T, userID int32) repo.RefreshToken {
	t.Helper()

	now := time.Now().Add(-RefreshTokenExp).UTC()

	token, err := s.queries.InsertRefreshToken(ctx, repo.InsertRefreshTokenParams{
		UserID: userID,
		IssuedAt: pgtype.Timestamptz{
			Time:  now,
			Valid: true,
		},
		ExpiresAt: pgtype.Timestamptz{
			Time:  now.Add(-RefreshTokenExp),
			Valid: true,
		},
	})
	if err != nil {
		t.Fatalf("failed to created factory expired refresh token: %v", err)
	}

	return token
}

// FactoryRole creates a new role for testing. If no name is given, generates a random one.
// Panics if role creation fails.
func (s *Store) FactoryRole(ctx context.Context, t *testing.T, roleName string) repo.Role {
	t.Helper()

	roleName = app.SetDefaultString(roleName, FactoryUsername())

	role, err := s.CreateNewRole(ctx, roleName)
	if err != nil {
		t.Fatalf("failed to create factory role: %v", err)
	}

	return role
}

// FactoryUserRole adds a role to a user for testing. Panics if the operation fails.
func (s *Store) FactoryUserRole(ctx context.Context, t *testing.T, userID int32, roleName string) repo.UserRole {
	t.Helper()

	userRole, err := s.AddRoleToUser(ctx, userID, roleName)
	if err != nil {
		t.Fatalf("failed to create factory user role: %v", err)
	}

	return userRole
}

// randomString generates a random alphanumeric string of 15 characters.
func randomString() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
	const length = 15

	b := make([]byte, length)
	for i := range length {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			console.NewError("failed to generate random string: %v", err)
		}
		b[i] = letters[num.Int64()]
	}

	return string(b)
}
