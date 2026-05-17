-- +goose Up
CREATE UNIQUE INDEX IF NOT EXISTS "uq_foods_user_lower_name"
ON "foods" ("user_id", lower("name"));

PRAGMA user_version = 20;

-- +goose Down
DROP INDEX IF EXISTS "uq_foods_user_lower_name";

PRAGMA user_version = 19;
