-- +goose Up
CREATE TABLE IF NOT EXISTS "categories" (
  "id" INTEGER PRIMARY KEY NOT NULL,
	"name" TEXT NOT NULL UNIQUE,
	"uid" TEXT NOT NULL UNIQUE,
	"created_at" INTEGER NOT NULL DEFAULT (strftime('%s','now')),
  "updated_at" INTEGER NOT NULL DEFAULT (strftime('%s','now'))
);

PRAGMA user_version = 2;

-- +goose Down
DROP TABLE IF EXISTS "categories";

PRAGMA user_version = 1;
