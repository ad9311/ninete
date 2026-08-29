package handlers

import (
	"net/http"

	"github.com/ad9311/ninete/internal/logic"
)

// PostLogout is a real form post rather than a fetch call, deliberately:
// Destroy/RenewToken are a session boundary, and a hard navigation is the
// simplest way to reset every piece of client-held state against it — the
// same reasoning routes/login and routes/register use for a successful
// sign-in (§5, Phase 6 of docs/spa-migration.md). Header.svelte posts to this
// directly.
func (h *Handler) PostLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.session.Destroy(ctx); err != nil {
		h.app.Logger.Errorf("failed to destroy session: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}
	if err := h.session.RenewToken(ctx); err != nil {
		h.app.Logger.Errorf("failed to renew session token: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, AppLoginPath, http.StatusSeeOther)
}

func getCurrentUser(r *http.Request) *logic.User {
	user, ok := r.Context().Value(KeyCurrentUser).(*logic.User)

	if !ok {
		panic("failed to get user context")
	}

	return user
}
