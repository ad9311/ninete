package repo

import (
	"context"
)

type MacroTemplate struct {
	ID         int
	UserID     int
	Name       string
	Kcal       float64
	ProteinG   float64
	CarbsG     float64
	FatG       float64
	Amount     float64
	AmountUnit string
	CreatedAt  int64
	UpdatedAt  int64
}

type InsertMacroTemplateParams struct {
	UserID     int
	Name       string
	Kcal       float64
	ProteinG   float64
	CarbsG     float64
	FatG       float64
	Amount     float64
	AmountUnit string
}

type UpdateMacroTemplateParams struct {
	ID         int
	Name       string
	Kcal       float64
	ProteinG   float64
	CarbsG     float64
	FatG       float64
	Amount     float64
	AmountUnit string
}

const selectMacroTemplates = `SELECT * FROM "macro_templates"`

func (q *Queries) SelectMacroTemplates(ctx context.Context, opts QueryOptions) ([]MacroTemplate, error) {
	var ts []MacroTemplate

	if err := opts.Validate(validMacroTemplateFields()); err != nil {
		return ts, err
	}

	subQuery, err := opts.Build()
	if err != nil {
		return ts, err
	}

	query := selectMacroTemplates + " " + subQuery
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
			var t MacroTemplate

			if err := rows.Scan(
				&t.ID,
				&t.UserID,
				&t.Name,
				&t.Kcal,
				&t.ProteinG,
				&t.CarbsG,
				&t.FatG,
				&t.Amount,
				&t.AmountUnit,
				&t.CreatedAt,
				&t.UpdatedAt,
			); err != nil {
				return err
			}

			ts = append(ts, t)
		}

		return rows.Err()
	})

	return ts, err
}

const selectMacroTemplate = `SELECT * FROM "macro_templates" WHERE "id" = ? AND "user_id" = ? LIMIT 1`

func (q *Queries) SelectMacroTemplate(ctx context.Context, id, userID int) (MacroTemplate, error) {
	var t MacroTemplate

	err := q.wrapQuery(selectMacroTemplate, func() error {
		row := q.db.QueryRowContext(ctx, selectMacroTemplate, id, userID)

		return row.Scan(
			&t.ID,
			&t.UserID,
			&t.Name,
			&t.Kcal,
			&t.ProteinG,
			&t.CarbsG,
			&t.FatG,
			&t.Amount,
			&t.AmountUnit,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
	})

	return t, err
}

const insertMacroTemplate = `
INSERT INTO "macro_templates" ("user_id", "name", "kcal", "protein_g", "carbs_g", "fat_g", "amount", "amount_unit")
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *`

func (q *TxQueries) InsertMacroTemplate(ctx context.Context, params InsertMacroTemplateParams) (MacroTemplate, error) {
	var t MacroTemplate

	err := q.wrapQuery(insertMacroTemplate, func() error {
		row := q.tx.QueryRowContext(
			ctx,
			insertMacroTemplate,
			params.UserID,
			params.Name,
			params.Kcal,
			params.ProteinG,
			params.CarbsG,
			params.FatG,
			params.Amount,
			params.AmountUnit,
		)

		return row.Scan(
			&t.ID,
			&t.UserID,
			&t.Name,
			&t.Kcal,
			&t.ProteinG,
			&t.CarbsG,
			&t.FatG,
			&t.Amount,
			&t.AmountUnit,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
	})

	return t, err
}

const updateMacroTemplate = `
UPDATE "macro_templates"
SET "name"        = ?,
    "kcal"        = ?,
    "protein_g"   = ?,
    "carbs_g"     = ?,
    "fat_g"       = ?,
    "amount"      = ?,
    "amount_unit" = ?,
    "updated_at"  = ?
WHERE "id" = ?
  AND "user_id" = ?
RETURNING *`

func (q *TxQueries) UpdateMacroTemplate(
	ctx context.Context,
	userID int,
	params UpdateMacroTemplateParams,
) (MacroTemplate, error) {
	var t MacroTemplate

	err := q.wrapQuery(updateMacroTemplate, func() error {
		row := q.tx.QueryRowContext(
			ctx,
			updateMacroTemplate,
			params.Name,
			params.Kcal,
			params.ProteinG,
			params.CarbsG,
			params.FatG,
			params.Amount,
			params.AmountUnit,
			newUpdatedAt(),
			params.ID,
			userID,
		)

		return row.Scan(
			&t.ID,
			&t.UserID,
			&t.Name,
			&t.Kcal,
			&t.ProteinG,
			&t.CarbsG,
			&t.FatG,
			&t.Amount,
			&t.AmountUnit,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
	})

	return t, err
}

const deleteMacroTemplate = `DELETE FROM "macro_templates" WHERE "id" = ? AND "user_id" = ? RETURNING "id"`

func (q *Queries) DeleteMacroTemplate(ctx context.Context, id, userID int) (int, error) {
	var i int

	err := q.wrapQuery(deleteMacroTemplate, func() error {
		row := q.db.QueryRowContext(ctx, deleteMacroTemplate, id, userID)

		return row.Scan(&i)
	})

	return i, err
}

func validMacroTemplateFields() []string {
	return []string{
		"id",
		"user_id",
		"name",
		"kcal",
		"protein_g",
		"carbs_g",
		"fat_g",
		"amount",
		"amount_unit",
		"created_at",
		"updated_at",
	}
}
