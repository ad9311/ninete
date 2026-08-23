# TODO

Known issues and follow-up work that is deliberately out of scope for the change
that surfaced it. Remove an entry once it is fixed.

## Deleting a single expense leaves its taggings behind

`Store.DeleteExpense` (`internal/logic/logic_expense.go`) removes the expense row
without deleting the matching `taggings` rows. Taggings key on `taggable_id`
rather than a foreign key, so nothing cascades: the orphaned rows survive, and
SQLite is free to hand that rowid to the next expense, which then inherits tags
it was never given.

The same hazard was fixed for recurrent expenses in `Store.DeleteRecurrentExpense`
by wrapping the delete and `DeleteTaggingsByTarget` in one transaction. The
expense path should follow that pattern.

This is pre-existing behaviour, but generating expenses from tagged recurrent
expenses widens the exposure: every generated expense now carries taggings, so
routine deletes produce orphans that previously only appeared when the user had
tagged an expense by hand.

Deleting *all* expenses is already safe — `DeleteAllExpensesByUser`
(`internal/repo/expense.go`) clears the taggings first.

## A unique-name collision will read as a server fault on the API

`Store.CreateFood` and `Store.UpdateFood` (`internal/logic/logic_food.go`) let SQLite's
`UNIQUE constraint failed: foods.name` propagate untouched, and the same holds for
every other table with a unique index. The pages render that string; the API's
`WriteAPIError` deliberately will not, so an unnamed collision becomes a generic
`500` instead — correct about not leaking, wrong about whose fault it is.

Whoever ports Foods in Phase 2 of `docs/spa-migration.md` should turn the
collision into a sentinel in `internal/logic/errs.go`, name it in the endpoint's
`userErrors`, and let it answer `422` with a message that does not quote the
constraint. Fixing it in the logic layer improves the existing pages too.
