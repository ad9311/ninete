package logic_test

import (
	"strings"
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestLogin(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_login_with_valid_credentials",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(
					t,
					"login_user_1",
					"login_user_1@example.com",
					"login_password_1",
				)

				loggedUser, err := s.Store.Login(ctx, logic.SessionParams{
					Email:    user.Email,
					Password: "login_password_1",
				})
				require.NoError(t, err)
				require.Equal(t, user.ID, loggedUser.ID)
				require.Equal(t, user.Username, loggedUser.Username)
				require.Equal(t, user.Email, loggedUser.Email)
				require.NotEmpty(t, loggedUser.PasswordHash)
			},
		},
		{
			name: "should_login_with_normalized_email",
			fn: func(t *testing.T) {
				user := s.CreateAuthUser(
					t,
					"login_user_2",
					"login_user_2@example.com",
					"login_password_2",
				)

				loggedUser, err := s.Store.Login(ctx, logic.SessionParams{
					Email:    " LOGIN_USER_2@EXAMPLE.COM ",
					Password: "login_password_2",
				})
				require.NoError(t, err)
				require.Equal(t, user.ID, loggedUser.ID)
			},
		},
		{
			name: "should_fail_when_user_not_found",
			fn: func(t *testing.T) {
				_, err := s.Store.Login(ctx, logic.SessionParams{
					Email:    "missing_user_1@example.com",
					Password: "login_password_3",
				})
				require.ErrorIs(t, err, logic.ErrWrongEmailOrPassword)
			},
		},
		{
			name: "should_fail_with_wrong_password",
			fn: func(t *testing.T) {
				s.CreateAuthUser(
					t,
					"login_user_3",
					"login_user_3@example.com",
					"login_password_3",
				)

				_, err := s.Store.Login(ctx, logic.SessionParams{
					Email:    "login_user_3@example.com",
					Password: "wrong_password_1",
				})
				require.ErrorIs(t, err, logic.ErrWrongEmailOrPassword)
			},
		},
		{
			name: "should_fail_validation_with_invalid_email",
			fn: func(t *testing.T) {
				_, err := s.Store.Login(ctx, logic.SessionParams{
					Email:    "invalid_email",
					Password: "login_password_4",
				})
				require.ErrorIs(t, err, logic.ErrValidationFailed)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestSignUp(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_signup_user_with_valid_params",
			fn: func(t *testing.T) {
				s.CreateInvitationCode(t, "invite_code_1")

				user, err := s.Store.SignUp(ctx, logic.SignUpParams{
					Username:             "newuser1",
					Email:                "new_user_1@example.com",
					Password:             "signup_password_1",
					PasswordConfirmation: "signup_password_1",
					InvitationCode:       "invite_code_1",
				})
				require.NoError(t, err)
				require.Positive(t, user.ID)
				require.Equal(t, "newuser1", user.Username)
				require.Equal(t, "new_user_1@example.com", user.Email)
			},
		},
		{
			name: "should_signup_with_normalized_email_username_and_invitation_code",
			fn: func(t *testing.T) {
				s.CreateInvitationCode(t, "invite_code_2")

				user, err := s.Store.SignUp(ctx, logic.SignUpParams{
					Username:             " NEWUSER2 ",
					Email:                " NEW_USER_2@EXAMPLE.COM ",
					Password:             "signup_password_2",
					PasswordConfirmation: "signup_password_2",
					InvitationCode:       " INVITE_CODE_2 ",
				})
				require.NoError(t, err)
				require.Equal(t, "newuser2", user.Username)
				require.Equal(t, "new_user_2@example.com", user.Email)
			},
		},
		{
			name: "should_fail_when_password_confirmation_does_not_match",
			fn: func(t *testing.T) {
				_, err := s.Store.SignUp(ctx, logic.SignUpParams{
					Username:             "newuser3",
					Email:                "new_user_3@example.com",
					Password:             "signup_password_3",
					PasswordConfirmation: "different_password_3",
					InvitationCode:       "invite_code_3",
				})
				require.ErrorIs(t, err, logic.ErrPasswordConfirmation)
			},
		},
		{
			name: "should_fail_when_invitation_code_is_invalid",
			fn: func(t *testing.T) {
				_, err := s.Store.SignUp(ctx, logic.SignUpParams{
					Username:             "newuser4",
					Email:                "new_user_4@example.com",
					Password:             "signup_password_4",
					PasswordConfirmation: "signup_password_4",
					InvitationCode:       "missing_invite_code_1",
				})
				require.ErrorIs(t, err, logic.ErrInvalidInvitationCode)
			},
		},
		{
			name: "should_fail_validation_with_invalid_email",
			fn: func(t *testing.T) {
				_, err := s.Store.SignUp(ctx, logic.SignUpParams{
					Username:             "newuser5",
					Email:                "invalid_email",
					Password:             "signup_password_5",
					PasswordConfirmation: "signup_password_5",
					InvitationCode:       "invite_code_5",
				})
				require.ErrorIs(t, err, logic.ErrValidationFailed)
			},
		},
		{
			name: "should_fail_with_duplicate_email",
			fn: func(t *testing.T) {
				s.CreateInvitationCode(t, "invite_code_6")
				s.CreateInvitationCode(t, "invite_code_7")

				_, err := s.Store.SignUp(ctx, logic.SignUpParams{
					Username:             "newuser6",
					Email:                "new_user_6@example.com",
					Password:             "signup_password_6",
					PasswordConfirmation: "signup_password_6",
					InvitationCode:       "invite_code_6",
				})
				require.NoError(t, err)

				_, err = s.Store.SignUp(ctx, logic.SignUpParams{
					Username:             "newuser7",
					Email:                "new_user_6@example.com",
					Password:             "signup_password_7",
					PasswordConfirmation: "signup_password_7",
					InvitationCode:       "invite_code_7",
				})
				require.Error(t, err)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestHashPassword(t *testing.T) {
	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_hash_valid_password",
			fn: func(t *testing.T) {
				rawPassword := "hash_password_1"

				passwordHash, err := logic.HashPassword(rawPassword)
				require.NoError(t, err)
				require.NotEmpty(t, passwordHash)
				require.NoError(
					t,
					bcrypt.CompareHashAndPassword(passwordHash, []byte(rawPassword)),
				)
			},
		},
		{
			name: "should_fail_when_password_is_too_long",
			fn: func(t *testing.T) {
				rawPassword := strings.Repeat("a", 73)

				_, err := logic.HashPassword(rawPassword)
				require.ErrorIs(t, err, logic.ErrWithPasswords)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}
