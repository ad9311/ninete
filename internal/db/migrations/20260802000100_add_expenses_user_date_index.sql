-- +goose Up
CREATE INDEX IF NOT EXISTS "idx_expenses_user_date"
ON "expenses" ("user_id", "date");

-- "idx_expenses_user_id" is a strict prefix of the index above, so it can never
-- be preferred by the planner while it still costs a write on every insert,
-- update and delete.
DROP INDEX IF EXISTS "idx_expenses_user_id";

PRAGMA user_version = 26;

-- +goose Down
CREATE INDEX IF NOT EXISTS "idx_expenses_user_id" ON "expenses" ("user_id");

DROP INDEX IF EXISTS "idx_expenses_user_date";

PRAGMA user_version = 25;
