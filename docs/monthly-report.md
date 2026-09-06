# Monthly Expense Report

## Purpose
A PDF summarising one billed month's expenses — grouped by tag, with category
totals and budget comparison — downloadable on demand and, later, emailed on
the 1st of each month.

This document is the plan and the rationale. `CLAUDE.md`'s Documentation map
points here.

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

`expenses.date` is a UTC-midnight month start, so every stored value is already
month-precision and timezone-free. The configured timezone decides exactly one
thing: which calendar month the *scheduled* job calls "last month" when it
fires. A job firing at 00:30 on 1 October in a UTC-5 zone still reads 05:30 UTC
on 1 October — but one firing at 20:00 on 30 September local would read 01:00
UTC on 1 October and report the wrong month, which is what the setting exists
to prevent.

The on-demand download takes its month from the client, so it does not consult
the setting. When no settings row exists the task falls back to UTC; the
settings form pre-fills from the browser's
`Intl.DateTimeFormat().resolvedOptions().timeZone`.

## Phases

### Phase 1 — settings table and UI

Ships standing alone: the page saves and reloads, and nothing consumes the
settings yet.

- Migration (`user_version` 31):
  - `report_settings` — `user_id` (unique, FK cascade), `timezone` TEXT,
    timestamps.
  - `report_setting_tags` — `report_setting_id` (FK cascade), `tag_id` (FK
    cascade), timestamps, unique index on the pair. Storing `tag_id` rather
    than a name means deleting a tag removes it from the report config for
    free. There is no `position` column: assignment is by earliest tagging and
    display is by total, so no configured order is ever read.
- `internal/repo/report_setting.go` with its columns constant, per the
  `SELECT *` invariant.
- `internal/logic/logic_report_setting.go`.
- `internal/handlers/handle_api_report_settings.go` — `GET`/`PUT`
  `/api/report-settings`.
- SPA route `/account/reports`: tag multi-select, timezone select, and the
  first-tag-wins note.

### Phase 2 — the PDF, on demand

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

### Phase 3 — email

- `report_deliveries` table recording the period sent, so a cron job that fires
  twice does not send twice.
- Resend, reached over SMTP with the standard library's `net/smtp`, so
  swapping providers is a config change and no vendor SDK enters the repo.
  Self-hosting a mail server on the VPS is rejected: a fresh IP with no
  reputation gets binned by the recipient's provider.
- Sender `reports@ninete.xyz`; recipient is the user's own `users.email`.
- Config from `/etc/ninete/env` — `prog.Load()` skips `.env` under
  `ENV=production`, so a `.env` on the host would be ignored.
- DNS records (SPF, DKIM, DMARC) go wherever `ninete.xyz`'s nameservers point.
- A `task.SendMonthlyReport` hook run by `cmd/task` from cron, alongside
  `CopyDueRecurrentExpenses`.
- Anything the task writes must live beside the database: the web unit runs
  under `ProtectSystem=strict`. See `docs/deployment.md`.

Accuracy of the report outranks delivery. A month where the email fails but
the download works is an acceptable outcome; a month where the numbers are
wrong is not.
