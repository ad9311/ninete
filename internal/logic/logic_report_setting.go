package logic

import (
	"context"
	"time"

	"github.com/ad9311/ninete/internal/repo"
)

// DefaultReportTimezone is what the scheduled report falls back to when the
// user has never saved settings. UTC rather than a guess at the owner's zone:
// expenses.date is already a UTC-midnight month start, so UTC is the one
// choice that cannot silently shift the reported month
// (docs/monthly-report.md, "Timezone").
const DefaultReportTimezone = "UTC"

// reportTagLimit bounds how many tags may group one report. Well past what a
// readable report can carry, and there only so a request cannot make the
// insert build an unbounded placeholder list.
const reportTagLimit = 20

// ReportSettingParams is the settings form's submission. Timezone is validated
// by loading it rather than by pattern, since only the zone database can say
// whether a name resolves on this host.
type ReportSettingParams struct {
	Timezone string `validate:"required,max=64"`
	TagIDs   []int
}

// ReportSetting is the settings as the app uses them: the stored row plus the
// grouping tag ids. A user with no stored row still gets a usable value, since
// every user has settings and some of them are only implicit — Configured is
// what tells the two apart, so the settings form can seed its timezone field
// from the browser instead of showing the UTC fallback as though it had been
// chosen.
type ReportSetting struct {
	Timezone   string
	TagIDs     []int
	Configured bool
}

// FindReportSetting returns the user's report settings, falling back to the
// defaults when nothing has been saved.
func (s *Store) FindReportSetting(ctx context.Context, userID int) (ReportSetting, error) {
	setting := ReportSetting{Timezone: DefaultReportTimezone}

	stored, found, err := s.queries.SelectReportSettingByUser(ctx, userID)
	if err != nil {
		return setting, err
	}

	if !found {
		return setting, nil
	}

	setting.Timezone = stored.Timezone
	setting.Configured = true

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
	if err := s.ValidateStruct(params); err != nil {
		return err
	}

	if _, err := time.LoadLocation(params.Timezone); err != nil {
		return ErrReportTimezone
	}

	if len(params.TagIDs) > reportTagLimit {
		return ErrReportTooManyTags
	}

	tagIDs := dedupeIDs(params.TagIDs)

	return s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		setting, err := tq.UpsertReportSetting(ctx, repo.UpsertReportSettingParams{
			UserID:   userID,
			Timezone: params.Timezone,
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
