# TODO

Known issues and follow-up work that is deliberately out of scope for the change
that surfaced it. Remove an entry once it is fixed.

- **MCP server (`mcp/`, separate Go module).** Phase 2 of the API-token work: a local stdio
  MCP binary that calls `/api/*` with a bearer token. Tools cover everything the UI can
  create, edit or read except deletion. Its module path must sit outside
  `github.com/ad9311/ninete/` so the compiler refuses any `internal/*` import.
- **Bulk tag rename/merge endpoint.** Retagging across every expense one `PUT` at a time is
  slow for the MCP. A `POST /api/tags/rename` (or `/merge`) would do it in one transaction.
  Deliberately left out of the API-token change.
