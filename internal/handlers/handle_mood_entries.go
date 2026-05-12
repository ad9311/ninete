package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/go-chi/chi/v5"
)

type moodEntryRow struct {
	ID       int
	Mood     string
	Notes    string
	LoggedAt int64
	Tags     []string
}

// ----------------------------------------------------------------------------- //
// Context Middleware
// ----------------------------------------------------------------------------- //

func (h *Handler) MoodEntryContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := getCurrentUser(r)

		id, err := prog.ParseID(chi.URLParam(r, "id"), "Mood Entry")
		if err != nil {
			h.NotFound(w, r)

			return
		}

		entry, err := h.store.FindMoodEntry(ctx, id, user.ID)
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}
		if err != nil {
			h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

			return
		}

		ctx = context.WithValue(ctx, KeyMoodEntry, &entry)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ----------------------------------------------------------------------------- //
// Handlers
// ----------------------------------------------------------------------------- //

func (h *Handler) GetMoodEntries(w http.ResponseWriter, r *http.Request) {
	data := h.tmplData(r)
	user := getCurrentUser(r)
	q := r.URL.Query()

	sortField := q.Get("sort_field")
	sortOrder := q.Get("sort_order")
	if sortField == "" {
		sortField = "logged_at"
	}
	if sortOrder == "" {
		sortOrder = "DESC"
	}

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}

	opts := repo.QueryOptions{
		Filters: repo.Filters{
			FilterFields: []repo.FilterField{
				{Name: "user_id", Value: user.ID, Operator: "="},
			},
			Connector: "AND",
		},
		Sorting: repo.Sorting{Field: sortField, Order: sortOrder},
		Pagination: repo.Pagination{
			Page:    page,
			PerPage: defaultPerPage,
		},
	}

	totalCount, err := h.store.CountMoodEntries(r.Context(), opts.Filters)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, MoodEntriesIndex, err)

		return
	}

	entries, err := h.store.ListMoodEntries(r.Context(), opts)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, MoodEntriesIndex, err)

		return
	}

	entryIDs := make([]int, 0, len(entries))
	for _, e := range entries {
		entryIDs = append(entryIDs, e.ID)
	}

	tagRows, err := h.store.FindTagRows(r.Context(), repo.TaggableTypeMoodEntry, "mood_entries", entryIDs, user.ID)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, MoodEntriesIndex, err)

		return
	}
	tagsByEntryID := tagNamesByTargetID(tagRows)

	rows := make([]moodEntryRow, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, moodEntryRow{
			ID:       e.ID,
			Mood:     e.Mood,
			Notes:    e.Notes,
			LoggedAt: e.LoggedAt,
			Tags:     tagsByEntryID[e.ID],
		})
	}

	data["moodEntries"] = rows
	data["pagination"] = newPaginationData(r, opts, totalCount, "")
	data["basePath"] = "/moods"

	h.render(w, http.StatusOK, MoodEntriesIndex, data)
}

func (h *Handler) GetMoodEntry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)
	entry := getMoodEntry(r)

	entryTags, err := h.store.FindMoodEntryTags(ctx, entry.ID, user.ID)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, MoodEntriesShow, err)

		return
	}

	data["moodEntry"] = moodEntryRow{
		ID:       entry.ID,
		Mood:     entry.Mood,
		Notes:    entry.Notes,
		LoggedAt: entry.LoggedAt,
		Tags:     logic.ExtractTagNames(entryTags),
	}

	h.render(w, http.StatusOK, MoodEntriesShow, data)
}

func (h *Handler) GetMoodEntriesNew(w http.ResponseWriter, r *http.Request) {
	data := h.tmplData(r)

	data["moodEntry"] = repo.MoodEntry{}
	data["moods"] = logic.Moods()
	data["tagsInput"] = ""

	h.render(w, http.StatusOK, MoodEntriesNew, data)
}

func (h *Handler) GetMoodEntriesEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)
	entry := getMoodEntry(r)

	entryTags, err := h.store.FindMoodEntryTags(ctx, entry.ID, user.ID)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, MoodEntriesEdit, err)

		return
	}

	data["moodEntry"] = entry
	data["moods"] = logic.Moods()
	data["tagsInput"] = logic.JoinTagNames(logic.ExtractTagNames(entryTags))

	h.render(w, http.StatusOK, MoodEntriesEdit, data)
}

func (h *Handler) PostMoodEntries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	rawTagsInput := r.FormValue("tags")

	params, err := parseMoodEntryForm(r)
	if err != nil {
		data["moodEntry"] = repo.MoodEntry{}
		data["moods"] = logic.Moods()
		data["tagsInput"] = rawTagsInput
		h.renderErr(w, r, http.StatusBadRequest, MoodEntriesNew, err)

		return
	}

	user := getCurrentUser(r)

	_, err = h.store.CreateMoodEntry(ctx, user.ID, params)
	if err != nil {
		data["moodEntry"] = repo.MoodEntry{Mood: params.Mood, Notes: params.Notes, LoggedAt: params.LoggedAt}
		data["moods"] = logic.Moods()
		data["tagsInput"] = logic.JoinTagNames(params.Tags)
		h.renderErr(w, r, http.StatusBadRequest, MoodEntriesNew, err)

		return
	}

	http.Redirect(w, r, "/moods", http.StatusSeeOther)
}

func (h *Handler) PostMoodEntriesUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)
	entry := *getMoodEntry(r)
	rawTagsInput := r.FormValue("tags")

	params, err := parseMoodEntryForm(r)
	if err != nil {
		data["moodEntry"] = entry
		data["moods"] = logic.Moods()
		data["tagsInput"] = rawTagsInput
		h.renderErr(w, r, http.StatusBadRequest, MoodEntriesEdit, err)

		return
	}

	_, err = h.store.UpdateMoodEntry(ctx, entry.ID, user.ID, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}

		entry.Mood = params.Mood
		entry.Notes = params.Notes
		entry.LoggedAt = params.LoggedAt
		data["moodEntry"] = entry
		data["moods"] = logic.Moods()
		data["tagsInput"] = logic.JoinTagNames(params.Tags)
		h.renderErr(w, r, http.StatusBadRequest, MoodEntriesEdit, err)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/moods/%d", entry.ID), http.StatusSeeOther)
}

func (h *Handler) PostMoodEntriesDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)
	entry := getMoodEntry(r)

	err := h.store.DeleteMoodEntry(ctx, entry.ID, user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}

		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	http.Redirect(w, r, "/moods", http.StatusSeeOther)
}

// ----------------------------------------------------------------------------- //
// Unexported Functions and Helpers
// ----------------------------------------------------------------------------- //

func parseMoodEntryForm(r *http.Request) (logic.MoodEntryParams, error) {
	var params logic.MoodEntryParams

	if err := r.ParseForm(); err != nil {
		return params, fmt.Errorf("failed to parse form: %w", err)
	}

	loggedAt, err := prog.StringToUnixDate(r.FormValue("logged_at"))
	if err != nil {
		return params, err
	}

	params.Mood = r.FormValue("mood")
	params.Notes = r.FormValue("notes")
	params.LoggedAt = loggedAt
	params.Tags = logic.ParseTagNames(r.FormValue("tags"))

	return params, nil
}

func (h *Handler) GetMoodEntriesStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)
	q := r.URL.Query()

	sortField := q.Get("sort_field")
	sortOrder := q.Get("sort_order")
	if sortField == "" {
		sortField = "count"
	}
	if sortOrder == "" {
		sortOrder = "DESC"
	}

	fromDate := q.Get("from_date")
	toDate := q.Get("to_date")

	filters := repo.Filters{
		FilterFields: []repo.FilterField{
			{Name: "user_id", Value: user.ID, Operator: "="},
		},
		Connector: "AND",
	}

	if fromDate != "" {
		if t, err := time.Parse(time.DateOnly, fromDate); err == nil {
			filters.FilterFields = append(filters.FilterFields, repo.FilterField{
				Name: "logged_at", Value: t.Unix(), Operator: ">=",
			})
		}
	}
	if toDate != "" {
		if t, err := time.Parse(time.DateOnly, toDate); err == nil {
			filters.FilterFields = append(filters.FilterFields, repo.FilterField{
				Name: "logged_at", Value: t.AddDate(0, 0, 1).Unix(), Operator: "<",
			})
		}
	}

	counts, err := h.store.FindMoodEntryCounts(ctx, filters)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, MoodEntriesStats, err)

		return
	}

	sort.Slice(counts, func(i, j int) bool {
		switch sortField {
		case "mood":
			if sortOrder == "ASC" {
				return counts[i].Mood < counts[j].Mood
			}

			return counts[i].Mood > counts[j].Mood
		default:
			if sortOrder == "ASC" {
				return counts[i].Count < counts[j].Count
			}

			return counts[i].Count > counts[j].Count
		}
	})

	type chartPoint struct {
		Name  string `json:"name"`
		Total int    `json:"total"`
	}
	chartPoints := make([]chartPoint, 0, len(counts))
	for _, c := range counts {
		chartPoints = append(chartPoints, chartPoint{Name: c.Mood, Total: c.Count})
	}

	chartDataBytes, err := json.Marshal(chartPoints)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, MoodEntriesStats, err)

		return
	}

	moodSortOrder := "ASC"
	if sortField == "mood" && sortOrder == "ASC" {
		moodSortOrder = "DESC"
	}

	countSortOrder := "DESC"
	if sortField == "count" && sortOrder == "DESC" {
		countSortOrder = "ASC"
	}

	baseURL := fmt.Sprintf("/moods/stats?from_date=%s&to_date=%s", fromDate, toDate)

	data["rows"] = counts
	data["chartData"] = string(chartDataBytes)
	data["sortField"] = sortField
	data["sortOrder"] = sortOrder
	data["fromDate"] = fromDate
	data["toDate"] = toDate
	data["moodSortURL"] = fmt.Sprintf("%s&sort_field=mood&sort_order=%s", baseURL, moodSortOrder)
	data["countSortURL"] = fmt.Sprintf("%s&sort_field=count&sort_order=%s", baseURL, countSortOrder)

	h.render(w, http.StatusOK, MoodEntriesStats, data)
}

func getMoodEntry(r *http.Request) *repo.MoodEntry {
	entry, ok := r.Context().Value(KeyMoodEntry).(*repo.MoodEntry)

	if !ok {
		panic("failed to get mood entry context")
	}

	return entry
}
