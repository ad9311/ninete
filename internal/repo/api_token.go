package repo

import (
	"context"
	"database/sql"
)

// APIToken is a personal access token as stored. The plaintext never reaches
// this struct: TokenHash is its SHA-256, and LastFour its last four characters,
// which the settings page shows masked.
type APIToken struct {
	ID         int
	UserID     int
	Name       string
	TokenHash  string
	LastFour   string
	Scope      string
	ExpiresAt  sql.NullInt64
	LastUsedAt sql.NullInt64
	RevokedAt  sql.NullInt64
	CreatedAt  int64
	UpdatedAt  int64
}

type InsertAPITokenParams struct {
	UserID    int
	Name      string
	TokenHash string
	LastFour  string
	Scope     string
	ExpiresAt sql.NullInt64
}

// apiTokenColumns pins the projection order the Scan calls in this file depend
// on. SELECT * would resolve to whatever order the table happens to have, so an
// ALTER TABLE could shift values into the wrong struct fields with no error.
const apiTokenColumns = `"id", "user_id", "name", "token_hash", "last_four", "scope", ` +
	`"expires_at", "last_used_at", "revoked_at", "created_at", "updated_at"`

func scanAPIToken(row interface{ Scan(dest ...any) error }, t *APIToken) error {
	return row.Scan(
		&t.ID,
		&t.UserID,
		&t.Name,
		&t.TokenHash,
		&t.LastFour,
		&t.Scope,
		&t.ExpiresAt,
		&t.LastUsedAt,
		&t.RevokedAt,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
}

// selectAPITokensByUser leaves revoked tokens out: revoking is the settings
// page's "remove", and the row is kept only so the hash keeps failing the
// lookup below. Expired tokens stay listed so the page can say they expired.
const selectAPITokensByUser = `SELECT ` + apiTokenColumns + `
FROM "api_tokens"
WHERE "user_id" = ? AND "revoked_at" IS NULL
ORDER BY "created_at" DESC, "id" DESC`

func (q *Queries) SelectAPITokensByUser(ctx context.Context, userID int) ([]APIToken, error) {
	var tokens []APIToken

	err := q.wrapQuery(selectAPITokensByUser, func() error {
		rows, err := q.db.QueryContext(ctx, selectAPITokensByUser, userID)
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				q.app.Logger.Error(closeErr)
			}
		}()

		for rows.Next() {
			var t APIToken
			if err := scanAPIToken(rows, &t); err != nil {
				return err
			}

			tokens = append(tokens, t)
		}

		return rows.Err()
	})

	return tokens, err
}

// selectAPITokenByHash is the authentication lookup. It returns the row
// whatever its state — revoked or expired is the caller's decision, not a
// missing row — and is the one query here not scoped by "user_id", because
// the token is what identifies the user.
const selectAPITokenByHash = `SELECT ` + apiTokenColumns + `
FROM "api_tokens" WHERE "token_hash" = ?`

func (q *Queries) SelectAPITokenByHash(ctx context.Context, tokenHash string) (APIToken, error) {
	var t APIToken

	err := q.wrapQuery(selectAPITokenByHash, func() error {
		row := q.db.QueryRowContext(ctx, selectAPITokenByHash, tokenHash)

		return scanAPIToken(row, &t)
	})

	return t, err
}

// countActiveAPITokens counts what the per-user limit applies to: neither
// revoked nor expired.
//
//nolint:gosec // G101: SQL naming a token table, not a credential
const countActiveAPITokens = `SELECT COUNT(*) FROM "api_tokens"
WHERE "user_id" = ? AND "revoked_at" IS NULL
AND ("expires_at" IS NULL OR "expires_at" > ?)`

func (q *TxQueries) CountActiveAPITokens(ctx context.Context, userID int, now int64) (int, error) {
	var count int

	err := q.wrapQuery(countActiveAPITokens, func() error {
		row := q.tx.QueryRowContext(ctx, countActiveAPITokens, userID, now)

		return row.Scan(&count)
	})

	return count, err
}

const insertAPIToken = `
INSERT INTO "api_tokens" ("user_id", "name", "token_hash", "last_four", "scope", "expires_at")
VALUES (?, ?, ?, ?, ?, ?)
RETURNING ` + apiTokenColumns

func (q *TxQueries) InsertAPIToken(ctx context.Context, params InsertAPITokenParams) (APIToken, error) {
	var t APIToken

	err := q.wrapQuery(insertAPIToken, func() error {
		row := q.tx.QueryRowContext(
			ctx,
			insertAPIToken,
			params.UserID,
			params.Name,
			params.TokenHash,
			params.LastFour,
			params.Scope,
			params.ExpiresAt,
		)

		return scanAPIToken(row, &t)
	})

	return t, err
}

// revokeAPIToken is scoped by "user_id" so an id belonging to another account
// matches no row, and skips a token already revoked so revoking twice reads as
// "not found" rather than moving its revocation time.
//
//nolint:gosec // G101: SQL naming a token table, not a credential
const revokeAPIToken = `
UPDATE "api_tokens" SET "revoked_at" = ?, "updated_at" = ?
WHERE "id" = ? AND "user_id" = ? AND "revoked_at" IS NULL`

// RevokeAPIToken returns the number of rows it revoked, zero when the token
// does not exist, belongs to someone else, or was already revoked.
func (q *Queries) RevokeAPIToken(ctx context.Context, id, userID int) (int64, error) {
	var affected int64

	err := q.wrapQuery(revokeAPIToken, func() error {
		now := newUpdatedAt()

		res, err := q.db.ExecContext(ctx, revokeAPIToken, now, now, id, userID)
		if err != nil {
			return err
		}

		affected, err = res.RowsAffected()

		return err
	})

	return affected, err
}

// updateAPITokenLastUsed deliberately leaves "updated_at" alone: using a token
// is not editing it.
//
//nolint:gosec // G101: SQL naming a token table, not a credential
const updateAPITokenLastUsed = `
UPDATE "api_tokens" SET "last_used_at" = ? WHERE "id" = ? AND "user_id" = ?`

func (q *Queries) UpdateAPITokenLastUsed(ctx context.Context, id, userID int, usedAt int64) error {
	return q.wrapQuery(updateAPITokenLastUsed, func() error {
		_, err := q.db.ExecContext(ctx, updateAPITokenLastUsed, usedAt, id, userID)

		return err
	})
}
