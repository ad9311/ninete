-- +goose Up
-- An optional free-text note, shown only on the expense's own page. NOT NULL
-- with an empty-string default so "no note" has one representation and the
-- Go field stays a plain string. Appended at the end of the table, so the
-- previous binary — still serving while this runs — keeps working: its
-- explicit column list never names "note", and its inserts take the default.
ALTER TABLE "expenses" ADD COLUMN "note" TEXT NOT NULL DEFAULT '';

PRAGMA user_version = 34;

-- +goose Down
-- Discards every saved note.
ALTER TABLE "expenses" DROP COLUMN "note";

PRAGMA user_version = 33;
