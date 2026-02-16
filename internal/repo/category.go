package repo

import "context"

type Category struct {
	ID        int
	Name      string
	UID       string
	CreatedAt int64
	UpdatedAt int64
}

const insertCategory = `
INSERT INTO "categories" ("name", "uid")
VALUES (?, ?)
RETURNING *`

func (q *Queries) InsertCategory(ctx context.Context, name, uid string) (Category, error) {
	var c Category

	err := q.wrapQuery(insertCategory, func() error {
		row := q.db.QueryRowContext(ctx, insertCategory, name, uid)

		return row.Scan(
			&c.ID,
			&c.Name,
			&c.UID,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
	})

	return c, err
}

const selectCategories = `
SELECT * FROM "categories" ORDER BY "name"`

func (q *Queries) SelectCategories(ctx context.Context) ([]Category, error) {
	var cs []Category

	err := q.wrapQuery(selectCategories, func() error {
		rows, err := q.db.QueryContext(ctx, selectCategories)
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				q.app.Logger.Error(closeErr)
			}
		}()

		for rows.Next() {
			var c Category

			if err := rows.Scan(
				&c.ID,
				&c.Name,
				&c.UID,
				&c.CreatedAt,
				&c.UpdatedAt,
			); err != nil {
				return err
			}

			cs = append(cs, c)
		}

		return rows.Err()
	})

	return cs, err
}
