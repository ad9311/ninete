package repo

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

const (
	TaggableTypeExpense   = "expense"
	TaggableTypeTask      = "task"
	TaggableTypeMoodEntry = "mood_entry"
)

type Tagging struct {
	ID           int
	TagID        int
	TaggableID   int
	TaggableType string
	CreatedAt    int64
	UpdatedAt    int64
}

type InsertTaggingParams struct {
	TagID        int
	TaggableID   int
	TaggableType string
}

type TagRow struct {
	TargetID int
	TagName  string
}

// TagNamesByTargetID groups TagRow values by target ID and returns each group
// sorted alphabetically so JSON/HTML output is stable.
func TagNamesByTargetID(rows []TagRow) map[int][]string {
	m := map[int][]string{}
	for _, row := range rows {
		m[row.TargetID] = append(m[row.TargetID], row.TagName)
	}

	for id := range m {
		sort.Strings(m[id])
	}

	return m
}

const insertOrIgnoreTagging = `
INSERT OR IGNORE INTO "taggings" ("tag_id", "taggable_id", "taggable_type")
VALUES (?, ?, ?)`

func (q *TxQueries) InsertOrIgnoreTagging(ctx context.Context, params InsertTaggingParams) error {
	return q.wrapQuery(insertOrIgnoreTagging, func() error {
		_, err := q.tx.ExecContext(
			ctx,
			insertOrIgnoreTagging,
			params.TagID,
			params.TaggableID,
			params.TaggableType,
		)

		return err
	})
}

const deleteTaggingsByTarget = `
DELETE FROM "taggings"
WHERE "taggable_type" = ?
  AND "taggable_id" = ?`

func (q *TxQueries) DeleteTaggingsByTarget(ctx context.Context, taggableType string, taggableID int) error {
	return q.wrapQuery(deleteTaggingsByTarget, func() error {
		_, err := q.tx.ExecContext(ctx, deleteTaggingsByTarget, taggableType, taggableID)

		return err
	})
}

const selectTagsForTaggableBase = `
SELECT t.*
FROM "taggings" tg
INNER JOIN "tags" t ON t."id" = tg."tag_id"
INNER JOIN "%s" o ON o."id" = tg."taggable_id"
WHERE tg."taggable_type" = ?
  AND tg."taggable_id" = ?
  AND o."user_id" = ?
ORDER BY t."name" ASC
`

// SelectTagsForTaggable returns the tags attached to a single taggable record,
// scoped to the owning user via ownerTable.
func (q *Queries) SelectTagsForTaggable(
	ctx context.Context,
	taggableType, ownerTable string,
	taggableID, userID int,
) ([]Tag, error) {
	var ts []Tag

	query := fmt.Sprintf(selectTagsForTaggableBase, ownerTable)

	err := q.wrapQuery(query, func() error {
		rows, err := q.db.QueryContext(ctx, query, taggableType, taggableID, userID)
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				q.app.Logger.Error(closeErr)
			}
		}()

		ts, err = scanTagRows(rows)
		if err != nil {
			return err
		}

		return nil
	})

	return ts, err
}

const selectTagRowsBase = `
SELECT tg."taggable_id", t."name"
FROM "taggings" tg
INNER JOIN "tags" t ON t."id" = tg."tag_id"
INNER JOIN "%s" r ON r."id" = tg."taggable_id"
WHERE tg."taggable_type" = ?
  AND r."user_id" = ?
  AND tg."taggable_id" IN (%s)
ORDER BY tg."taggable_id" ASC, t."name" ASC
`

func (q *Queries) SelectTagRows(
	ctx context.Context,
	taggableType string,
	joinTable string,
	targetIDs []int,
	userID int,
) ([]TagRow, error) {
	var rowsResult []TagRow
	if len(targetIDs) == 0 {
		return rowsResult, nil
	}

	query, values := selectTagRowsQuery(taggableType, joinTable, targetIDs, userID)

	err := q.wrapQuery(query, func() error {
		rows, err := q.db.QueryContext(ctx, query, values...)
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				q.app.Logger.Error(closeErr)
			}
		}()

		for rows.Next() {
			var row TagRow

			if err := rows.Scan(&row.TargetID, &row.TagName); err != nil {
				return err
			}

			rowsResult = append(rowsResult, row)
		}

		return rows.Err()
	})

	return rowsResult, err
}

func selectTagRowsQuery(taggableType, joinTable string, targetIDs []int, userID int) (string, []any) {
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(targetIDs)), ",")
	query := fmt.Sprintf(selectTagRowsBase, joinTable, placeholders)

	values := make([]any, 0, len(targetIDs)+2)
	values = append(values, taggableType, userID)
	for _, id := range targetIDs {
		values = append(values, id)
	}

	return query, values
}
