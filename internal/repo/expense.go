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

const selectExpenses = `SELECT * FROM "expenses"`

func (q *Queries) SelectExpenses(ctx context.Context, opts QueryOptions) ([]Expense, error) {
	var es []Expense

	subQuery, err := opts.Build()
	if err != nil {
		return es, err
	}

	if err := opts.Validate(validExpenseFields()); err != nil {
		return es, err
	}

	query := selectExpenses + " " + subQuery
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
			var e Expense

			if err := rows.Scan(
				&e.ID,
				&e.UserID,
				&e.CategoryID,
				&e.Description,
				&e.Amount,
				&e.Date,
				&e.CreatedAt,
				&e.UpdatedAt,
			); err != nil {
				return err
			}

			es = append(es, e)
		}

		return err
	})

	return es, err
}

const countExpenses = `SELECT COUNT(*) FROM "expenses"`

func (q *Queries) CountExpenses(ctx context.Context, filters Filters) (int, error) {
	var c int

	subQuery, err := filters.Build()
	if err != nil {
		return 0, err
	}

	query := countExpenses + " " + subQuery
	values := filters.Values()

	err = q.wrapQuery(query, func() error {
		row := q.db.QueryRowContext(ctx, query, values...)

		return row.Scan(&c)
	})

	return c, err
}

const selectExpense = `SELECT * FROM "expenses" WHERE "id" = ? AND "user_id" = ? LIMIT 1`

func (q *Queries) SelectExpense(ctx context.Context, id, userID int) (Expense, error) {
	var e Expense

	err := q.wrapQuery(selectExpense, func() error {
		row := q.db.QueryRowContext(ctx, selectExpense, id, userID)

		return row.Scan(
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
VALUES (?, ?, ?, ?, ?)
RETURNING *`

func (q *Queries) InsertExpense(ctx context.Context, params InsertExpenseParams) (Expense, error) {
	var e Expense

	err := q.wrapQuery(insertExpense, func() error {
		row := q.db.QueryRowContext(
			ctx,
			insertExpense,
			params.UserID,
			params.CategoryID,
			params.Description,
			params.Amount,
			params.Date,
		)

		return row.Scan(
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

func (q *TxQueries) InsertExpense(ctx context.Context, params InsertExpenseParams) (Expense, error) {
	var e Expense

	err := q.wrapQuery(insertExpense, func() error {
		row := q.tx.QueryRowContext(
			ctx,
			insertExpense,
			params.UserID,
			params.CategoryID,
			params.Description,
			params.Amount,
			params.Date,
		)

		return row.Scan(
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
SET "category_id" = ?,
    "description" = ?,
    "amount"      = ?,
    "date"        = ?,
    "updated_at"  = ?
WHERE "id" = ?
RETURNING *;
`

func (q *Queries) UpdateExpense(ctx context.Context, params UpdateExpenseParams) (Expense, error) {
	var e Expense

	err := q.wrapQuery(updateExpense, func() error {
		row := q.db.QueryRowContext(
			ctx,
			updateExpense,
			params.CategoryID,
			params.Description,
			params.Amount,
			params.Date,
			newUpdatedAt(),
			params.ID,
		)

		return row.Scan(
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

const deleteExpense = `DELETE FROM "expenses" WHERE "id" = ? RETURNING "id"`

func (q *Queries) DeleteExpense(ctx context.Context, id int) (int, error) {
	var i int

	err := q.wrapQuery(deleteExpense, func() error {
		row := q.db.QueryRowContext(ctx, deleteExpense, id)

		return row.Scan(&i)
	})

	return i, err
}

func validExpenseFields() []string {
	return []string{
		"id",
		"user_id",
		"category_id",
		"description",
		"amount",
		"date",
		"created_at",
		"updated_at",
	}
}
