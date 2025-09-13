package repo

import (
	"context"
	"database/sql"
)

type User struct {
	ID           int
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    string
	UpdatedAt    string
	D            sql.NullBool
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
INSERT INTO users (username, email, password_hash)
VALUES ($1, $2, $3)
RETURNING *
`

func (q *Queries) InsertUser(ctx context.Context, params InsertUserParams) (User, error) {
	var u User
	row := q.db.QueryRowContext(
		ctx,
		insertUser,
		params.Username,
		params.Email,
		params.PasswordHash,
	)

	err := row.Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	return u, err
}
