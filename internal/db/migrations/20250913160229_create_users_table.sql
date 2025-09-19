-- +goose Up
CREATE TABLE "users" (
  "id" INTEGER PRIMARY KEY NOT NULL,
  'username' TEXT NOT NULL UNIQUE,
  "email" TEXT NOT NULL UNIQUE,
  "password_hash" BLOB NOT NULL,
  "created_at" INTEGER NOT NULL DEFAULT (strftime('%s','now')),
  "updated_at" INTEGER NOT NULL DEFAULT (strftime('%s','now'))
);

PRAGMA user_version = 1;

-- +goose Down
DROP TABLE IF EXISTS "users";

PRAGMA user_version = 0;
