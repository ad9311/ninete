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

type ExpenseParams struct {
	UserID      int    `validate:"required"`
	CategoryID  int    `json:"categoryId" validate:"required"`
	Description string `json:"description" validate:"required,min=3,max=50"`
	Amount      uint64 `json:"amount" validate:"required,gt=0"`
	Date        int64  `json:"date" validate:"required"`
}

const insertExpense = `
INSERT INTO "expenses" ("user_id", "category_id", "description", "amount", "date")
VALUES ($1, $2, $3, $4, $5)
RETURNING *`

func (q *Queries) InsertExpense(ctx context.Context, params ExpenseParams) (Expense, error) {
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
SET "category_id" = $2,
	"description" = $3,
	"amount" = $4,
	"date" = $5,
	"updated_at" = strftime('%s','now')
WHERE "id" = $1
RETURNING *`

func (q *Queries) UpdateExpense(ctx context.Context, id int, params ExpenseParams) (Expense, error) {
	var e Expense
	var err error

	q.wrapQuery(updateExpense, func() {
		row := q.db.QueryRowContext(
			ctx,
			updateExpense,
			id,
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

const deleteExpense = `
DELETE FROM "expenses"
WHERE "id" = $1
RETURNING id`

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
