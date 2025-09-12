package serve

import "net/http"

func (s *Server) getHealthz(w http.ResponseWriter, _ *http.Request) {
	s.respondNoContent(w)
}

func (s *Server) getReadyz(w http.ResponseWriter, _ *http.Request) {
	s.respond(w, http.StatusOK, "OK")
}
