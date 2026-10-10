# TODO

Known issues and follow-up work that is deliberately out of scope for the change
that surfaced it. Remove an entry once it is fixed.

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
