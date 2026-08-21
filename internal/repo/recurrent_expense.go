package repo

import (
	"context"
	"database/sql"
)

type recurrentExpense struct {
	ID                int
	UserID            int
	CategoryID        int
	Description       string
	Amount            uint64
	Period            uint
	LastCopyCreatedAt sql.NullInt64
	CreatedAt         int64
	UpdatedAt         int64
	OccurrenceLimit   uint
	OccurrenceCount   uint
	ArchivedAt        sql.NullInt64
}

type RecurrentExpense struct {
	ID                int
	UserID            int
	CategoryID        int
	Description       string
	Amount            uint64
	Period            uint
	LastCopyCreatedAt *int64
	CreatedAt         int64
	UpdatedAt         int64
	OccurrenceLimit   uint
	OccurrenceCount   uint
	ArchivedAt        *int64
}

func (re recurrentExpense) toRecurrentExpense() RecurrentExpense {
	var lastCopy *int64
	if re.LastCopyCreatedAt.Valid {
		value := re.LastCopyCreatedAt.Int64
		lastCopy = &value
	}

	var archivedAt *int64
	if re.ArchivedAt.Valid {
		value := re.ArchivedAt.Int64
		archivedAt = &value
	}

	return RecurrentExpense{
		ID:                re.ID,
		UserID:            re.UserID,
		CategoryID:        re.CategoryID,
		Description:       re.Description,
		Amount:            re.Amount,
		Period:            re.Period,
		LastCopyCreatedAt: lastCopy,
		CreatedAt:         re.CreatedAt,
		UpdatedAt:         re.UpdatedAt,
		OccurrenceLimit:   re.OccurrenceLimit,
		OccurrenceCount:   re.OccurrenceCount,
		ArchivedAt:        archivedAt,
	}
}

func NullInt64FromPtr(value *int64) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{Valid: false}
	}

	return sql.NullInt64{Int64: *value, Valid: true}
}

type InsertRecurrentExpenseParams struct {
	UserID          int
	CategoryID      int
	Description     string
	Amount          uint64
	Period          uint
	OccurrenceLimit uint
}

type UpdateRecurrentExpenseParams struct {
	ID                int
	UserID            int
	CategoryID        int
	Description       string
	Amount            uint64
	Period            uint
	LastCopyCreatedAt sql.NullInt64
	OccurrenceLimit   uint
}

// RecurrentExpenseArchivedFilter builds the predicate splitting the active list
// from the archived one. Both lists reuse the same select, so the split lives in
// a filter rather than in two near-identical queries.
func RecurrentExpenseArchivedFilter(archived bool) FilterField {
	if archived {
		return FilterField{Expr: `"archived_at" IS NOT NULL`}
	}

	return FilterField{Expr: `"archived_at" IS NULL`}
}

// recurrentExpenseColumns pins the projection order the Scan calls in this file depend on.
// SELECT * would resolve to whatever order the table happens to have, so an
// ALTER TABLE could shift values into the wrong struct fields with no error.
const recurrentExpenseColumns = `"id", "user_id", "category_id", "description", "amount", "period",
"last_copy_created_at", "created_at", "updated_at", "occurrence_limit", "occurrence_count", "archived_at"`

const insertRecurrentExpense = `
INSERT INTO "recurrent_expenses" ("user_id", "category_id", "description", "amount", "period", "occurrence_limit")
VALUES (?, ?, ?, ?, ?, ?)
RETURNING ` + recurrentExpenseColumns

const selectRecurrentExpenses = `SELECT ` + recurrentExpenseColumns + ` FROM "recurrent_expenses"`

func (q *Queries) SelectRecurrentExpenses(ctx context.Context, opts QueryOptions) ([]RecurrentExpense, error) {
	var res []RecurrentExpense

	if err := opts.Validate(validRecurrentExpenseFields()); err != nil {
		return res, err
	}

	subQuery, err := opts.Build()
	if err != nil {
		return res, err
	}

	query := selectRecurrentExpenses + " " + subQuery
	values := opts.Filters.Values()

	err = q.wrapQuery(query, func() error {
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
			var re recurrentExpense

			if err := rows.Scan(
				&re.ID,
				&re.UserID,
				&re.CategoryID,
				&re.Description,
				&re.Amount,
				&re.Period,
				&re.LastCopyCreatedAt,
				&re.CreatedAt,
				&re.UpdatedAt,
				&re.OccurrenceLimit,
				&re.OccurrenceCount,
				&re.ArchivedAt,
			); err != nil {
				return err
			}

			res = append(res, re.toRecurrentExpense())
		}

		return rows.Err()
	})

	return res, err
}

const countRecurrentExpenses = `SELECT COUNT(*) FROM "recurrent_expenses"`

func (q *Queries) CountRecurrentExpenses(ctx context.Context, filters Filters) (int, error) {
	var c int

	subQuery, err := filters.Build()
	if err != nil {
		return 0, err
	}

	query := countRecurrentExpenses + " " + subQuery
	values := filters.Values()

	err = q.wrapQuery(query, func() error {
		row := q.db.QueryRowContext(ctx, query, values...)

		return row.Scan(&c)
	})

	return c, err
}

func (q *TxQueries) InsertRecurrentExpense(
	ctx context.Context,
	params InsertRecurrentExpenseParams,
) (RecurrentExpense, error) {
	var re recurrentExpense

	err := q.wrapQuery(insertRecurrentExpense, func() error {
		row := q.tx.QueryRowContext(
			ctx,
			insertRecurrentExpense,
			params.UserID,
			params.CategoryID,
			params.Description,
			params.Amount,
			params.Period,
			params.OccurrenceLimit,
		)

		return row.Scan(
			&re.ID,
			&re.UserID,
			&re.CategoryID,
			&re.Description,
			&re.Amount,
			&re.Period,
			&re.LastCopyCreatedAt,
			&re.CreatedAt,
			&re.UpdatedAt,
			&re.OccurrenceLimit,
			&re.OccurrenceCount,
			&re.ArchivedAt,
		)
	})

	return re.toRecurrentExpense(), err
}

const updateRecurrentExpense = `
UPDATE "recurrent_expenses"
SET "category_id"          = ?,
    "description"          = ?,
    "amount"               = ?,
    "period"               = ?,
    "last_copy_created_at" = COALESCE(?, "last_copy_created_at"),
    "occurrence_limit"     = ?,
    "updated_at"           = ?
WHERE "id" = ? AND "user_id" = ?
RETURNING ` + recurrentExpenseColumns + `;
`

const deleteRecurrentExpense = `DELETE FROM "recurrent_expenses" WHERE "id" = ? AND "user_id" = ? RETURNING "id"`

func (q *Queries) UpdateRecurrentExpense(
	ctx context.Context,
	params UpdateRecurrentExpenseParams,
) (RecurrentExpense, error) {
	var re recurrentExpense

	err := q.wrapQuery(updateRecurrentExpense, func() error {
		row := q.db.QueryRowContext(
			ctx,
			updateRecurrentExpense,
			params.CategoryID,
			params.Description,
			params.Amount,
			params.Period,
			params.LastCopyCreatedAt,
			params.OccurrenceLimit,
			newUpdatedAt(),
			params.ID,
			params.UserID,
		)

		return row.Scan(
			&re.ID,
			&re.UserID,
			&re.CategoryID,
			&re.Description,
			&re.Amount,
			&re.Period,
			&re.LastCopyCreatedAt,
			&re.CreatedAt,
			&re.UpdatedAt,
			&re.OccurrenceLimit,
			&re.OccurrenceCount,
			&re.ArchivedAt,
		)
	})

	return re.toRecurrentExpense(), err
}

func (q *TxQueries) UpdateRecurrentExpense(
	ctx context.Context,
	params UpdateRecurrentExpenseParams,
) (RecurrentExpense, error) {
	var re recurrentExpense

	err := q.wrapQuery(updateRecurrentExpense, func() error {
		row := q.tx.QueryRowContext(
			ctx,
			updateRecurrentExpense,
			params.CategoryID,
			params.Description,
			params.Amount,
			params.Period,
			params.LastCopyCreatedAt,
			params.OccurrenceLimit,
			newUpdatedAt(),
			params.ID,
			params.UserID,
		)

		return row.Scan(
			&re.ID,
			&re.UserID,
			&re.CategoryID,
			&re.Description,
			&re.Amount,
			&re.Period,
			&re.LastCopyCreatedAt,
			&re.CreatedAt,
			&re.UpdatedAt,
			&re.OccurrenceLimit,
			&re.OccurrenceCount,
			&re.ArchivedAt,
		)
	})

	return re.toRecurrentExpense(), err
}

func (q *TxQueries) DeleteRecurrentExpense(ctx context.Context, id, userID int) (int, error) {
	var i int

	err := q.wrapQuery(deleteRecurrentExpense, func() error {
		row := q.tx.QueryRowContext(ctx, deleteRecurrentExpense, id, userID)

		return row.Scan(&i)
	})

	return i, err
}

const countRecurrentExpensesByUser = `SELECT COUNT(*) FROM "recurrent_expenses" WHERE "user_id" = ?`

func (q *Queries) CountRecurrentExpensesByUser(ctx context.Context, userID int) (int, error) {
	var c int

	err := q.wrapQuery(countRecurrentExpensesByUser, func() error {
		row := q.db.QueryRowContext(ctx, countRecurrentExpensesByUser, userID)

		return row.Scan(&c)
	})

	return c, err
}

const deleteRecurrentExpenseTaggingsByUser = `
DELETE FROM "taggings"
WHERE "taggable_type" = 'recurrent_expense'
  AND "taggable_id" IN (SELECT "id" FROM "recurrent_expenses" WHERE "user_id" = ?)`

const deleteAllRecurrentExpensesByUser = `DELETE FROM "recurrent_expenses" WHERE "user_id" = ?`

func (q *TxQueries) DeleteAllRecurrentExpensesByUser(ctx context.Context, userID int) error {
	return q.wrapQuery(deleteAllRecurrentExpensesByUser, func() error {
		if _, err := q.tx.ExecContext(ctx, deleteRecurrentExpenseTaggingsByUser, userID); err != nil {
			return err
		}

		_, err := q.tx.ExecContext(ctx, deleteAllRecurrentExpensesByUser, userID)

		return err
	})
}

const selectRecurrentExpense = `
SELECT ` + recurrentExpenseColumns + `
FROM "recurrent_expenses" WHERE "id" = ? AND "user_id" = ? LIMIT 1
`

func (q *Queries) SelectRecurrentExpense(ctx context.Context, id, userID int) (RecurrentExpense, error) {
	var re recurrentExpense

	err := q.wrapQuery(selectRecurrentExpense, func() error {
		row := q.db.QueryRowContext(ctx, selectRecurrentExpense, id, userID)

		return row.Scan(
			&re.ID,
			&re.UserID,
			&re.CategoryID,
			&re.Description,
			&re.Amount,
			&re.Period,
			&re.LastCopyCreatedAt,
			&re.CreatedAt,
			&re.UpdatedAt,
			&re.OccurrenceLimit,
			&re.OccurrenceCount,
			&re.ArchivedAt,
		)
	})

	return re.toRecurrentExpense(), err
}

const selectAllDueRecurrentExpenses = `
SELECT ` + recurrentExpenseColumns + `
FROM "recurrent_expenses"
WHERE "archived_at" IS NULL
  AND ("occurrence_limit" = 0 OR "occurrence_count" < "occurrence_limit")
  AND ("last_copy_created_at" IS NULL
   OR (
        (CAST(strftime('%Y', datetime(?, 'unixepoch')) AS int) -
         CAST(strftime('%Y', datetime("last_copy_created_at", 'unixepoch')) AS int)) * 12 +
        (CAST(strftime('%m', datetime(?, 'unixepoch')) AS int) -
         CAST(strftime('%m', datetime("last_copy_created_at", 'unixepoch')) AS int))
      ) >= "period")
ORDER BY "id" ASC
`

func (q *Queries) SelectAllDueRecurrentExpenses(ctx context.Context, nowUnix int64) ([]RecurrentExpense, error) {
	var res []RecurrentExpense

	err := q.wrapQuery(selectAllDueRecurrentExpenses, func() error {
		rows, err := q.db.QueryContext(ctx, selectAllDueRecurrentExpenses, nowUnix, nowUnix)
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				q.app.Logger.Error(closeErr)
			}
		}()

		for rows.Next() {
			var re recurrentExpense

			if err := rows.Scan(
				&re.ID,
				&re.UserID,
				&re.CategoryID,
				&re.Description,
				&re.Amount,
				&re.Period,
				&re.LastCopyCreatedAt,
				&re.CreatedAt,
				&re.UpdatedAt,
				&re.OccurrenceLimit,
				&re.OccurrenceCount,
				&re.ArchivedAt,
			); err != nil {
				return err
			}

			res = append(res, re.toRecurrentExpense())
		}

		return rows.Err()
	})

	return res, err
}

// recordRecurrentExpenseOccurrence closes out one generated copy: it stamps the
// copy date, bumps the counter and archives the row in the same statement when
// the bumped counter reaches a non-zero limit. Doing it in one UPDATE keeps the
// count and the archived flag from disagreeing.
const recordRecurrentExpenseOccurrence = `
UPDATE "recurrent_expenses"
SET "last_copy_created_at" = ?,
    "occurrence_count"     = "occurrence_count" + 1,
    "archived_at"          = CASE
                               WHEN "occurrence_limit" > 0
                                AND "occurrence_count" + 1 >= "occurrence_limit"
                               THEN ?
                               ELSE "archived_at"
                             END,
    "updated_at"           = ?
WHERE "id" = ? AND "user_id" = ?
RETURNING ` + recurrentExpenseColumns + `;
`

func (q *TxQueries) RecordRecurrentExpenseOccurrence(
	ctx context.Context,
	id, userID int,
	copiedAt int64,
) (RecurrentExpense, error) {
	var re recurrentExpense
	now := newUpdatedAt()

	err := q.wrapQuery(recordRecurrentExpenseOccurrence, func() error {
		row := q.tx.QueryRowContext(
			ctx,
			recordRecurrentExpenseOccurrence,
			copiedAt,
			now,
			now,
			id,
			userID,
		)

		return row.Scan(
			&re.ID,
			&re.UserID,
			&re.CategoryID,
			&re.Description,
			&re.Amount,
			&re.Period,
			&re.LastCopyCreatedAt,
			&re.CreatedAt,
			&re.UpdatedAt,
			&re.OccurrenceLimit,
			&re.OccurrenceCount,
			&re.ArchivedAt,
		)
	})

	return re.toRecurrentExpense(), err
}

// unarchiveRecurrentExpense resets the counter along with the flag. Leaving the
// count at the limit would archive the row again on the very next cron run.
const unarchiveRecurrentExpense = `
UPDATE "recurrent_expenses"
SET "archived_at"      = NULL,
    "occurrence_count" = 0,
    "updated_at"       = ?
WHERE "id" = ? AND "user_id" = ?
RETURNING ` + recurrentExpenseColumns + `;
`

func (q *TxQueries) UnarchiveRecurrentExpense(ctx context.Context, id, userID int) (RecurrentExpense, error) {
	var re recurrentExpense

	err := q.wrapQuery(unarchiveRecurrentExpense, func() error {
		row := q.tx.QueryRowContext(ctx, unarchiveRecurrentExpense, newUpdatedAt(), id, userID)

		return row.Scan(
			&re.ID,
			&re.UserID,
			&re.CategoryID,
			&re.Description,
			&re.Amount,
			&re.Period,
			&re.LastCopyCreatedAt,
			&re.CreatedAt,
			&re.UpdatedAt,
			&re.OccurrenceLimit,
			&re.OccurrenceCount,
			&re.ArchivedAt,
		)
	})

	return re.toRecurrentExpense(), err
}

func validRecurrentExpenseFields() []string {
	return []string{
		"id",
		"user_id",
		"category_id",
		"description",
		"amount",
		"period",
		"last_copy_created_at",
		"created_at",
		"updated_at",
		"occurrence_limit",
		"occurrence_count",
		"archived_at",
	}
}
