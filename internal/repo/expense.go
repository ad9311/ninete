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
	UserID      int    `validate:"required"`
	CategoryID  int    `validate:"required"`
	Description string `validate:"required,min=4,max=100"`
	Amount      uint64 `validate:"required,gt=0"`
	Date        int64  `validate:"required"`
}

type UpdateExpenseParams struct {
	ID          int    `validate:"required"`
	UserID      int    `validate:"required"`
	CategoryID  int    `validate:"required"`
	Description string `validate:"required,min=4,max=100"`
	Amount      uint64 `validate:"required,gt=0"`
	Date        int64  `validate:"required"`
}

const insertExpense = `
INSERT INTO "expenses" ("user_id", "category_id", "description", "amount", "date")
VALUES ($1, $2, $3, $4, $5)
RETURNING *`

func (q *Queries) InsertExpense(ctx context.Context, arg InsertExpenseParams) (Expense, error) {
	var e Expense
	var err error

	q.wrapQuery(insertExpense, func() {
		row := q.db.QueryRowContext(
			ctx,
			insertExpense,
			arg.UserID,
			arg.CategoryID,
			arg.Description,
			arg.Amount,
			arg.Date,
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
WHERE "id" = $1 AND "user_id" = $6
RETURNING *`

func (q *Queries) UpdateExpense(ctx context.Context, arg UpdateExpenseParams) (Expense, error) {
	var e Expense
	var err error

	q.wrapQuery(updateExpense, func() {
		row := q.db.QueryRowContext(
			ctx,
			updateExpense,
			arg.ID,
			arg.CategoryID,
			arg.Description,
			arg.Amount,
			arg.Date,
			arg.UserID,
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
