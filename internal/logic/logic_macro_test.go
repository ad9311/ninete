package logic_test

import (
	"database/sql"
	"testing"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/ad9311/ninete/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestCreateMacroEntry(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "macro_user_1",
		Email:        "macro_user_1@example.com",
		PasswordHash: []byte("macro_user_hash_1"),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_create_macro_entry",
			fn: func(t *testing.T) {
				params := newMacroEntryParams("chicken breast", 165, 31, 0, 4, 1742083200)
				entry, err := s.Store.CreateMacroEntry(ctx, user.ID, params)
				require.NoError(t, err)
				require.Positive(t, entry.ID)
				require.Equal(t, user.ID, entry.UserID)
				require.Equal(t, "chicken breast", entry.Name)
				require.Equal(t, 165.0, entry.Kcal)
				require.Equal(t, 31.0, entry.ProteinG)
			},
		},
		{
			name: "should_create_macro_entry_with_decimal_values",
			fn: func(t *testing.T) {
				params := newMacroEntryParams("greek yogurt", 133.22, 12.5, 8.75, 3.33, 1742083200)
				entry, err := s.Store.CreateMacroEntry(ctx, user.ID, params)
				require.NoError(t, err)
				require.Equal(t, 133.22, entry.Kcal)
				require.Equal(t, 12.5, entry.ProteinG)
				require.Equal(t, 8.75, entry.CarbsG)
				require.Equal(t, 3.33, entry.FatG)
			},
		},
		{
			name: "should_fail_validation_for_invalid_params",
			fn: func(t *testing.T) {
				_, err := s.Store.CreateMacroEntry(ctx, user.ID, newMacroEntryParams("", 0, 0, 0, 0, 0))
				require.ErrorIs(t, err, logic.ErrValidationFailed)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestUpdateMacroEntry(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "macro_user_2",
		Email:        "macro_user_2@example.com",
		PasswordHash: []byte("macro_user_hash_2"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "macro_user_3",
		Email:        "macro_user_3@example.com",
		PasswordHash: []byte("macro_user_hash_3"),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_update_macro_entry",
			fn: func(t *testing.T) {
				entry := s.CreateMacroEntry(t, user.ID, newMacroEntryParams("oats", 150, 5, 27, 3, 1742083200))
				updParams := newMacroEntryParams("oats updated", 160, 6, 28, 3, 1742083200)
				updated, err := s.Store.UpdateMacroEntry(ctx, entry.ID, user.ID, updParams)
				require.NoError(t, err)
				require.Equal(t, "oats updated", updated.Name)
				require.Equal(t, 160.0, updated.Kcal)
			},
		},
		{
			name: "should_fail_when_entry_does_not_belong_to_user",
			fn: func(t *testing.T) {
				entry := s.CreateMacroEntry(t, user.ID, newMacroEntryParams("rice", 200, 4, 44, 1, 1742083200))
				updParams := newMacroEntryParams("rice updated", 210, 4, 45, 1, 1742083200)
				_, err := s.Store.UpdateMacroEntry(ctx, entry.ID, otherUser.ID, updParams)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestDeleteMacroEntry(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "macro_user_4",
		Email:        "macro_user_4@example.com",
		PasswordHash: []byte("macro_user_hash_4"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "macro_user_5",
		Email:        "macro_user_5@example.com",
		PasswordHash: []byte("macro_user_hash_5"),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_delete_macro_entry_for_owner",
			fn: func(t *testing.T) {
				entry := s.CreateMacroEntry(t, user.ID, newMacroEntryParams("banana", 89, 1, 23, 0, 1742083200))
				deletedID, err := s.Store.DeleteMacroEntry(ctx, entry.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, entry.ID, deletedID)
			},
		},
		{
			name: "should_fail_when_deleting_entry_of_another_user",
			fn: func(t *testing.T) {
				entry := s.CreateMacroEntry(t, user.ID, newMacroEntryParams("apple", 52, 0, 14, 0, 1742083200))
				_, err := s.Store.DeleteMacroEntry(ctx, entry.ID, otherUser.ID)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestFindMacroDayTotals(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "macro_user_6",
		Email:        "macro_user_6@example.com",
		PasswordHash: []byte("macro_user_hash_6"),
	})

	// 2026-03-16 00:00:00 UTC
	const dayStart int64 = 1742083200
	const nextDayStart int64 = dayStart + 86400

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_return_correct_sums_for_day_window",
			fn: func(t *testing.T) {
				s.CreateMacroEntry(t, user.ID, newMacroEntryParams("food_a", 100, 10, 20, 5, dayStart))
				s.CreateMacroEntry(t, user.ID, newMacroEntryParams("food_b", 200, 20, 30, 10, dayStart+3600))

				totals, err := s.Store.FindMacroDayTotals(ctx, user.ID, dayStart, nextDayStart, "")
				require.NoError(t, err)
				require.Equal(t, 300.0, totals.Kcal)
				require.Equal(t, 30.0, totals.ProteinG)
				require.Equal(t, 50.0, totals.CarbsG)
				require.Equal(t, 15.0, totals.FatG)
			},
		},
		{
			name: "should_return_zeros_for_empty_window",
			fn: func(t *testing.T) {
				totals, err := s.Store.FindMacroDayTotals(ctx, user.ID, nextDayStart, nextDayStart+86400, "")
				require.NoError(t, err)
				require.Zero(t, totals.Kcal)
				require.Zero(t, totals.ProteinG)
				require.Zero(t, totals.CarbsG)
				require.Zero(t, totals.FatG)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestSaveMacroGoal(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "macro_user_7",
		Email:        "macro_user_7@example.com",
		PasswordHash: []byte("macro_user_hash_7"),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_create_goal_on_first_call_and_upsert_on_second",
			fn: func(t *testing.T) {
				goal, err := s.Store.SaveMacroGoal(ctx, user.ID,
					logic.MacroGoalParams{Kcal: 2000, ProteinG: 150, CarbsG: 200, FatG: 70})
				require.NoError(t, err)
				require.Positive(t, goal.ID)
				require.Equal(t, 2000.0, goal.Kcal)

				updated, err := s.Store.SaveMacroGoal(ctx, user.ID,
					logic.MacroGoalParams{Kcal: 2200, ProteinG: 160, CarbsG: 220, FatG: 80})
				require.NoError(t, err)
				require.Equal(t, goal.ID, updated.ID)
				require.Equal(t, 2200.0, updated.Kcal)
				require.Equal(t, 160.0, updated.ProteinG)
			},
		},
		{
			name: "should_save_goal_with_decimal_values",
			fn: func(t *testing.T) {
				goal, err := s.Store.SaveMacroGoal(ctx, user.ID,
					logic.MacroGoalParams{Kcal: 1950.5, ProteinG: 147.25, CarbsG: 195.75, FatG: 65.5})
				require.NoError(t, err)
				require.Equal(t, 1950.5, goal.Kcal)
				require.Equal(t, 147.25, goal.ProteinG)
			},
		},
		{
			name: "should_fail_validation_when_any_value_is_zero",
			fn: func(t *testing.T) {
				_, err := s.Store.SaveMacroGoal(ctx, user.ID,
					logic.MacroGoalParams{Kcal: 0, ProteinG: 150, CarbsG: 200, FatG: 70})
				require.ErrorIs(t, err, logic.ErrValidationFailed)
			},
		},
		{
			name: "should_fail_validation_when_any_value_is_negative",
			fn: func(t *testing.T) {
				_, err := s.Store.SaveMacroGoal(ctx, user.ID,
					logic.MacroGoalParams{Kcal: 2000, ProteinG: -1, CarbsG: 200, FatG: 70})
				require.ErrorIs(t, err, logic.ErrValidationFailed)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestFindMacroGoal(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "macro_user_8",
		Email:        "macro_user_8@example.com",
		PasswordHash: []byte("macro_user_hash_8"),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_return_sql_err_no_rows_when_not_set",
			fn: func(t *testing.T) {
				_, err := s.Store.FindMacroGoal(ctx, user.ID)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
		{
			name: "should_return_goal_when_set",
			fn: func(t *testing.T) {
				s.SaveMacroGoal(t, user.ID, logic.MacroGoalParams{Kcal: 1800, ProteinG: 130, CarbsG: 180, FatG: 60})
				goal, err := s.Store.FindMacroGoal(ctx, user.ID)
				require.NoError(t, err)
				require.Equal(t, 1800.0, goal.Kcal)
				require.Equal(t, user.ID, goal.UserID)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestCreateMacroTemplate(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "macro_tmpl_user_1",
		Email:        "macro_tmpl_user_1@example.com",
		PasswordHash: []byte("macro_tmpl_hash_1"),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_create_macro_template",
			fn: func(t *testing.T) {
				params := newMacroTemplateParams("chicken breast", 165, 31, 0, 4, 100, "g")
				tmpl, err := s.Store.CreateMacroTemplate(ctx, user.ID, params)
				require.NoError(t, err)
				require.Positive(t, tmpl.ID)
				require.Equal(t, user.ID, tmpl.UserID)
				require.Equal(t, "chicken breast", tmpl.Name)
				require.Equal(t, 165.0, tmpl.Kcal)
				require.Equal(t, 31.0, tmpl.ProteinG)
				require.Equal(t, 100.0, tmpl.Amount)
				require.Equal(t, "g", tmpl.AmountUnit)
			},
		},
		{
			name: "should_create_macro_template_with_decimal_values",
			fn: func(t *testing.T) {
				params := newMacroTemplateParams("greek yogurt", 133.22, 12.5, 8.75, 3.33, 170.5, "g")
				tmpl, err := s.Store.CreateMacroTemplate(ctx, user.ID, params)
				require.NoError(t, err)
				require.Equal(t, 133.22, tmpl.Kcal)
				require.Equal(t, 12.5, tmpl.ProteinG)
				require.Equal(t, 8.75, tmpl.CarbsG)
				require.Equal(t, 3.33, tmpl.FatG)
				require.Equal(t, 170.5, tmpl.Amount)
			},
		},
		{
			name: "should_fail_validation_for_empty_name",
			fn: func(t *testing.T) {
				_, err := s.Store.CreateMacroTemplate(ctx, user.ID, newMacroTemplateParams("", 0, 0, 0, 0, 100, "g"))
				require.ErrorIs(t, err, logic.ErrValidationFailed)
			},
		},
		{
			name: "should_create_macro_template_with_ml_unit",
			fn: func(t *testing.T) {
				params := newMacroTemplateParams("olive oil", 120, 0, 0, 14, 15, "ml")
				tmpl, err := s.Store.CreateMacroTemplate(ctx, user.ID, params)
				require.NoError(t, err)
				require.Equal(t, "ml", tmpl.AmountUnit)
				require.Equal(t, 15.0, tmpl.Amount)
			},
		},
		{
			name: "should_fail_validation_for_zero_amount",
			fn: func(t *testing.T) {
				_, err := s.Store.CreateMacroTemplate(ctx, user.ID, newMacroTemplateParams("food", 100, 10, 20, 5, 0, "g"))
				require.ErrorIs(t, err, logic.ErrValidationFailed)
			},
		},
		{
			name: "should_fail_validation_for_invalid_unit",
			fn: func(t *testing.T) {
				_, err := s.Store.CreateMacroTemplate(ctx, user.ID, newMacroTemplateParams("food", 100, 10, 20, 5, 100, "kg"))
				require.ErrorIs(t, err, logic.ErrValidationFailed)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestUpdateMacroTemplate(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "macro_tmpl_user_2",
		Email:        "macro_tmpl_user_2@example.com",
		PasswordHash: []byte("macro_tmpl_hash_2"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "macro_tmpl_user_3",
		Email:        "macro_tmpl_user_3@example.com",
		PasswordHash: []byte("macro_tmpl_hash_3"),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_update_macro_template",
			fn: func(t *testing.T) {
				tmpl := s.CreateMacroTemplate(t, user.ID, newMacroTemplateParams("oats", 150, 5, 27, 3, 80, "g"))
				updParams := newMacroTemplateParams("oats updated", 160, 6, 28, 3, 85, "ml")
				updated, err := s.Store.UpdateMacroTemplate(ctx, tmpl.ID, user.ID, updParams)
				require.NoError(t, err)
				require.Equal(t, "oats updated", updated.Name)
				require.Equal(t, 160.0, updated.Kcal)
				require.Equal(t, 85.0, updated.Amount)
				require.Equal(t, "ml", updated.AmountUnit)
			},
		},
		{
			name: "should_fail_when_template_does_not_belong_to_user",
			fn: func(t *testing.T) {
				tmpl := s.CreateMacroTemplate(t, user.ID, newMacroTemplateParams("rice", 200, 4, 44, 1, 150, "g"))
				updParams := newMacroTemplateParams("rice updated", 210, 4, 45, 1, 155, "g")
				_, err := s.Store.UpdateMacroTemplate(ctx, tmpl.ID, otherUser.ID, updParams)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestDeleteMacroTemplate(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "macro_tmpl_user_4",
		Email:        "macro_tmpl_user_4@example.com",
		PasswordHash: []byte("macro_tmpl_hash_4"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "macro_tmpl_user_5",
		Email:        "macro_tmpl_user_5@example.com",
		PasswordHash: []byte("macro_tmpl_hash_5"),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_delete_macro_template_for_owner",
			fn: func(t *testing.T) {
				tmpl := s.CreateMacroTemplate(t, user.ID, newMacroTemplateParams("banana shake", 250, 8, 40, 5, 300, "g"))
				deletedID, err := s.Store.DeleteMacroTemplate(ctx, tmpl.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, tmpl.ID, deletedID)
			},
		},
		{
			name: "should_fail_when_deleting_template_of_another_user",
			fn: func(t *testing.T) {
				tmpl := s.CreateMacroTemplate(t, user.ID, newMacroTemplateParams("protein bar", 200, 20, 25, 8, 60, "g"))
				_, err := s.Store.DeleteMacroTemplate(ctx, tmpl.ID, otherUser.ID)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestFindMacroTemplate(t *testing.T) {
	s := spec.New(t)
	ctx := t.Context()
	user := s.CreateUser(t, repo.InsertUserParams{
		Username:     "macro_tmpl_user_6",
		Email:        "macro_tmpl_user_6@example.com",
		PasswordHash: []byte("macro_tmpl_hash_6"),
	})
	otherUser := s.CreateUser(t, repo.InsertUserParams{
		Username:     "macro_tmpl_user_7",
		Email:        "macro_tmpl_user_7@example.com",
		PasswordHash: []byte("macro_tmpl_hash_7"),
	})

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_find_template_for_owner",
			fn: func(t *testing.T) {
				created := s.CreateMacroTemplate(t, user.ID, newMacroTemplateParams("eggs", 155, 13, 1, 11, 100, "g"))
				found, err := s.Store.FindMacroTemplate(ctx, created.ID, user.ID)
				require.NoError(t, err)
				require.Equal(t, created.ID, found.ID)
				require.Equal(t, "eggs", found.Name)
			},
		},
		{
			name: "should_fail_when_template_belongs_to_another_user",
			fn: func(t *testing.T) {
				created := s.CreateMacroTemplate(t, user.ID, newMacroTemplateParams("milk", 42, 3, 5, 1, 100, "g"))
				_, err := s.Store.FindMacroTemplate(ctx, created.ID, otherUser.ID)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
		{
			name: "should_fail_when_template_does_not_exist",
			fn: func(t *testing.T) {
				_, err := s.Store.FindMacroTemplate(ctx, 999999, user.ID)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func newMacroEntryParams(name string, kcal, proteinG, carbsG, fatG float64, date int64) logic.MacroEntryParams {
	return logic.MacroEntryParams{
		Name:     name,
		Kcal:     kcal,
		ProteinG: proteinG,
		CarbsG:   carbsG,
		FatG:     fatG,
		Date:     date,
		MealType: "other",
	}
}

func newMacroTemplateParams(
	name string,
	kcal, proteinG, carbsG, fatG, amount float64,
	amountUnit string,
) logic.MacroTemplateParams {
	return logic.MacroTemplateParams{
		Name:       name,
		Kcal:       kcal,
		ProteinG:   proteinG,
		CarbsG:     carbsG,
		FatG:       fatG,
		Amount:     amount,
		AmountUnit: amountUnit,
	}
}
