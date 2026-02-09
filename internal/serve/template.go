package serve

import (
	"fmt"
	"html/template"
	"path/filepath"
	"strings"

	"github.com/ad9311/ninete/internal/webtmpl"
)

const (
	layoutPath   = "./web/views/layout.html"
	viewsPath    = "./web/views/**/*.html"
	partialsPath = "./web/views/**/_*.html"
)

func (s *Server) LoadTemplates() error {
	views, err := parseTemplates()
	if err != nil {
		return err
	}
	s.templates = views

	return nil
}

func parseTemplates() (map[webtmpl.Name]*template.Template, error) {
	vc := map[webtmpl.Name]*template.Template{}

	views, err := filepath.Glob(viewsPath)
	if err != nil {
		return vc, err
	}

	for _, v := range views {
		file := filepath.Base(v)
		newView, err := template.New(file).Funcs(templateFuncMap()).ParseFiles(v)
		if err != nil {
			return vc, err
		}

		layouts, err := filepath.Glob(layoutPath)
		if err != nil {
			return vc, err
		}

		if len(layouts) == 0 {
			return vc, ErrLayoutNotFound
		}

		partials, err := filepath.Glob(partialsPath)
		if err != nil {
			return vc, err
		}

		newView, err = newView.ParseGlob(layoutPath)
		if err != nil {
			return vc, err
		}

		if len(partials) > 0 {
			newView, err = newView.ParseGlob(partialsPath)
			if err != nil {
				return vc, err
			}
		}

		name := viewKey(v)
		vc[name] = newView
	}

	return vc, nil
}

func viewKey(path string) webtmpl.Name {
	dir := strings.Split(filepath.Dir(path), "/")
	action := strings.Split(filepath.Base(path), ".")

	return webtmpl.Name(fmt.Sprintf("%s/%s", dir[len(dir)-1], action[0]))
}

func templateFuncMap() template.FuncMap {
	return template.FuncMap{}
}
