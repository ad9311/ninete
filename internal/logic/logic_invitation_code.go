package logic

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"golang.org/x/crypto/bcrypt"
)

type invitationCodeParams struct {
	Code string `validate:"required"`
}

func (s *Store) CreateInvitationCode(ctx context.Context, rawCode string) (repo.InvitationCode, error) {
	var invitationCode repo.InvitationCode

	params := invitationCodeParams{
		Code: prog.NormalizeLowerTrim(rawCode),
	}
	if err := s.ValidateStruct(params); err != nil {
		return invitationCode, err
	}

	fingerprint := invitationCodeFingerprint(params.Code)
	_, err := s.queries.SelectInvitationCodeByFingerprint(ctx, fingerprint)
	if err == nil {
		return invitationCode, ErrInvitationCodeExists
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return invitationCode, err
	}

	codeHash, err := HashPassword(params.Code)
	if err != nil {
		return invitationCode, err
	}

	invitationCode, err = s.queries.InsertInvitationCode(ctx, repo.InsertInvitationCodeParams{
		CodeHash:        codeHash,
		CodeFingerprint: fingerprint,
	})
	if err != nil {
		return invitationCode, err
	}

	return invitationCode, nil
}

func (s *Store) ValidateInvitationCode(ctx context.Context, rawCode string) error {
	params := invitationCodeParams{
		Code: prog.NormalizeLowerTrim(rawCode),
	}
	if err := s.ValidateStruct(params); err != nil {
		return err
	}

	fingerprint := invitationCodeFingerprint(params.Code)
	invitationCode, err := s.queries.SelectInvitationCodeByFingerprint(ctx, fingerprint)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidInvitationCode
		}

		return err
	}

	if err := compareInvitationCode(params.Code, invitationCode.CodeHash); err != nil {
		return err
	}

	return nil
}

func compareInvitationCode(rawCode, codeHash string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(codeHash), []byte(rawCode)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidInvitationCode
		}

		return fmt.Errorf("%w: %v", ErrInvitationCodeVerify, err)
	}

	return nil
}

func invitationCodeFingerprint(code string) string {
	sum := sha256.Sum256([]byte(code))

	return hex.EncodeToString(sum[:])
}
