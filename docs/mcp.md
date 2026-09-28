# MCP server

`mcp/` holds `ninete-mcp`, a local [MCP](https://modelcontextprotocol.io) server that lets an MCP
client (Claude Code, Claude Desktop) work with a Ninete account: search and read expenses, create
and edit them, fix their tags, manage recurrent expenses, budgets and the monthly report's grouping
tags. **It cannot delete anything.**

It runs on your machine, speaks MCP over stdio, and reaches the app only through `/api/*` with a
personal access token — the same checks the web UI goes through. The token side is documented in
`docs/architecture.md` ("Bearer tokens").

## Why it is shaped this way

- **A separate Go module, named outside the app's import path.** `mcp/go.mod` declares
  `github.com/ad9311/ninete-mcp`, not `github.com/ad9311/ninete/mcp`. Go's `internal` rule goes by
  import path, so a module under `github.com/ad9311/ninete/` could import `internal/logic` given a
  `require` and a `replace`; one outside it cannot, even then — the compiler answers "use of
  internal package … not allowed". Keep the module path as it is.
- **HTTP only.** The binary has no database and no server secrets; its one credential is the token.
  Every action therefore passes the same auth, `user_id` scoping, validation and token-scope checks
  as the UI. A shortcut into `internal/*` would skip all of them.
- **Its own copies of the API's JSON shapes** (`mcp/internal/api/types.go`). A change to a
  handler's JSON has to be mirrored here, and `contract/api.json` is what makes a missed one fail —
  see "Keeping the API contract" below.
- **No delete, twice over.** The server refuses `DELETE` to every token scope, the API client has no
  `Delete` method, and no tool is named for deleting. `TestToolList` fails if one appears.

## Building and checking

`mcp/` is outside the root module, so the root `go test ./...` and `golangci-lint run` never enter
it. Use the Makefile targets, which CI mirrors (`mcp-tests` in `tests.yml`, the second golangci step
in `linters.yml`):

| Target | What it does |
| --- | --- |
| `make build-mcp` | Builds `./build/ninete-mcp`, stamping `main.version` from `git describe` |
| `make test-mcp` | Runs the module's tests. No database, bundle or `.env` needed |
| `make lint-mcp` | golangci-lint over the module, with the repository's config. `make lint` and `make lint-fix` run it too |

Dependencies are managed inside `mcp/` (`cd mcp && go mod tidy`); the app's `go.mod` does not know
about the MCP SDK.

## Keeping the API contract

The app and this module cannot import each other's types, so they meet in a file:
`contract/api.json` lists, for every route a token can reach, the query keys it reads and the JSON
fields of its request body and response, flattened to `"path": "type"`
(`"data[].amount": "integer"`). It is generated from the handler structs and committed.

Three tests hold it in place:

| Test | Side | Fails when |
| --- | --- | --- |
| `TestAPIContract` (`internal/handlers/api_contract_internal_test.go`) | app | A handler's JSON changed and the file was not regenerated |
| `TestAPIContractCoversTokenRoutes` (`internal/serve/contract_internal_test.go`) | app | A token-reachable, non-`DELETE` route is missing from the file, or the file lists one the router no longer serves |
| `TestContract` (`mcp/internal/api/contract_test.go`) | mcp | This module's types disagree with the file |

`TestContract` is strict in one direction only:

- **Request bodies must match exactly.** The server's decoder ignores unknown fields and a `PUT`
  replaces the whole record, so a field the server reads but the MCP does not send would be reset
  to empty on every update — silently. That is the drift that loses data.
- **Responses may be a subset.** The MCP may ignore a field, but every field it reads must exist
  with the same type. This is why the list and detail expense are two types (`ListedExpense`,
  `Expense`): the list never sends `note`.

The tools test adds two more guards, both run over every request a tool sends to the fake API:
the route must be in `contracttest.Endpoints`, so a new call cannot bypass `TestContract`, and
every query key must be one the file lists for that route (`contracttest.UnknownQueryKeys`).

The workflow after changing a token-reachable handler's JSON:

```sh
make contract     # regenerate contract/api.json; review its diff
make test-mcp     # fails until mcp/internal/api/types.go (and the tools) follow
```

Both sides walk types with the same rules (json tags, `-` skipped, untagged embedded structs
inlined, `int`/`uint` both `integer`). The two walkers are copies —
`internal/handlers/api_contract_internal_test.go` and `mcp/internal/contracttest` — and the file
keeps them honest: if they disagreed, one side's test would fail on the next run.

### Query keys

A handler reads its query string only through `decodeQuery` (`internal/handlers/query.go`), which
fills a struct of raw strings tagged `query:"…"`: `apiExpenseListQuery`, `apiDashboardQuery` and
so on. `apiContract` names that struct for each route, and `TestAPIContract` writes its keys into
the file as a sorted `"query"` list. The tags are therefore the only declaration of a route's keys,
and renaming one changes the file by itself. Two things keep it that way:

- **`forbidigo` rejects any other read.** `r.URL.Query()`, `RawQuery`, `FormValue`, `ParseForm` and
  `Form` are forbidden in `internal/handlers` outside `query.go` and tests (`.golangci.yml`). A
  `q.Get("…")` literal would never reach the contract.
- **`TestAPIContract` rejects a tagged field decodeQuery would skip.** One that is not a string, or a
  key declared twice across embedded structs, fails the test rather than silently not decoding.

On the MCP side keys are a subset, like responses. Every key is optional to the server, but one it
does not read is ignored, and the tool gets an unfiltered answer instead of an error. That is the
drift this catches.

The check sees only keys a tools test actually sends. When a tool gains a query parameter, add it
to that tool's "sends every key" test (`should_send_the_tag_category_and_page_filters`,
`should_send_every_recurrent_expense_list_filter`, …), or a rename of it goes unnoticed.

**Values are not covered.** The file records `mode`, not that it accepts `month` or `months`. The
same goes for `sort_field` and `sort_order` values and the `archived` format. A server that renamed
an accepted value would fall back to its default without the MCP knowing. `TODO.md` tracks it.

## Configuration

The MCP client launches the binary and passes three environment variables:

| Variable | Required | Meaning |
| --- | --- | --- |
| `NINETE_URL` | yes | The app's origin, e.g. `https://ninete.example.com`. Must be `https`; plain `http` is accepted only for `localhost`/loopback |
| `NINETE_TOKEN` | yes | A token from `/account/tokens` (`nin_…`). Use a `read` token for a look-only assistant, `write` to let it create and edit |
| `NINETE_TZ` | no | IANA zone, e.g. `America/Bogota`. Decides what "this month" and "today" mean and which instants a created-on day covers — the role the browser's zone plays for the web app. Defaults to the machine's local zone |

### Claude Code

```sh
make build-mcp
claude mcp add ninete --scope user \
  -e NINETE_URL=https://ninete.example.com \
  -e NINETE_TOKEN=nin_… \
  -e NINETE_TZ=America/Bogota \
  -- /absolute/path/to/ninete/build/ninete-mcp
```

The token ends up in plain text in Claude Code's config (`~/.claude.json` for `--scope user`). Never
use `--scope project`, which writes `.mcp.json` into the repository — this repository is public.

### Claude Desktop

In `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "ninete": {
      "command": "/absolute/path/to/ninete/build/ninete-mcp",
      "env": {
        "NINETE_URL": "https://ninete.example.com",
        "NINETE_TOKEN": "nin_…",
        "NINETE_TZ": "America/Bogota"
      }
    }
  }
}
```

If the token leaks, revoke it at `/account/tokens`; it stops working on the next request.

## Tools

Amounts are decimal strings in the account's currency (`"12.50"`), converted to and from the API's
integer cents without floats. An expense's `billed_month` is `YYYY-MM` — the month it counts toward,
sent to the API as UTC midnight of the 1st (the API stores the epoch as given), so it needs no zone.

| Tool | Scope | Notes |
| --- | --- | --- |
| `search_expenses` | read | Text, tag, category; billed-month range **or** created-day range, not both (the server would silently drop the first); pages of 15/25/50/100. `sort_field=date` sorts by billed month with ties in id order — an old row carrying a mid-month day sorts after its month's other rows ascending and before them descending (see `CLAUDE.md` on the billed date). A created-day range may span months, e.g. a card cycle Aug 22 – Sep 21 |
| `get_expense` | read | Includes the note, which list results never carry |
| `create_expense` | write | `billed_month` defaults to the current month in `NINETE_TZ` |
| `quick_add_expense` | write | The app's one-line quick-add. Sends `tz_offset` from `NINETE_TZ`. When the description has no remembered category, fails asking for `category_id` |
| `update_expense` | write | Partial: reads the expense, applies only the fields given, `PUT`s the whole thing. Tags via `add_tags`/`remove_tags`, or `tags` to replace |
| `list_recurrent_expenses` | read | Active by default, `archived=true` for the rest |
| `get_recurrent_expense` | read | |
| `create_recurrent_expense` | write | `period_months`, optional `occurrence_limit` |
| `update_recurrent_expense` | write | Partial, same pattern as `update_expense` |
| `unarchive_recurrent_expense` | write | |
| `list_categories` | read | Categories are global and fixed |
| `list_tags` | read | `GET /api/tags`: every tag with its expense and recurrent-expense counts; both at zero means unused |
| `get_dashboard` | read | One month against the previous one; defaults to the current month |
| `get_expense_stats` | read | Totals per category over a month range, or all time |
| `get_budgets` | read | Spending against budgets over a month range, `month` or `months` mode |
| `set_budgets` | write | Only the categories listed change; `0` removes a budget |
| `get_report_settings` | read | The monthly report's grouping tags, by name |
| `set_report_settings` | write | Takes tag names and resolves them to ids; an unknown name is an error rather than silently dropped |
| `retag` | write | One `POST /api/tags/retag`: moves every expense and recurrent expense from the `from` tags onto `to`, for renames and merges. The `from` tags are kept, unused — the owner deletes them on `/account/tags` |

Text the owner typed (descriptions, notes, tag names) comes back as data. The tool descriptions say
so, so a model reading a note that looks like an instruction treats it as content.

## Adding a tool

1. Check the route is on `tokenAPIPrefixes` (`internal/serve/middleware.go`). A route that is not
   answers `403` to every token — that allowlist is deliberate, see `CLAUDE.md`. If you add it
   there, add it to `apiContract` in `internal/handlers/api_contract_internal_test.go` and run
   `make contract`.
2. Copy the JSON shape into `mcp/internal/api/types.go` and list the route in
   `mcp/internal/contracttest.Endpoints`.
3. Add the tool in `mcp/internal/tools/`, converting amounts and dates through
   `mcp/internal/units`, and mark it with `readOnly`, `creates` or `overwrites`.
4. Add it to the list in `TestToolList` and give it a case against the fake API.
