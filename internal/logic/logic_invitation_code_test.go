package logic_test

import (
	"crypto/sha256"
	"encoding/hex"
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
				invitationCode, err := s.Store.CreateInvitationCode(ctx, "invitation_code_1")
				require.NoError(t, err)
				require.Positive(t, invitationCode.ID)
				require.NotEmpty(t, invitationCode.CodeHash)
				require.Equal(
					t,
					hashString("invitation_code_1"),
					invitationCode.CodeFingerprint,
				)
				require.NotZero(t, invitationCode.CreatedAt)
				require.NotZero(t, invitationCode.UpdatedAt)
			},
		},
		{
			name: "should_fail_with_duplicate_code",
			fn: func(t *testing.T) {
				_, err := s.Store.CreateInvitationCode(ctx, "invitation_code_2")
				require.NoError(t, err)

				_, err = s.Store.CreateInvitationCode(ctx, " INVITATION_CODE_2 ")
				require.ErrorIs(t, err, logic.ErrInvitationCodeExists)
			},
		},
		{
			name: "should_fail_validation_when_code_is_empty",
			fn: func(t *testing.T) {
				_, err := s.Store.CreateInvitationCode(ctx, "   ")
				require.ErrorIs(t, err, logic.ErrValidationFailed)
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
				s.CreateInvitationCode(t, "invitation_code_3")

				err := s.Store.ValidateInvitationCode(ctx, "invitation_code_3")
				require.NoError(t, err)
			},
		},
		{
			name: "should_validate_with_normalized_input",
			fn: func(t *testing.T) {
				s.CreateInvitationCode(t, "invitation_code_4")

				err := s.Store.ValidateInvitationCode(ctx, " INVITATION_CODE_4 ")
				require.NoError(t, err)
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
