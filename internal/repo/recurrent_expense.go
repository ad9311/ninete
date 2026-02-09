package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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
}

func (re recurrentExpense) toRecurrentExpense() RecurrentExpense {
	var lastCopy *int64
	if re.LastCopyCreatedAt.Valid {
		value := re.LastCopyCreatedAt.Int64
		lastCopy = &value
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
	}
}

func NullInt64FromPtr(value *int64) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{Valid: false}
	}

	return sql.NullInt64{Int64: *value, Valid: true}
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

const selectRecurrentExpenses = `SELECT * FROM "recurrent_expenses"`

func (q *Queries) SelectRecurrentExpenses(ctx context.Context, opts QueryOptions) ([]RecurrentExpense, error) {
	var res []RecurrentExpense

	subQuery, err := opts.Build()
	if err != nil {
		return res, err
	}

	if err := opts.Validate(validRecurrentExpenseFields()); err != nil {
		return res, err
	}

	query := strings.TrimSpace(selectRecurrentExpenses + " " + subQuery)
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
			); err != nil {
				return err
			}

			res = append(res, re.toRecurrentExpense())
		}

		return nil
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

	query := strings.TrimSpace(countRecurrentExpenses + " " + subQuery)
	values := filters.Values()

	err = q.wrapQuery(query, func() error {
		row := q.db.QueryRowContext(ctx, query, values...)

		return row.Scan(&c)
	})

	return c, err
}

func (q *Queries) InsertRecurrentExpense(
	ctx context.Context,
	params InsertRecurrentExpenseParams,
) (RecurrentExpense, error) {
	var re recurrentExpense

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

	return re.toRecurrentExpense(), err
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

	return re.toRecurrentExpense(), err
}

func (q *Queries) DeleteRecurrentExpense(ctx context.Context, id, userID int) (int, error) {
	var i int

	err := q.wrapQuery(deleteRecurrentExpense, func() error {
		row := q.db.QueryRowContext(ctx, deleteRecurrentExpense, id, userID)

		return row.Scan(&i)
	})

	return i, err
}

const selectRecurrentExpense = `
SELECT * FROM "recurrent_expenses" WHERE "id" = ? AND "user_id" = ? LIMIT 1
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
		)
	})

	return re.toRecurrentExpense(), err
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
`

func (q *Queries) SelectDueRecurrentExpenses(
	ctx context.Context,
	nowUnix int64,
	sorting Sorting,
	pagination Pagination,
) ([]RecurrentExpense, error) {
	var res []RecurrentExpense

	if sorting.Field == "" && sorting.Order == "" {
		sorting = Sorting{
			Field: "id",
			Order: "ASC",
		}
	}

	if !sorting.ValidField(validRecurrentExpenseFields()) {
		availableFields := strings.Join(validRecurrentExpenseFields(), ",")

		return nil, fmt.Errorf("%w, valid fields for sorting are: %s", ErrInvalidField, availableFields)
	}

	sortingQuery, err := sorting.Build()
	if err != nil {
		return nil, err
	}

	paginationQuery, err := pagination.Build()
	if err != nil {
		return nil, err
	}

	query := strings.TrimSpace(selectDueRecurrentExpenses + " " + sortingQuery + " " + paginationQuery)

	err = q.wrapQuery(query, func() error {
		rows, err := q.db.QueryContext(
			ctx,
			query,
			nowUnix,
			nowUnix,
			nowUnix,
			nowUnix,
			nowUnix,
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
			); err != nil {
				return err
			}

			res = append(res, re.toRecurrentExpense())
		}

		return nil
	})

	return res, err
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
	}
}
