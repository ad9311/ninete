# Monthly Expense Report

## Purpose
A PDF summarising one billed month's expenses — grouped by tag, with category
totals and budget comparison — downloaded on demand from
`/account/reports`.

This document is the rationale. `CLAUDE.md`'s Documentation map points here.

**The report is finished.** Both phases below are built, and the scheduled
email that was once Phase 3 has been dropped — see "The email, and why there
is none".

## Why the period is one calendar month of `date`

The owner pays for most things with two credit cards, one cutting off on the
21st and one on the 25th, and settles the rest by bank transfer. A card
purchase made on or after its cutoff day bills to the *next* month, and a
purchase made a day or two before the cutoff can still slip into the next cycle
because settlement takes one or two business days. Bank transfers have no cycle
at all: the money leaves now, so they belong to the current calendar month.

Taken as *purchase* dates, one report period is therefore a ragged ~41-day
window — for a September report, roughly 21 August through 30 September, since
it must span the earliest card purchase that bills to September and the last
transfer made inside it.

**None of that reaches the query.** `expenses.date` already stores the *billed
month*: it is picked and displayed at month precision, and the owner files each
expense against the month it actually bills to, checking the bank statement
when the cutoff is ambiguous. Production bears this out — August's expenses
have `created_at` values spanning 26 July to 31 August, all filed under the
same `date`.

So the report period is exactly one calendar month of `date`, and `created_at`
never enters it:

```sql
WHERE "user_id" = ? AND "date" >= :startOfMonth AND "date" < :startOfNextMonth
```

The ragged window is a fact about the owner's banking, not about this code. Do
not reintroduce it as purchase-date arithmetic — the manual billed month is the
authority, and a mistake in it is corrected by editing the expense, not by
widening a query.

## What the app does not model

**Cards are not a concept here.** There is no `payment_method` column, no
`cards` table, and no cutoff-day arithmetic. The owner tracks which card or
transfer paid for something with ordinary tags, and the report groups by tags
it is told to use without knowing what they mean. Keep it that way: adding
card awareness would require storing a true purchase date alongside the billed
month, which is a schema change buying nothing the tags do not already give.

## Grouping rules

The report's tag sections are configured in the settings page, as a set of
tags with no order.

- **An expense lands in exactly one section.** Its section is the selected tag
  whose *tagging* was created first — `ORDER BY "taggings"."created_at",
  "taggings"."id"`. The `id` tie-break is load-bearing: tagging several tags in
  one form submit writes the same `strftime('%s','now')`, so `created_at` alone
  is not a total order.
- **No expense is ever counted twice**, so section subtotals sum to the grand
  total. The settings page states this, because a reader who knows an expense
  carries two group tags would otherwise expect it in both sections.
- **Expenses matching no selected tag** go to a section named `Untagged`,
  pinned last regardless of its total.
- **Sections are ordered by total, descending**, `Untagged` excepted.
- **A selected tag with no expenses in the period is omitted**, not printed as
  a `$0.00` heading.
- **With no tags selected the report is one flat list** and no sections appear.

## Timezone

**There is no timezone anywhere in this feature, and none is needed.**
`expenses.date` is a UTC-midnight month start, so every stored value is already
month-precision and zone-free. The download's month arrives from the client as
a plain `YYYY-MM`; turning it into `[start, end)` is calendar arithmetic on
values that carry no instant, so unlike the `/api/expenses*` ranges (§3.6 of
`docs/spa-migration.md`) there is no client zone to resolve.

`report_settings` did carry a `timezone` column, and the settings page a zone
select, for the scheduled send: a job firing at 20:00 on 30 September local
reads 01:00 UTC on 1 October and would have reported the wrong month. With the
send dropped nothing read the column, so migration 33 removed it along with the
select, `time/tzdata`, and the zone validation in
`internal/logic/logic_report_setting.go`. Do not reintroduce a zone setting
without a reader for it.

One rough edge is left deliberately. `parseReportMonth`
(`internal/handlers/handle_reports.go`) defaults an omitted `?month=` in UTC,
while the picker that normally supplies it resolves last month in the browser's
zone (`lib/dates.ts`'s `lastCalendarMonth`), so on the 1st the two can name
different months for a few hours. The default is reachable only by typing the
URL with no query — the picker always sends one — and the month it chose is
printed in the report's own header, so a reader sees it and re-downloads.

## The email, and why there is none

The plan's third phase was a monthly send: a `report_deliveries` table so a
cron job firing twice did not send twice, Resend over `net/smtp` so no vendor
SDK entered the repo, SPF/DKIM/DMARC on `ninete.xyz`, and a
`task.SendMonthlyReport` hook run by `cmd/task` beside
`CopyDueRecurrentExpenses`. It was never started, and the owner dropped it: the
report is one click from `/account/reports` for the one person who reads it,
and a delivery pipeline is a standing source of failures — bounces, reputation,
a cron job silently not firing — in exchange for saving that click.

What the phase would have cost is the useful part of the record. It was the
only reason the app stored a timezone, and the only thing that would have made
the app send mail at all. Reviving it means bringing back both, plus the
`report_deliveries` idempotency table; reviving the timezone alone buys
nothing.

## Phases

Both are built. They are kept as a record of what was decided and why, not as
a plan with work left in it.

### Phase 1 — settings table and UI — **done** (PR #151)

Shipped standing alone, before anything consumed the settings: at the time the
page only saved and reloaded. Phase 2 is what reads them.

- Migration (`user_version` 32):
  - `report_settings` — `user_id` (unique, FK cascade), `timezone` TEXT,
    timestamps. The `timezone` column was dropped again by migration 33 with
    the email; the row now holds nothing but its `user_id` and timestamps, and
    exists to give the tag join rows a parent.
  - `report_setting_tags` — `report_setting_id` (FK cascade), `tag_id` (FK
    cascade), timestamps, unique index on the pair. Storing `tag_id` rather
    than a name means deleting a tag removes it from the report config for
    free. There is no `position` column: assignment is by earliest tagging and
    display is by total, so no configured order is ever read.
- `internal/repo/report_setting.go` with its columns constant, per the
  `SELECT *` invariant.
- `internal/logic/logic_report_setting.go`. Its zone validation — rejecting
  `"Local"` by name, and the blank `time/tzdata` import that let a deployed
  binary resolve a zone without the host's tzdata — went with the column in
  migration 33.
- `report_settings` joins the tables `Store.DeleteAllUserData` clears. The
  join rows would cascade away with the tags regardless, but the parent row
  must go too: an orphan config row surviving "delete all my data" is wrong on
  its face, and nothing above the repo can see it, which is why the test
  asserts on `SelectReportSettingByUser` directly.
- `internal/handlers/handle_api_report_settings.go` — `GET`/`PUT`
  `/api/report-settings`.
- SPA route `/account/reports`: tag multi-select and the first-tag-wins note
  (the timezone select is gone with migration 33). The form stays disabled
  until the `GET` lands, because
  `PUT` replaces the tag list wholesale — a failed load would otherwise leave
  an empty, enabled form whose Save wipes every grouping tag, the page having
  no way to tell "none selected" from "never found out". The tag limit rides
  along in the response rather than being repeated in the client, so the
  checkboxes cap where the server does.

### Phase 2 — the PDF, on demand — **done**

- `github.com/go-pdf/fpdf`.
- A new `internal/report` package, kept pure: rows in, PDF bytes out, no HTTP
  and no database. That is what makes the layout testable without a server.
- `GET /reports/expenses.pdf` on the **page chain, not `/api/*`** — it is
  reached by a plain anchor, so an expired session must answer with a redirect
  the browser can follow rather than the API chain's `401`. Same reasoning as
  `GET /exports/expenses.json`; see the route map in `CLAUDE.md`. The link
  needs `rel="external"` or `router.ts`'s `onLinkClick` swallows it.
- A month picker defaulting to last month.
- Contents: period header, tag sections with subtotals, `Untagged`, grand
  total, a category-totals bar chart, and the single-month budget table
  (spent, budget, left, over) matching the budgets page in `month` mode.
- The chart is drawn as fpdf rectangles. Chart.js is a browser library and
  cannot render here; a headless-Chrome HTML-to-PDF pipeline is rejected —
  it would add a large external binary and fight the unit's
  `ProtectSystem=strict` sandbox.

As built, with the parts that were not obvious from the plan:

- **The month travels as `?month=YYYY-MM`, not as resolved bounds.** Every
  `/api/expenses*` range sends epoch bounds the client resolved, because
  `created_at` is an *instant* and only the browser knows the zone (§3.6 of
  `docs/spa-migration.md`). The billed date is not an instant: it is
  month-precision and carries no zone, so the month is the whole of the input
  and the server needs nothing else. An omitted `month` means last month.
- **`budgetLeft`/`budgetPercent` moved to `internal/logic` as `BudgetLeft` and
  `BudgetPercent`.** The budgets endpoint and the report both need them, and a
  second copy is the kind that drifts — the report would go on saying "under
  budget" after the endpoint's rule changed.
- **The bars are scaled against the largest category, not the month's total.**
  With one dominant category every other bar rounds to nothing otherwise.
- **The core fonts are Latin-1**, so the renderer runs every string through
  fpdf's cp1252 translator. An expense described in Spanish is the normal case
  here, and without it the accents reach the page as mojibake.
- **The expense table's column header prints once**, under the "Expenses"
  heading rather than under every tag section — repeated, it turned the list
  into stripes.
- **Long text is truncated with an ellipsis** rather than left to fpdf, which
  clips silently: a cut word reads as a missing word. The fit is measured on
  the *translated* string, not the raw one — `GetStringWidth` walks bytes for a
  core font, so a UTF-8 `ñ` bills as two glyphs against the one cp1252 byte
  that reaches the page, and measuring the raw string cut Spanish descriptions
  short of columns they fit in.

Accuracy of the report outranks delivery. A month where the numbers are wrong
is the one outcome this feature cannot have.
