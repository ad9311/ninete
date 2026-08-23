package handlers_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ad9311/ninete/internal/handlers"
	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/stretchr/testify/require"
)

type apiErrorParams struct {
	Name          string `validate:"required"`
	SaturatedFatG int    `validate:"gte=0"`
}

func apiErrorHandler(t *testing.T) *handlers.Handler {
	t.Helper()

	return handlers.New(handlers.Deps{
		App: &prog.App{
			ENV: prog.ENVTest,
			Logger: prog.NewLogger(prog.LogOptions{
				EnableColor: false,
				EnableQuery: false,
			}),
		},
	})
}

func writeAPIError(t *testing.T, err error, userErrors ...error) (int, handlers.APIError) {
	t.Helper()

	rec := httptest.NewRecorder()
	apiErrorHandler(t).WriteAPIError(rec, err, userErrors...)

	var body handlers.APIError
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))

	return rec.Code, body
}

func TestWriteAPIError(t *testing.T) {
	store := logic.New(nil, repo.Queries{})
	validationErr := store.ValidateStruct(apiErrorParams{SaturatedFatG: -1})
	require.Error(t, validationErr)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_answer_a_validation_failure_with_422_and_the_field_rules",
			fn: func(t *testing.T) {
				status, body := writeAPIError(t, validationErr)

				require.Equal(t, http.StatusUnprocessableEntity, status)
				require.Equal(t, map[string]string{
					"name":            "required",
					"saturated_fat_g": "gte",
				}, body.Fields)
				require.Equal(t, validationErr.Error(), body.Error)
			},
		},
		{
			name: "should_answer_a_missing_row_with_404",
			fn: func(t *testing.T) {
				status, body := writeAPIError(t, fmt.Errorf("select food: %w", sql.ErrNoRows))

				require.Equal(t, http.StatusNotFound, status)
				require.Equal(t, "resource not found", body.Error)
				require.Empty(t, body.Fields)
			},
		},
		{
			name: "should_answer_a_named_user_error_with_its_own_message",
			fn: func(t *testing.T) {
				err := fmt.Errorf("sign up: %w", logic.ErrAccountExists)
				status, body := writeAPIError(t, err, logic.ErrAccountExists)

				require.Equal(t, http.StatusUnprocessableEntity, status)
				require.Equal(t, logic.ErrAccountExists.Error(), body.Error)
			},
		},
		{
			// The one that matters. Today a failed insert re-renders the form
			// with the driver's message, so a UNIQUE violation names the table
			// and column on a public page. Nothing is user-facing by default
			// here.
			name: "should_not_leak_an_unnamed_error_to_the_client",
			fn: func(t *testing.T) {
				// The driver's error is dynamic; a static one stands in for
				// it so the message under test is the one that must not leak.
				dbErr := fmt.Errorf("UNIQUE constraint failed: foods.name: %w", sql.ErrConnDone)
				status, body := writeAPIError(t, dbErr)

				require.Equal(t, http.StatusInternalServerError, status)
				require.NotContains(t, body.Error, "constraint")
				require.Equal(t, "the request could not be completed, please try again", body.Error)
			},
		},
		{
			// A user error the endpoint did not name is still a fault, not a
			// message to pass through.
			name: "should_treat_an_unnamed_user_error_as_a_fault",
			fn: func(t *testing.T) {
				err := fmt.Errorf("sign up: %w", logic.ErrAccountExists)
				status, body := writeAPIError(t, err)

				require.Equal(t, http.StatusInternalServerError, status)
				require.NotContains(t, body.Error, logic.ErrAccountExists.Error())
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}
