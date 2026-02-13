-- +goose Up
CREATE TABLE IF NOT EXISTS "taggings" (
  "id" INTEGER PRIMARY KEY NOT NULL,
  "tag_id" INTEGER NOT NULL REFERENCES "tags"("id") ON DELETE CASCADE,
  "taggable_id" INTEGER NOT NULL,
  "taggable_type" TEXT NOT NULL,
  "created_at" INTEGER NOT NULL DEFAULT (strftime('%s','now')),
  "updated_at" INTEGER NOT NULL DEFAULT (strftime('%s','now')),
  CHECK ("taggable_id" > 0),
  CHECK (length(trim("taggable_type")) > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS "uq_taggings_tag_target"
ON "taggings" ("tag_id", "taggable_type", "taggable_id");

CREATE INDEX IF NOT EXISTS "idx_taggings_taggable"
ON "taggings" ("taggable_type", "taggable_id");

CREATE INDEX IF NOT EXISTS "idx_taggings_type_tag_id"
ON "taggings" ("taggable_type", "tag_id");

PRAGMA user_version = 6;

-- +goose Down
DROP INDEX IF EXISTS "idx_taggings_type_tag_id";
DROP INDEX IF EXISTS "idx_taggings_taggable";
DROP INDEX IF EXISTS "uq_taggings_tag_target";
DROP TABLE IF EXISTS "taggings";

PRAGMA user_version = 5;
