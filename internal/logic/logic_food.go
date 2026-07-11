package logic

import (
	"context"

	"github.com/ad9311/ninete/internal/repo"
)

type FoodParams struct {
	Name          string  `validate:"required,min=1,max=100"`
	Kcal          float64 `validate:"gte=0"`
	ProteinG      float64 `validate:"gte=0"`
	CarbsG        float64 `validate:"gte=0"`
	FatG          float64 `validate:"gte=0"`
	FiberG        float64 `validate:"gte=0"`
	SodiumG       float64 `validate:"gte=0"`
	SaturatedFatG float64 `validate:"gte=0"`
}

func (s *Store) FindFoods(ctx context.Context, opts repo.QueryOptions) ([]repo.Food, error) {
	foods, err := s.queries.SelectFoods(ctx, opts)
	if err != nil {
		return foods, err
	}

	return foods, nil
}

func (s *Store) FindFood(ctx context.Context, id, userID int) (repo.Food, error) {
	food, err := s.queries.SelectFood(ctx, id, userID)
	if err != nil {
		return food, err
	}

	return food, nil
}

func (s *Store) CreateFood(ctx context.Context, userID int, params FoodParams) (repo.Food, error) {
	var food repo.Food

	if err := s.ValidateStruct(params); err != nil {
		return food, err
	}

	err := s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		var txErr error

		food, txErr = tq.InsertFood(ctx, repo.InsertFoodParams{
			UserID:        userID,
			Name:          params.Name,
			Kcal:          params.Kcal,
			ProteinG:      params.ProteinG,
			CarbsG:        params.CarbsG,
			FatG:          params.FatG,
			FiberG:        params.FiberG,
			SodiumG:       params.SodiumG,
			SaturatedFatG: params.SaturatedFatG,
		})

		return txErr
	})
	if err != nil {
		return food, err
	}

	return food, nil
}

func (s *Store) UpdateFood(
	ctx context.Context,
	id, userID int,
	params FoodParams,
) (repo.Food, error) {
	var food repo.Food

	if err := s.ValidateStruct(params); err != nil {
		return food, err
	}

	err := s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		var txErr error

		food, txErr = tq.UpdateFood(ctx, userID, repo.UpdateFoodParams{
			ID:            id,
			Name:          params.Name,
			Kcal:          params.Kcal,
			ProteinG:      params.ProteinG,
			CarbsG:        params.CarbsG,
			FatG:          params.FatG,
			FiberG:        params.FiberG,
			SodiumG:       params.SodiumG,
			SaturatedFatG: params.SaturatedFatG,
		})

		return txErr
	})
	if err != nil {
		return food, err
	}

	return food, nil
}

func (s *Store) DeleteFood(ctx context.Context, id, userID int) (int, error) {
	i, err := s.queries.DeleteFood(ctx, id, userID)
	if err != nil {
		return 0, err
	}

	return i, nil
}

func (s *Store) DeleteAllFoods(ctx context.Context, userID int) error {
	return s.queries.WithTx(ctx, func(tq *repo.TxQueries) error {
		return tq.DeleteAllFoodsByUser(ctx, userID)
	})
}
