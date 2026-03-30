package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/go-chi/chi/v5"
)

// ----------------------------------------------------------------------------- //
// Context Middleware
// ----------------------------------------------------------------------------- //

func (h *Handler) FoodContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := getCurrentUser(r)
		foodID := chi.URLParam(r, "food_id")

		id, err := prog.ParseID(foodID, "Food")
		if err != nil {
			h.NotFound(w, r)

			return
		}

		food, err := h.store.FindFood(ctx, id, user.ID)
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}
		if err != nil {
			h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

			return
		}

		ctx = context.WithValue(ctx, KeyFood, &food)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ----------------------------------------------------------------------------- //
// Handlers
// ----------------------------------------------------------------------------- //

func (h *Handler) GetFoods(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)

	q := r.URL.Query()
	sortField := q.Get("sort_field")
	sortOrder := q.Get("sort_order")
	if sortField == "" {
		sortField = "name"
	}
	if sortOrder == "" {
		sortOrder = "ASC"
	}

	opts := repo.QueryOptions{
		Filters: repo.Filters{
			FilterFields: []repo.FilterField{
				{Name: "user_id", Value: user.ID, Operator: "="},
			},
			Connector: "AND",
		},
		Sorting: repo.Sorting{Field: sortField, Order: sortOrder},
	}

	foods, err := h.store.FindFoods(ctx, opts)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, FoodsIndex, err)

		return
	}

	data["foods"] = foods
	data["sortField"] = sortField
	data["sortOrder"] = sortOrder

	h.render(w, http.StatusOK, FoodsIndex, data)
}

func (h *Handler) GetFoodsNew(w http.ResponseWriter, r *http.Request) {
	data := h.tmplData(r)
	data["food"] = repo.Food{}

	h.render(w, http.StatusOK, FoodsNew, data)
}

func (h *Handler) PostFoods(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)

	params, err := parseFoodForm(r)
	if err != nil {
		data["food"] = repo.Food{}
		h.renderErr(w, r, http.StatusBadRequest, FoodsNew, err)

		return
	}

	food, err := h.store.CreateFood(ctx, user.ID, params)
	if err != nil {
		data["food"] = repo.Food{
			Name:     params.Name,
			Kcal:     params.Kcal,
			ProteinG: params.ProteinG,
			CarbsG:   params.CarbsG,
			FatG:     params.FatG,
		}
		h.renderErr(w, r, http.StatusBadRequest, FoodsNew, err)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/foods/%d", food.ID), http.StatusSeeOther)
}

func (h *Handler) GetFood(w http.ResponseWriter, r *http.Request) {
	data := h.tmplData(r)
	data["food"] = getFood(r)

	h.render(w, http.StatusOK, FoodsShow, data)
}

func (h *Handler) GetFoodEdit(w http.ResponseWriter, r *http.Request) {
	data := h.tmplData(r)
	data["food"] = getFood(r)

	h.render(w, http.StatusOK, FoodsEdit, data)
}

func (h *Handler) PostFoodUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)
	food := *getFood(r)

	params, err := parseFoodForm(r)
	if err != nil {
		data["food"] = food
		h.renderErr(w, r, http.StatusBadRequest, FoodsEdit, err)

		return
	}

	_, err = h.store.UpdateFood(ctx, food.ID, user.ID, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}

		food.Name = params.Name
		food.Kcal = params.Kcal
		food.ProteinG = params.ProteinG
		food.CarbsG = params.CarbsG
		food.FatG = params.FatG
		data["food"] = food
		h.renderErr(w, r, http.StatusBadRequest, FoodsEdit, err)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/foods/%d", food.ID), http.StatusSeeOther)
}

func (h *Handler) PostFoodDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)
	food := getFood(r)

	_, err := h.store.DeleteFood(ctx, food.ID, user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}
		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	http.Redirect(w, r, "/foods", http.StatusSeeOther)
}

// ----------------------------------------------------------------------------- //
// Unexported Functions and Helpers
// ----------------------------------------------------------------------------- //

func parseFoodForm(r *http.Request) (logic.FoodParams, error) {
	var params logic.FoodParams

	if err := r.ParseForm(); err != nil {
		return params, fmt.Errorf("failed to parse form, %w", err)
	}

	kcal, err := strconv.ParseFloat(r.FormValue("kcal"), 64)
	if err != nil {
		return params, fmt.Errorf("kcal: %w", err)
	}

	proteinG, err := strconv.ParseFloat(r.FormValue("protein_g"), 64)
	if err != nil {
		return params, fmt.Errorf("protein_g: %w", err)
	}

	carbsG, err := strconv.ParseFloat(r.FormValue("carbs_g"), 64)
	if err != nil {
		return params, fmt.Errorf("carbs_g: %w", err)
	}

	fatG, err := strconv.ParseFloat(r.FormValue("fat_g"), 64)
	if err != nil {
		return params, fmt.Errorf("fat_g: %w", err)
	}

	params.Name = r.FormValue("name")
	params.Kcal = kcal
	params.ProteinG = proteinG
	params.CarbsG = carbsG
	params.FatG = fatG

	return params, nil
}

func getFood(r *http.Request) *repo.Food {
	food, ok := r.Context().Value(KeyFood).(*repo.Food)

	if !ok {
		panic("failed to get food context")
	}

	return food
}
