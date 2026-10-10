# TODO

Known issues and follow-up work that is deliberately out of scope for the change
that surfaced it. Remove an entry once it is fixed.

- **A structured amount for the part paid on someone else's behalf.** Reconciling a card statement
  means explaining why an expense is lower than its statement line (a shared purchase where someone
  repaid their part in cash). The MCP now reads notes in bulk through `search_expenses`, but a
  field such as a "paid for others" amount would make the statement line `amount` plus that field,
  so reconciliation is arithmetic rather than reading free text. Bigger: a migration (append the
  column), the expense forms, the export, and the contract. Only worth it if shared purchases stay a
  monthly occurrence. Surfaced while reconciling the September 2026 Lulo and Nu statements through
  the MCP.
