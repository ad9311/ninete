package repo

import (
	"context"
	"strings"
)

// ExpenseGroupTag is one (expense, grouping tag) pair in the order the monthly
// report resolves sections with: the first row for an expense names the
// section it belongs to.
type ExpenseGroupTag struct {
	ExpenseID int
	TagName   string
}

// selectExpenseGroupTagsPrefix is completed by buildSelectExpenseGroupTags,
// which appends one placeholder per grouping tag.
//
// The ORDER BY is the whole point of the query: an expense's report section is
// the grouping tag whose *tagging* was created first, and the "id" tie-break is
// load-bearing because tagging several tags in one form submit writes the same
// strftime('%s','now') — created_at alone is not a total order
// (docs/monthly-report.md, "Grouping rules").
//
// Rows are bounded by the tag list rather than by a list of expense ids, so
// there is no chunking to do: the grouping tags are capped well below SQLite's
// parameter limit, however many expenses the month holds.
const selectExpenseGroupTagsPrefix = `
SELECT tg."taggable_id", t."name"
FROM "taggings" tg
INNER JOIN "tags" t ON t."id" = tg."tag_id"
INNER JOIN "expenses" e ON e."id" = tg."taggable_id"
WHERE tg."taggable_type" = ?
  AND e."user_id" = ?
  AND e."date" >= ?
  AND e."date" < ?
  AND tg."tag_id" IN (`

const selectExpenseGroupTagsSuffix = `)
ORDER BY tg."taggable_id" ASC, tg."created_at" ASC, tg."id" ASC`

func buildSelectExpenseGroupTags(count int) string {
	var b strings.Builder

	b.WriteString(selectExpenseGroupTagsPrefix)
	b.WriteString(strings.TrimSuffix(strings.Repeat("?,", count), ","))
	b.WriteString(selectExpenseGroupTagsSuffix)

	return b.String()
}

// SelectExpenseGroupTags returns every tagging linking an expense billed in
// [start, end) to one of tagIDs, ordered so the first row for an expense is
// the tag that decides its section. An empty tagIDs returns nothing: a report
// configured with no grouping tags is one flat list, not an error.
func (q *Queries) SelectExpenseGroupTags(
	ctx context.Context,
	userID int,
	start, end int64,
	tagIDs []int,
) ([]ExpenseGroupTag, error) {
	var out []ExpenseGroupTag

	if len(tagIDs) == 0 {
		return out, nil
	}

	taggable := TaggableExpense()
	if err := taggable.validate(); err != nil {
		return nil, err
	}

	query := buildSelectExpenseGroupTags(len(tagIDs))

	args := make([]any, 0, len(tagIDs)+4)
	args = append(args, taggable.Type(), userID, start, end)

	for _, id := range tagIDs {
		args = append(args, id)
	}

	err := q.wrapQuery(query, func() error {
		rows, err := q.db.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				q.app.Logger.Error(closeErr)
			}
		}()

		for rows.Next() {
			var row ExpenseGroupTag

			if err := rows.Scan(&row.ExpenseID, &row.TagName); err != nil {
				return err
			}

			out = append(out, row)
		}

		return rows.Err()
	})

	return out, err
}
