package repo

import (
	"context"
)

type List struct {
	ID        int
	UserID    int
	Name      string
	CreatedAt int64
	UpdatedAt int64
}

type InsertListParams struct {
	UserID int
	Name   string
}

type UpdateListParams struct {
	ID   int
	Name string
}

const selectLists = `SELECT * FROM "lists"`

func (q *Queries) SelectLists(ctx context.Context, opts QueryOptions) ([]List, error) {
	var ls []List

	subQuery, err := opts.Build()
	if err != nil {
		return ls, err
	}

	if err := opts.Validate(validListFields()); err != nil {
		return ls, err
	}

	query := selectLists + " " + subQuery
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
			var l List

			if err := rows.Scan(
				&l.ID,
				&l.UserID,
				&l.Name,
				&l.CreatedAt,
				&l.UpdatedAt,
			); err != nil {
				return err
			}

			ls = append(ls, l)
		}

		return rows.Err()
	})

	return ls, err
}

const countLists = `SELECT COUNT(*) FROM "lists"`

func (q *Queries) CountLists(ctx context.Context, filters Filters) (int, error) {
	var c int

	subQuery, err := filters.Build()
	if err != nil {
		return 0, err
	}

	query := countLists + " " + subQuery
	values := filters.Values()

	err = q.wrapQuery(query, func() error {
		row := q.db.QueryRowContext(ctx, query, values...)

		return row.Scan(&c)
	})

	return c, err
}

const selectList = `SELECT * FROM "lists" WHERE "id" = ? AND "user_id" = ? LIMIT 1`

func (q *Queries) SelectList(ctx context.Context, id, userID int) (List, error) {
	var l List

	err := q.wrapQuery(selectList, func() error {
		row := q.db.QueryRowContext(ctx, selectList, id, userID)

		return row.Scan(
			&l.ID,
			&l.UserID,
			&l.Name,
			&l.CreatedAt,
			&l.UpdatedAt,
		)
	})

	return l, err
}

const insertList = `
INSERT INTO "lists" ("user_id", "name")
VALUES (?, ?)
RETURNING *`

func (q *Queries) InsertList(ctx context.Context, params InsertListParams) (List, error) {
	var l List

	err := q.wrapQuery(insertList, func() error {
		row := q.db.QueryRowContext(
			ctx,
			insertList,
			params.UserID,
			params.Name,
		)

		return row.Scan(
			&l.ID,
			&l.UserID,
			&l.Name,
			&l.CreatedAt,
			&l.UpdatedAt,
		)
	})

	return l, err
}

const updateList = `
UPDATE "lists"
SET "name"       = ?,
    "updated_at" = ?
WHERE "id" = ?
  AND "user_id" = ?
RETURNING *;
`

func (q *Queries) UpdateList(ctx context.Context, userID int, params UpdateListParams) (List, error) {
	var l List

	err := q.wrapQuery(updateList, func() error {
		row := q.db.QueryRowContext(
			ctx,
			updateList,
			params.Name,
			newUpdatedAt(),
			params.ID,
			userID,
		)

		return row.Scan(
			&l.ID,
			&l.UserID,
			&l.Name,
			&l.CreatedAt,
			&l.UpdatedAt,
		)
	})

	return l, err
}

const deleteList = `DELETE FROM "lists" WHERE "id" = ? AND "user_id" = ? RETURNING "id"`

func (q *Queries) DeleteList(ctx context.Context, id, userID int) (int, error) {
	var i int

	err := q.wrapQuery(deleteList, func() error {
		row := q.db.QueryRowContext(ctx, deleteList, id, userID)

		return row.Scan(&i)
	})

	return i, err
}

func validListFields() []string {
	return []string{
		"id",
		"user_id",
		"name",
		"created_at",
		"updated_at",
	}
}
