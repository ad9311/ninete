-- +goose Up
CREATE TABLE IF NOT EXISTS "report_settings" (
  "id" INTEGER PRIMARY KEY NOT NULL,
  "user_id" INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "timezone" TEXT NOT NULL,
  "created_at" INTEGER NOT NULL DEFAULT (strftime('%s','now')),
  "updated_at" INTEGER NOT NULL DEFAULT (strftime('%s','now'))
);

CREATE UNIQUE INDEX IF NOT EXISTS "uq_report_settings_user"
ON "report_settings" ("user_id");

-- The tags the monthly report groups by. There is no position column on
-- purpose: an expense's section is decided by which of its taggings was
-- created first, and sections print in descending total order, so no
-- configured order is ever read (docs/monthly-report.md, "Grouping rules").
-- The tag is referenced by id rather than by name so deleting a tag drops it
-- from the report configuration by cascade.
CREATE TABLE IF NOT EXISTS "report_setting_tags" (
  "id" INTEGER PRIMARY KEY NOT NULL,
  "report_setting_id" INTEGER NOT NULL REFERENCES "report_settings"("id") ON DELETE CASCADE,
  "tag_id" INTEGER NOT NULL REFERENCES "tags"("id") ON DELETE CASCADE,
  "created_at" INTEGER NOT NULL DEFAULT (strftime('%s','now')),
  "updated_at" INTEGER NOT NULL DEFAULT (strftime('%s','now'))
);

CREATE UNIQUE INDEX IF NOT EXISTS "uq_report_setting_tags_setting_tag"
ON "report_setting_tags" ("report_setting_id", "tag_id");

PRAGMA user_version = 32;

-- +goose Down
DROP INDEX IF EXISTS "uq_report_setting_tags_setting_tag";
DROP TABLE IF EXISTS "report_setting_tags";
DROP INDEX IF EXISTS "uq_report_settings_user";
DROP TABLE IF EXISTS "report_settings";

PRAGMA user_version = 31;
