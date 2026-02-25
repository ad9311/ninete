package repo

import (
	"context"
	"fmt"
	"strings"
)

type Task struct {
	ID          int
	ListID      int
	UserID      int
	Description string
	Priority    int
	Done        bool
	CreatedAt   int64
	UpdatedAt   int64
}

type InsertTaskParams struct {
	ListID      int
	UserID      int
	Description string
	Priority    int
}

type UpdateTaskParams struct {
	ID          int
	Description string
	Priority    int
	Done        bool
}

const selectTasks = `SELECT * FROM "tasks"`

func (q *Queries) SelectTasks(ctx context.Context, opts QueryOptions) ([]Task, error) {
	var ts []Task

	subQuery, err := opts.Build()
	if err != nil {
		return ts, err
	}

	if err := opts.Validate(validTaskFields()); err != nil {
		return ts, err
	}

	query := selectTasks + " " + subQuery
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
			var t Task

			if err := rows.Scan(
				&t.ID,
				&t.ListID,
				&t.UserID,
				&t.Description,
				&t.Priority,
				&t.Done,
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

const countTasks = `SELECT COUNT(*) FROM "tasks"`

func (q *Queries) CountTasks(ctx context.Context, filters Filters) (int, error) {
	var c int

	subQuery, err := filters.Build()
	if err != nil {
		return 0, err
	}

	query := countTasks + " " + subQuery
	values := filters.Values()

	err = q.wrapQuery(query, func() error {
		row := q.db.QueryRowContext(ctx, query, values...)

		return row.Scan(&c)
	})

	return c, err
}

const selectTask = `SELECT * FROM "tasks" WHERE "id" = ? AND "user_id" = ? LIMIT 1`

func (q *Queries) SelectTask(ctx context.Context, id, userID int) (Task, error) {
	var t Task

	err := q.wrapQuery(selectTask, func() error {
		row := q.db.QueryRowContext(ctx, selectTask, id, userID)

		return row.Scan(
			&t.ID,
			&t.ListID,
			&t.UserID,
			&t.Description,
			&t.Priority,
			&t.Done,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
	})

	return t, err
}

const insertTask = `
INSERT INTO "tasks" ("list_id", "user_id", "description", "priority")
VALUES (?, ?, ?, ?)
RETURNING *`

func (q *TxQueries) InsertTask(ctx context.Context, params InsertTaskParams) (Task, error) {
	var t Task

	err := q.wrapQuery(insertTask, func() error {
		row := q.tx.QueryRowContext(
			ctx,
			insertTask,
			params.ListID,
			params.UserID,
			params.Description,
			params.Priority,
		)

		return row.Scan(
			&t.ID,
			&t.ListID,
			&t.UserID,
			&t.Description,
			&t.Priority,
			&t.Done,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
	})

	return t, err
}

const updateTask = `
UPDATE "tasks"
SET "description" = ?,
    "priority"    = ?,
    "done"        = ?,
    "updated_at"  = ?
WHERE "id" = ?
  AND "user_id" = ?
RETURNING *;
`

func (q *TxQueries) UpdateTask(ctx context.Context, userID int, params UpdateTaskParams) (Task, error) {
	var t Task

	err := q.wrapQuery(updateTask, func() error {
		row := q.tx.QueryRowContext(
			ctx,
			updateTask,
			params.Description,
			params.Priority,
			params.Done,
			newUpdatedAt(),
			params.ID,
			userID,
		)

		return row.Scan(
			&t.ID,
			&t.ListID,
			&t.UserID,
			&t.Description,
			&t.Priority,
			&t.Done,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
	})

	return t, err
}

const toggleTaskDone = `
UPDATE "tasks"
SET "done"       = NOT "done",
    "updated_at" = ?
WHERE "id" = ?
  AND "user_id" = ?
RETURNING *`

func (q *Queries) ToggleTaskDone(ctx context.Context, id, userID int) (Task, error) {
	var t Task

	err := q.wrapQuery(toggleTaskDone, func() error {
		row := q.db.QueryRowContext(ctx, toggleTaskDone, newUpdatedAt(), id, userID)

		return row.Scan(
			&t.ID,
			&t.ListID,
			&t.UserID,
			&t.Description,
			&t.Priority,
			&t.Done,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
	})

	return t, err
}

const deleteTask = `DELETE FROM "tasks" WHERE "id" = ? AND "user_id" = ? RETURNING "id"`

func (q *Queries) DeleteTask(ctx context.Context, id, userID int) (int, error) {
	var i int

	err := q.wrapQuery(deleteTask, func() error {
		row := q.db.QueryRowContext(ctx, deleteTask, id, userID)

		return row.Scan(&i)
	})

	return i, err
}

const countTasksByListIDsBase = `
SELECT "list_id", COUNT(*)
FROM "tasks"
WHERE "user_id" = ?
  AND "list_id" IN (%s)
GROUP BY "list_id"
`

func (q *Queries) CountTasksByListIDs(ctx context.Context, listIDs []int, userID int) (map[int]int, error) {
	counts := make(map[int]int, len(listIDs))
	if len(listIDs) == 0 {
		return counts, nil
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(listIDs)), ",")
	query := fmt.Sprintf(countTasksByListIDsBase, placeholders)

	values := make([]any, 0, len(listIDs)+1)
	values = append(values, userID)
	for _, id := range listIDs {
		values = append(values, id)
	}

	err := q.wrapQuery(query, func() error {
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
			var listID, count int

			if err := rows.Scan(&listID, &count); err != nil {
				return err
			}

			counts[listID] = count
		}

		return rows.Err()
	})

	return counts, err
}

func validTaskFields() []string {
	return []string{
		"id",
		"list_id",
		"user_id",
		"description",
		"priority",
		"done",
		"created_at",
		"updated_at",
	}
}
