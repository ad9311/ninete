# TODO

Known issues and follow-up work that is deliberately out of scope for the change
that surfaced it. Remove an entry once it is fixed.

- **Bulk tag rename/merge endpoint.** Retagging across every expense one `PUT` at a time is
  slow for the MCP. A `POST /api/tags/rename` (or `/merge`) would do it in one transaction.
  Deliberately left out of the API-token change.
