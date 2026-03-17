package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/go-chi/chi/v5"
)

type macroProgressData struct {
	KcalPct    int
	ProteinPct int
	CarbsPct   int
	FatPct     int
}

// ----------------------------------------------------------------------------- //
// Context Middleware
// ----------------------------------------------------------------------------- //

func (h *Handler) MacroEntryContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := getCurrentUser(r)
		entryID := chi.URLParam(r, "id")

		id, err := prog.ParseID(entryID, "MacroEntry")
		if err != nil {
			h.NotFound(w, r)

			return
		}

		entry, err := h.store.FindMacroEntry(ctx, id, user.ID)
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}
		if err != nil {
			h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

			return
		}

		ctx = context.WithValue(ctx, KeyMacroEntry, &entry)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ----------------------------------------------------------------------------- //
// Handlers
// ----------------------------------------------------------------------------- //

func (h *Handler) GetMacros(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)

	q := r.URL.Query()
	dayStart, nextDayStart, selectedDate := computeDayWindow(q.Get("date"))

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
				{Name: "date", Value: dayStart, Operator: ">="},
				{Name: "date", Value: nextDayStart, Operator: "<"},
			},
			Connector: "AND",
		},
		Sorting: repo.Sorting{Field: sortField, Order: sortOrder},
	}

	entries, err := h.store.FindMacroEntries(ctx, opts)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, MacrosIndex, err)

		return
	}

	totals, err := h.store.FindMacroDayTotals(ctx, user.ID, dayStart, nextDayStart)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, MacrosIndex, err)

		return
	}

	goal, goalErr := h.store.FindMacroGoal(ctx, user.ID)
	hasGoal := !errors.Is(goalErr, sql.ErrNoRows)
	if goalErr != nil && !errors.Is(goalErr, sql.ErrNoRows) {
		h.renderErr(w, r, http.StatusInternalServerError, MacrosIndex, goalErr)

		return
	}

	data["entries"] = entries
	data["totals"] = totals
	data["selectedDate"] = selectedDate
	data["hasGoal"] = hasGoal
	data["sortField"] = sortField
	data["sortOrder"] = sortOrder

	if hasGoal {
		data["goal"] = goal
		data["progress"] = computeMacroProgress(totals, goal)
	}

	h.render(w, http.StatusOK, MacrosIndex, data)
}

func (h *Handler) GetMacrosNew(w http.ResponseWriter, r *http.Request) {
	data := h.tmplData(r)
	data["entry"] = repo.MacroEntry{}
	data["selectedDate"] = time.Now().UTC().Format("2006-01-02")

	h.render(w, http.StatusOK, MacrosNew, data)
}

func (h *Handler) PostMacros(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)

	params, err := parseMacroEntryForm(r)
	if err != nil {
		data["entry"] = repo.MacroEntry{}
		h.renderErr(w, r, http.StatusBadRequest, MacrosNew, err)

		return
	}

	_, err = h.store.CreateMacroEntry(ctx, user.ID, params)
	if err != nil {
		data["entry"] = repo.MacroEntry{
			Name:     params.Name,
			Kcal:     params.Kcal,
			ProteinG: params.ProteinG,
			CarbsG:   params.CarbsG,
			FatG:     params.FatG,
			Date:     params.Date,
		}
		h.renderErr(w, r, http.StatusBadRequest, MacrosNew, err)

		return
	}

	dateStr := time.Unix(params.Date, 0).UTC().Format("2006-01-02")
	http.Redirect(w, r, fmt.Sprintf("/macros?date=%s", dateStr), http.StatusSeeOther)
}

func (h *Handler) GetMacroEntryEdit(w http.ResponseWriter, r *http.Request) {
	data := h.tmplData(r)
	data["entry"] = getMacroEntry(r)

	h.render(w, http.StatusOK, MacrosEdit, data)
}

func (h *Handler) PostMacroEntryUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)
	entry := *getMacroEntry(r)

	params, err := parseMacroEntryForm(r)
	if err != nil {
		data["entry"] = entry
		h.renderErr(w, r, http.StatusBadRequest, MacrosEdit, err)

		return
	}

	_, err = h.store.UpdateMacroEntry(ctx, entry.ID, user.ID, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}

		entry.Name = params.Name
		entry.Kcal = params.Kcal
		entry.ProteinG = params.ProteinG
		entry.CarbsG = params.CarbsG
		entry.FatG = params.FatG
		entry.Date = params.Date
		data["entry"] = entry
		h.renderErr(w, r, http.StatusBadRequest, MacrosEdit, err)

		return
	}

	dateStr := time.Unix(params.Date, 0).UTC().Format("2006-01-02")
	http.Redirect(w, r, fmt.Sprintf("/macros?date=%s", dateStr), http.StatusSeeOther)
}

func (h *Handler) GetMacroEntry(w http.ResponseWriter, r *http.Request) {
	data := h.tmplData(r)
	data["entry"] = getMacroEntry(r)

	h.render(w, http.StatusOK, MacrosShow, data)
}

func (h *Handler) PostMacroEntryDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)
	entry := getMacroEntry(r)

	_, err := h.store.DeleteMacroEntry(ctx, entry.ID, user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}
		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	http.Redirect(w, r, "/macros", http.StatusSeeOther)
}

func (h *Handler) GetMacrosGoals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)

	goal, err := h.store.FindMacroGoal(ctx, user.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		h.renderErr(w, r, http.StatusInternalServerError, MacrosGoals, err)

		return
	}

	data["goal"] = goal

	h.render(w, http.StatusOK, MacrosGoals, data)
}

func (h *Handler) PostMacrosGoals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)

	params, err := parseMacroGoalForm(r)
	if err != nil {
		data["goal"] = repo.MacroGoal{}
		h.renderErr(w, r, http.StatusBadRequest, MacrosGoals, err)

		return
	}

	_, err = h.store.SaveMacroGoal(ctx, user.ID, params)
	if err != nil {
		data["goal"] = repo.MacroGoal{
			Kcal:     params.Kcal,
			ProteinG: params.ProteinG,
			CarbsG:   params.CarbsG,
			FatG:     params.FatG,
		}
		h.renderErr(w, r, http.StatusBadRequest, MacrosGoals, err)

		return
	}

	http.Redirect(w, r, "/macros", http.StatusSeeOther)
}

// ----------------------------------------------------------------------------- //
// Unexported Functions and Helpers
// ----------------------------------------------------------------------------- //

func computeDayWindow(dateStr string) (dayStart, nextDayStart int64, selectedDate string) {
	var t time.Time

	if dateStr == "" {
		t = time.Now()
	} else {
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			t = time.Now()
		} else {
			t = parsed
		}
	}

	y, m, d := t.Date()
	start := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	dayStart = start.Unix()
	nextDayStart = dayStart + 86400
	selectedDate = time.Unix(dayStart, 0).UTC().Format("2006-01-02")

	return dayStart, nextDayStart, selectedDate
}

func parseMacroEntryForm(r *http.Request) (logic.MacroEntryParams, error) {
	var params logic.MacroEntryParams

	if err := r.ParseForm(); err != nil {
		return params, fmt.Errorf("failed to parse form, %w", err)
	}

	date, err := prog.StringToUnixDate(r.FormValue("date"))
	if err != nil {
		return params, err
	}

	kcal, err := strconv.ParseFloat(r.FormValue("kcal"), 64)
	if err != nil {
		return params, fmt.Errorf("kcal: %w", err)
	}

	proteinG, err := strconv.ParseFloat(r.FormValue("protein_g"), 64)
	if err != nil {
		return params, fmt.Errorf("protein_g: %w", err)
	}

	carbsG, err := strconv.ParseFloat(r.FormValue("carbs_g"), 64)
	if err != nil {
		return params, fmt.Errorf("carbs_g: %w", err)
	}

	fatG, err := strconv.ParseFloat(r.FormValue("fat_g"), 64)
	if err != nil {
		return params, fmt.Errorf("fat_g: %w", err)
	}

	params.Name = r.FormValue("name")
	params.Kcal = kcal
	params.ProteinG = proteinG
	params.CarbsG = carbsG
	params.FatG = fatG
	params.Date = date

	return params, nil
}

func parseMacroGoalForm(r *http.Request) (logic.MacroGoalParams, error) {
	var params logic.MacroGoalParams

	if err := r.ParseForm(); err != nil {
		return params, fmt.Errorf("failed to parse form, %w", err)
	}

	kcal, err := strconv.ParseFloat(r.FormValue("kcal"), 64)
	if err != nil {
		return params, fmt.Errorf("kcal: %w", err)
	}

	proteinG, err := strconv.ParseFloat(r.FormValue("protein_g"), 64)
	if err != nil {
		return params, fmt.Errorf("protein_g: %w", err)
	}

	carbsG, err := strconv.ParseFloat(r.FormValue("carbs_g"), 64)
	if err != nil {
		return params, fmt.Errorf("carbs_g: %w", err)
	}

	fatG, err := strconv.ParseFloat(r.FormValue("fat_g"), 64)
	if err != nil {
		return params, fmt.Errorf("fat_g: %w", err)
	}

	params.Kcal = kcal
	params.ProteinG = proteinG
	params.CarbsG = carbsG
	params.FatG = fatG

	return params, nil
}

func computeMacroProgress(totals repo.MacroDayTotals, goal repo.MacroGoal) macroProgressData {
	pct := func(total, g float64) int {
		if g <= 0 {
			return 0
		}

		v := int(total * 100 / g)
		if v > 100 {
			return 100
		}

		return v
	}

	return macroProgressData{
		KcalPct:    pct(totals.Kcal, goal.Kcal),
		ProteinPct: pct(totals.ProteinG, goal.ProteinG),
		CarbsPct:   pct(totals.CarbsG, goal.CarbsG),
		FatPct:     pct(totals.FatG, goal.FatG),
	}
}

func getMacroEntry(r *http.Request) *repo.MacroEntry {
	entry, ok := r.Context().Value(KeyMacroEntry).(*repo.MacroEntry)

	if !ok {
		panic("failed to get macro entry context")
	}

	return entry
}
