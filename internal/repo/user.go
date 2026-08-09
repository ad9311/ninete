package repo

import (
	"context"
)

type User struct {
	ID           int
	Username     string
	Email        string
	PasswordHash []byte
	CreatedAt    int64
	UpdatedAt    int64
}

type InsertUserParams struct {
	Username     string
	Email        string
	PasswordHash []byte
}

// userColumns pins the projection order the Scan calls in this file depend on.
// SELECT * would resolve to whatever order the table happens to have, so an
// ALTER TABLE could shift values into the wrong struct fields with no error.
const userColumns = `"id", "username", "email", "password_hash", "created_at", "updated_at"`

const insertUser = `
INSERT INTO "users" ("username", "email", "password_hash")
VALUES (?, ?, ?)
RETURNING ` + userColumns

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

const selectUser = `SELECT ` + userColumns + ` FROM "users" WHERE "id" = ? LIMIT 1`

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

const selectUserByEmail = `SELECT ` + userColumns + ` FROM "users" WHERE "email" = ? LIMIT 1`

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
