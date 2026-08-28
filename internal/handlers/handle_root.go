package handlers

import "net/http"

// GetRoot sends the front door to the SPA. Phase 6 of docs/spa-migration.md
// rather than Phase 7, which is the real flip: leaving it on the template
// dashboard meant a person typing the bare domain landed in the old UI while
// every other redirect had already moved. The templates stay reachable by URL,
// which is what keeps them usable as the Phase 3 oracle. Phase 7 deletes this
// handler along with the catch-all that serves the shell from "/".
func (*Handler) GetRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, AppDashboardPath, http.StatusSeeOther)
}
