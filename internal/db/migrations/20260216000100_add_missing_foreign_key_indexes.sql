-- +goose Up
CREATE INDEX IF NOT EXISTS "idx_expenses_user_id" ON "expenses" ("user_id");
CREATE INDEX IF NOT EXISTS "idx_expenses_category_id" ON "expenses" ("category_id");
CREATE INDEX IF NOT EXISTS "idx_recurrent_expenses_user_id" ON "recurrent_expenses" ("user_id");
CREATE INDEX IF NOT EXISTS "idx_recurrent_expenses_category_id" ON "recurrent_expenses" ("category_id");

-- +goose Down
DROP INDEX IF EXISTS "idx_expenses_user_id";
DROP INDEX IF EXISTS "idx_expenses_category_id";
DROP INDEX IF EXISTS "idx_recurrent_expenses_user_id";
DROP INDEX IF EXISTS "idx_recurrent_expenses_category_id";
