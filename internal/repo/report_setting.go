package repo

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

// ReportSetting is the per-user configuration of the monthly expense report.
// The grouping tags hang off it in "report_setting_tags" rather than living
// here, since there are many of them per row.
type ReportSetting struct {
	ID        int
	UserID    int
	Timezone  string
	CreatedAt int64
	UpdatedAt int64
}

type UpsertReportSettingParams struct {
	UserID   int
	Timezone string
}

// reportSettingColumns pins the projection order the Scan calls in this file
// depend on. SELECT * would resolve to whatever order the table happens to
// have, so an ALTER TABLE could shift values into the wrong struct fields with
// no error.
const reportSettingColumns = `"id", "user_id", "timezone", "created_at", "updated_at"`

// reportSettingTagColumns exists for the same reason, even though nothing
// selects the whole row today: the constant is what TestColumnConstantsMatchSchema
// checks, so declaring it now is what keeps a future SELECT honest.
const reportSettingTagColumns = `"id", "report_setting_id", "tag_id", "created_at", "updated_at"`

const selectReportSettingByUser = `SELECT ` + reportSettingColumns + `
FROM "report_settings" WHERE "user_id" = ?`

// SelectReportSettingByUser returns the user's report settings. A user who has
// never opened the settings page has no row, which is not an error: the second
// return value reports whether one was found, so the caller can apply its own
// defaults instead of distinguishing sql.ErrNoRows itself.
func (q *Queries) SelectReportSettingByUser(ctx context.Context, userID int) (ReportSetting, bool, error) {
	var s ReportSetting

	found := true

	err := q.wrapQuery(selectReportSettingByUser, func() error {
		row := q.db.QueryRowContext(ctx, selectReportSettingByUser, userID)

		err := row.Scan(&s.ID, &s.UserID, &s.Timezone, &s.CreatedAt, &s.UpdatedAt)
		if errors.Is(err, sql.ErrNoRows) {
			found = false

			return nil
		}

		return err
	})

	return s, found, err
}

// selectReportSettingTagIDs reads the grouping tags through the join rather
// than from "report_setting_tags" alone, so a tag belonging to another user
// could never surface here even if a row were written wrongly.
const selectReportSettingTagIDs = `
SELECT rst."tag_id" FROM "report_setting_tags" rst
INNER JOIN "report_settings" rs ON rs."id" = rst."report_setting_id"
INNER JOIN "tags" t ON t."id" = rst."tag_id" AND t."user_id" = rs."user_id"
WHERE rs."user_id" = ?
ORDER BY rst."tag_id"`

func (q *Queries) SelectReportSettingTagIDs(ctx context.Context, userID int) ([]int, error) {
	var ids []int

	err := q.wrapQuery(selectReportSettingTagIDs, func() error {
		rows, err := q.db.QueryContext(ctx, selectReportSettingTagIDs, userID)
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				q.app.Logger.Error(closeErr)
			}
		}()

		for rows.Next() {
			var id int

			if err := rows.Scan(&id); err != nil {
				return err
			}

			ids = append(ids, id)
		}

		return rows.Err()
	})

	return ids, err
}

const upsertReportSetting = `
INSERT INTO "report_settings" ("user_id","timezone")
VALUES (?,?)
ON CONFLICT ("user_id") DO UPDATE SET
  "timezone"   = excluded."timezone",
  "updated_at" = strftime('%s','now')
RETURNING ` + reportSettingColumns

func (q *TxQueries) UpsertReportSetting(
	ctx context.Context,
	params UpsertReportSettingParams,
) (ReportSetting, error) {
	var s ReportSetting

	err := q.wrapQuery(upsertReportSetting, func() error {
		row := q.tx.QueryRowContext(ctx, upsertReportSetting, params.UserID, params.Timezone)

		return row.Scan(&s.ID, &s.UserID, &s.Timezone, &s.CreatedAt, &s.UpdatedAt)
	})

	return s, err
}

const deleteReportSettingTags = `
DELETE FROM "report_setting_tags" WHERE "report_setting_id" = ?`

func (q *TxQueries) DeleteReportSettingTags(ctx context.Context, reportSettingID int) error {
	return q.wrapQuery(deleteReportSettingTags, func() error {
		_, err := q.tx.ExecContext(ctx, deleteReportSettingTags, reportSettingID)

		return err
	})
}

// insertReportSettingTagsPrefix is completed by buildInsertReportSettingTags,
// which appends one placeholder row per tag. The SELECT guard is what makes
// the tag ids safe to trust: a row is written only when the tag really belongs
// to the settings row's owner, so a request naming someone else's tag id
// inserts nothing rather than linking it.
const insertReportSettingTagsPrefix = `
INSERT INTO "report_setting_tags" ("report_setting_id","tag_id")
SELECT ?, "id" FROM "tags" WHERE "user_id" = ? AND "id" IN (`

func buildInsertReportSettingTags(count int) string {
	var b strings.Builder

	b.WriteString(insertReportSettingTagsPrefix)
	b.WriteString(strings.TrimSuffix(strings.Repeat("?,", count), ","))
	b.WriteString(`)`)

	return b.String()
}

// InsertReportSettingTags links tagIDs to the settings row. It is a no-op for
// an empty list: "IN ()" is not valid SQLite, and "no grouping tags" is a
// legitimate configuration meaning the report prints one flat list.
func (q *TxQueries) InsertReportSettingTags(
	ctx context.Context,
	reportSettingID, userID int,
	tagIDs []int,
) error {
	if len(tagIDs) == 0 {
		return nil
	}

	query := buildInsertReportSettingTags(len(tagIDs))

	args := make([]any, 0, len(tagIDs)+2)
	args = append(args, reportSettingID, userID)

	for _, id := range tagIDs {
		args = append(args, id)
	}

	return q.wrapQuery(query, func() error {
		_, err := q.tx.ExecContext(ctx, query, args...)

		return err
	})
}
