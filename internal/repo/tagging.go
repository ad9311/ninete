package repo

import (
	"context"
	"fmt"
	"strings"
)

const TaggableTypeExpense = "expense"

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

type ExpenseTagRow struct {
	ExpenseID int
	TagName   string
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

const selectExpenseTags = `
SELECT t.*
FROM "taggings" tg
INNER JOIN "tags" t ON t."id" = tg."tag_id"
INNER JOIN "expenses" e ON e."id" = tg."taggable_id"
WHERE tg."taggable_type" = ?
  AND tg."taggable_id" = ?
  AND e."user_id" = ?
ORDER BY t."name" ASC
`

func (q *Queries) SelectExpenseTags(ctx context.Context, expenseID, userID int) ([]Tag, error) {
	var ts []Tag

	err := q.wrapQuery(selectExpenseTags, func() error {
		rows, err := q.db.QueryContext(ctx, selectExpenseTags, TaggableTypeExpense, expenseID, userID)
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

const selectExpenseTagRowsBase = `
SELECT tg."taggable_id", t."name"
FROM "taggings" tg
INNER JOIN "tags" t ON t."id" = tg."tag_id"
INNER JOIN "expenses" e ON e."id" = tg."taggable_id"
WHERE tg."taggable_type" = ?
  AND e."user_id" = ?
  AND tg."taggable_id" IN (%s)
ORDER BY tg."taggable_id" ASC, t."name" ASC
`

func (q *Queries) SelectExpenseTagRows(
	ctx context.Context,
	expenseIDs []int,
	userID int,
) ([]ExpenseTagRow, error) {
	var rowsResult []ExpenseTagRow
	if len(expenseIDs) == 0 {
		return rowsResult, nil
	}

	query, values := selectExpenseTagRowsQuery(expenseIDs, userID)

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
			var row ExpenseTagRow

			if err := rows.Scan(&row.ExpenseID, &row.TagName); err != nil {
				return err
			}

			rowsResult = append(rowsResult, row)
		}

		return nil
	})

	return rowsResult, err
}

func selectExpenseTagRowsQuery(expenseIDs []int, userID int) (string, []any) {
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(expenseIDs)), ",")
	query := fmt.Sprintf(selectExpenseTagRowsBase, placeholders)

	values := make([]any, 0, len(expenseIDs)+2)
	values = append(values, TaggableTypeExpense, userID)
	for _, id := range expenseIDs {
		values = append(values, id)
	}

	return query, values
}
