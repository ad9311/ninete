-- +goose Up
-- Invitation codes are now generated with 128 bits of entropy, stored only as
-- their SHA-256 ("code_fingerprint"), and deleted by the sign-up that uses
-- them. The bcrypt "code_hash" is dropped: next to an unsalted SHA-256 of the
-- same code it protected nothing, since a short code could be cracked from
-- the fingerprint alone.
--
-- Every existing code is deleted. They were chosen by hand, so they may be
-- short enough to crack from a copy of the database, and they were reusable,
-- so any of them may already have been passed around. Generate a new one with
-- the create_invitation_code task when someone needs to register.
DELETE FROM "invitation_codes";

ALTER TABLE "invitation_codes" DROP COLUMN "code_hash";

PRAGMA user_version = 36;

-- +goose Down
-- Restores the previous layout, empty. A code created after the Up has no
-- bcrypt hash for the previous binary to check, so none can be carried back.
DROP INDEX IF EXISTS "uq_invitation_codes_code_fingerprint";
DROP TABLE IF EXISTS "invitation_codes";

CREATE TABLE IF NOT EXISTS "invitation_codes" (
  "id" INTEGER PRIMARY KEY NOT NULL,
  "code_hash" BLOB NOT NULL,
  "code_fingerprint" TEXT NOT NULL,
  "created_at" INTEGER NOT NULL DEFAULT (strftime('%s','now')),
  "updated_at" INTEGER NOT NULL DEFAULT (strftime('%s','now'))
);

CREATE UNIQUE INDEX IF NOT EXISTS "uq_invitation_codes_code_fingerprint"
ON "invitation_codes" ("code_fingerprint");

PRAGMA user_version = 35;
