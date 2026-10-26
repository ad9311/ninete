package logic_test

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestCreateInvitationCode(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_create_invitation_code",
			fn: func(t *testing.T) {
				invitationCode, code, err := s.Store.CreateInvitationCode(ctx)
				require.NoError(t, err)
				require.Positive(t, invitationCode.ID)
				require.Equal(t, hashString(code), invitationCode.CodeFingerprint)
				require.NotZero(t, invitationCode.CreatedAt)
				require.NotZero(t, invitationCode.UpdatedAt)
			},
		},
		{
			// 16 random bytes as unpadded base32: 26 characters, already in
			// the lower case SignUp normalizes to.
			name: "should_generate_a_lowercase_26_character_code",
			fn: func(t *testing.T) {
				_, code, err := s.Store.CreateInvitationCode(ctx)
				require.NoError(t, err)
				require.Len(t, code, 26)
				require.Equal(t, strings.ToLower(code), code)
			},
		},
		{
			name: "should_generate_a_different_code_each_time",
			fn: func(t *testing.T) {
				_, first, err := s.Store.CreateInvitationCode(ctx)
				require.NoError(t, err)

				_, second, err := s.Store.CreateInvitationCode(ctx)
				require.NoError(t, err)

				require.NotEqual(t, first, second)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestValidateInvitationCode(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_validate_existing_invitation_code",
			fn: func(t *testing.T) {
				code := s.CreateInvitationCode(t)

				err := s.Store.ValidateInvitationCode(ctx, code)
				require.NoError(t, err)
			},
		},
		{
			name: "should_validate_with_normalized_input",
			fn: func(t *testing.T) {
				code := s.CreateInvitationCode(t)

				err := s.Store.ValidateInvitationCode(ctx, " "+strings.ToUpper(code)+" ")
				require.NoError(t, err)
			},
		},
		{
			name: "should_not_consume_the_code",
			fn: func(t *testing.T) {
				code := s.CreateInvitationCode(t)

				require.NoError(t, s.Store.ValidateInvitationCode(ctx, code))
				require.NoError(t, s.Store.ValidateInvitationCode(ctx, code))
			},
		},
		{
			name: "should_fail_when_invitation_code_does_not_exist",
			fn: func(t *testing.T) {
				err := s.Store.ValidateInvitationCode(ctx, "missing_invitation_code_1")
				require.ErrorIs(t, err, logic.ErrInvalidInvitationCode)
			},
		},
		{
			name: "should_fail_validation_when_code_is_empty",
			fn: func(t *testing.T) {
				err := s.Store.ValidateInvitationCode(ctx, " ")
				require.ErrorIs(t, err, logic.ErrValidationFailed)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func hashString(v string) string {
	sum := sha256.Sum256([]byte(v))

	return hex.EncodeToString(sum[:])
}
