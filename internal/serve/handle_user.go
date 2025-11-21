package serve

import (
	"fmt"
	"net/http"

	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
)

func (s *Server) GetMe(w http.ResponseWriter, r *http.Request) {
	user := getUserContext(r)

	s.respond(w, http.StatusOK, user)
}

func getUserContext(r *http.Request) *repo.SafeUser {
	user, ok := r.Context().Value(prog.KeyCurrentUser).(*repo.SafeUser)

	if !ok {
		panic(fmt.Sprintf("failed to extract user context, %v", user))
	}

	return user
}
