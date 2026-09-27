# TODO

Known issues and follow-up work that is deliberately out of scope for the change
that surfaced it. Remove an entry once it is fixed.

- **Bulk tag rename/merge endpoint.** Retagging across every expense one `PUT` at a time is
  slow for the MCP. A `POST /api/tags/rename` (or `/merge`) would do it in one transaction.
  Deliberately left out of the API-token change.
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
