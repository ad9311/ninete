package repo

import (
	"context"
	"database/sql"
	"errors"
)

type ExpenseCategoryMapping struct {
	ID             int
	UserID         int
	CategoryID     int
	DescriptionKey string
	CreatedAt      int64
	UpdatedAt      int64
}

type UpsertExpenseCategoryMappingParams struct {
	UserID         int
	CategoryID     int
	DescriptionKey string
}

const selectExpenseCategoryMapping = `
SELECT * FROM "expense_category_mappings"
WHERE "user_id" = ? AND "description_key" = ?`

// SelectExpenseCategoryMapping returns the mapping for a user's description key.
// The second return value is false when no mapping exists.
func (q *Queries) SelectExpenseCategoryMapping(
	ctx context.Context,
	userID int,
	descriptionKey string,
) (ExpenseCategoryMapping, bool, error) {
	var m ExpenseCategoryMapping

	err := q.wrapQuery(selectExpenseCategoryMapping, func() error {
		row := q.db.QueryRowContext(ctx, selectExpenseCategoryMapping, userID, descriptionKey)

		return row.Scan(
			&m.ID,
			&m.UserID,
			&m.CategoryID,
			&m.DescriptionKey,
			&m.CreatedAt,
			&m.UpdatedAt,
		)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return m, false, nil
	}
	if err != nil {
		return m, false, err
	}

	return m, true, nil
}

const upsertExpenseCategoryMapping = `
INSERT INTO "expense_category_mappings" ("user_id", "category_id", "description_key")
VALUES (?, ?, ?)
ON CONFLICT ("user_id", "description_key")
DO UPDATE SET "category_id" = excluded."category_id", "updated_at" = ?
RETURNING *`

func (q *TxQueries) UpsertExpenseCategoryMapping(
	ctx context.Context,
	params UpsertExpenseCategoryMappingParams,
) (ExpenseCategoryMapping, error) {
	var m ExpenseCategoryMapping

	err := q.wrapQuery(upsertExpenseCategoryMapping, func() error {
		row := q.tx.QueryRowContext(
			ctx,
			upsertExpenseCategoryMapping,
			params.UserID,
			params.CategoryID,
			params.DescriptionKey,
			newUpdatedAt(),
		)

		return row.Scan(
			&m.ID,
			&m.UserID,
			&m.CategoryID,
			&m.DescriptionKey,
			&m.CreatedAt,
			&m.UpdatedAt,
		)
	})

	return m, err
}
