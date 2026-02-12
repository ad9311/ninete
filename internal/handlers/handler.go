package handlers

import (
	"html/template"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/alexedwards/scs/v2"
)

type (
	TemplateLookupFunc func(TemplateName) *template.Template
	TemplateReloadFunc func() error
)

type Deps struct {
	App             *prog.App
	Store           *logic.Store
	Session         *scs.SessionManager
	TemplateByName  TemplateLookupFunc
	ReloadTemplates TemplateReloadFunc
}

type Handler struct {
	app             *prog.App
	store           *logic.Store
	session         *scs.SessionManager
	templateByName  TemplateLookupFunc
	reloadTemplates TemplateReloadFunc
}

func New(deps Deps) *Handler {
	return &Handler{
		app:             deps.App,
		store:           deps.Store,
		session:         deps.Session,
		templateByName:  deps.TemplateByName,
		reloadTemplates: deps.ReloadTemplates,
	}
}
