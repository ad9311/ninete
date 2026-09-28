# TODO

Known issues and follow-up work that is deliberately out of scope for the change
that surfaced it. Remove an entry once it is fixed.

- **List tags with usage counts and delete the unused ones.** `POST /api/tags/retag` leaves its
  `from` tags in place, unused, on purpose — nothing is deleted without the owner seeing it — so
  orphaned tags accumulate and show up in `list_tags` and the report settings page. The app needs
  a tag page listing each tag with how many expenses and recurrent expenses carry it, and a way to
  delete one. That means a `GET /api/tags` with counts and a session-only `DELETE /api/tags/{id}`
  (tokens can never `DELETE`, so the MCP keeps its no-delete guarantee). Once `GET /api/tags`
  exists, `list_tags` in `mcp/` should read it instead of `/api/report-settings`. Left out of the
  retag change as a separate UI feature.
- **Put query-parameter names under the API contract.** `contract/api.json` records JSON
  bodies only, so the query keys the MCP server sends (`start`/`end`, `created_start`/`created_end`,
  `this_start`…`last_end`, `archived`, `mode`, `sort_field`/`sort_order`, `page`/`per_page`,
  `category_id`, `q`, `tag`) are unchecked: if a handler renamed one, the server would ignore the
  old name, answer with an unfiltered result, and no test would fail. The handlers have no single
  declaration of the keys they read — they are `q.Get("…")` literals across `shared.go`,
  `expense_search.go`, `expense_shared.go` and the `handle_api_*` files — so there is nothing for
  `TestAPIContract` to reflect over. The intended fix is to move each token-reachable handler's
  query parsing onto tagged query structs (or named constants) that `apiContract` references, so a
  renamed key changes the generated file by itself; then extend `contracttest` to fail when the
  MCP sends a key the file does not list. A hand-maintained key list in `apiContract` was
  rejected: it guards the MCP side but not a handler that renames a literal. Left out of the MCP
  server change because it refactors query parsing in about seven root-module handler files.
- **Read notes in bulk for statement reconciliation.** Reconciling a card statement through the
  MCP means explaining why an expense is lower than its statement line (a shared purchase where
  someone repaid their part in cash), but `search_expenses` omits `note` — the list's `apiExpense`
  leaves it out by design — so reading notes takes one `get_expense` call per expense. Two options,
  not mutually exclusive:
  - An opt-in `include_notes` query parameter on `GET /api/expenses`, surfaced as a flag on
    `search_expenses` and off by default, so the list keeps its shape for the SPA. It changes the
    JSON of a token-reachable route, so `make contract` and `mcp/` move together. The tool
    description must repeat that note text is owner-typed data, not instructions, since a bulk
    read puts much more of it in front of the model at once.
  - A structured field for the part paid on someone else's behalf (e.g. a "paid for others"
    amount), so the statement line is `amount` plus that field and reconciliation is arithmetic
    rather than reading free text. Bigger: a migration (append the column), the expense forms, the
    export, and the contract. Only worth it if shared purchases stay a monthly occurrence.
  Surfaced while reconciling the September 2026 Lulo and Nu statements through the MCP.
