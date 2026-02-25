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

type taskRow struct {
	ID            int
	ListID        int
	Description   string
	Priority      int
	PriorityLabel string
	Done          bool
	CreatedAt     int64
	Tags          []string
}

// ----------------------------------------------------------------------------- //
// Context Middleware
// ----------------------------------------------------------------------------- //

func (h *Handler) TaskContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := getCurrentUser(r)

		id, err := prog.ParseID(chi.URLParam(r, "task_id"), "Task")
		if err != nil {
			h.NotFound(w, r)

			return
		}

		task, err := h.store.FindTask(ctx, id, user.ID)
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}
		if err != nil {
			h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

			return
		}

		ctx = context.WithValue(ctx, KeyTask, &task)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ----------------------------------------------------------------------------- //
// Handlers
// ----------------------------------------------------------------------------- //

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)
	list := getList(r)
	task := getTask(r)

	taskTags, err := h.store.FindTaskTags(ctx, task.ID, user.ID)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, TasksShow, err)

		return
	}

	row := taskRow{
		ID:            task.ID,
		ListID:        task.ListID,
		Description:   task.Description,
		Priority:      task.Priority,
		PriorityLabel: priorityLabel(task.Priority),
		Done:          task.Done,
		CreatedAt:     task.CreatedAt,
		Tags:          logic.ExtractTagNames(taskTags),
	}

	data["list"] = list
	data["task"] = row

	h.render(w, http.StatusOK, TasksShow, data)
}

func (h *Handler) GetTasksNew(w http.ResponseWriter, r *http.Request) {
	data := h.tmplData(r)
	list := getList(r)

	data["list"] = list
	data["task"] = repo.Task{}
	data["tagsInput"] = ""

	h.render(w, http.StatusOK, TasksNew, data)
}

func (h *Handler) PostTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)
	list := getList(r)
	rawTagsInput := r.FormValue("tags")

	params, err := parseTaskForm(r)
	if err != nil {
		data["list"] = list
		data["task"] = repo.Task{}
		data["tagsInput"] = rawTagsInput
		h.renderErr(w, r, http.StatusBadRequest, TasksNew, err)

		return
	}

	_, err = h.store.CreateTask(ctx, list.ID, user.ID, params)
	if err != nil {
		data["list"] = list
		data["task"] = repo.Task{Description: params.Description, Priority: params.Priority}
		data["tagsInput"] = logic.JoinTagNames(params.Tags)
		h.renderErr(w, r, http.StatusBadRequest, TasksNew, err)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/lists/%d", list.ID), http.StatusSeeOther)
}

func (h *Handler) GetTasksEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)
	list := getList(r)
	task := getTask(r)

	taskTags, err := h.store.FindTaskTags(ctx, task.ID, user.ID)
	if err != nil {
		h.renderErr(w, r, http.StatusInternalServerError, TasksEdit, err)

		return
	}

	data["list"] = list
	data["task"] = task
	data["tagsInput"] = logic.JoinTagNames(logic.ExtractTagNames(taskTags))

	h.render(w, http.StatusOK, TasksEdit, data)
}

func (h *Handler) PostTasksUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := h.tmplData(r)
	user := getCurrentUser(r)
	list := getList(r)
	task := *getTask(r)
	rawTagsInput := r.FormValue("tags")

	params, err := parseTaskForm(r)
	if err != nil {
		data["list"] = list
		data["task"] = task
		data["tagsInput"] = rawTagsInput
		h.renderErr(w, r, http.StatusBadRequest, TasksEdit, err)

		return
	}

	_, err = h.store.UpdateTask(ctx, task.ID, user.ID, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}

		task.Description = params.Description
		task.Priority = params.Priority
		task.Done = params.Done
		data["list"] = list
		data["task"] = task
		data["tagsInput"] = logic.JoinTagNames(params.Tags)
		h.renderErr(w, r, http.StatusBadRequest, TasksEdit, err)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/lists/%d", list.ID), http.StatusSeeOther)
}

func (h *Handler) PostTasksDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)
	list := getList(r)
	task := getTask(r)

	_, err := h.store.DeleteTask(ctx, task.ID, user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}

		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/lists/%d", list.ID), http.StatusSeeOther)
}

func (h *Handler) PostTasksDone(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := getCurrentUser(r)
	list := getList(r)
	task := getTask(r)

	_, err := h.store.ToggleTaskDone(ctx, task.ID, user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.NotFound(w, r)

			return
		}

		h.renderErr(w, r, http.StatusInternalServerError, ErrorIndex, err)

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/lists/%d", list.ID), http.StatusSeeOther)
}

// ----------------------------------------------------------------------------- //
// Unexported Functions and Helpers
// ----------------------------------------------------------------------------- //

func parseTaskForm(r *http.Request) (logic.TaskParams, error) {
	var params logic.TaskParams

	if err := r.ParseForm(); err != nil {
		return params, fmt.Errorf("failed to parse form: %w", err)
	}

	priority, err := strconv.Atoi(r.FormValue("priority"))
	if err != nil {
		return params, fmt.Errorf("invalid priority value: %w", err)
	}

	params.Description = r.FormValue("description")
	params.Priority = priority
	params.Done = r.FormValue("done") == "true"
	params.Tags = logic.ParseTagNames(r.FormValue("tags"))

	return params, nil
}

func priorityLabel(priority int) string {
	switch priority {
	case 1:
		return "Low"
	case 2:
		return "Medium"
	case 3:
		return "High"
	default:
		return "Low"
	}
}

func getTask(r *http.Request) *repo.Task {
	task, ok := r.Context().Value(KeyTask).(*repo.Task)

	if !ok {
		panic("failed to get task context")
	}

	return task
}
