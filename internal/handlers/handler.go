package handlers

import (
	"net/http"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/alexedwards/scs/v2"
)

type (
	RenderFunc       func(http.ResponseWriter, int, TemplateName, map[string]any)
	TemplateDataFunc func(*http.Request) map[string]any
)

type Deps struct {
	App          *prog.App
	Store        *logic.Store
	Session      *scs.SessionManager
	Render       RenderFunc
	TemplateData TemplateDataFunc
}

type Handler struct {
	app      *prog.App
	store    *logic.Store
	session  *scs.SessionManager
	render   RenderFunc
	tmplData TemplateDataFunc
}

func New(deps Deps) *Handler {
	return &Handler{
		app:      deps.App,
		store:    deps.Store,
		session:  deps.Session,
		render:   deps.Render,
		tmplData: deps.TemplateData,
	}
}

func (h *Handler) renderError(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	tmplName TemplateName,
	err error,
) {
	data := h.tmplData(r)
	data["error"] = err.Error()
	h.render(w, status, tmplName, data)
}
