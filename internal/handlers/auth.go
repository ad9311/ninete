package handlers

import (
	"net/http"

	"github.com/ad9311/ninete/internal/logic"
)

func (h *Handler) GetLogin(w http.ResponseWriter, r *http.Request) {
	h.render(w, http.StatusOK, LoginIndex, h.tmplData(r))
}

func (h *Handler) PostLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		h.renderError(w, r, http.StatusBadRequest, ErrorIndex, err)

		return
	}

	user, err := h.store.Login(ctx, logic.SessionParams{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	})
	if err != nil {
		h.renderError(w, r, http.StatusBadRequest, LoginIndex, err)

		return
	}

	h.session.Put(ctx, SessionIsUserSignedIn, true)
	h.session.Put(ctx, SessionUserID, user.ID)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) PostLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.session.Destroy(ctx); err != nil {
		h.renderError(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}
	if err := h.session.RenewToken(ctx); err != nil {
		h.renderError(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
