package serve

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

func (*Server) respondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) respond(w http.ResponseWriter, status int, res any) {
	body := map[string]any{
		"error": nil,
		"data":  res,
	}

	s.setHeaderAndWrite(w, status, body)
}

func (s *Server) respondError(w http.ResponseWriter, status int, err error) {
	body := map[string]any{
		"error": err.Error(),
		"data":  nil,
	}

	s.setHeaderAndWrite(w, status, body)
}

func (s *Server) setHeaderAndWrite(w http.ResponseWriter, status int, body any) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		s.app.Logger.Errorf("failed to encode to JSON: %v", err)
		http.Error(
			w,
			`{"error":{"code":"INTERNAL_ERROR","message":"failed to encode response"},"data":null}`,
			http.StatusInternalServerError,
		)

		return
	}

	w.WriteHeader(status)
	if _, err := io.Copy(w, &buf); err != nil {
		s.app.Logger.Errorf("failed to write response: %v", err)
	}
}

func decodeJSONBody(r *http.Request, params any) error {
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		return ErrFormParsing
	}

	return nil
}
