package logic

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ad9311/ninete/internal/repo"
)

// Token scopes. A read token may only GET; a write token may also POST and
// PUT. No scope permits DELETE — the bearer chain refuses it outright (see
// tokenAuth in internal/serve), so a leaked token can never destroy data.
const (
	APITokenScopeRead  = "read"
	APITokenScopeWrite = "write"
)

const (
	// APITokenLimit bounds the tokens one user may hold at once. Revoked and
	// expired tokens do not count against it.
	APITokenLimit = 5

	// apiTokenMarker opens every token, so one pasted into the wrong place is
	// recognizable, and a secret scanner can match it.
	apiTokenMarker = "nin_"
	// apiTokenBytes is the token's entropy. 32 bytes puts guessing out of
	// reach, which is why a single SHA-256 is enough to store it: a slow hash
	// exists to protect low-entropy secrets, and this one is not.
	apiTokenBytes = 32
	// apiTokenVisibleChars is how much of the secret the settings page shows,
	// taken from the end and displayed masked: enough to tell tokens apart, and
	// 24 of 256 bits, far too little to authenticate with.
	apiTokenVisibleChars = 4
	// apiTokenMask stands in for the hidden part of a displayed token.
	apiTokenMask = "••••" //nolint:gosec // G101: display mask, not a credential

	// apiTokenTouchInterval throttles the last-used write. Without it every
	// token-authenticated request would cost a database write just to record
	// that it happened.
	apiTokenTouchInterval = int64(60)
)

// APITokenExpiryDays returns the lifetimes the settings form offers, in days.
// Zero means the token never expires. It must agree with the oneof rule on
// APITokenParams.ExpiresInDays.
func APITokenExpiryDays() []int {
	return []int{30, 90, 365, 0}
}

type APITokenParams struct {
	Name          string `validate:"required,max=50"`
	Scope         string `validate:"required,oneof=read write"`
	ExpiresInDays int    `validate:"oneof=0 30 90 365"`
}

// FindAPITokens lists the user's tokens that have not been revoked, newest
// first, expired ones included.
func (s *Store) FindAPITokens(ctx context.Context, userID int) ([]repo.APIToken, error) {
	return s.queries.SelectAPITokensByUser(ctx, userID)
}

// CreateAPIToken mints a token and returns it alongside its plaintext. The
// plaintext exists only in this return value: the caller shows it once, and
// nothing that could reproduce it is stored.
//
// The limit is counted inside the insert's transaction so two concurrent
// creations cannot both pass it.
func (s *Store) CreateAPIToken(
	ctx context.Context,
	userID int,
	params APITokenParams,
) (repo.APIToken, string, error) {
	var token repo.APIToken

	params.Name = strings.TrimSpace(params.Name)
	if err := s.ValidateStruct(params); err != nil {
		return token, "", err
	}

	plaintext, err := generateAPIToken()
	if err != nil {
		return token, "", err
	}

	now := time.Now()

	var expiresAt sql.NullInt64
	if params.ExpiresInDays > 0 {
		expiresAt = sql.NullInt64{
			Int64: now.AddDate(0, 0, params.ExpiresInDays).Unix(),
			Valid: true,
		}
	}

	err = s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		count, err := tq.CountActiveAPITokens(ctx, userID, now.Unix())
		if err != nil {
			return err
		}

		if count >= APITokenLimit {
			return ErrAPITokenLimit
		}

		token, err = tq.InsertAPIToken(ctx, repo.InsertAPITokenParams{
			UserID:    userID,
			Name:      params.Name,
			TokenHash: hashAPIToken(plaintext),
			LastFour:  plaintext[len(plaintext)-apiTokenVisibleChars:],
			Scope:     params.Scope,
			ExpiresAt: expiresAt,
		})

		return err
	})
	if err != nil {
		return repo.APIToken{}, "", err
	}

	return token, plaintext, nil
}

// RevokeAPIToken revokes one of the user's tokens. A token that does not
// exist, belongs to someone else or is already revoked answers sql.ErrNoRows,
// so the three are indistinguishable to the caller.
func (s *Store) RevokeAPIToken(ctx context.Context, id, userID int) error {
	affected, err := s.queries.RevokeAPIToken(ctx, id, userID)
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// AuthenticateAPIToken resolves a bearer token to its stored row. Unknown,
// revoked and expired tokens all answer ErrAPITokenInvalid, so a caller
// cannot tell which it presented.
func (s *Store) AuthenticateAPIToken(ctx context.Context, plaintext string) (repo.APIToken, error) {
	if !strings.HasPrefix(plaintext, apiTokenMarker) {
		return repo.APIToken{}, ErrAPITokenInvalid
	}

	token, err := s.queries.SelectAPITokenByHash(ctx, hashAPIToken(plaintext))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repo.APIToken{}, ErrAPITokenInvalid
		}

		return repo.APIToken{}, err
	}

	now := time.Now().Unix()

	if token.RevokedAt.Valid || (token.ExpiresAt.Valid && token.ExpiresAt.Int64 <= now) {
		return repo.APIToken{}, ErrAPITokenInvalid
	}

	if !token.LastUsedAt.Valid || now-token.LastUsedAt.Int64 >= apiTokenTouchInterval {
		// Recording use is bookkeeping: failing it must not fail a request the
		// token was valid for.
		if err := s.queries.UpdateAPITokenLastUsed(ctx, token.ID, token.UserID, now); err != nil {
			s.app.Logger.Errorf("failed to record api token use: %v", err)
		} else {
			token.LastUsedAt = sql.NullInt64{Int64: now, Valid: true}
		}
	}

	return token, nil
}

// MaskedAPIToken is how a stored token is displayed once its plaintext is
// gone — "nin_••••wTEF", the marker, a mask and the last four characters.
// Built here so the marker is spelled in one place.
func MaskedAPIToken(t repo.APIToken) string {
	return apiTokenMarker + apiTokenMask + t.LastFour
}

func generateAPIToken() (string, error) {
	buf := make([]byte, apiTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("%w: %w", ErrAPITokenGeneration, err)
	}

	return apiTokenMarker + base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashAPIToken(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))

	return hex.EncodeToString(sum[:])
}
