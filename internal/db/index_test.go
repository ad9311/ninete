package db_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ad9311/ninete/internal/db"
	"github.com/stretchr/testify/require"
)

// openMigrated runs the migrations against a throwaway file and returns a pool
// on it, so the query plans below are checked against the real schema.
func openMigrated(t *testing.T) *sql.DB {
	t.Helper()

	t.Setenv("DATABASE_URL", filepath.Join(t.TempDir(), "index_test.db"))
	require.NoError(t, db.RunMigrationsUp())

	sqlDB, err := db.Open()
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, sqlDB.Close())
	})

	return sqlDB
}

func queryPlan(ctx context.Context, t *testing.T, sqlDB *sql.DB, query string, args ...any) string {
	t.Helper()

	rows, err := sqlDB.QueryContext(ctx, "EXPLAIN QUERY PLAN "+query, args...)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, rows.Close())
	}()

	var plan strings.Builder

	for rows.Next() {
		var id, parent, notUsed int
		var detail string

		require.NoError(t, rows.Scan(&id, &parent, &notUsed, &detail))
		plan.WriteString(detail)
		plan.WriteString("\n")
	}

	require.NoError(t, rows.Err())

	return plan.String()
}

// TestUserScopedQueryPlans pins the access path for the queries that run on
// every dashboard and listing view. A plan that falls back to SCAN means an
// index was dropped or a WHERE clause drifted away from it.
func TestUserScopedQueryPlans(t *testing.T) {
	cases := []struct {
		name      string
		query     string
		args      []any
		wantIndex string
	}{
		{
			name: "macro day totals",
			query: `SELECT COALESCE(SUM("kcal"), 0) FROM "macro_entries"
			        WHERE "user_id" = ? AND "date" >= ? AND "date" < ?`,
			args:      []any{1, 0, 1},
			wantIndex: "idx_macro_entries_user_date",
		},
		{
			name: "macro daily totals",
			query: `SELECT "date", COALESCE(SUM("kcal"), 0) FROM "macro_entries"
			        WHERE "user_id" = ? AND "date" >= ? AND "date" < ? GROUP BY "date"`,
			args:      []any{1, 0, 1},
			wantIndex: "idx_macro_entries_user_date",
		},
		{
			name:      "macro entries listing",
			query:     `SELECT "id" FROM "macro_entries" WHERE "user_id" = ? ORDER BY "date" DESC LIMIT 15`,
			args:      []any{1},
			wantIndex: "idx_macro_entries_user_date",
		},
		{
			name:      "mood entries listing",
			query:     `SELECT "id" FROM "mood_entries" WHERE "user_id" = ? ORDER BY "logged_at" DESC LIMIT 15`,
			args:      []any{1},
			wantIndex: "idx_mood_entries_user_logged_at",
		},
		{
			name: "expenses category month totals",
			query: `SELECT "category_id", strftime('%Y-%m', "date", 'unixepoch') AS "month",
			        SUM("amount") FROM "expenses"
			        WHERE "user_id" = ? AND "date" >= ? AND "date" < ?
			        GROUP BY "category_id", "month"`,
			args:      []any{1, 0, 1},
			wantIndex: "idx_expenses_user_date",
		},
		{
			name:      "expenses listing",
			query:     `SELECT "id" FROM "expenses" WHERE "user_id" = ? ORDER BY "date" DESC LIMIT 15`,
			args:      []any{1},
			wantIndex: "idx_expenses_user_date",
		},
	}

	ctx := t.Context()
	sqlDB := openMigrated(t)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			plan := queryPlan(ctx, t, sqlDB, c.query, c.args...)

			require.Contains(t, plan, c.wantIndex, "plan was:\n%s", plan)
			require.NotContains(t, plan, "SCAN", "plan was:\n%s", plan)
		})
	}
}
