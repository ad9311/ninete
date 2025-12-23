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
	ID          int
	CategoryID  int
	Description string
	Amount      uint64
	Period      uint
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

const updateLastCopyCreated = `
UPDATE "recurrent_expenses"
SET "last_copy_created_at"  = ?,
    "updated_at"            = ?
WHERE "id" = ?
RETURNING "last_copy_created_at";
`

func (q *Queries) UpdateLastCopyCreated(
	ctx context.Context,
	id int,
	lastCopyCreated int64,
) (RecurrentExpense, error) {
	var re RecurrentExpense

	err := q.wrapQuery(updateLastCopyCreated, func() error {
		row := q.db.QueryRowContext(
			ctx,
			updateLastCopyCreated,
			lastCopyCreated,
			newUpdatedAt(),
			id,
		)

		return row.Scan(&re.LastCopyCreatedAt)
	})

	return re, err
}

const selectRecurrentExpense = `SELECT * FROM "recurrent_expenses" WHERE "id" = ? LIMIT 1`

func (q *Queries) SelectRecurrentExpense(ctx context.Context, id int) (RecurrentExpense, error) {
	var re RecurrentExpense

	err := q.wrapQuery(selectRecurrentExpense, func() error {
		row := q.db.QueryRowContext(ctx, selectRecurrentExpense, id)

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
