# TODO

Known issues and follow-up work that is deliberately out of scope for the change
that surfaced it. Remove an entry once it is fixed.

- **Put accepted query values under the API contract.** `contract/api.json` lists each route's
  query keys (`docs/mcp.md`, "Query keys"), not the values a key accepts. Several handlers fall back
  silently on a value they do not know: `mode` becomes `month`, an unparseable `archived` becomes
  `false`, and the sort builder handles `sort_field`/`sort_order` its own way. So if the server
  renamed `months`, the MCP would keep sending the old value, get the default back, and no test
  would fail. The fix would record enum values in the contract (an `enum:"month,months"` tag on the
  query struct field, or a constant list it references) and have `contracttest` check the values the
  tools send as well as the keys. Left out of the query-key change to keep that one a pure
  refactor.
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
