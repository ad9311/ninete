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
VALUES (?, ?, ?, ?)
RETURNING *`

func (q *Queries) InsertRefreshToken(ctx context.Context, arg InsertRefreshTokenParams) (RefreshToken, error) {
	var rt RefreshToken

	err := q.wrapQuery(insertRefreshToken, func() error {
		row := q.db.QueryRowContext(
			ctx,
			insertRefreshToken,
			arg.UserID,
			arg.TokenHash,
			arg.IssuedAt,
			arg.ExpiresAt,
		)

		return row.Scan(
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
DELETE FROM "refresh_tokens" WHERE "token_hash" = ? RETURNING id`

func (q *Queries) DeleteRefreshToken(ctx context.Context, tokenHash []byte) (int, error) {
	var id int

	err := q.wrapQuery(deleteRefreshToken, func() error {
		row := q.db.QueryRowContext(ctx, deleteRefreshToken, tokenHash)

		return row.Scan(&id)
	})

	return id, err
}

const selectRefreshToken = `
SELECT * FROM "refresh_tokens" WHERE "token_hash" = ? LIMIT 1`

func (q *Queries) SelectRefreshToken(ctx context.Context, tokenHash []byte) (RefreshToken, error) {
	var rt RefreshToken

	err := q.wrapQuery(selectRefreshToken, func() error {
		row := q.db.QueryRowContext(ctx, selectRefreshToken, tokenHash)

		return row.Scan(
			&rt.ID,
			&rt.UserID,
			&rt.TokenHash,
			&rt.IssuedAt,
			&rt.ExpiresAt,
		)
	})

	return rt, err
}

const deleteRefreshTokensAt = `
DELETE FROM "refresh_tokens" WHERE "expires_at" <= ?`

func (q *Queries) DeleteRefreshTokensAt(ctx context.Context, nowUnix int64) (int, error) {
	var deleted int64

	err := q.wrapQuery(deleteRefreshTokensAt, func() error {
		result, err := q.db.ExecContext(ctx, deleteRefreshTokensAt, nowUnix)
		if err != nil {
			return err
		}

		deleted, err = result.RowsAffected()
		if err != nil {
			return err
		}

		return nil
	})

	return int(deleted), err
}
