package handlers

import (
	"net/http"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
)

type apiReportTag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// apiReportSettingsResponse carries the user's whole tag list alongside the
// selection, so the settings page needs one request rather than two. There is
// no GET /api/tags to fetch them from — tags are created as free text on the
// expense forms and have never had a listing endpoint of their own.
type apiReportSettingsResponse struct {
	Timezone string `json:"timezone"`
	// Configured is false until the user saves for the first time, which is
	// what lets the form seed its timezone field from the browser rather than
	// presenting the UTC fallback as a deliberate choice.
	Configured     bool           `json:"configured"`
	SelectedTagIDs []int          `json:"selected_tag_ids"`
	Tags           []apiReportTag `json:"tags"`
}

// GetAPIReportSettings answers the monthly report's settings page. A user who
// has never saved settings gets the defaults rather than a 404, since
// FindReportSetting treats "no row" as "implicit settings"
// (docs/monthly-report.md, "Timezone").
func (h *Handler) GetAPIReportSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	setting, err := h.store.FindReportSetting(ctx, user.ID)
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	tags, err := h.store.FindTags(ctx, repo.QueryOptions{
		Filters: repo.Filters{
			FilterFields: []repo.FilterField{
				{Name: "user_id", Value: user.ID, Operator: "="},
			},
		},
		Sorting: repo.Sorting{Field: "name", Order: "ASC"},
	})
	if err != nil {
		h.WriteAPIError(w, err)

		return
	}

	apiTags := make([]apiReportTag, 0, len(tags))
	for _, tag := range tags {
		apiTags = append(apiTags, apiReportTag{ID: tag.ID, Name: tag.Name})
	}

	selected := setting.TagIDs
	if selected == nil {
		selected = []int{}
	}

	h.WriteJSON(w, http.StatusOK, apiReportSettingsResponse{
		Timezone:       setting.Timezone,
		Configured:     setting.Configured,
		SelectedTagIDs: selected,
		Tags:           apiTags,
	})
}

type reportSettingsRequestBody struct {
	Timezone string `json:"timezone"`
	TagIDs   []int  `json:"tag_ids"`
}

// PutAPIReportSettings saves the settings. The submitted tag list replaces the
// stored one wholesale, so an empty list is meaningful — it configures a report
// with no tag sections, printing one flat list.
func (h *Handler) PutAPIReportSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)

	var body reportSettingsRequestBody
	if !h.DecodeJSON(w, r, &body) {
		return
	}

	err := h.store.SaveReportSetting(ctx, user.ID, logic.ReportSettingParams{
		Timezone: body.Timezone,
		TagIDs:   body.TagIDs,
	})
	if err != nil {
		h.WriteAPIError(w, err, logic.ErrReportTimezone, logic.ErrReportTooManyTags)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
