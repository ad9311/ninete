package handlers

import (
	"html/template"
	"time"

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

const templateReloadInterval = 2 * time.Second

type Handler struct {
	app             *prog.App
	store           *logic.Store
	session         *scs.SessionManager
	templateByName  TemplateLookupFunc
	reloadTemplates TemplateReloadFunc
	lastReload      time.Time
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
