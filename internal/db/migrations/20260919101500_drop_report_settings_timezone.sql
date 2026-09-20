-- +goose Up
-- The monthly report's scheduled email was dropped (docs/monthly-report.md),
-- and the timezone existed for that one job: it decided which month a run
-- firing near midnight called "last month". Nothing else ever read it — the
-- on-demand download takes its month from the client — so with the send gone
-- the column was write-only configuration, and a settings field nobody
-- consumes reads as a promise the app does not keep.
ALTER TABLE "report_settings" DROP COLUMN "timezone";

PRAGMA user_version = 33;

-- +goose Down
-- Recreates the column empty, defaulted to UTC. This is not a restore: the
-- saved zone does not come back.
--
-- It also returns at the *end* of the table rather than in its original third
-- position, ALTER TABLE having no way to insert a column mid-list. Code that
-- expects the pre-33 physical order — reportSettingColumns as it read then —
-- would scan values into the wrong fields, so rolling back past this point
-- means rebuilding the table, not just running this Down.
ALTER TABLE "report_settings" ADD COLUMN "timezone" TEXT NOT NULL DEFAULT 'UTC';

PRAGMA user_version = 32;
