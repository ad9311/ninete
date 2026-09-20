-- +goose Up
-- The monthly report's scheduled email was dropped (docs/monthly-report.md),
-- and the timezone existed for that one job: it decided which month a run
-- firing near midnight called "last month". Nothing else ever read it — the
-- on-demand download takes its month from the client — so with the send gone
-- the column was write-only configuration, and a settings field nobody
-- consumes reads as a promise the app does not keep.
--
-- Migrations run while the previous binary is still serving, and that binary
-- still names "timezone" in its column constant, so GET/PUT
-- /api/report-settings and GET /reports/expenses.pdf fail with "no such
-- column: timezone" until the restart lands.
ALTER TABLE "report_settings" DROP COLUMN "timezone";

PRAGMA user_version = 33;

-- +goose Down
-- Recreates the column empty, defaulted to UTC. This is not a restore: the
-- saved zone does not come back.
--
-- It also returns at the *end* of the table rather than in its original third
-- position, ALTER TABLE having no way to insert a column mid-list. The Scans
-- still read correctly — every statement names its columns through
-- reportSettingColumns, and SQLite returns an explicit projection in the order
-- written — but TestColumnConstantsMatchSchema compares the constant against
-- pragma_table_info order, so a pre-33 checkout's suite fails against a
-- rolled-back database even though its queries work.
ALTER TABLE "report_settings" ADD COLUMN "timezone" TEXT NOT NULL DEFAULT 'UTC';

PRAGMA user_version = 32;
