package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
	"github.com/go-chi/chi/v5"
)

type listRow struct {
	ID        int
	Name      string
	TaskCount int
}

// ----------------------------------------------------------------------------- //
// Context Middleware
// ----------------------------------------------------------------------------- //

func (h *Handler) ListContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := getCurrentUser(r)

		id, err := prog.ParseID(chi.URLParam(r, "id"), "List")
		if err != nil {
			h.NotFound(w, r)

			return
		}

		list, err := h.store.FindList(ctx, id, user.ID)
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}
		if err != nil {
			h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

			return
		}

		ctx = context.WithValue(ctx, KeyList, &list)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ----------------------------------------------------------------------------- //
// Handlers
// ----------------------------------------------------------------------------- //

func (h *Handler) GetLists(w http.ResponseWriter, r *http.Request) {
	data := h.tmplData(r)
	user := getCurrentUser(r)

	opts := userScopedQueryOpts(r, user.ID, repo.Sorting{Field: "name", Order: "ASC"})

	totalCount, err := h.store.CountLists(r.Context(), opts.Filters)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ListsIndex, err)

		return
	}

	lists, err := h.store.FindLists(r.Context(), opts)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ListsIndex, err)

		return
	}

	listIDs := make([]int, 0, len(lists))
	for _, l := range lists {
		listIDs = append(listIDs, l.ID)
	}

	taskCountByListID, err := h.countTasksByListIDs(r.Context(), listIDs, user.ID)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ListsIndex, err)

		return
	}

	rows := make([]listRow, 0, len(lists))
	for _, l := range lists {
		rows = append(rows, listRow{
			ID:        l.ID,
			Name:      l.Name,
			TaskCount: taskCountByListID[l.ID],
		})
	}

	data["lists"] = rows
	data["pagination"] = newPaginationData(r, opts, totalCount)
	data["basePath"] = "/lists"

	h.render(w, http.StatusOK, ListsIndex, data)
}

func (h *Handler) GetListsNew(w http.ResponseWriter, r *http.Request) {
	data := h.tmplData(r)
	data["list"] = repo.List{}

	h.render(w, http.StatusOK, ListsNew, data)
}

func (h *Handler) PostLists(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)

	params, err := parseListForm(r)
	if err != nil {
		data["list"] = repo.List{}
		h.renderErr(w, r, http.StatusBadRequest, ListsNew, err)

		return
	}

	_, err = h.store.CreateList(ctx, user.ID, params)
	if err != nil {
		data["list"] = repo.List{Name: params.Name}
		h.renderErr(w, r, http.StatusBadRequest, ListsNew, err)

		return
	}

	http.Redirect(w, r, "/lists", http.StatusSeeOther)
}

func (h *Handler) GetList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)
	list := getList(r)

	opts := repo.QueryOptions{
		Sorting: repo.Sorting{Field: "created_at", Order: "ASC"},
	}
	opts.Filters.Connector = "AND"
	opts.Filters.FilterFields = []repo.FilterField{
		{Name: "list_id", Value: list.ID, Operator: "="},
		{Name: "user_id", Value: user.ID, Operator: "="},
	}

	tasks, err := h.store.FindTasks(ctx, opts)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ListsShow, err)

		return
	}

	taskIDs := make([]int, 0, len(tasks))
	for _, t := range tasks {
		taskIDs = append(taskIDs, t.ID)
	}

	taskTagRows, err := h.store.FindTaskTagRows(ctx, taskIDs, user.ID)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, ListsShow, err)

		return
	}

	taskTagNames := map[int][]string{}
	for _, row := range taskTagRows {
		taskTagNames[row.TaskID] = append(taskTagNames[row.TaskID], row.TagName)
	}

	taskRows := make([]taskRow, 0, len(tasks))
	for _, t := range tasks {
		taskRows = append(taskRows, taskRow{
			ID:            t.ID,
			ListID:        t.ListID,
			Description:   t.Description,
			Priority:      t.Priority,
			PriorityLabel: priorityLabel(t.Priority),
			Done:          t.Done,
			Tags:          taskTagNames[t.ID],
		})
	}

	data["list"] = list
	data["tasks"] = taskRows

	h.render(w, http.StatusOK, ListsShow, data)
}

func (h *Handler) GetListsEdit(w http.ResponseWriter, r *http.Request) {
	data := h.tmplData(r)
	data["list"] = getList(r)

	h.render(w, http.StatusOK, ListsEdit, data)
}

func (h *Handler) PostListsUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)
	list := *getList(r)

	params, err := parseListForm(r)
	if err != nil {
		data["list"] = list
		h.renderErr(w, r, http.StatusBadRequest, ListsEdit, err)

		return
	}

	_, err = h.store.UpdateList(ctx, list.ID, user.ID, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}

		list.Name = params.Name
		data["list"] = list
		h.renderErr(w, r, http.StatusBadRequest, ListsEdit, err)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/lists/%d", list.ID), http.StatusSeeOther)
}

func (h *Handler) PostListsDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)
	list := getList(r)

	_, err := h.store.DeleteList(ctx, list.ID, user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}

		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	http.Redirect(w, r, "/lists", http.StatusSeeOther)
}

// ----------------------------------------------------------------------------- //
// Unexported Functions and Helpers
// ----------------------------------------------------------------------------- //

func parseListForm(r *http.Request) (logic.ListParams, error) {
	var params logic.ListParams

	if err := r.ParseForm(); err != nil {
		return params, fmt.Errorf("failed to parse form: %w", err)
	}

	params.Name = r.FormValue("name")

	return params, nil
}

func getList(r *http.Request) *repo.List {
	list, ok := r.Context().Value(KeyList).(*repo.List)

	if !ok {
		panic("failed to get list context")
	}

	return list
}

func (h *Handler) countTasksByListIDs(
	ctx context.Context,
	listIDs []int,
	userID int,
) (map[int]int, error) {
	counts := make(map[int]int, len(listIDs))
	for _, id := range listIDs {
		counts[id] = 0
	}

	for _, listID := range listIDs {
		filters := repo.Filters{
			Connector: "AND",
			FilterFields: []repo.FilterField{
				{Name: "list_id", Value: listID, Operator: "="},
				{Name: "user_id", Value: userID, Operator: "="},
			},
		}

		count, err := h.store.CountTasks(ctx, filters)
		if err != nil {
			return nil, err
		}

		counts[listID] = count
	}

	return counts, nil
}
