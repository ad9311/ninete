package repo

import (
	"context"
	"fmt"
	"strings"
)

type Expense struct {
	ID          int
	UserID      int
	CategoryID  int
	Description string
	Amount      uint64
	Date        int64
	CreatedAt   int64
	UpdatedAt   int64
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

// expenseColumns pins the projection order the Scan calls in this file depend on.
// SELECT * would resolve to whatever order the table happens to have, so an
// ALTER TABLE could shift values into the wrong struct fields with no error.
const expenseColumns = `"id", "user_id", "category_id", "description", "amount", "date", "created_at", "updated_at"`

const selectExpenses = `SELECT ` + expenseColumns + ` FROM "expenses"`

// expenseDescriptionExpr matches a description substring. SQLite's LIKE is
// case-insensitive for ASCII, so no lower() wrapper is needed (and a leading
// wildcard rules out an index either way).
const expenseDescriptionExpr = `"description" LIKE ? ESCAPE '\'`

// expenseTagExpr keeps the tag lookup index-friendly: the inner SELECT resolves
// the tag through "uq_tags_user_lower_name" (user_id, lower(name)), then walks
// "uq_taggings_tag_target" (tag_id, taggable_type, taggable_id) to collect the
// tagged ids. The subquery is uncorrelated on purpose — SQLite materializes the
// (small) id set once and drives from it, instead of probing the taggings index
// once per candidate expense row as a correlated EXISTS would.
const expenseTagExpr = `"expenses"."id" IN (
  SELECT "taggings"."taggable_id" FROM "taggings"
  WHERE "taggings"."taggable_type" = 'expense'
    AND "taggings"."tag_id" IN (
      SELECT "id" FROM "tags" WHERE "user_id" = ? AND lower("name") = ?
    )
)`

// ExpenseDescriptionFilter builds a case-insensitive "description contains"
// predicate. LIKE wildcards in term are escaped so they match literally.
func ExpenseDescriptionFilter(term string) FilterField {
	return FilterField{
		Expr: expenseDescriptionExpr,
		Args: []any{"%" + escapeLikePattern(term) + "%"},
	}
}

// ExpenseTagFilter builds a predicate matching expenses tagged with tagName for
// the given user. tagName is matched case-insensitively.
func ExpenseTagFilter(userID int, tagName string) FilterField {
	return FilterField{
		Expr: expenseTagExpr,
		Args: []any{userID, strings.ToLower(tagName)},
	}
}

var likePatternEscaper = strings.NewReplacer( //nolint:gochecknoglobals // static replacer
	`\`, `\\`,
	`%`, `\%`,
	`_`, `\_`,
)

func escapeLikePattern(s string) string {
	return likePatternEscaper.Replace(s)
}

func (q *Queries) SelectExpenses(ctx context.Context, opts QueryOptions) ([]Expense, error) {
	var es []Expense

	if err := opts.Validate(validExpenseFields()); err != nil {
		return es, err
	}

	subQuery, err := opts.Build()
	if err != nil {
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

		return rows.Err()
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

const selectExpense = `SELECT ` + expenseColumns + `
FROM "expenses" WHERE "id" = ? AND "user_id" = ? LIMIT 1`

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
RETURNING ` + expenseColumns

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
  AND "user_id" = ?
RETURNING ` + expenseColumns + `;
`

func (q *Queries) UpdateExpense(
	ctx context.Context,
	userID int,
	params UpdateExpenseParams,
) (Expense, error) {
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
			userID,
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

func (q *TxQueries) UpdateExpense(
	ctx context.Context,
	userID int,
	params UpdateExpenseParams,
) (Expense, error) {
	var e Expense

	err := q.wrapQuery(updateExpense, func() error {
		row := q.tx.QueryRowContext(
			ctx,
			updateExpense,
			params.CategoryID,
			params.Description,
			params.Amount,
			params.Date,
			newUpdatedAt(),
			params.ID,
			userID,
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

const deleteExpense = `DELETE FROM "expenses" WHERE "id" = ? AND "user_id" = ? RETURNING "id"`

func (q *Queries) DeleteExpense(ctx context.Context, id, userID int) (int, error) {
	var i int

	err := q.wrapQuery(deleteExpense, func() error {
		row := q.db.QueryRowContext(ctx, deleteExpense, id, userID)

		return row.Scan(&i)
	})

	return i, err
}

const countExpensesByUser = `SELECT COUNT(*) FROM "expenses" WHERE "user_id" = ?`

func (q *Queries) CountExpensesByUser(ctx context.Context, userID int) (int, error) {
	var c int

	err := q.wrapQuery(countExpensesByUser, func() error {
		row := q.db.QueryRowContext(ctx, countExpensesByUser, userID)

		return row.Scan(&c)
	})

	return c, err
}

const deleteExpenseTaggingsByUser = `
DELETE FROM "taggings"
WHERE "taggable_type" = 'expense'
  AND "taggable_id" IN (SELECT "id" FROM "expenses" WHERE "user_id" = ?)`

const deleteAllExpensesByUser = `DELETE FROM "expenses" WHERE "user_id" = ?`

func (q *TxQueries) DeleteAllExpensesByUser(ctx context.Context, userID int) error {
	return q.wrapQuery(deleteAllExpensesByUser, func() error {
		if _, err := q.tx.ExecContext(ctx, deleteExpenseTaggingsByUser, userID); err != nil {
			return err
		}

		_, err := q.tx.ExecContext(ctx, deleteAllExpensesByUser, userID)

		return err
	})
}

const selectExpensesCategoryTotals = `
SELECT "category_id", SUM("amount") AS "total"
FROM "expenses"
%s
GROUP BY "category_id"`

type ExpenseCategoryTotal struct {
	CategoryID int
	Total      uint64
}

func (q *Queries) SelectExpensesCategoryTotals(ctx context.Context, filters Filters) ([]ExpenseCategoryTotal, error) {
	var totals []ExpenseCategoryTotal

	filterSubQuery, err := filters.Build()
	if err != nil {
		return totals, err
	}

	query := fmt.Sprintf(selectExpensesCategoryTotals, filterSubQuery)
	values := filters.Values()

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
			var t ExpenseCategoryTotal

			if err := rows.Scan(&t.CategoryID, &t.Total); err != nil {
				return err
			}

			totals = append(totals, t)
		}

		return rows.Err()
	})

	return totals, err
}

// selectExpensesCategoryMonthTotals groups by calendar month as well as
// category. strftime with 'unixepoch' yields UTC months, matching the UTC
// range boundaries the client resolves and sends (§3.6 of
// docs/spa-migration.md) — do not switch one side to the client zone.
const selectExpensesCategoryMonthTotals = `
SELECT "category_id", strftime('%%Y-%%m', "date", 'unixepoch') AS "month", SUM("amount") AS "total"
FROM "expenses"
%s
GROUP BY "category_id", "month"`

type ExpenseCategoryMonthTotal struct {
	CategoryID int
	Month      string
	Total      uint64
}

func (q *Queries) SelectExpensesCategoryMonthTotals(
	ctx context.Context,
	filters Filters,
) ([]ExpenseCategoryMonthTotal, error) {
	var totals []ExpenseCategoryMonthTotal

	filterSubQuery, err := filters.Build()
	if err != nil {
		return totals, err
	}

	query := fmt.Sprintf(selectExpensesCategoryMonthTotals, filterSubQuery)
	values := filters.Values()

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
			var t ExpenseCategoryMonthTotal

			if err := rows.Scan(&t.CategoryID, &t.Month, &t.Total); err != nil {
				return err
			}

			totals = append(totals, t)
		}

		return rows.Err()
	})

	return totals, err
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
