-- +goose Up
-- Supports bounding the expense search on "created_at" instead of the billed
-- date. Without it the planner can only narrow on "user_id" and then scans
-- every row the user owns.
CREATE INDEX IF NOT EXISTS "idx_expenses_user_created_at"
ON "expenses" ("user_id", "created_at");

PRAGMA user_version = 27;

-- +goose Down
DROP INDEX IF EXISTS "idx_expenses_user_created_at";

PRAGMA user_version = 26;
