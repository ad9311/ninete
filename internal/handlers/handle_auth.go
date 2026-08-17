package handlers

import (
	"errors"
	"net/http"

	"github.com/ad9311/ninete/internal/logic"
)

func (h *Handler) GetRegister(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, r, http.StatusOK, RegisterIndex)
}

func (h *Handler) GetLogin(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, r, http.StatusOK, LoginIndex)
}

func (h *Handler) PostRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		h.renderErr(w, r, http.StatusBadRequest, ErrorIndex, err)

		return
	}

	user, err := h.store.SignUp(ctx, logic.SignUpParams{
		Username:             r.FormValue("username"),
		Email:                r.FormValue("email"),
		Password:             r.FormValue("password"),
		PasswordConfirmation: r.FormValue("passwordConfirmation"),
		InvitationCode:       r.FormValue("invitationCode"),
	})
	if err != nil {
		// A sign-up that failed on the insert for anything other than a
		// collision is a server fault, so it must not be rendered back as a
		// rejected form.
		if errors.Is(err, logic.ErrSignUpFailed) {
			h.app.Logger.Errorf("%v", err)
			h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, ErrRegistrationUnavailable)

			return
		}

		h.renderErr(w, r, http.StatusBadRequest, RegisterIndex, err)

		return
	}

	if err := h.session.RenewToken(ctx); err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	h.session.Put(ctx, SessionIsUserSignedIn, true)
	h.session.Put(ctx, SessionUserID, user.ID)

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *Handler) PostLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		h.renderErr(w, r, http.StatusBadRequest, ErrorIndex, err)

		return
	}

	user, err := h.store.Login(ctx, logic.SessionParams{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	})
	if err != nil {
		// A failed lookup is a server fault, so it must not be rendered back as
		// a rejected credential on the login form.
		if errors.Is(err, logic.ErrLoginLookup) {
			h.app.Logger.Errorf("%v", err)
			h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, ErrLoginUnavailable)

			return
		}

		h.renderErr(w, r, http.StatusBadRequest, LoginIndex, err)

		return
	}

	if err := h.session.RenewToken(ctx); err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	h.session.Put(ctx, SessionIsUserSignedIn, true)
	h.session.Put(ctx, SessionUserID, user.ID)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) PostLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.session.Destroy(ctx); err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}
	if err := h.session.RenewToken(ctx); err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func getCurrentUser(r *http.Request) *logic.User {
	user, ok := r.Context().Value(KeyCurrentUser).(*logic.User)

	if !ok {
		panic("failed to get user context")
	}

	return user
}
