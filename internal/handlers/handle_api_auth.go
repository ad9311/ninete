package handlers

import (
	"errors"
	"net/http"

	"github.com/ad9311/ninete/internal/logic"
)

type apiLoginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type apiRegisterBody struct {
	Username             string `json:"username"`
	Email                string `json:"email"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
	InvitationCode       string `json:"invitation_code"`
}

// PostAPILogin signs a person in (Phase 6 of docs/spa-migration.md): the
// session write the retired form post used to do, answered with a 204 instead
// of a redirect. The client does a full
// page load on success rather than staying inside the SPA's client state —
// that is what actually resets the session boundary; this handler only has
// to leave the session itself correct.
func (h *Handler) PostAPILogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body apiLoginBody
	if !h.DecodeJSON(w, r, &body) {
		return
	}

	user, err := h.store.Login(ctx, logic.SessionParams{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		// A lookup that failed for any reason other than "no such row" is a
		// server fault, not a rejected credential — see PostLogin.
		if errors.Is(err, logic.ErrLoginLookup) {
			h.app.Logger.Errorf("%v", err)
			h.WriteJSONError(w, http.StatusInternalServerError, ErrLoginUnavailable)

			return
		}

		h.WriteAPIError(w, err, logic.ErrWrongEmailOrPassword)

		return
	}

	if err := h.session.RenewToken(ctx); err != nil {
		h.WriteAPIError(w, err)

		return
	}

	h.session.Put(ctx, SessionIsUserSignedIn, true)
	h.session.Put(ctx, SessionUserID, user.ID)

	w.WriteHeader(http.StatusNoContent)
}

// PostAPIRegister is PostAPILogin's sign-up counterpart.
func (h *Handler) PostAPIRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body apiRegisterBody
	if !h.DecodeJSON(w, r, &body) {
		return
	}

	user, err := h.store.SignUp(ctx, logic.SignUpParams{
		Username:             body.Username,
		Email:                body.Email,
		Password:             body.Password,
		PasswordConfirmation: body.PasswordConfirmation,
		InvitationCode:       body.InvitationCode,
	})
	if err != nil {
		// An insert that failed for anything other than a collision is a
		// server fault — see PostRegister.
		if errors.Is(err, logic.ErrSignUpFailed) {
			h.app.Logger.Errorf("%v", err)
			h.WriteJSONError(w, http.StatusInternalServerError, ErrRegistrationUnavailable)

			return
		}

		h.WriteAPIError(w, err,
			logic.ErrPasswordConfirmation,
			logic.ErrAccountExists,
			logic.ErrInvalidInvitationCode,
		)

		return
	}

	if err := h.session.RenewToken(ctx); err != nil {
		h.WriteAPIError(w, err)

		return
	}

	h.session.Put(ctx, SessionIsUserSignedIn, true)
	h.session.Put(ctx, SessionUserID, user.ID)

	w.WriteHeader(http.StatusNoContent)
}
