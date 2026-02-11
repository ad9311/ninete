package logic

import (
	"context"

	"github.com/ad9311/ninete/internal/repo"
)

type User struct {
	ID        int
	Username  string
	Email     string
	CreatedAt int64
	UpdatedAt int64
}

func (s *Store) FindUser(ctx context.Context, id int) (User, error) {
	var user repo.User
	var safeUser User

	user, err := s.queries.SelectUser(ctx, id)
	if err != nil {
		return safeUser, err
	}

	safeUser.fromRepoUser(user)

	return safeUser, nil
}

func (s *Store) FindUserForAuth(ctx context.Context, email string) (repo.User, error) {
	var user repo.User

	user, err := s.queries.SelectUserByEmail(ctx, email)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (s *Store) CreateUser(ctx context.Context, params repo.InsertUserParams) (User, error) {
	var user repo.User
	var safeUser User

	user, err := s.queries.InsertUser(ctx, params)
	if err != nil {
		return safeUser, err
	}

	safeUser.fromRepoUser(user)

	return safeUser, nil
}

func (u *User) fromRepoUser(user repo.User) {
	u.ID = user.ID
	u.Username = user.Username
	u.Email = user.Email
	u.CreatedAt = user.CreatedAt
	u.UpdatedAt = user.UpdatedAt
}
