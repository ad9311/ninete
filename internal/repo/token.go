package repo

import (
	"context"
)

type RefreshToken struct {
	ID        int    `json:"id"`
	UserID    int    `json:"userId"`
	UUID      []byte `json:"uuid"`
	IssuedAt  int64  `json:"issuedAt"`
	ExpiresAt int64  `json:"expiresAt"`
	Revoked   bool   `json:"revoked"`
}

type InsertRefreshTokenParams struct {
	UserID    int
	UUID      []byte
	IssuedAt  int64
	ExpiresAt int64
}

const insertRefreshToken = `
INSERT INTO refresh_tokens (user_id, issued_at, expires_at)
VALUES ($1, $2, $3)
RETURNING *;`

func (q *Queries) InsertRefreshToken(ctx context.Context, arg InsertRefreshTokenParams) (RefreshToken, error) {
	var rf RefreshToken
	var err error

	q.wrapQuery(insertRefreshToken, func() {
		row := q.db.QueryRowContext(
			ctx,
			insertRefreshToken,
			arg.UserID,
			arg.UUID,
			arg.IssuedAt,
			arg.ExpiresAt,
		)

		err = row.Scan(
			&rf.ID,
			&rf.UserID,
			&rf.UUID,
			&rf.IssuedAt,
			&rf.ExpiresAt,
			&rf.Revoked,
		)
	})

	return rf, err
}
