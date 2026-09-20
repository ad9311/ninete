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
	SelectedTagIDs []int          `json:"selected_tag_ids"`
	Tags           []apiReportTag `json:"tags"`
	// TagLimit is what SaveReportSetting enforces, sent so the form can cap
	// its checkboxes at the same number rather than letting the user build a
	// submission the server will refuse.
	TagLimit int `json:"tag_limit"`
}

// GetAPIReportSettings answers the monthly report's settings page. A user who
// has never saved settings gets an empty selection rather than a 404, since
// FindReportSetting treats "no row" as "no grouping tags" — a valid
// configuration printing one flat list.
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
		SelectedTagIDs: selected,
		Tags:           apiTags,
		TagLimit:       logic.ReportTagLimit,
	})
}

type reportSettingsRequestBody struct {
	TagIDs []int `json:"tag_ids"`
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
		TagIDs: body.TagIDs,
	})
	if err != nil {
		h.WriteAPIError(w, err, logic.ErrReportTooManyTags)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
