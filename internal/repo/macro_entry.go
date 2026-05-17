package repo

import (
	"context"
)

const (
	MacroEntryMealTypeBreakfast = "breakfast"
	MacroEntryMealTypeLunch     = "lunch"
	MacroEntryMealTypeDinner    = "dinner"
	MacroEntryMealTypeSnack     = "snack"
	MacroEntryMealTypeOther     = "other"
)

type MacroEntry struct {
	ID        int
	UserID    int
	Name      string
	Kcal      float64
	ProteinG  float64
	CarbsG    float64
	FatG      float64
	Date      int64
	CreatedAt int64
	UpdatedAt int64
	MealType  string
}

type InsertMacroEntryParams struct {
	UserID   int
	Name     string
	Kcal     float64
	ProteinG float64
	CarbsG   float64
	FatG     float64
	Date     int64
	MealType string
}

type UpdateMacroEntryParams struct {
	ID       int
	Name     string
	Kcal     float64
	ProteinG float64
	CarbsG   float64
	FatG     float64
	Date     int64
	MealType string
}

type MacroDayTotals struct {
	Kcal     float64
	ProteinG float64
	CarbsG   float64
	FatG     float64
}

const selectMacroEntries = `SELECT * FROM "macro_entries"`

func (q *Queries) SelectMacroEntries(ctx context.Context, opts QueryOptions) ([]MacroEntry, error) {
	var es []MacroEntry

	if err := opts.Validate(validMacroEntryFields()); err != nil {
		return es, err
	}

	subQuery, err := opts.Build()
	if err != nil {
		return es, err
	}

	query := selectMacroEntries + " " + subQuery
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
			var e MacroEntry

			if err := rows.Scan(
				&e.ID,
				&e.UserID,
				&e.Name,
				&e.Kcal,
				&e.ProteinG,
				&e.CarbsG,
				&e.FatG,
				&e.Date,
				&e.CreatedAt,
				&e.UpdatedAt,
				&e.MealType,
			); err != nil {
				return err
			}

			es = append(es, e)
		}

		return rows.Err()
	})

	return es, err
}

const countMacroEntries = `SELECT COUNT(*) FROM "macro_entries"`

func (q *Queries) CountMacroEntries(ctx context.Context, filters Filters) (int, error) {
	var c int

	subQuery, err := filters.Build()
	if err != nil {
		return 0, err
	}

	query := countMacroEntries + " " + subQuery
	values := filters.Values()

	err = q.wrapQuery(query, func() error {
		row := q.db.QueryRowContext(ctx, query, values...)

		return row.Scan(&c)
	})

	return c, err
}

const selectMacroEntry = `SELECT * FROM "macro_entries" WHERE "id" = ? AND "user_id" = ? LIMIT 1`

func (q *Queries) SelectMacroEntry(ctx context.Context, id, userID int) (MacroEntry, error) {
	var e MacroEntry

	err := q.wrapQuery(selectMacroEntry, func() error {
		row := q.db.QueryRowContext(ctx, selectMacroEntry, id, userID)

		return row.Scan(
			&e.ID,
			&e.UserID,
			&e.Name,
			&e.Kcal,
			&e.ProteinG,
			&e.CarbsG,
			&e.FatG,
			&e.Date,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.MealType,
		)
	})

	return e, err
}

const insertMacroEntry = `
INSERT INTO "macro_entries" ("user_id", "name", "kcal", "protein_g", "carbs_g", "fat_g", "date", "meal_type")
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *`

func (q *TxQueries) InsertMacroEntry(ctx context.Context, params InsertMacroEntryParams) (MacroEntry, error) {
	var e MacroEntry

	err := q.wrapQuery(insertMacroEntry, func() error {
		row := q.tx.QueryRowContext(
			ctx,
			insertMacroEntry,
			params.UserID,
			params.Name,
			params.Kcal,
			params.ProteinG,
			params.CarbsG,
			params.FatG,
			params.Date,
			params.MealType,
		)

		return row.Scan(
			&e.ID,
			&e.UserID,
			&e.Name,
			&e.Kcal,
			&e.ProteinG,
			&e.CarbsG,
			&e.FatG,
			&e.Date,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.MealType,
		)
	})

	return e, err
}

const updateMacroEntry = `
UPDATE "macro_entries"
SET "name"       = ?,
    "kcal"       = ?,
    "protein_g"  = ?,
    "carbs_g"    = ?,
    "fat_g"      = ?,
    "date"       = ?,
    "meal_type"  = ?,
    "updated_at" = ?
WHERE "id" = ?
  AND "user_id" = ?
RETURNING *`

func (q *TxQueries) UpdateMacroEntry(
	ctx context.Context,
	userID int,
	params UpdateMacroEntryParams,
) (MacroEntry, error) {
	var e MacroEntry

	err := q.wrapQuery(updateMacroEntry, func() error {
		row := q.tx.QueryRowContext(
			ctx,
			updateMacroEntry,
			params.Name,
			params.Kcal,
			params.ProteinG,
			params.CarbsG,
			params.FatG,
			params.Date,
			params.MealType,
			newUpdatedAt(),
			params.ID,
			userID,
		)

		return row.Scan(
			&e.ID,
			&e.UserID,
			&e.Name,
			&e.Kcal,
			&e.ProteinG,
			&e.CarbsG,
			&e.FatG,
			&e.Date,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.MealType,
		)
	})

	return e, err
}

const deleteMacroEntry = `DELETE FROM "macro_entries" WHERE "id" = ? AND "user_id" = ? RETURNING "id"`

func (q *Queries) DeleteMacroEntry(ctx context.Context, id, userID int) (int, error) {
	var i int

	err := q.wrapQuery(deleteMacroEntry, func() error {
		row := q.db.QueryRowContext(ctx, deleteMacroEntry, id, userID)

		return row.Scan(&i)
	})

	return i, err
}

const selectMacroDayTotals = `
SELECT COALESCE(SUM("kcal"),0), COALESCE(SUM("protein_g"),0),
       COALESCE(SUM("carbs_g"),0), COALESCE(SUM("fat_g"),0)
FROM "macro_entries" WHERE "user_id"=? AND "date">=? AND "date"<?`

const selectMacroDayTotalsByMealType = selectMacroDayTotals + ` AND "meal_type"=?`

func (q *Queries) SelectMacroDayTotals(
	ctx context.Context,
	userID int,
	dayStart, nextDayStart int64,
	mealType string,
) (MacroDayTotals, error) {
	var t MacroDayTotals

	query := selectMacroDayTotals
	args := []any{userID, dayStart, nextDayStart}

	if mealType != "" {
		query = selectMacroDayTotalsByMealType
		args = append(args, mealType)
	}

	err := q.wrapQuery(query, func() error {
		row := q.db.QueryRowContext(ctx, query, args...)

		return row.Scan(&t.Kcal, &t.ProteinG, &t.CarbsG, &t.FatG)
	})

	return t, err
}

const selectMacroDailyTotals = `
SELECT "date",
       COALESCE(SUM("kcal"), 0),
       COALESCE(SUM("protein_g"), 0),
       COALESCE(SUM("carbs_g"), 0),
       COALESCE(SUM("fat_g"), 0)
FROM "macro_entries"
WHERE "user_id" = ? AND "date" >= ? AND "date" < ?
GROUP BY "date"
ORDER BY "date" ASC`

type MacroDailyTotal struct {
	Date     int64
	Kcal     float64
	ProteinG float64
	CarbsG   float64
	FatG     float64
}

func (q *Queries) SelectMacroDailyTotals(ctx context.Context, userID int, start, end int64) ([]MacroDailyTotal, error) {
	var totals []MacroDailyTotal

	err := q.wrapQuery(selectMacroDailyTotals, func() error {
		rows, err := q.db.QueryContext(ctx, selectMacroDailyTotals, userID, start, end)
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				q.app.Logger.Error(closeErr)
			}
		}()

		for rows.Next() {
			var t MacroDailyTotal
			if err := rows.Scan(&t.Date, &t.Kcal, &t.ProteinG, &t.CarbsG, &t.FatG); err != nil {
				return err
			}
			totals = append(totals, t)
		}

		return rows.Err()
	})

	return totals, err
}

func validMacroEntryFields() []string {
	return []string{
		"id",
		"user_id",
		"name",
		"kcal",
		"protein_g",
		"carbs_g",
		"fat_g",
		"date",
		"created_at",
		"updated_at",
		"meal_type",
	}
}
