package handlers

import (
	"bytes"
	"net/http"
	"time"
)

const templateExecErr = "ERROR EXECUTING TEMPLATE"

func (h *Handler) render(
	w http.ResponseWriter,
	status int,
	tmplName TemplateName,
	data map[string]any,
) {
	if h.app.IsDevelopment() && time.Since(h.lastReload) > templateReloadInterval {
		h.app.Logger.Log("Rebuilding templates...")
		if err := h.reloadTemplates(); err != nil {
			h.app.Logger.Errorf("failed to reload templates: %v", err)
			http.Error(w, templateExecErr, http.StatusInternalServerError)

			return
		}
		h.lastReload = time.Now()
	}

	view := h.templateByName(tmplName)
	if view == nil {
		h.app.Logger.Errorf("missing template: %s", tmplName)
		http.Error(w, templateExecErr, http.StatusInternalServerError)

		return
	}

	buff := new(bytes.Buffer)
	if err := view.Execute(buff, data); err != nil {
		h.app.Logger.Errorf("failed to write template: %v", err)
		http.Error(w, templateExecErr, http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if _, err := buff.WriteTo(w); err != nil {
		h.app.Logger.Errorf("failed to write response: %v", err)
	}
}

func (h *Handler) renderPage(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	tmplName TemplateName,
) {
	h.render(w, status, tmplName, h.tmplData(r))
}

func (*Handler) tmplData(r *http.Request) map[string]any {
	templateMap, ok := r.Context().Value(KeyTemplateData).(map[string]any)
	if !ok {
		panic("failed to retrieve template data map")
	}

	return templateMap
}
