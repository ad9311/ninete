package handlers

import (
	"net/http"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/go-chi/chi/v5"
)

// apiTokenView is a token as the settings page sees it. The hash never leaves
// the server; the prefix is all the page needs to tell tokens apart.
type apiTokenView struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Prefix     string `json:"prefix"`
	Scope      string `json:"scope"`
	ExpiresAt  *int64 `json:"expires_at"`
	LastUsedAt *int64 `json:"last_used_at"`
	CreatedAt  int64  `json:"created_at"`
	// Expired is computed here rather than by the page, so the page does not
	// have to compare against a client clock that may disagree with the
	// server's.
	Expired bool `json:"expired"`
}

type apiTokensResponse struct {
	Data []apiTokenView `json:"data"`
	// Limit and ExpiryDays are what CreateAPIToken enforces, sent so the form
	// offers exactly the choices the server accepts.
	Limit      int   `json:"limit"`
	ExpiryDays []int `json:"expiry_days"`
}

type apiTokenCreateBody struct {
	Name          string `json:"name"`
	Scope         string `json:"scope"`
	ExpiresInDays int    `json:"expires_in_days"`
}

// apiTokenCreatedResponse carries the plaintext exactly once. Nothing the
// server keeps can reproduce it, so a page that loses it has to revoke the
// token and create another.
type apiTokenCreatedResponse struct {
	Token  apiTokenView `json:"token"`
	Secret string       `json:"secret"`
}

func toAPITokenView(t repo.APIToken, now int64) apiTokenView {
	view := apiTokenView{
		ID:        t.ID,
		Name:      t.Name,
		Prefix:    t.Prefix,
		Scope:     t.Scope,
		CreatedAt: t.CreatedAt,
	}

	if t.ExpiresAt.Valid {
		expiresAt := t.ExpiresAt.Int64
		view.ExpiresAt = &expiresAt
		view.Expired = expiresAt <= now
	}

	if t.LastUsedAt.Valid {
		lastUsedAt := t.LastUsedAt.Int64
		view.LastUsedAt = &lastUsedAt
	}

	return view
}

// GetAPITokens lists the user's tokens. Like every /api/tokens route it is
// reachable from a browser session only — tokenAuth refuses the whole group
// to a bearer token, so a leaked token cannot mint more or hide by revoking
// the others.
func (h *Handler) GetAPITokens(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	tokens, err := h.store.FindAPITokens(ctx, user.ID)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	now := time.Now().Unix()

	views := make([]apiTokenView, 0, len(tokens))
	for _, t := range tokens {
		views = append(views, toAPITokenView(t, now))
	}

	h.WriteJSON(w, http.StatusOK, apiTokensResponse{
		Data:       views,
		Limit:      logic.APITokenLimit,
		ExpiryDays: logic.APITokenExpiryDays(),
	})
}

func (h *Handler) PostAPITokens(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	var body apiTokenCreateBody
	if !h.DecodeJSON(w, r, &body) {
		return
	}

	token, secret, err := h.store.CreateAPIToken(ctx, user.ID, logic.APITokenParams{
		Name:          body.Name,
		Scope:         body.Scope,
		ExpiresInDays: body.ExpiresInDays,
	})
	if err != nil {
		h.WriteAPIError(w, err, logic.ErrAPITokenLimit)

		return
	}

	h.WriteJSON(w, http.StatusCreated, apiTokenCreatedResponse{
		Token:  toAPITokenView(token, time.Now().Unix()),
		Secret: secret,
	})
}

func (h *Handler) DeleteAPIToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	id, err := prog.ParseID(chi.URLParam(r, "id"), "Token")
	if err != nil {
		h.APINotFound(w, r)

		return
	}

	if err := h.store.RevokeAPIToken(ctx, id, user.ID); err != nil {
		h.WriteAPIError(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
