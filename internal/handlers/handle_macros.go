package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
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

	mealType := q.Get("meal_type")

	filterFields := []repo.FilterField{
		{Name: "user_id", Value: user.ID, Operator: "="},
		{Name: "date", Value: dayStart, Operator: ">="},
		{Name: "date", Value: nextDayStart, Operator: "<"},
	}
	if mealType != "" {
		filterFields = append(filterFields, repo.FilterField{Name: "meal_type", Value: mealType, Operator: "="})
	}

	opts := repo.QueryOptions{
		Filters: repo.Filters{
			FilterFields: filterFields,
			Connector:    "AND",
		},
		Sorting: repo.Sorting{Field: sortField, Order: sortOrder},
	}

	entries, err := h.store.FindMacroEntries(ctx, opts)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, MacrosIndex, err)

		return
	}

	totals, err := h.store.FindMacroDayTotals(ctx, user.ID, dayStart, nextDayStart, mealType)
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
	data["selectedMealType"] = mealType

	if hasGoal {
		data["goal"] = goal
		data["progress"] = computeMacroProgress(totals, goal)
	}

	h.render(w, http.StatusOK, MacrosIndex, data)
}

func (h *Handler) GetMacrosNew(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)

	entry := repo.MacroEntry{}

	if fromFoodStr := r.URL.Query().Get("from_food"); fromFoodStr != "" {
		foodID, err := prog.ParseID(fromFoodStr, "Food")
		if err == nil {
			food, err := h.store.FindFood(ctx, foodID, user.ID)
			if err == nil {
				scale := 1.0
				if amountStr := r.URL.Query().Get("amount"); amountStr != "" {
					if amount, err := strconv.ParseFloat(amountStr, 64); err == nil && amount > 0 {
						scale = amount / foodBaseAmountG
					}
				}

				entry.Name = food.Name
				entry.Kcal = roundMacro(food.Kcal * scale)
				entry.ProteinG = roundMacro(food.ProteinG * scale)
				entry.CarbsG = roundMacro(food.CarbsG * scale)
				entry.FatG = roundMacro(food.FatG * scale)
			}
		}
	}

	if fromTemplateStr := r.URL.Query().Get("from_template"); fromTemplateStr != "" {
		tmplID, err := prog.ParseID(fromTemplateStr, "MacroTemplate")
		if err == nil {
			tmpl, err := h.store.FindMacroTemplate(ctx, tmplID, user.ID)
			if err == nil {
				entry.Name = tmpl.Name
				entry.Kcal = tmpl.Kcal
				entry.ProteinG = tmpl.ProteinG
				entry.CarbsG = tmpl.CarbsG
				entry.FatG = tmpl.FatG
				data["template"] = tmpl
			}
		}
	}

	data["entry"] = entry
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

	newEntry, err := h.store.CreateMacroEntry(ctx, user.ID, params)
	if err != nil {
		data["entry"] = repo.MacroEntry{
			Name:     params.Name,
			Kcal:     params.Kcal,
			ProteinG: params.ProteinG,
			CarbsG:   params.CarbsG,
			FatG:     params.FatG,
			Date:     params.Date,
			MealType: params.MealType,
		}
		h.renderErr(w, r, http.StatusBadRequest, MacrosNew, err)

		return
	}

	if r.FormValue("save_as_template") == "on" {
		http.Redirect(w, r, fmt.Sprintf("/macros/templates/new?from_entry=%d", newEntry.ID), http.StatusSeeOther)

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
		entry.MealType = params.MealType
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
		return params, fmt.Errorf("%w: %w", ErrParseForm, err)
	}

	date, err := prog.StringToUnixDate(r.FormValue("date"))
	if err != nil {
		return params, err
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

	params.Name = r.FormValue("name")
	params.Kcal = kcal
	params.ProteinG = proteinG
	params.CarbsG = carbsG
	params.FatG = fatG
	params.Date = date
	params.MealType = r.FormValue("meal_type")

	return params, nil
}

func parseMacroGoalForm(r *http.Request) (logic.MacroGoalParams, error) {
	var params logic.MacroGoalParams

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

type macroTrendSummary struct {
	ActiveDays   int
	TotalKcal    float64
	AvgKcal      float64
	TotalProtein float64
	AvgProtein   float64
	TotalCarbs   float64
	AvgCarbs     float64
	TotalFat     float64
	AvgFat       float64
}

func computeMacroTrendSummary(dailyTotals []repo.MacroDailyTotal) macroTrendSummary {
	var s macroTrendSummary
	s.ActiveDays = len(dailyTotals)

	for _, t := range dailyTotals {
		s.TotalKcal += t.Kcal
		s.TotalProtein += t.ProteinG
		s.TotalCarbs += t.CarbsG
		s.TotalFat += t.FatG
	}

	if s.ActiveDays > 0 {
		n := float64(s.ActiveDays)
		s.AvgKcal = s.TotalKcal / n
		s.AvgProtein = s.TotalProtein / n
		s.AvgCarbs = s.TotalCarbs / n
		s.AvgFat = s.TotalFat / n
	}

	return s
}

type macroTrendDataset struct {
	Label string    `json:"label"`
	Data  []float64 `json:"data"`
}

type macroTrendChartData struct {
	Labels   []string            `json:"labels"`
	Datasets []macroTrendDataset `json:"datasets"`
}

func (h *Handler) GetMacrosStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)

	period := r.URL.Query().Get("period")

	var days int
	switch period {
	case "week":
		days = 7
	case "six_months":
		days = 180
	default:
		period = "month"
		days = 30
	}

	now := time.Now().UTC()
	y, m, d := now.Date()
	todayMidnight := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	start := todayMidnight.AddDate(0, 0, -(days - 1))
	end := todayMidnight.Add(24 * time.Hour)

	dailyTotals, err := h.store.FindMacroDailyTotals(ctx, user.ID, start.Unix(), end.Unix())
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, MacrosStats, err)

		return
	}

	totalsMap := make(map[int64]repo.MacroDailyTotal, len(dailyTotals))
	for _, t := range dailyTotals {
		totalsMap[t.Date] = t
	}

	labels := make([]string, 0, days)
	kcalData := make([]float64, 0, days)
	proteinData := make([]float64, 0, days)
	carbsData := make([]float64, 0, days)
	fatData := make([]float64, 0, days)

	for i := 0; i < days; i++ {
		day := start.AddDate(0, 0, i)
		labels = append(labels, day.Format("Jan 2"))

		if t, ok := totalsMap[day.Unix()]; ok {
			kcalData = append(kcalData, t.Kcal)
			proteinData = append(proteinData, t.ProteinG)
			carbsData = append(carbsData, t.CarbsG)
			fatData = append(fatData, t.FatG)
		} else {
			kcalData = append(kcalData, 0)
			proteinData = append(proteinData, 0)
			carbsData = append(carbsData, 0)
			fatData = append(fatData, 0)
		}
	}

	chartData := macroTrendChartData{
		Labels: labels,
		Datasets: []macroTrendDataset{
			{Label: "Kcal", Data: kcalData},
			{Label: "Protein (g)", Data: proteinData},
			{Label: "Carbs (g)", Data: carbsData},
			{Label: "Fat (g)", Data: fatData},
		},
	}

	chartDataBytes, err := json.Marshal(chartData)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, MacrosStats, err)

		return
	}

	data["chartData"] = string(chartDataBytes)
	data["period"] = period
	data["summary"] = computeMacroTrendSummary(dailyTotals)

	h.render(w, http.StatusOK, MacrosStats, data)
}

func getMacroEntry(r *http.Request) *repo.MacroEntry {
	entry, ok := r.Context().Value(KeyMacroEntry).(*repo.MacroEntry)

	if !ok {
		panic("failed to get macro entry context")
	}

	return entry
}
