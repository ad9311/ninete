package serve

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// respondNoContent sets the 204 status code and renders no content
func (*Server) respondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// respond sends a standardized JSON response to the client with the specified HTTP status code and response data.
// It is used for successful responses.
func (s *Server) respond(w http.ResponseWriter, status int, res any) {
	body := map[string]any{
		"error": nil,
		"data":  res,
	}

	s.setHeaderAndWrite(w, status, body)
}

// respondError sends a standardized JSON error response to the client with the specified HTTP status code,
// error code, and error message.
// It is used for error responses.
// func (s *Server) respondError(w http.ResponseWriter, status int, code string, err error) {
// 	body := map[string]any{
// 		"error": map[string]string{
// 			"code":    code,
// 			"message": err.Error(),
// 		},
// 		"data": nil,
// 	}

// 	s.setHeaderAndWrite(w, status, body)
// }

// setHeaderAndWrite encodes the provided body as JSON, sets the HTTP status code,
// and writes the response to the provided http.ResponseWriter. If JSON encoding fails,
// it logs the error and writes a standardized internal error response with status 500.
// Any errors encountered during writing the response are also logged.
func (s *Server) setHeaderAndWrite(w http.ResponseWriter, status int, body any) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		s.app.Logger.Error("failed to encode to JSON: %v", err)
		http.Error(
			w,
			`{"error":{"code":"INTERNAL_ERROR","message":"failed to encode response"},"data":null}`,
			http.StatusInternalServerError,
		)

		return
	}

	w.WriteHeader(status)
	if _, err := io.Copy(w, &buf); err != nil {
		s.app.Logger.Error("failed to write response: %v", err)
	}
}
