package logic_test

import (
	"database/sql"
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestCreateFood(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "food_user_1",
		Email:        "food_user_1@example.com",
		PasswordHash: []byte("food_user_hash_1"),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_create_food",
			fn: func(t *testing.T) {
				params := newFoodParams("chicken breast", 165, 31, 0, 3.6)
				food, err := s.Store.CreateFood(ctx, user.ID, params)
				require.NoError(t, err)
				require.Positive(t, food.ID)
				require.Equal(t, user.ID, food.UserID)
				require.Equal(t, "chicken breast", food.Name)
				require.Equal(t, 165.0, food.Kcal)
				require.Equal(t, 31.0, food.ProteinG)
				require.Equal(t, 0.0, food.CarbsG)
				require.Equal(t, 3.6, food.FatG)
			},
		},
		{
			name: "should_create_food_with_decimal_values",
			fn: func(t *testing.T) {
				params := newFoodParams("greek yogurt", 133.22, 12.5, 8.75, 3.33)
				food, err := s.Store.CreateFood(ctx, user.ID, params)
				require.NoError(t, err)
				require.Equal(t, 133.22, food.Kcal)
				require.Equal(t, 12.5, food.ProteinG)
				require.Equal(t, 8.75, food.CarbsG)
				require.Equal(t, 3.33, food.FatG)
			},
		},
		{
			name: "should_fail_validation_for_empty_name",
			fn: func(t *testing.T) {
				_, err := s.Store.CreateFood(ctx, user.ID, newFoodParams("", 0, 0, 0, 0))
				require.ErrorIs(t, err, logic.ErrValidationFailed)
			},
		},
		{
			name: "should_fail_for_duplicate_name_same_user",
			fn: func(t *testing.T) {
				_, err := s.Store.CreateFood(ctx, user.ID, newFoodParams("oats", 380, 13, 67, 7))
				require.NoError(t, err)

				_, err = s.Store.CreateFood(ctx, user.ID, newFoodParams("oats", 100, 1, 1, 1))
				require.Error(t, err)
			},
		},
		{
			name: "should_fail_for_duplicate_name_case_insensitive",
			fn: func(t *testing.T) {
				_, err := s.Store.CreateFood(ctx, user.ID, newFoodParams("Banana", 89, 1.1, 23, 0.3))
				require.NoError(t, err)

				_, err = s.Store.CreateFood(ctx, user.ID, newFoodParams("banana", 89, 1.1, 23, 0.3))
				require.Error(t, err)

				_, err = s.Store.CreateFood(ctx, user.ID, newFoodParams("BANANA", 89, 1.1, 23, 0.3))
				require.Error(t, err)
			},
		},
		{
			name: "should_allow_same_name_for_different_users",
			fn: func(t *testing.T) {
				otherUser := s.CreateUser(t, repo.InsertUserParams{
					Username:     "food_user_1b",
					Email:        "food_user_1b@example.com",
					PasswordHash: []byte("food_user_hash_1b"),
				})

				_, err := s.Store.CreateFood(ctx, user.ID, newFoodParams("rice", 130, 2.7, 28, 0.3))
				require.NoError(t, err)

				_, err = s.Store.CreateFood(ctx, otherUser.ID, newFoodParams("rice", 130, 2.7, 28, 0.3))
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestUpdateFood(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "food_user_2",
		Email:        "food_user_2@example.com",
		PasswordHash: []byte("food_user_hash_2"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "food_user_3",
		Email:        "food_user_3@example.com",
		PasswordHash: []byte("food_user_hash_3"),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_update_food",
			fn: func(t *testing.T) {
				food := s.CreateFood(t, user.ID, newFoodParams("oats", 150, 5, 27, 3))
				updParams := newFoodParams("oats updated", 160, 6, 28, 3)
				updated, err := s.Store.UpdateFood(ctx, food.ID, user.ID, updParams)
				require.NoError(t, err)
				require.Equal(t, "oats updated", updated.Name)
				require.Equal(t, 160.0, updated.Kcal)
			},
		},
		{
			name: "should_fail_when_food_does_not_belong_to_user",
			fn: func(t *testing.T) {
				food := s.CreateFood(t, user.ID, newFoodParams("rice", 200, 4, 44, 1))
				updParams := newFoodParams("rice updated", 210, 4, 45, 1)
				_, err := s.Store.UpdateFood(ctx, food.ID, otherUser.ID, updParams)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestDeleteFood(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "food_user_4",
		Email:        "food_user_4@example.com",
		PasswordHash: []byte("food_user_hash_4"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "food_user_5",
		Email:        "food_user_5@example.com",
		PasswordHash: []byte("food_user_hash_5"),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_delete_food_for_owner",
			fn: func(t *testing.T) {
				food := s.CreateFood(t, user.ID, newFoodParams("banana", 89, 1.1, 23, 0.3))
				deletedID, err := s.Store.DeleteFood(ctx, food.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, food.ID, deletedID)
			},
		},
		{
			name: "should_fail_when_deleting_food_of_another_user",
			fn: func(t *testing.T) {
				food := s.CreateFood(t, user.ID, newFoodParams("avocado", 160, 2, 9, 15))
				_, err := s.Store.DeleteFood(ctx, food.ID, otherUser.ID)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestFindFood(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "food_user_6",
		Email:        "food_user_6@example.com",
		PasswordHash: []byte("food_user_hash_6"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "food_user_7",
		Email:        "food_user_7@example.com",
		PasswordHash: []byte("food_user_hash_7"),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_find_food_for_owner",
			fn: func(t *testing.T) {
				created := s.CreateFood(t, user.ID, newFoodParams("eggs", 155, 13, 1, 11))
				found, err := s.Store.FindFood(ctx, created.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, created.ID, found.ID)
				require.Equal(t, "eggs", found.Name)
			},
		},
		{
			name: "should_fail_when_food_belongs_to_another_user",
			fn: func(t *testing.T) {
				created := s.CreateFood(t, user.ID, newFoodParams("milk", 42, 3, 5, 1))
				_, err := s.Store.FindFood(ctx, created.ID, otherUser.ID)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
		{
			name: "should_fail_when_food_does_not_exist",
			fn: func(t *testing.T) {
				_, err := s.Store.FindFood(ctx, 999999, user.ID)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func newFoodParams(name string, kcal, proteinG, carbsG, fatG float64) logic.FoodParams {
	return logic.FoodParams{
		Name:     name,
		Kcal:     kcal,
		ProteinG: proteinG,
		CarbsG:   carbsG,
		FatG:     fatG,
	}
}
