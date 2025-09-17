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
	Revoked   bool   `json:"revoked"`
}

type InsertRefreshTokenParams struct {
	UserID    int
	TokenHash []byte
	IssuedAt  int64
	ExpiresAt int64
}

const insertRefreshToken = `
INSERT INTO refresh_tokens (user_id, token_hash, issued_at, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *`

func (q *Queries) InsertRefreshToken(ctx context.Context, arg InsertRefreshTokenParams) (RefreshToken, error) {
	var rf RefreshToken
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
			&rf.ID,
			&rf.UserID,
			&rf.TokenHash,
			&rf.IssuedAt,
			&rf.ExpiresAt,
			&rf.Revoked,
		)
	})

	return rf, err
}

const deleteRefreshTokens = `
DELETE FROM refresh_tokens WHERE user_id = $1 AND token_hash = $2 RETURNING id`

func (q *Queries) DeleteRefreshToken(ctx context.Context, userID int, tokenHash []byte) (int, error) {
	var id int
	var err error

	q.wrapQuery(deleteRefreshTokens, func() {
		row := q.db.QueryRowContext(ctx, deleteRefreshTokens, userID, tokenHash)
		err = row.Scan(&id)
	})

	return id, err
}
