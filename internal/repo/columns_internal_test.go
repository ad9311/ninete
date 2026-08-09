package repo

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ad9311/ninete/internal/prog"
	_ "github.com/mattn/go-sqlite3" // Database driver, normally registered by internal/db
	"github.com/stretchr/testify/require"
)

// buildSchema applies the Up half of every migration to a throwaway database.
//
// The migrations are read off disk rather than through internal/db: that package
// imports logic, which imports this one, so a test inside package repo cannot
// reach it. This only understands plain goose files, which is all the project
// has — a migration using +goose StatementBegin would need handling here.
func buildSchema(t *testing.T) *sql.DB {
	t.Helper()

	root, ok := prog.FindRoot()
	require.True(t, ok, "failed to locate project root")

	migrationsDir := filepath.Join(root, "internal", "db", "migrations")
	entries, err := os.ReadDir(migrationsDir)
	require.NoError(t, err)

	sqlDB, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "columns_test.db"))
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, sqlDB.Close())
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		//nolint:gosec // path is the project's own migrations directory
		content, err := os.ReadFile(filepath.Join(migrationsDir, entry.Name()))
		require.NoError(t, err)

		_, up, found := strings.Cut(string(content), "-- +goose Up")
		require.True(t, found, "%s has no Up section", entry.Name())

		up, _, _ = strings.Cut(up, "-- +goose Down")

		_, err = sqlDB.Exec(up)
		require.NoError(t, err, "failed to apply %s", entry.Name())
	}

	return sqlDB
}

// normalizeColumns flattens a column constant onto one line so a wrapped
// constant compares equal to the list built from the schema.
func normalizeColumns(columns string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(columns, "\n", " ")), " ")
}

// TestColumnConstantsMatchSchema is the guard that makes the explicit column
// lists worth having. Each constant must name every column of its table in
// physical order, because the Scan calls read them positionally: a column added
// mid-table by a future migration, or a constant left un-updated, would land
// values in the wrong struct fields.
func TestColumnConstantsMatchSchema(t *testing.T) {
	sqlDB := buildSchema(t)
	ctx := t.Context()

	cases := []struct {
		table   string
		columns string
	}{
		{"categories", categoryColumns},
		{"expense_category_mappings", expenseCategoryMappingColumns},
		{"expenses", expenseColumns},
		{"foods", foodColumns},
		{"invitation_codes", invitationCodeColumns},
		{"macro_entries", macroEntryColumns},
		{"macro_goals", macroGoalColumns},
		{"mood_entries", moodEntryColumns},
		{"recurrent_expenses", recurrentExpenseColumns},
		{"tags", tagColumns},
		{"users", userColumns},
	}

	for _, c := range cases {
		t.Run(c.table, func(t *testing.T) {
			rows, err := sqlDB.QueryContext(ctx, `SELECT "name" FROM pragma_table_info(?)`, c.table)
			require.NoError(t, err)
			defer func() {
				require.NoError(t, rows.Close())
			}()

			var want []string

			for rows.Next() {
				var name string

				require.NoError(t, rows.Scan(&name))
				want = append(want, `"`+name+`"`)
			}

			require.NoError(t, rows.Err())
			require.NotEmpty(t, want, "table %q has no columns", c.table)

			require.Equal(t, strings.Join(want, ", "), normalizeColumns(c.columns))
		})
	}
}

// TestAliasedTagColumnsMatchTagColumns keeps the joined projection in tagging.go
// in step with the plain one, since both feed the same scanTagRows.
func TestAliasedTagColumnsMatchTagColumns(t *testing.T) {
	require.Equal(t, tagColumns, strings.ReplaceAll(tagColumnsAliased, "t.", ""))
}
