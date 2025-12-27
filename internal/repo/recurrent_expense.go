package repo

import (
	"context"
	"database/sql"
)

type RecurrentExpense struct {
	ID                int           `json:"id"`
	UserID            int           `json:"userId"`
	CategoryID        int           `json:"categoryId"`
	Description       string        `json:"description"`
	Amount            uint64        `json:"amount"`
	Period            uint          `json:"period"`
	LastCopyCreatedAt sql.NullInt64 `json:"lastCopyCreated"`
	CreatedAt         int64         `json:"createdAt"`
	UpdatedAt         int64         `json:"updatedAt"`
}

type InsertRecurrentExpenseParams struct {
	UserID      int
	CategoryID  int
	Description string
	Amount      uint64
	Period      uint
}

type UpdateRecurrentExpenseParams struct {
	ID                int
	UserID            int
	Description       string
	Amount            uint64
	Period            uint
	LastCopyCreatedAt sql.NullInt64
}

const insertRecurrentExpense = `
INSERT INTO "recurrent_expenses" ("user_id", "category_id", "description", "amount", "period")
VALUES (?, ?, ?, ?, ?)
RETURNING *`

func (q *Queries) InsertRecurrentExpense(
	ctx context.Context,
	params InsertRecurrentExpenseParams,
) (RecurrentExpense, error) {
	var re RecurrentExpense

	err := q.wrapQuery(insertRecurrentExpense, func() error {
		row := q.db.QueryRowContext(
			ctx,
			insertRecurrentExpense,
			params.UserID,
			params.CategoryID,
			params.Description,
			params.Amount,
			params.Period,
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
		)
	})

	return re, err
}

const updateRecurrentExpense = `
UPDATE "recurrent_expenses"
SET "description"          = ?,
    "amount"               = ?,
    "period"               = ?,
    "last_copy_created_at" = ?,
    "updated_at"           = ?
WHERE "id" = ? AND "user_id" = ?
RETURNING *;
`

func (q *Queries) UpdateRecurrentExpense(
	ctx context.Context,
	params UpdateRecurrentExpenseParams,
) (RecurrentExpense, error) {
	var re RecurrentExpense

	err := q.wrapQuery(updateRecurrentExpense, func() error {
		row := q.db.QueryRowContext(
			ctx,
			updateRecurrentExpense,
			params.Description,
			params.Amount,
			params.Period,
			params.LastCopyCreatedAt,
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
		)
	})

	return re, err
}

func (q *TxQueries) UpdateRecurrentExpense(
	ctx context.Context,
	params UpdateRecurrentExpenseParams,
) (RecurrentExpense, error) {
	var re RecurrentExpense

	err := q.wrapQuery(updateRecurrentExpense, func() error {
		row := q.tx.QueryRowContext(
			ctx,
			updateRecurrentExpense,
			params.Description,
			params.Amount,
			params.Period,
			params.LastCopyCreatedAt,
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
		)
	})

	return re, err
}

const selectRecurrentExpense = `
SELECT * FROM "recurrent_expenses" WHERE "id" = ? AND "user_id" = ? LIMIT 1
`

func (q *Queries) SelectRecurrentExpense(ctx context.Context, id, userID int) (RecurrentExpense, error) {
	var re RecurrentExpense

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
		)
	})

	return re, err
}

const selectDueRecurrentExpenses = `
SELECT *
FROM "recurrent_expenses"
WHERE "last_copy_created_at" IS NULL
   OR (
        CASE
          WHEN CAST(strftime('%d', datetime(?, 'unixepoch')) AS int) <
               CAST(strftime('%d', datetime("last_copy_created_at", 'unixepoch')) AS int)
          THEN (
            (CAST(strftime('%Y', datetime(?, 'unixepoch')) AS int) -
             CAST(strftime('%Y', datetime("last_copy_created_at", 'unixepoch')) AS int)) * 12 +
            (CAST(strftime('%m', datetime(?, 'unixepoch')) AS int) -
             CAST(strftime('%m', datetime("last_copy_created_at", 'unixepoch')) AS int)) - 1
          )
          ELSE (
            (CAST(strftime('%Y', datetime(?, 'unixepoch')) AS int) -
             CAST(strftime('%Y', datetime("last_copy_created_at", 'unixepoch')) AS int)) * 12 +
            (CAST(strftime('%m', datetime(?, 'unixepoch')) AS int) -
             CAST(strftime('%m', datetime("last_copy_created_at", 'unixepoch')) AS int))
          )
        END
      ) >= "period"
ORDER BY "id" ASC
LIMIT ? OFFSET ?;
`

func (q *Queries) SelectDueRecurrentExpenses(
	ctx context.Context,
	nowUnix int64,
	limit int,
	offset int,
) ([]RecurrentExpense, error) {
	var res []RecurrentExpense

	err := q.wrapQuery(selectDueRecurrentExpenses, func() error {
		rows, err := q.db.QueryContext(
			ctx,
			selectDueRecurrentExpenses,
			nowUnix,
			nowUnix,
			nowUnix,
			nowUnix,
			nowUnix,
			limit,
			offset,
		)
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				q.app.Logger.Error(closeErr)
			}
		}()

		for rows.Next() {
			var re RecurrentExpense

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
			); err != nil {
				return err
			}

			res = append(res, re)
		}

		return nil
	})

	return res, err
}
