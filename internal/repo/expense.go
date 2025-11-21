package repo

import (
	"context"
)

type Expense struct {
	ID          int    `json:"id"`
	UserID      int    `json:"userId"`
	CategoryID  int    `json:"categoryId"`
	Description string `json:"description"`
	Amount      uint64 `json:"amount"`
	Date        int64  `json:"date"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}

type InsertExpenseParams struct {
	UserID      int
	CategoryID  int
	Description string
	Amount      uint64
	Date        int64
}

type UpdateExpenseParams struct {
	ID          int
	CategoryID  int
	Description string
	Amount      uint64
	Date        int64
}

const selectExpense = `SELECT * FROM "expenses" WHERE "id" = $1 AND "user_id" = $2 LIMIT 1`

func (q *Queries) SelectExpense(ctx context.Context, id, userID int) (Expense, error) {
	var e Expense
	var err error

	q.wrapQuery(selectExpense, func() {
		row := q.db.QueryRowContext(
			ctx,
			selectExpense,
			id,
			userID,
		)

		err = row.Scan(
			&e.ID,
			&e.UserID,
			&e.CategoryID,
			&e.Description,
			&e.Amount,
			&e.Date,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
	})

	return e, err
}

const insertExpense = `
INSERT INTO "expenses" ("user_id", "category_id", "description", "amount", "date")
VALUES ($1, $2, $3, $4, $5)
RETURNING *`

func (q *Queries) InsertExpense(ctx context.Context, params InsertExpenseParams) (Expense, error) {
	var e Expense
	var err error

	q.wrapQuery(insertExpense, func() {
		row := q.db.QueryRowContext(
			ctx,
			insertExpense,
			params.UserID,
			params.CategoryID,
			params.Description,
			params.Amount,
			params.Date,
		)

		err = row.Scan(
			&e.ID,
			&e.UserID,
			&e.CategoryID,
			&e.Description,
			&e.Amount,
			&e.Date,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
	})

	return e, err
}

const updateExpense = `
UPDATE "expenses"
SET "category_id" = $2, "description" = $3, "amount" = $4, "date" = $5, "updated_at" = strftime('%s','now')
WHERE "id" = $1
RETURNING *`

func (q *Queries) UpdateExpense(ctx context.Context, params UpdateExpenseParams) (Expense, error) {
	var e Expense
	var err error

	q.wrapQuery(updateExpense, func() {
		row := q.db.QueryRowContext(
			ctx,
			updateExpense,
			params.ID,
			params.CategoryID,
			params.Description,
			params.Amount,
			params.Date,
		)

		err = row.Scan(
			&e.ID,
			&e.UserID,
			&e.CategoryID,
			&e.Description,
			&e.Amount,
			&e.Date,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
	})

	return e, err
}

const deleteExpense = `DELETE FROM "expenses" WHERE "id" = $1 RETURNING *`

func (q *Queries) DeleteExpense(ctx context.Context, id int) (Expense, error) {
	var e Expense
	var err error

	q.wrapQuery(deleteExpense, func() {
		row := q.db.QueryRowContext(
			ctx,
			deleteExpense,
			id,
		)

		err = row.Scan(
			&e.ID,
			&e.UserID,
			&e.CategoryID,
			&e.Description,
			&e.Amount,
			&e.Date,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
	})

	return e, err
}
