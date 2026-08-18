package repo

import (
	"context"
)

type ExpenseBudget struct {
	ID         int
	UserID     int
	CategoryID int
	Amount     uint64
	CreatedAt  int64
	UpdatedAt  int64
}

type UpsertExpenseBudgetParams struct {
	UserID     int
	CategoryID int
	Amount     uint64
}

// expenseBudgetColumns pins the projection order the Scan calls in this file
// depend on. SELECT * would resolve to whatever order the table happens to
// have, so an ALTER TABLE could shift values into the wrong struct fields with
// no error.
const expenseBudgetColumns = `"id", "user_id", "category_id", "amount",
"created_at", "updated_at"`

const selectExpenseBudgetsByUser = `SELECT ` + expenseBudgetColumns + `
FROM "expense_budgets" WHERE "user_id" = ?`

func (q *Queries) SelectExpenseBudgetsByUser(ctx context.Context, userID int) ([]ExpenseBudget, error) {
	var budgets []ExpenseBudget

	err := q.wrapQuery(selectExpenseBudgetsByUser, func() error {
		rows, err := q.db.QueryContext(ctx, selectExpenseBudgetsByUser, userID)
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				q.app.Logger.Error(closeErr)
			}
		}()

		for rows.Next() {
			var b ExpenseBudget

			if err := rows.Scan(
				&b.ID,
				&b.UserID,
				&b.CategoryID,
				&b.Amount,
				&b.CreatedAt,
				&b.UpdatedAt,
			); err != nil {
				return err
			}

			budgets = append(budgets, b)
		}

		return rows.Err()
	})

	return budgets, err
}

const countExpenseBudgetsByUser = `SELECT COUNT(*) FROM "expense_budgets" WHERE "user_id" = ?`

func (q *Queries) CountExpenseBudgetsByUser(ctx context.Context, userID int) (int, error) {
	var c int

	err := q.wrapQuery(countExpenseBudgetsByUser, func() error {
		row := q.db.QueryRowContext(ctx, countExpenseBudgetsByUser, userID)

		return row.Scan(&c)
	})

	return c, err
}

const upsertExpenseBudget = `
INSERT INTO "expense_budgets" ("user_id","category_id","amount")
VALUES (?,?,?)
ON CONFLICT ("user_id","category_id") DO UPDATE SET
  "amount"     = excluded."amount",
  "updated_at" = strftime('%s','now')
RETURNING ` + expenseBudgetColumns

func (q *TxQueries) UpsertExpenseBudget(
	ctx context.Context,
	params UpsertExpenseBudgetParams,
) (ExpenseBudget, error) {
	var b ExpenseBudget

	err := q.wrapQuery(upsertExpenseBudget, func() error {
		row := q.tx.QueryRowContext(
			ctx,
			upsertExpenseBudget,
			params.UserID,
			params.CategoryID,
			params.Amount,
		)

		return row.Scan(
			&b.ID,
			&b.UserID,
			&b.CategoryID,
			&b.Amount,
			&b.CreatedAt,
			&b.UpdatedAt,
		)
	})

	return b, err
}

const deleteExpenseBudget = `
DELETE FROM "expense_budgets" WHERE "user_id" = ? AND "category_id" = ?`

func (q *TxQueries) DeleteExpenseBudget(ctx context.Context, userID, categoryID int) error {
	return q.wrapQuery(deleteExpenseBudget, func() error {
		_, err := q.tx.ExecContext(ctx, deleteExpenseBudget, userID, categoryID)

		return err
	})
}

const deleteAllExpenseBudgetsByUser = `DELETE FROM "expense_budgets" WHERE "user_id" = ?`

func (q *TxQueries) DeleteAllExpenseBudgetsByUser(ctx context.Context, userID int) error {
	return q.wrapQuery(deleteAllExpenseBudgetsByUser, func() error {
		_, err := q.tx.ExecContext(ctx, deleteAllExpenseBudgetsByUser, userID)

		return err
	})
}
