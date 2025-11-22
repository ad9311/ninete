package repo

import (
	"context"
)

type User struct {
	ID           int
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    int64
	UpdatedAt    int64
}

type SafeUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type InsertUserParams struct {
	Username     string
	Email        string
	PasswordHash []byte
}

func (u *User) ToSafe() SafeUser {
	return SafeUser{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
	}
}

const insertUser = `
INSERT INTO "users" ("username", "email", "password_hash")
VALUES (?, ?, ?)
RETURNING *`

func (q *Queries) InsertUser(ctx context.Context, params InsertUserParams) (User, error) {
	var u User

	err := q.wrapQuery(insertUser, func() error {
		row := q.db.QueryRowContext(
			ctx,
			insertUser,
			params.Username,
			params.Email,
			params.PasswordHash,
		)

		return row.Scan(
			&u.ID,
			&u.Username,
			&u.Email,
			&u.PasswordHash,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
	})

	return u, err
}

const selectUser = `SELECT * FROM "users" WHERE "id" = ? LIMIT 1`

func (q *Queries) SelectUser(ctx context.Context, id int) (User, error) {
	var u User

	err := q.wrapQuery(selectUser, func() error {
		row := q.db.QueryRowContext(ctx, selectUser, id)

		return row.Scan(
			&u.ID,
			&u.Username,
			&u.Email,
			&u.PasswordHash,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
	})

	return u, err
}

const selectUserByEmail = `SELECT * FROM "users" WHERE "email" = ? LIMIT 1`

func (q *Queries) SelectUserByEmail(ctx context.Context, email string) (User, error) {
	var u User

	err := q.wrapQuery(selectUserByEmail, func() error {
		row := q.db.QueryRowContext(ctx, selectUserByEmail, email)

		return row.Scan(
			&u.ID,
			&u.Username,
			&u.Email,
			&u.PasswordHash,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
	})

	return u, err
}
