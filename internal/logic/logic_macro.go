package logic

import (
	"context"

	"github.com/ad9311/ninete/internal/repo"
)

type MacroEntryParams struct {
	Name     string  `validate:"required,min=1,max=100"`
	Kcal     float64 `validate:"gte=0"`
	ProteinG float64 `validate:"gte=0"`
	CarbsG   float64 `validate:"gte=0"`
	FatG     float64 `validate:"gte=0"`
	Date     int64   `validate:"required,gt=0"`
	MealType string  `validate:"required,oneof=breakfast lunch dinner snack other"`
}

type MacroGoalParams struct {
	Kcal     float64 `validate:"gt=0"`
	ProteinG float64 `validate:"gt=0"`
	CarbsG   float64 `validate:"gt=0"`
	FatG     float64 `validate:"gt=0"`
}

func (s *Store) FindMacroEntries(ctx context.Context, opts repo.QueryOptions) ([]repo.MacroEntry, error) {
	entries, err := s.queries.SelectMacroEntries(ctx, opts)
	if err != nil {
		return entries, err
	}

	return entries, nil
}

func (s *Store) CountMacroEntries(ctx context.Context, filters repo.Filters) (int, error) {
	count, err := s.queries.CountMacroEntries(ctx, filters)
	if err != nil {
		return count, err
	}

	return count, nil
}

func (s *Store) FindMacroEntry(ctx context.Context, id, userID int) (repo.MacroEntry, error) {
	entry, err := s.queries.SelectMacroEntry(ctx, id, userID)
	if err != nil {
		return entry, err
	}

	return entry, nil
}

func (s *Store) CreateMacroEntry(ctx context.Context, userID int, params MacroEntryParams) (repo.MacroEntry, error) {
	var entry repo.MacroEntry

	if err := s.ValidateStruct(params); err != nil {
		return entry, err
	}

	err := s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		var txErr error

		entry, txErr = tq.InsertMacroEntry(ctx, repo.InsertMacroEntryParams{
			UserID:   userID,
			Name:     params.Name,
			Kcal:     params.Kcal,
			ProteinG: params.ProteinG,
			CarbsG:   params.CarbsG,
			FatG:     params.FatG,
			Date:     params.Date,
			MealType: params.MealType,
		})

		return txErr
	})
	if err != nil {
		return entry, err
	}

	return entry, nil
}

func (s *Store) UpdateMacroEntry(
	ctx context.Context,
	id, userID int,
	params MacroEntryParams,
) (repo.MacroEntry, error) {
	var entry repo.MacroEntry

	if err := s.ValidateStruct(params); err != nil {
		return entry, err
	}

	err := s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		var txErr error

		entry, txErr = tq.UpdateMacroEntry(ctx, userID, repo.UpdateMacroEntryParams{
			ID:       id,
			Name:     params.Name,
			Kcal:     params.Kcal,
			ProteinG: params.ProteinG,
			CarbsG:   params.CarbsG,
			FatG:     params.FatG,
			Date:     params.Date,
			MealType: params.MealType,
		})

		return txErr
	})
	if err != nil {
		return entry, err
	}

	return entry, nil
}

func (s *Store) DeleteMacroEntry(ctx context.Context, id, userID int) (int, error) {
	i, err := s.queries.DeleteMacroEntry(ctx, id, userID)
	if err != nil {
		return 0, err
	}

	return i, nil
}

func (s *Store) FindMacroDayTotals(
	ctx context.Context,
	userID int,
	dayStart, nextDayStart int64,
) (repo.MacroDayTotals, error) {
	return s.queries.SelectMacroDayTotals(ctx, userID, dayStart, nextDayStart)
}

func (s *Store) FindMacroGoal(ctx context.Context, userID int) (repo.MacroGoal, error) {
	return s.queries.SelectMacroGoal(ctx, userID)
}

func (s *Store) SaveMacroGoal(ctx context.Context, userID int, params MacroGoalParams) (repo.MacroGoal, error) {
	var goal repo.MacroGoal

	if err := s.ValidateStruct(params); err != nil {
		return goal, err
	}

	err := s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		var txErr error

		goal, txErr = tq.UpsertMacroGoal(ctx, repo.UpsertMacroGoalParams{
			UserID:   userID,
			Kcal:     params.Kcal,
			ProteinG: params.ProteinG,
			CarbsG:   params.CarbsG,
			FatG:     params.FatG,
		})

		return txErr
	})
	if err != nil {
		return goal, err
	}

	return goal, nil
}
