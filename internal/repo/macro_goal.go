package repo

import (
	"context"
)

type MacroGoal struct {
	ID        int
	UserID    int
	Kcal      int
	ProteinG  int
	CarbsG    int
	FatG      int
	CreatedAt int64
	UpdatedAt int64
}

type UpsertMacroGoalParams struct {
	UserID   int
	Kcal     int
	ProteinG int
	CarbsG   int
	FatG     int
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
		)
	})

	return g, err
}

const upsertMacroGoal = `
INSERT INTO "macro_goals" ("user_id","kcal","protein_g","carbs_g","fat_g") VALUES (?,?,?,?,?)
ON CONFLICT ("user_id") DO UPDATE SET
  "kcal"       = excluded."kcal",
  "protein_g"  = excluded."protein_g",
  "carbs_g"    = excluded."carbs_g",
  "fat_g"      = excluded."fat_g",
  "updated_at" = strftime('%s','now')
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
		)
	})

	return g, err
}
