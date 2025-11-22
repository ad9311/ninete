package repo

import "context"

type Category struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	UID       string `json:"uid"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

const insertCategory = `
INSERT INTO "categories" ("name", "uid")
VALUES (?, ?)
RETURNING *`

func (q *Queries) InserCategory(ctx context.Context, name, uid string) (Category, error) {
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
