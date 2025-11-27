package serve

import "net/http"

func (s *Server) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := s.store.FindCategories(r.Context())
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err)
	}

	s.respond(w, http.StatusOK, categories)
}
