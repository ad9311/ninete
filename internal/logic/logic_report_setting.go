package logic

import (
	"context"

	"github.com/ad9311/ninete/internal/repo"
)

// ReportTagLimit bounds how many tags may group one report. Well past what a
// readable report can carry, and there only so a request cannot make the
// insert build an unbounded placeholder list. Exported because the settings
// form caps its checkboxes at the same number.
const ReportTagLimit = 20

// ReportSettingParams is the settings form's submission. The grouping tags
// are the whole of the configuration: the report period comes from the
// download's month parameter, and the sections order themselves by total.
type ReportSettingParams struct {
	TagIDs []int
}

// ReportSetting is the settings as the app uses them. A user with no stored
// row is not an error — an unconfigured report is a valid one, printing a
// single flat list.
type ReportSetting struct {
	TagIDs []int
}

// FindReportSetting returns the user's grouping tags. Nothing saved means an
// empty list, not an error.
func (s *Store) FindReportSetting(ctx context.Context, userID int) (ReportSetting, error) {
	var setting ReportSetting

	// No existence probe: selectReportSettingTagIDs joins through
	// "report_settings" on "user_id", so a user with no row selects no tags and
	// the empty list falls out of the same query.
	tagIDs, err := s.queries.SelectReportSettingTagIDs(ctx, userID)
	if err != nil {
		return setting, err
	}

	setting.TagIDs = tagIDs

	return setting, nil
}

// SaveReportSetting upserts the settings row and replaces its grouping tags
// wholesale. Replacing rather than diffing keeps the submitted list the single
// description of the configuration: a tag the form dropped disappears without
// the caller having to say so.
//
// The tag ids are not checked for ownership here — InsertReportSettingTags
// filters them through the user's own tags in SQL, so a foreign id inserts
// nothing rather than being rejected with an error that would confirm it
// exists.
func (s *Store) SaveReportSetting(ctx context.Context, userID int, params ReportSettingParams) error {
	// No ValidateStruct: the tag ids are the only input left, and they are not
	// checked by tag rules but by dedupeIDs and the limit below — ownership is
	// enforced in SQL by InsertReportSettingTags.
	//
	// Deduped before the limit is applied, so the error means what it says:
	// twenty-one copies of one tag is one grouping tag, not twenty-one.
	tagIDs := dedupeIDs(params.TagIDs)
	if len(tagIDs) > ReportTagLimit {
		return ErrReportTooManyTags
	}

	return s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		setting, err := tq.UpsertReportSetting(ctx, repo.UpsertReportSettingParams{
			UserID: userID,
		})
		if err != nil {
			return err
		}

		if err := tq.DeleteReportSettingTags(ctx, setting.ID); err != nil {
			return err
		}

		return tq.InsertReportSettingTags(ctx, setting.ID, userID, tagIDs)
	})
}

// dedupeIDs drops repeats and anything non-positive, preserving order. A
// duplicate would otherwise collide with the unique index and fail the whole
// save over an input the form should never have produced.
func dedupeIDs(ids []int) []int {
	seen := make(map[int]struct{}, len(ids))
	out := make([]int, 0, len(ids))

	for _, id := range ids {
		if id < 1 {
			continue
		}

		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}
		out = append(out, id)
	}

	return out
}
