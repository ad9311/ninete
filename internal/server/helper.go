package server

import (
	"encoding/json"
	"net/http"

	"github.com/ad9311/go-api-base/internal/app"
	"github.com/ad9311/go-api-base/internal/console"
	"github.com/ad9311/go-api-base/internal/errs"
	"github.com/ad9311/go-api-base/internal/service"
)

// Data is an alias for response data map
type Data map[string]any

// Response is the structure of a successful response
type Response struct {
	Code Code `json:"code"`
	Data any  `json:"data"`
}

// ErrorResponse is the structure of an error response
type ErrorResponse struct {
	Code  Code   `json:"code"`
	Error string `json:"error"`
}

// write sends a JSON response with the given status code and data
func write(w http.ResponseWriter, status int, res any) {
	body := map[string]any{
		"code": successCode,
		"data": res,
	}

	setHeaderAndWrite(w, status, body)
}

// writeNoContent sets the 204 status code and renders no content
func writeNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// writeError sends a JSON error response with http status code, interal app code and a list of errors
func writeError(w http.ResponseWriter, status int, code Code, err error) {
	body := map[string]any{
		"code":  code,
		"error": err.Error(),
	}

	setHeaderAndWrite(w, status, body)
}

func setHeaderAndWrite(w http.ResponseWriter, status int, body any) {
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		logger := console.New(nil, nil, false)
		logger.Error("failed to encode response: %v", err)

		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(`{"code":"INTERNAL_ERROR","error":"error writing response"}`))
		if err != nil {
			logger.Error("failed to write response: %v", err)
		}
	}
}

func decodeJSONBody(r *http.Request, params any) error {
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		return errs.ErrFormParsing
	}

	return nil
}

// func getUserIDContext(r *http.Request) (int32, bool) {
// 	userID, ok := r.Context().Value(app.CurrentUserIDKey).(int32)

// 	return userID, ok
// }

func getUserContext(r *http.Request) (service.SafeUser, bool) {
	user, ok := r.Context().Value(app.CurrentUserKey).(service.SafeUser)

	return user, ok
}
