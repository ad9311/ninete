package repo

import (
	"context"
	"fmt"
	"strings"
)

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

		ts, err = scanTagRows(rows)
		if err != nil {
			return err
		}

		return nil
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

const insertOrIgnoreTag = `
INSERT OR IGNORE INTO "tags" ("user_id", "name")
VALUES (?, ?)`

func (q *TxQueries) InsertOrIgnoreTag(ctx context.Context, params InsertTagParams) error {
	return q.wrapQuery(insertOrIgnoreTag, func() error {
		_, err := q.tx.ExecContext(ctx, insertOrIgnoreTag, params.UserID, params.Name)

		return err
	})
}

const selectTagsByUserAndNames = `
SELECT * FROM "tags"
WHERE "user_id" = ?
  AND "name" IN (%s)
ORDER BY "name" ASC`

func (q *Queries) SelectTagsByUserAndNames(ctx context.Context, userID int, names []string) ([]Tag, error) {
	var ts []Tag
	if len(names) == 0 {
		return ts, nil
	}

	query, values := selectTagsByUserAndNamesQuery(userID, names)

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

		ts, err = scanTagRows(rows)
		if err != nil {
			return err
		}

		return nil
	})

	return ts, err
}

func (q *TxQueries) SelectTagsByUserAndNames(ctx context.Context, userID int, names []string) ([]Tag, error) {
	var ts []Tag
	if len(names) == 0 {
		return ts, nil
	}

	query, values := selectTagsByUserAndNamesQuery(userID, names)

	err := q.wrapQuery(query, func() error {
		rows, err := q.tx.QueryContext(ctx, query, values...)
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				q.app.Logger.Error(closeErr)
			}
		}()

		ts, err = scanTagRows(rows)
		if err != nil {
			return err
		}

		return nil
	})

	return ts, err
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

type tagRows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

func scanTagRows(rows tagRows) ([]Tag, error) {
	var ts []Tag

	for rows.Next() {
		var t Tag

		if err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.Name,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return ts, err
		}

		ts = append(ts, t)
	}

	return ts, rows.Err()
}

func selectTagsByUserAndNamesQuery(userID int, names []string) (string, []any) {
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(names)), ",")
	query := fmt.Sprintf(selectTagsByUserAndNames, placeholders)

	values := make([]any, 0, len(names)+1)
	values = append(values, userID)
	for _, name := range names {
		values = append(values, name)
	}

	return query, values
}
