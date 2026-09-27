package logic_test

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestAPITokens(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()

	user := s.CreateAuthUser(t, "api_token_user", "api_token_user@example.com", "api_token_pass_1")
	other := s.CreateAuthUser(t, "api_token_other", "api_token_other@example.com", "api_token_pass_2")
	limitUser := s.CreateAuthUser(t, "api_token_limit", "api_token_limit@example.com", "api_token_pass_3")

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_store_only_the_hash_and_the_last_four_characters",
			fn: func(t *testing.T) {
				token, secret := s.CreateAPIToken(t, user.ID, "hash only", logic.APITokenScopeRead)

				require.True(t, strings.HasPrefix(secret, "nin_"))
				require.NotEqual(t, secret, token.TokenHash)

				sum := sha256.Sum256([]byte(secret))
				require.Equal(t, hex.EncodeToString(sum[:]), token.TokenHash)
				require.Len(t, token.LastFour, 4)
				require.True(t, strings.HasSuffix(secret, token.LastFour))
				require.Equal(t, "nin_••••"+token.LastFour, logic.MaskedAPIToken(token))
			},
		},
		{
			name: "should_set_an_expiry_only_when_one_is_chosen",
			fn: func(t *testing.T) {
				never, _, err := s.Store.CreateAPIToken(ctx, user.ID, logic.APITokenParams{
					Name: "never", Scope: logic.APITokenScopeRead,
				})
				require.NoError(t, err)
				require.False(t, never.ExpiresAt.Valid)

				month, _, err := s.Store.CreateAPIToken(ctx, user.ID, logic.APITokenParams{
					Name: "month", Scope: logic.APITokenScopeRead, ExpiresInDays: 30,
				})
				require.NoError(t, err)
				require.True(t, month.ExpiresAt.Valid)
				require.InDelta(t, time.Now().AddDate(0, 0, 30).Unix(), month.ExpiresAt.Int64, 5)

				require.NoError(t, s.Store.RevokeAPIToken(ctx, never.ID, user.ID))
				require.NoError(t, s.Store.RevokeAPIToken(ctx, month.ID, user.ID))
			},
		},
		{
			name: "should_reject_invalid_params",
			fn: func(t *testing.T) {
				invalid := []logic.APITokenParams{
					{Name: "   ", Scope: logic.APITokenScopeRead},
					{Name: strings.Repeat("n", 51), Scope: logic.APITokenScopeRead},
					{Name: "bad scope", Scope: "admin"},
					{Name: "bad expiry", Scope: logic.APITokenScopeRead, ExpiresInDays: 7},
				}

				for _, params := range invalid {
					_, _, err := s.Store.CreateAPIToken(ctx, user.ID, params)
					require.ErrorIs(t, err, logic.ErrValidationFailed, "%+v", params)
				}
			},
		},
		{
			name: "should_cap_active_tokens_and_free_a_slot_on_revoke",
			fn: func(t *testing.T) {
				var ids []int

				for range logic.APITokenLimit {
					token, _ := s.CreateAPIToken(t, limitUser.ID, "slot", logic.APITokenScopeRead)
					ids = append(ids, token.ID)
				}

				_, _, err := s.Store.CreateAPIToken(ctx, limitUser.ID, logic.APITokenParams{
					Name: "one too many", Scope: logic.APITokenScopeRead,
				})
				require.ErrorIs(t, err, logic.ErrAPITokenLimit)

				require.NoError(t, s.Store.RevokeAPIToken(ctx, ids[0], limitUser.ID))

				_, _, err = s.Store.CreateAPIToken(ctx, limitUser.ID, logic.APITokenParams{
					Name: "after revoke", Scope: logic.APITokenScopeRead,
				})
				require.NoError(t, err)
			},
		},
		{
			name: "should_authenticate_a_live_token_and_record_its_use",
			fn: func(t *testing.T) {
				created, secret := s.CreateAPIToken(t, user.ID, "live", logic.APITokenScopeWrite)
				require.False(t, created.LastUsedAt.Valid)

				token, err := s.Store.AuthenticateAPIToken(ctx, secret)
				require.NoError(t, err)
				require.Equal(t, created.ID, token.ID)
				require.Equal(t, user.ID, token.UserID)
				require.True(t, token.LastUsedAt.Valid)

				tokens, err := s.Store.FindAPITokens(ctx, user.ID)
				require.NoError(t, err)

				for _, listed := range tokens {
					if listed.ID == created.ID {
						require.True(t, listed.LastUsedAt.Valid, "use was not persisted")
					}
				}
			},
		},
		{
			name: "should_reject_unknown_malformed_and_revoked_tokens",
			fn: func(t *testing.T) {
				_, err := s.Store.AuthenticateAPIToken(ctx, "nin_doesnotexist")
				require.ErrorIs(t, err, logic.ErrAPITokenInvalid)

				_, err = s.Store.AuthenticateAPIToken(ctx, "not-a-token")
				require.ErrorIs(t, err, logic.ErrAPITokenInvalid)

				token, secret := s.CreateAPIToken(t, user.ID, "revoked", logic.APITokenScopeRead)
				require.NoError(t, s.Store.RevokeAPIToken(ctx, token.ID, user.ID))

				_, err = s.Store.AuthenticateAPIToken(ctx, secret)
				require.ErrorIs(t, err, logic.ErrAPITokenInvalid)
			},
		},
		{
			name: "should_reject_an_expired_token",
			fn: func(t *testing.T) {
				// CreateAPIToken only offers future expiries, so the expired
				// row is written directly.
				secret := "nin_expired_token_for_logic_test"
				sum := sha256.Sum256([]byte(secret))

				err := s.Queries.WithTx(ctx, func(tq *repo.TxQueries) error {
					_, err := tq.InsertAPIToken(ctx, repo.InsertAPITokenParams{
						UserID:    other.ID,
						Name:      "expired",
						TokenHash: hex.EncodeToString(sum[:]),
						LastFour:  secret[len(secret)-4:],
						Scope:     logic.APITokenScopeRead,
						ExpiresAt: sql.NullInt64{Int64: time.Now().Add(-time.Hour).Unix(), Valid: true},
					})

					return err
				})
				require.NoError(t, err)

				_, err = s.Store.AuthenticateAPIToken(ctx, secret)
				require.ErrorIs(t, err, logic.ErrAPITokenInvalid)
			},
		},
		{
			name: "should_not_revoke_another_users_token",
			fn: func(t *testing.T) {
				token, secret := s.CreateAPIToken(t, user.ID, "not yours", logic.APITokenScopeRead)

				err := s.Store.RevokeAPIToken(ctx, token.ID, other.ID)
				require.ErrorIs(t, err, sql.ErrNoRows)

				_, err = s.Store.AuthenticateAPIToken(ctx, secret)
				require.NoError(t, err, "a foreign revoke must leave the token working")
			},
		},
		{
			name: "should_list_only_the_users_unrevoked_tokens",
			fn: func(t *testing.T) {
				revoked, _ := s.CreateAPIToken(t, other.ID, "other revoked", logic.APITokenScopeRead)
				require.NoError(t, s.Store.RevokeAPIToken(ctx, revoked.ID, other.ID))

				tokens, err := s.Store.FindAPITokens(ctx, other.ID)
				require.NoError(t, err)

				for _, token := range tokens {
					require.Equal(t, other.ID, token.UserID)
					require.NotEqual(t, revoked.ID, token.ID)
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}
