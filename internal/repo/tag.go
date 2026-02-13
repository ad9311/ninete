package repo

import "context"

type Tag struct {
	ID        int
	UserID    int
	Name      string
	CreatedAt int64
	UpdatedAt int64
}

type InsertTagParams struct {
	UserID int
	Name   string
}

const selectTags = `SELECT * FROM "tags"`

func (q *Queries) SelectTags(ctx context.Context, opts QueryOptions) ([]Tag, error) {
	var ts []Tag

	subQuery, err := opts.Build()
	if err != nil {
		return ts, err
	}

	if err := opts.Validate(validTagFields()); err != nil {
		return ts, err
	}

	query := selectTags + " " + subQuery
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
			var t Tag

			if err := rows.Scan(
				&t.ID,
				&t.UserID,
				&t.Name,
				&t.CreatedAt,
				&t.UpdatedAt,
			); err != nil {
				return err
			}

			ts = append(ts, t)
		}

		return err
	})

	return ts, err
}

const insertTag = `
INSERT INTO "tags" ("user_id", "name")
VALUES (?, ?)
RETURNING *`

func (q *Queries) InsertTag(ctx context.Context, params InsertTagParams) (Tag, error) {
	var t Tag

	err := q.wrapQuery(insertTag, func() error {
		row := q.db.QueryRowContext(ctx, insertTag, params.UserID, params.Name)

		return row.Scan(
			&t.ID,
			&t.UserID,
			&t.Name,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
	})

	return t, err
}

const deleteTag = `DELETE FROM "tags" WHERE "id" = ? AND "user_id" = ? RETURNING "id"`

func (q *Queries) DeleteTag(ctx context.Context, id, userID int) (int, error) {
	var i int

	err := q.wrapQuery(deleteTag, func() error {
		row := q.db.QueryRowContext(ctx, deleteTag, id, userID)

		return row.Scan(&i)
	})

	return i, err
}

func validTagFields() []string {
	return []string{
		"id",
		"user_id",
		"name",
		"created_at",
		"updated_at",
	}
}
