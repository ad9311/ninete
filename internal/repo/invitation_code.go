package repo

import "context"

type InvitationCode struct {
	ID              int
	CodeHash        []byte
	CodeFingerprint string
	CreatedAt       int64
	UpdatedAt       int64
}

type InsertInvitationCodeParams struct {
	CodeHash        []byte
	CodeFingerprint string
}

// invitationCodeColumns pins the projection order the Scan calls in this file depend on.
// SELECT * would resolve to whatever order the table happens to have, so an
// ALTER TABLE could shift values into the wrong struct fields with no error.
const invitationCodeColumns = `"id", "code_hash", "code_fingerprint", "created_at", "updated_at"`

const insertInvitationCode = `
INSERT INTO "invitation_codes" ("code_hash", "code_fingerprint")
VALUES (?, ?)
RETURNING ` + invitationCodeColumns

func (q *Queries) InsertInvitationCode(ctx context.Context, params InsertInvitationCodeParams) (InvitationCode, error) {
	var c InvitationCode

	err := q.wrapQuery(insertInvitationCode, func() error {
		row := q.db.QueryRowContext(ctx, insertInvitationCode, params.CodeHash, params.CodeFingerprint)

		return row.Scan(
			&c.ID,
			&c.CodeHash,
			&c.CodeFingerprint,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
	})

	return c, err
}

const selectInvitationCodeByFingerprint = `
SELECT ` + invitationCodeColumns + ` FROM "invitation_codes"
WHERE "code_fingerprint" = ?
LIMIT 1`

func (q *Queries) SelectInvitationCodeByFingerprint(
	ctx context.Context,
	codeFingerprint string,
) (InvitationCode, error) {
	var c InvitationCode

	err := q.wrapQuery(selectInvitationCodeByFingerprint, func() error {
		row := q.db.QueryRowContext(ctx, selectInvitationCodeByFingerprint, codeFingerprint)

		return row.Scan(
			&c.ID,
			&c.CodeHash,
			&c.CodeFingerprint,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
	})

	return c, err
}
