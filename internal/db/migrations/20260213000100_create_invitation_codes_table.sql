-- +goose Up
CREATE TABLE IF NOT EXISTS "invitation_codes" (
  "id" INTEGER PRIMARY KEY NOT NULL,
  "code_hash" BLOB NOT NULL,
  "code_fingerprint" TEXT NOT NULL,
  "created_at" INTEGER NOT NULL DEFAULT (strftime('%s','now')),
  "updated_at" INTEGER NOT NULL DEFAULT (strftime('%s','now'))
);

CREATE UNIQUE INDEX IF NOT EXISTS "uq_invitation_codes_code_fingerprint"
ON "invitation_codes" ("code_fingerprint");

PRAGMA user_version = 7;

-- +goose Down
DROP INDEX IF EXISTS "uq_invitation_codes_code_fingerprint";
DROP TABLE IF EXISTS "invitation_codes";

PRAGMA user_version = 6;
