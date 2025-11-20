package serve

import "net/http"

func (s *Server) GetHealthz(w http.ResponseWriter, _ *http.Request) {
	s.respondNoContent(w)
}

func (s *Server) GetReadyz(w http.ResponseWriter, _ *http.Request) {
	stats, err := s.store.AppStatus()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, err)

		return
	}

	s.respond(w, http.StatusOK, stats)
}
