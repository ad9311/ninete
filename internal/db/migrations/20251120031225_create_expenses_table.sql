-- +goose Up
CREATE TABLE IF NOT EXISTS "expenses" (
  "id" INTEGER PRIMARY KEY NOT NULL,
	"user_id" INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
	"category_id" INTEGER NOT NULL REFERENCES "categories"("id") ON DELETE CASCADE,
	"description" TEXT NOT NULL,
	"amount" REAL NOT NULL,
  "date" INTEGER NOT NULL,
  "created_at" INTEGER NOT NULL DEFAULT (strftime('%s','now')),
  "updated_at" INTEGER NOT NULL DEFAULT (strftime('%s','now'))
);

PRAGMA user_version = 4;

-- +goose Down
DROP TABLE IF EXISTS "expenses";

PRAGMA user_version = 3;
