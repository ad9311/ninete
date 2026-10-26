package repo

import "context"

// InvitationCode is a single-use code that lets one person register. The code
// itself is never stored: CodeFingerprint is its SHA-256, which is enough
// because the code is generated with 128 bits of entropy (see
// logic.CreateInvitationCode), the same reasoning as api_tokens.token_hash.
type InvitationCode struct {
	ID              int
	CodeFingerprint string
	CreatedAt       int64
	UpdatedAt       int64
}

type InsertInvitationCodeParams struct {
	CodeFingerprint string
}

// invitationCodeColumns pins the projection order the Scan calls in this file depend on.
// SELECT * would resolve to whatever order the table happens to have, so an
// ALTER TABLE could shift values into the wrong struct fields with no error.
const invitationCodeColumns = `"id", "code_fingerprint", "created_at", "updated_at"`

const insertInvitationCode = `
INSERT INTO "invitation_codes" ("code_fingerprint")
VALUES (?)
RETURNING ` + invitationCodeColumns

func (q *Queries) InsertInvitationCode(ctx context.Context, params InsertInvitationCodeParams) (InvitationCode, error) {
	var c InvitationCode

	err := q.wrapQuery(insertInvitationCode, func() error {
		row := q.db.QueryRowContext(ctx, insertInvitationCode, params.CodeFingerprint)

		return row.Scan(
			&c.ID,
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
			&c.CodeFingerprint,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
	})

	return c, err
}

const deleteInvitationCodeByFingerprint = `
DELETE FROM "invitation_codes" WHERE "code_fingerprint" = ? RETURNING "id"`

// DeleteInvitationCodeByFingerprint consumes a code. It answers sql.ErrNoRows
// when no such code exists, so inside the sign-up transaction the delete is
// also the check: of two sign-ups racing on one code, only one deletes it.
func (q *TxQueries) DeleteInvitationCodeByFingerprint(ctx context.Context, codeFingerprint string) error {
	var id int

	return q.wrapQuery(deleteInvitationCodeByFingerprint, func() error {
		row := q.tx.QueryRowContext(ctx, deleteInvitationCodeByFingerprint, codeFingerprint)

		return row.Scan(&id)
	})
}
