package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/go-chi/chi/v5"
)

// ----------------------------------------------------------------------------- //
// Context Middleware
// ----------------------------------------------------------------------------- //

func (h *Handler) MacroTemplateContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := getCurrentUser(r)
		id, err := prog.ParseID(chi.URLParam(r, "id"), "MacroTemplate")
		if err != nil {
			h.NotFound(w, r)

			return
		}

		tmpl, err := h.store.FindMacroTemplate(ctx, id, user.ID)
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}
		if err != nil {
			h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

			return
		}

		ctx = context.WithValue(ctx, KeyMacroTemplate, &tmpl)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ----------------------------------------------------------------------------- //
// Handlers
// ----------------------------------------------------------------------------- //

func (h *Handler) GetMacroTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)

	q := r.URL.Query()
	sortField := q.Get("sort_field")
	sortOrder := q.Get("sort_order")
	if sortField == "" {
		sortField = "name"
	}
	if sortOrder == "" {
		sortOrder = "ASC"
	}

	opts := repo.QueryOptions{
		Filters: repo.Filters{
			FilterFields: []repo.FilterField{
				{Name: "user_id", Value: user.ID, Operator: "="},
			},
			Connector: "AND",
		},
		Sorting: repo.Sorting{Field: sortField, Order: sortOrder},
	}

	templates, err := h.store.FindMacroTemplates(ctx, opts)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, MacroTemplatesIndex, err)

		return
	}

	data["templates"] = templates
	data["sortField"] = sortField
	data["sortOrder"] = sortOrder

	h.render(w, http.StatusOK, MacroTemplatesIndex, data)
}

func (h *Handler) GetMacroTemplatesNew(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)

	tmpl := repo.MacroTemplate{}

	if fromEntryStr := r.URL.Query().Get("from_entry"); fromEntryStr != "" {
		entryID, err := prog.ParseID(fromEntryStr, "MacroEntry")
		if err == nil {
			entry, err := h.store.FindMacroEntry(ctx, entryID, user.ID)
			if err == nil {
				tmpl.Name = entry.Name
				tmpl.Kcal = entry.Kcal
				tmpl.ProteinG = entry.ProteinG
				tmpl.CarbsG = entry.CarbsG
				tmpl.FatG = entry.FatG
			}
		}
	}

	data["template"] = tmpl
	data["amountUnitOptions"] = macroAmountUnitOptions()

	h.render(w, http.StatusOK, MacroTemplatesNew, data)
}

func (h *Handler) PostMacroTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)

	params, err := parseMacroTemplateForm(r)
	if err != nil {
		data["template"] = repo.MacroTemplate{}
		data["amountUnitOptions"] = macroAmountUnitOptions()
		h.renderErr(w, r, http.StatusBadRequest, MacroTemplatesNew, err)

		return
	}

	tmpl, err := h.store.CreateMacroTemplate(ctx, user.ID, params)
	if err != nil {
		data["template"] = repo.MacroTemplate{
			Name:       params.Name,
			Kcal:       params.Kcal,
			ProteinG:   params.ProteinG,
			CarbsG:     params.CarbsG,
			FatG:       params.FatG,
			Amount:     params.Amount,
			AmountUnit: params.AmountUnit,
		}
		data["amountUnitOptions"] = macroAmountUnitOptions()
		h.renderErr(w, r, http.StatusBadRequest, MacroTemplatesNew, err)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/macros/templates/%d", tmpl.ID), http.StatusSeeOther)
}

func (h *Handler) GetMacroTemplate(w http.ResponseWriter, r *http.Request) {
	data := h.tmplData(r)
	data["template"] = getMacroTemplate(r)

	h.render(w, http.StatusOK, MacroTemplatesShow, data)
}

func (h *Handler) GetMacroTemplateEdit(w http.ResponseWriter, r *http.Request) {
	data := h.tmplData(r)
	data["template"] = getMacroTemplate(r)
	data["amountUnitOptions"] = macroAmountUnitOptions()

	h.render(w, http.StatusOK, MacroTemplatesEdit, data)
}

func (h *Handler) PostMacroTemplateUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)
	tmpl := *getMacroTemplate(r)

	params, err := parseMacroTemplateForm(r)
	if err != nil {
		data["template"] = tmpl
		data["amountUnitOptions"] = macroAmountUnitOptions()
		h.renderErr(w, r, http.StatusBadRequest, MacroTemplatesEdit, err)

		return
	}

	_, err = h.store.UpdateMacroTemplate(ctx, tmpl.ID, user.ID, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}

		tmpl.Name = params.Name
		tmpl.Kcal = params.Kcal
		tmpl.ProteinG = params.ProteinG
		tmpl.CarbsG = params.CarbsG
		tmpl.FatG = params.FatG
		tmpl.Amount = params.Amount
		tmpl.AmountUnit = params.AmountUnit
		data["template"] = tmpl
		data["amountUnitOptions"] = macroAmountUnitOptions()
		h.renderErr(w, r, http.StatusBadRequest, MacroTemplatesEdit, err)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/macros/templates/%d", tmpl.ID), http.StatusSeeOther)
}

func (h *Handler) PostMacroTemplateDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)
	tmpl := getMacroTemplate(r)

	_, err := h.store.DeleteMacroTemplate(ctx, tmpl.ID, user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}
		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	http.Redirect(w, r, "/macros/templates", http.StatusSeeOther)
}

// ----------------------------------------------------------------------------- //
// Unexported Functions and Helpers
// ----------------------------------------------------------------------------- //

func parseMacroTemplateForm(r *http.Request) (logic.MacroTemplateParams, error) {
	var params logic.MacroTemplateParams

	if err := r.ParseForm(); err != nil {
		return params, fmt.Errorf("%w: %w", ErrParseForm, err)
	}

	kcal, err := parseFloatField(r, "kcal")
	if err != nil {
		return params, err
	}

	proteinG, err := parseFloatField(r, "protein_g")
	if err != nil {
		return params, err
	}

	carbsG, err := parseFloatField(r, "carbs_g")
	if err != nil {
		return params, err
	}

	fatG, err := parseFloatField(r, "fat_g")
	if err != nil {
		return params, err
	}

	amount, err := parseFloatField(r, "amount")
	if err != nil {
		return params, err
	}

	params.Name = r.FormValue("name")
	params.Kcal = kcal
	params.ProteinG = proteinG
	params.CarbsG = carbsG
	params.FatG = fatG
	params.Amount = amount
	params.AmountUnit = r.FormValue("amount_unit")

	return params, nil
}

func getMacroTemplate(r *http.Request) *repo.MacroTemplate {
	tmpl, ok := r.Context().Value(KeyMacroTemplate).(*repo.MacroTemplate)

	if !ok {
		panic("failed to get macro template context")
	}

	return tmpl
}
