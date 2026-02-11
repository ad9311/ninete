package serve

import (
	"bytes"
	"net/http"

	"github.com/ad9311/ninete/internal/handlers"
)

func (s *Server) renderTemplate(
	w http.ResponseWriter,
	status int,
	tmplName handlers.TemplateName,
	data map[string]any,
) {
	if s.app.IsDevelopment() {
		s.app.Logger.Log("Rebuilding templates...")
		err := s.LoadTemplates()
		if err != nil {
			s.app.Logger.Errorf("failed to reload templates: %v", err)
			http.Error(w, `ERROR EXECUTING TEMPLATE`, http.StatusInternalServerError)

			return
		}
	}

	template := s.templates[tmplName]
	buff := new(bytes.Buffer)

	err := template.Execute(buff, data)
	if err != nil {
		s.app.Logger.Errorf("failed to write template: %v", err)
		http.Error(w, `ERROR EXECUTING TEMPLATE`, http.StatusInternalServerError)

		return
	}

	w.WriteHeader(status)
	if _, err := buff.WriteTo(w); err != nil {
		s.app.Logger.Errorf("failed to write response: %v", err)
	}
}

func (s *Server) tmplData(r *http.Request) map[string]any {
	templateMap, ok := r.Context().Value(handlers.TemplateData).(map[string]any)
	if !ok {
		panic("failed to retrieve template data map")
	}

	return templateMap
}
