package repo

import (
	"context"
)

type MacroGoal struct {
	ID            int
	UserID        int
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

type UpsertMacroGoalParams struct {
	UserID        int
	Kcal          float64
	ProteinG      float64
	CarbsG        float64
	FatG          float64
	FiberG        float64
	SodiumG       float64
	SaturatedFatG float64
}

const selectMacroGoal = `SELECT * FROM "macro_goals" WHERE "user_id" = ? LIMIT 1`

func (q *Queries) SelectMacroGoal(ctx context.Context, userID int) (MacroGoal, error) {
	var g MacroGoal

	err := q.wrapQuery(selectMacroGoal, func() error {
		row := q.db.QueryRowContext(ctx, selectMacroGoal, userID)

		return row.Scan(
			&g.ID,
			&g.UserID,
			&g.Kcal,
			&g.ProteinG,
			&g.CarbsG,
			&g.FatG,
			&g.CreatedAt,
			&g.UpdatedAt,
			&g.FiberG,
			&g.SodiumG,
			&g.SaturatedFatG,
		)
	})

	return g, err
}

const upsertMacroGoal = `
INSERT INTO "macro_goals"
  ("user_id","kcal","protein_g","carbs_g","fat_g","fiber_g","sodium_g","saturated_fat_g")
VALUES (?,?,?,?,?,?,?,?)
ON CONFLICT ("user_id") DO UPDATE SET
  "kcal"            = excluded."kcal",
  "protein_g"       = excluded."protein_g",
  "carbs_g"         = excluded."carbs_g",
  "fat_g"           = excluded."fat_g",
  "fiber_g"         = excluded."fiber_g",
  "sodium_g"        = excluded."sodium_g",
  "saturated_fat_g" = excluded."saturated_fat_g",
  "updated_at"      = strftime('%s','now')
RETURNING *`

func (q *TxQueries) UpsertMacroGoal(ctx context.Context, params UpsertMacroGoalParams) (MacroGoal, error) {
	var g MacroGoal

	err := q.wrapQuery(upsertMacroGoal, func() error {
		row := q.tx.QueryRowContext(
			ctx,
			upsertMacroGoal,
			params.UserID,
			params.Kcal,
			params.ProteinG,
			params.CarbsG,
			params.FatG,
			params.FiberG,
			params.SodiumG,
			params.SaturatedFatG,
		)

		return row.Scan(
			&g.ID,
			&g.UserID,
			&g.Kcal,
			&g.ProteinG,
			&g.CarbsG,
			&g.FatG,
			&g.CreatedAt,
			&g.UpdatedAt,
			&g.FiberG,
			&g.SodiumG,
			&g.SaturatedFatG,
		)
	})

	return g, err
}
