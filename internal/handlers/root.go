package handlers

import "net/http"

func (*Handler) GetRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
