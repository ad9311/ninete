-- +goose Up
CREATE TABLE "sessions" (
  "token"  TEXT PRIMARY KEY NOT NULL,
  "data"   BLOB NOT NULL,
  "expiry" REAL NOT NULL
);
CREATE INDEX "sessions_expiry_idx" ON "sessions" ("expiry");
PRAGMA user_version = 15;

-- +goose Down
DROP TABLE "sessions";
PRAGMA user_version = 14;
