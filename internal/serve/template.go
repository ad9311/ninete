package serve

import (
	"fmt"
	"html/template"
	"path/filepath"
	"strings"

	"github.com/ad9311/ninete/internal/handlers"
)

// viewsPath globs every view. Phase 7 of docs/spa-migration.md left exactly
// one — the SPA shell — which defines its own top-level "layout" template
// rather than sharing one parsed separately, so there is nothing left to glob
// a base or partials from.
const viewsPath = "./web/views/**/*.html"

func (s *Server) LoadTemplates() error {
	views, err := parseTemplates()
	if err != nil {
		return err
	}
	s.templates = views

	return nil
}

func parseTemplates() (map[handlers.TemplateName]*template.Template, error) {
	vc := map[handlers.TemplateName]*template.Template{}

	views, err := filepath.Glob(viewsPath)
	if err != nil {
		return vc, err
	}
	if len(views) == 0 {
		return vc, nil
	}

	for _, v := range views {
		tmpl, err := template.New("layout").ParseFiles(v)
		if err != nil {
			return vc, err
		}

		vc[viewKey(v)] = tmpl
	}

	return vc, nil
}

func viewKey(path string) handlers.TemplateName {
	dir := strings.Split(filepath.Dir(path), "/")
	action := strings.Split(filepath.Base(path), ".")

	return handlers.TemplateName(fmt.Sprintf("%s/%s", dir[len(dir)-1], action[0]))
}

func (s *Server) templateByName(name handlers.TemplateName) *template.Template {
	return s.templates[name]
}
