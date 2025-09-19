package repo

import (
	"context"
)

type RefreshToken struct {
	ID        int    `json:"id"`
	UserID    int    `json:"userId"`
	TokenHash []byte `json:"token_hash"`
	IssuedAt  int64  `json:"issuedAt"`
	ExpiresAt int64  `json:"expiresAt"`
}

type InsertRefreshTokenParams struct {
	UserID    int
	TokenHash []byte
	IssuedAt  int64
	ExpiresAt int64
}

const insertRefreshToken = `
INSERT INTO "refresh_tokens" ("user_id", "token_hash", "issued_at", "expires_at")
VALUES ($1, $2, $3, $4)
RETURNING *`

func (q *Queries) InsertRefreshToken(ctx context.Context, arg InsertRefreshTokenParams) (RefreshToken, error) {
	var rt RefreshToken
	var err error

	q.wrapQuery(insertRefreshToken, func() {
		row := q.db.QueryRowContext(
			ctx,
			insertRefreshToken,
			arg.UserID,
			arg.TokenHash,
			arg.IssuedAt,
			arg.ExpiresAt,
		)

		err = row.Scan(
			&rt.ID,
			&rt.UserID,
			&rt.TokenHash,
			&rt.IssuedAt,
			&rt.ExpiresAt,
		)
	})

	return rt, err
}

const deleteRefreshToken = `
DELETE FROM "refresh_tokens" WHERE "token_hash" = $1 RETURNING id`

func (q *Queries) DeleteRefreshToken(ctx context.Context, tokenHash []byte) (int, error) {
	var id int
	var err error

	q.wrapQuery(deleteRefreshToken, func() {
		row := q.db.QueryRowContext(ctx, deleteRefreshToken, tokenHash)
		err = row.Scan(&id)
	})

	return id, err
}

const selectRefreshToken = `
SELECT * FROM "refresh_tokens" WHERE "token_hash" = $1 LIMIT 1`

func (q *Queries) SelectRefreshToken(ctx context.Context, tokenHash []byte) (RefreshToken, error) {
	var rt RefreshToken
	var err error

	q.wrapQuery(selectRefreshToken, func() {
		row := q.db.QueryRowContext(ctx, selectRefreshToken, tokenHash)

		err = row.Scan(
			&rt.ID,
			&rt.UserID,
			&rt.TokenHash,
			&rt.IssuedAt,
			&rt.ExpiresAt,
		)
	})

	return rt, err
}
