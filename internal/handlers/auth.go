package handlers

import (
	"net/http"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/webkeys"
	"github.com/ad9311/ninete/internal/webtmpl"
)

func (h *Handler) GetLogin(w http.ResponseWriter, r *http.Request) {
	h.render(w, http.StatusOK, webtmpl.LoginIndex, h.tmplData(r))
}

func (h *Handler) PostLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		h.renderError(w, r, http.StatusBadRequest, webtmpl.ErrorIndex, err)

		return
	}

	user, err := h.store.Login(ctx, logic.SessionParams{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	})
	if err != nil {
		h.renderError(w, r, http.StatusBadRequest, webtmpl.LoginIndex, err)

		return
	}

	h.session.Put(ctx, webkeys.SessionIsUserSignedIn, true)
	h.session.Put(ctx, webkeys.SessionUserID, user.ID)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) PostLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.session.Destroy(ctx); err != nil {
		h.renderError(w, r, http.StatusInternalServerError, webtmpl.ErrorIndex, err)

		return
	}
	if err := h.session.RenewToken(ctx); err != nil {
		h.renderError(w, r, http.StatusInternalServerError, webtmpl.ErrorIndex, err)

		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
