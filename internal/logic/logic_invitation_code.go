package logic

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
)

// invitationCodeBytes is a generated code's entropy. 16 bytes puts guessing
// out of reach, which is why a single SHA-256 is enough to store it — the
// reasoning apiTokenBytes documents. A code chosen by hand had no such floor,
// and the unsalted SHA-256 kept beside its bcrypt hash let anyone holding the
// database crack a short one without touching bcrypt at all.
const invitationCodeBytes = 16

// invitationCodeEncoding renders a code as 26 base32 characters, lower-cased
// on output. Codes are compared after prog.NormalizeLowerTrim, so the alphabet
// must survive lower-casing: base64 would not, and hex would be 32 characters
// to type instead of 26.
var invitationCodeEncoding = base32.StdEncoding.WithPadding(base32.NoPadding) //nolint:gochecknoglobals // static

type invitationCodeParams struct {
	Code string `validate:"required"`
}

// CreateInvitationCode generates a single-use invitation code and returns it
// alongside its plaintext. The plaintext exists only in this return value:
// the caller shows it once, and only its SHA-256 is stored.
func (s *Store) CreateInvitationCode(ctx context.Context) (repo.InvitationCode, string, error) {
	buf := make([]byte, invitationCodeBytes)
	if _, err := rand.Read(buf); err != nil {
		return repo.InvitationCode{}, "", fmt.Errorf("%w: %w", ErrInvitationCodeGeneration, err)
	}

	code := strings.ToLower(invitationCodeEncoding.EncodeToString(buf))

	invitationCode, err := s.queries.InsertInvitationCode(ctx, repo.InsertInvitationCodeParams{
		CodeFingerprint: invitationCodeFingerprint(code),
	})
	if err != nil {
		return repo.InvitationCode{}, "", err
	}

	return invitationCode, code, nil
}

// ValidateInvitationCode reports whether rawCode names an unused code, without
// consuming it. SignUp calls it before paying for the password hash; the code
// is consumed only by the sign-up transaction itself.
func (s *Store) ValidateInvitationCode(ctx context.Context, rawCode string) error {
	params := invitationCodeParams{
		Code: prog.NormalizeLowerTrim(rawCode),
	}
	if err := s.ValidateStruct(params); err != nil {
		return err
	}

	_, err := s.queries.SelectInvitationCodeByFingerprint(ctx, invitationCodeFingerprint(params.Code))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidInvitationCode
		}

		return err
	}

	return nil
}

// consumeInvitationCodeTx deletes the code inside tq's transaction, so the
// code is spent exactly when the account it admitted is created.
func consumeInvitationCodeTx(ctx context.Context, tq *repo.TxQueries, code string) error {
	err := tq.DeleteInvitationCodeByFingerprint(ctx, invitationCodeFingerprint(code))
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalidInvitationCode
	}

	return err
}

func invitationCodeFingerprint(code string) string {
	sum := sha256.Sum256([]byte(code))

	return hex.EncodeToString(sum[:])
}
