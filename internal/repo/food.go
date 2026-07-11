package repo

import (
	"context"
)

type Food struct {
	ID            int
	UserID        int
	Name          string
	Kcal          float64
	ProteinG      float64
	CarbsG        float64
	FatG          float64
	CreatedAt     int64
	UpdatedAt     int64
	FiberG        float64
	SodiumG       float64
	SaturatedFatG float64
}

type InsertFoodParams struct {
	UserID        int
	Name          string
	Kcal          float64
	ProteinG      float64
	CarbsG        float64
	FatG          float64
	FiberG        float64
	SodiumG       float64
	SaturatedFatG float64
}

type UpdateFoodParams struct {
	ID            int
	Name          string
	Kcal          float64
	ProteinG      float64
	CarbsG        float64
	FatG          float64
	FiberG        float64
	SodiumG       float64
	SaturatedFatG float64
}

const selectFoods = `SELECT * FROM "foods"`

func (q *Queries) SelectFoods(ctx context.Context, opts QueryOptions) ([]Food, error) {
	var fs []Food

	if err := opts.Validate(validFoodFields()); err != nil {
		return fs, err
	}

	subQuery, err := opts.Build()
	if err != nil {
		return fs, err
	}

	query := selectFoods + " " + subQuery
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
			var f Food

			if err := rows.Scan(
				&f.ID,
				&f.UserID,
				&f.Name,
				&f.Kcal,
				&f.ProteinG,
				&f.CarbsG,
				&f.FatG,
				&f.CreatedAt,
				&f.UpdatedAt,
				&f.FiberG,
				&f.SodiumG,
				&f.SaturatedFatG,
			); err != nil {
				return err
			}

			fs = append(fs, f)
		}

		return rows.Err()
	})

	return fs, err
}

const selectFood = `SELECT * FROM "foods" WHERE "id" = ? AND "user_id" = ? LIMIT 1`

func (q *Queries) SelectFood(ctx context.Context, id, userID int) (Food, error) {
	var f Food

	err := q.wrapQuery(selectFood, func() error {
		row := q.db.QueryRowContext(ctx, selectFood, id, userID)

		return row.Scan(
			&f.ID,
			&f.UserID,
			&f.Name,
			&f.Kcal,
			&f.ProteinG,
			&f.CarbsG,
			&f.FatG,
			&f.CreatedAt,
			&f.UpdatedAt,
			&f.FiberG,
			&f.SodiumG,
			&f.SaturatedFatG,
		)
	})

	return f, err
}

const insertFood = `
INSERT INTO "foods"
  ("user_id", "name", "kcal", "protein_g", "carbs_g", "fat_g",
   "fiber_g", "sodium_g", "saturated_fat_g")
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *`

func (q *TxQueries) InsertFood(ctx context.Context, params InsertFoodParams) (Food, error) {
	var f Food

	err := q.wrapQuery(insertFood, func() error {
		row := q.tx.QueryRowContext(
			ctx,
			insertFood,
			params.UserID,
			params.Name,
			params.Kcal,
			params.ProteinG,
			params.CarbsG,
			params.FatG,
			params.FiberG,
			params.SodiumG,
			params.SaturatedFatG,
		)

		return row.Scan(
			&f.ID,
			&f.UserID,
			&f.Name,
			&f.Kcal,
			&f.ProteinG,
			&f.CarbsG,
			&f.FatG,
			&f.CreatedAt,
			&f.UpdatedAt,
			&f.FiberG,
			&f.SodiumG,
			&f.SaturatedFatG,
		)
	})

	return f, err
}

const updateFood = `
UPDATE "foods"
SET "name"            = ?,
    "kcal"            = ?,
    "protein_g"       = ?,
    "carbs_g"         = ?,
    "fat_g"           = ?,
    "fiber_g"         = ?,
    "sodium_g"        = ?,
    "saturated_fat_g" = ?,
    "updated_at"      = ?
WHERE "id" = ?
  AND "user_id" = ?
RETURNING *`

func (q *TxQueries) UpdateFood(
	ctx context.Context,
	userID int,
	params UpdateFoodParams,
) (Food, error) {
	var f Food

	err := q.wrapQuery(updateFood, func() error {
		row := q.tx.QueryRowContext(
			ctx,
			updateFood,
			params.Name,
			params.Kcal,
			params.ProteinG,
			params.CarbsG,
			params.FatG,
			params.FiberG,
			params.SodiumG,
			params.SaturatedFatG,
			newUpdatedAt(),
			params.ID,
			userID,
		)

		return row.Scan(
			&f.ID,
			&f.UserID,
			&f.Name,
			&f.Kcal,
			&f.ProteinG,
			&f.CarbsG,
			&f.FatG,
			&f.CreatedAt,
			&f.UpdatedAt,
			&f.FiberG,
			&f.SodiumG,
			&f.SaturatedFatG,
		)
	})

	return f, err
}

const deleteFood = `DELETE FROM "foods" WHERE "id" = ? AND "user_id" = ? RETURNING "id"`

func (q *Queries) DeleteFood(ctx context.Context, id, userID int) (int, error) {
	var i int

	err := q.wrapQuery(deleteFood, func() error {
		row := q.db.QueryRowContext(ctx, deleteFood, id, userID)

		return row.Scan(&i)
	})

	return i, err
}

const countFoodsByUser = `SELECT COUNT(*) FROM "foods" WHERE "user_id" = ?`

func (q *Queries) CountFoodsByUser(ctx context.Context, userID int) (int, error) {
	var c int

	err := q.wrapQuery(countFoodsByUser, func() error {
		row := q.db.QueryRowContext(ctx, countFoodsByUser, userID)

		return row.Scan(&c)
	})

	return c, err
}

const deleteAllFoodsByUser = `DELETE FROM "foods" WHERE "user_id" = ?`

func (q *TxQueries) DeleteAllFoodsByUser(ctx context.Context, userID int) error {
	return q.wrapQuery(deleteAllFoodsByUser, func() error {
		_, err := q.tx.ExecContext(ctx, deleteAllFoodsByUser, userID)

		return err
	})
}

func validFoodFields() []string {
	return []string{
		"id",
		"user_id",
		"name",
		"kcal",
		"protein_g",
		"carbs_g",
		"fat_g",
		"created_at",
		"updated_at",
		"fiber_g",
		"sodium_g",
		"saturated_fat_g",
	}
}
