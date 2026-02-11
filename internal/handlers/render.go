package handlers

import (
	"bytes"
	"errors"
	"net/http"
)

const templateExecErr = "ERROR EXECUTING TEMPLATE"

var (
	ErrNotAllowed = errors.New("request not allowed")
)

func (h *Handler) render(
	w http.ResponseWriter,
	status int,
	tmplName TemplateName,
	data map[string]any,
) {
	if h.app.IsDevelopment() {
		h.app.Logger.Log("Rebuilding templates...")
		if err := h.reloadTemplates(); err != nil {
			h.app.Logger.Errorf("failed to reload templates: %v", err)
			http.Error(w, templateExecErr, http.StatusInternalServerError)

			return
		}
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

func (h *Handler) renderErr(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	tmplName TemplateName,
	err error,
) {
	data := h.tmplData(r)
	if err != nil {
		data["error"] = err.Error()
	}
	h.render(w, status, tmplName, data)
}

func (*Handler) tmplData(r *http.Request) map[string]any {
	templateMap, ok := r.Context().Value(TemplateData).(map[string]any)
	if !ok {
		panic("failed to retrieve template data map")
	}

	return templateMap
}

func (h *Handler) NotFound(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, r, http.StatusNotFound, NotFoundIndex)
}

func (h *Handler) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	h.renderErr(w, r, http.StatusMethodNotAllowed, ErrorIndex, ErrNotAllowed)
}
