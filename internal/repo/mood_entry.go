package repo

import (
	"context"
	"fmt"
)

type MoodEntry struct {
	ID        int
	UserID    int
	Mood      string
	Notes     string
	LoggedAt  int64
	CreatedAt int64
	UpdatedAt int64
}

type InsertMoodEntryParams struct {
	UserID   int
	Mood     string
	Notes    string
	LoggedAt int64
}

type UpdateMoodEntryParams struct {
	ID       int
	UserID   int
	Mood     string
	Notes    string
	LoggedAt int64
}

const selectMoodEntries = `SELECT * FROM "mood_entries"`

func (q *Queries) SelectMoodEntries(ctx context.Context, opts QueryOptions) ([]MoodEntry, error) {
	var es []MoodEntry

	if err := opts.Validate(validMoodEntryFields()); err != nil {
		return es, err
	}

	subQuery, err := opts.Build()
	if err != nil {
		return es, err
	}

	query := selectMoodEntries + " " + subQuery
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
			var e MoodEntry

			if err := rows.Scan(
				&e.ID,
				&e.UserID,
				&e.Mood,
				&e.Notes,
				&e.LoggedAt,
				&e.CreatedAt,
				&e.UpdatedAt,
			); err != nil {
				return err
			}

			es = append(es, e)
		}

		return rows.Err()
	})

	return es, err
}

const countMoodEntries = `SELECT COUNT(*) FROM "mood_entries"`

func (q *Queries) CountMoodEntries(ctx context.Context, filters Filters) (int, error) {
	var c int

	subQuery, err := filters.Build()
	if err != nil {
		return 0, err
	}

	query := countMoodEntries + " " + subQuery
	values := filters.Values()

	err = q.wrapQuery(query, func() error {
		row := q.db.QueryRowContext(ctx, query, values...)

		return row.Scan(&c)
	})

	return c, err
}

const selectMoodEntry = `SELECT * FROM "mood_entries" WHERE "id" = ? AND "user_id" = ? LIMIT 1`

func (q *Queries) SelectMoodEntry(ctx context.Context, id, userID int) (MoodEntry, error) {
	var e MoodEntry

	err := q.wrapQuery(selectMoodEntry, func() error {
		row := q.db.QueryRowContext(ctx, selectMoodEntry, id, userID)

		return row.Scan(
			&e.ID,
			&e.UserID,
			&e.Mood,
			&e.Notes,
			&e.LoggedAt,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
	})

	return e, err
}

const insertMoodEntry = `
INSERT INTO "mood_entries" ("user_id", "mood", "notes", "logged_at")
VALUES (?, ?, ?, ?)
RETURNING *`

func (q *TxQueries) InsertMoodEntry(ctx context.Context, params InsertMoodEntryParams) (MoodEntry, error) {
	var e MoodEntry

	err := q.wrapQuery(insertMoodEntry, func() error {
		row := q.tx.QueryRowContext(
			ctx,
			insertMoodEntry,
			params.UserID,
			params.Mood,
			params.Notes,
			params.LoggedAt,
		)

		return row.Scan(
			&e.ID,
			&e.UserID,
			&e.Mood,
			&e.Notes,
			&e.LoggedAt,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
	})

	return e, err
}

const updateMoodEntry = `
UPDATE "mood_entries"
SET "mood"       = ?,
    "notes"      = ?,
    "logged_at"  = ?,
    "updated_at" = ?
WHERE "id" = ?
  AND "user_id" = ?
RETURNING *`

func (q *TxQueries) UpdateMoodEntry(ctx context.Context, params UpdateMoodEntryParams) (MoodEntry, error) {
	var e MoodEntry

	err := q.wrapQuery(updateMoodEntry, func() error {
		row := q.tx.QueryRowContext(
			ctx,
			updateMoodEntry,
			params.Mood,
			params.Notes,
			params.LoggedAt,
			newUpdatedAt(),
			params.ID,
			params.UserID,
		)

		return row.Scan(
			&e.ID,
			&e.UserID,
			&e.Mood,
			&e.Notes,
			&e.LoggedAt,
			&e.CreatedAt,
			&e.UpdatedAt,
		)
	})

	return e, err
}

const deleteMoodEntry = `DELETE FROM "mood_entries" WHERE "id" = ? AND "user_id" = ? RETURNING "id"`

func (q *Queries) DeleteMoodEntry(ctx context.Context, id, userID int) (int, error) {
	var i int

	err := q.wrapQuery(deleteMoodEntry, func() error {
		row := q.db.QueryRowContext(ctx, deleteMoodEntry, id, userID)

		return row.Scan(&i)
	})

	return i, err
}

const countMoodEntriesByUser = `SELECT COUNT(*) FROM "mood_entries" WHERE "user_id" = ?`

func (q *Queries) CountMoodEntriesByUser(ctx context.Context, userID int) (int, error) {
	var c int

	err := q.wrapQuery(countMoodEntriesByUser, func() error {
		row := q.db.QueryRowContext(ctx, countMoodEntriesByUser, userID)

		return row.Scan(&c)
	})

	return c, err
}

const deleteMoodEntryTaggingsByUser = `
DELETE FROM "taggings"
WHERE "taggable_type" = 'mood_entry'
  AND "taggable_id" IN (SELECT "id" FROM "mood_entries" WHERE "user_id" = ?)`

const deleteAllMoodEntriesByUser = `DELETE FROM "mood_entries" WHERE "user_id" = ?`

func (q *TxQueries) DeleteAllMoodEntriesByUser(ctx context.Context, userID int) error {
	return q.wrapQuery(deleteAllMoodEntriesByUser, func() error {
		if _, err := q.tx.ExecContext(ctx, deleteMoodEntryTaggingsByUser, userID); err != nil {
			return err
		}

		_, err := q.tx.ExecContext(ctx, deleteAllMoodEntriesByUser, userID)

		return err
	})
}

type MoodCount struct {
	Mood  string
	Count int
}

const selectMoodEntryCountsBase = `
SELECT "mood", COUNT(*) AS "count"
FROM "mood_entries"
%s
GROUP BY "mood"`

func (q *Queries) SelectMoodEntryCounts(ctx context.Context, filters Filters) ([]MoodCount, error) {
	var counts []MoodCount

	filterSubQuery, err := filters.Build()
	if err != nil {
		return counts, err
	}

	query := fmt.Sprintf(selectMoodEntryCountsBase, filterSubQuery)
	values := filters.Values()

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
			var c MoodCount

			if err := rows.Scan(&c.Mood, &c.Count); err != nil {
				return err
			}

			counts = append(counts, c)
		}

		return rows.Err()
	})

	return counts, err
}

func validMoodEntryFields() []string {
	return []string{
		"id",
		"user_id",
		"mood",
		"notes",
		"logged_at",
		"created_at",
		"updated_at",
	}
}
