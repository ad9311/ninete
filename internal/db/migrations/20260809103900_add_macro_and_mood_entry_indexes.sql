-- +goose Up
-- Both tables are only ever read scoped to a user and bounded by a timestamp:
-- the macro day/daily totals, the macro and mood listings, and the mood counts.
-- SQLite creates no index for an inline REFERENCES column, so every one of
-- those queries scanned the whole table.
CREATE INDEX IF NOT EXISTS "idx_macro_entries_user_date"
ON "macro_entries" ("user_id", "date");

CREATE INDEX IF NOT EXISTS "idx_mood_entries_user_logged_at"
ON "mood_entries" ("user_id", "logged_at");

PRAGMA user_version = 28;

-- +goose Down
DROP INDEX IF EXISTS "idx_macro_entries_user_date";
DROP INDEX IF EXISTS "idx_mood_entries_user_logged_at";

PRAGMA user_version = 27;
