package tools_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ad9311/ninete-mcp/internal/api"
	"github.com/ad9311/ninete-mcp/internal/contracttest"
	"github.com/ad9311/ninete-mcp/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// fakeAPI stands in for Ninete: routes map "METHOD /path" to a canned JSON
// answer, and every request is recorded so a test can assert on what a tool
// actually sent.
type fakeAPI struct {
	mu       sync.Mutex
	routes   map[string]func(body []byte) (int, string)
	requests []recorded
}

type recorded struct {
	method string
	path   string
	query  url.Values
	body   map[string]any
}

func (f *fakeAPI) handle(w http.ResponseWriter, r *http.Request) {
	data, _ := io.ReadAll(r.Body)

	var body map[string]any
	_ = json.Unmarshal(data, &body)

	f.mu.Lock()
	f.requests = append(f.requests, recorded{r.Method, r.URL.Path, r.URL.Query(), body})
	route := f.routes[r.Method+" "+r.URL.Path]
	f.mu.Unlock()

	if route == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"resource not found"}`))

		return
	}

	status, answer := route(data)
	if status == http.StatusNoContent {
		w.WriteHeader(status)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(answer))
}

func (f *fakeAPI) last(method, path string) recorded {
	f.mu.Lock()
	defer f.mu.Unlock()

	for i := len(f.requests) - 1; i >= 0; i-- {
		if f.requests[i].method == method && f.requests[i].path == path {
			return f.requests[i]
		}
	}

	return recorded{}
}

func (f *fakeAPI) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return len(f.requests)
}

func ok(body string) func([]byte) (int, string) {
	return func([]byte) (int, string) { return http.StatusOK, body }
}

// now is 2026-09-30 20:00 UTC, which is already October in Auckland — so a
// default month resolved in UTC instead of the configured zone shows up.
var now = time.Date(2026, 9, 30, 20, 0, 0, 0, time.UTC) //nolint:gochecknoglobals // fixed test clock

func connect(t *testing.T, routes map[string]func([]byte) (int, string)) (*mcp.ClientSession, *fakeAPI) {
	t.Helper()

	fake := &fakeAPI{routes: routes}
	srv := httptest.NewServer(http.HandlerFunc(fake.handle))
	t.Cleanup(srv.Close)

	// Every route a tool calls must be one whose JSON shapes TestContract
	// checks (mcp/internal/api), or a new call could skip the contract.
	t.Cleanup(func() {
		endpoints := contracttest.Endpoints()

		fake.mu.Lock()
		defer fake.mu.Unlock()

		for _, req := range fake.requests {
			route := contracttest.Route(req.method, req.path)
			_, checked := endpoints[route]
			require.True(t, checked, "a tool called %s, which contracttest.Endpoints does not list", route)
		}
	})

	auckland, err := time.LoadLocation("Pacific/Auckland")
	require.NoError(t, err)

	server := mcp.NewServer(&mcp.Implementation{Name: "ninete", Version: "test"}, nil)
	tools.Register(server, tools.Deps{
		API:      api.New(srv.URL, "nin_tools_test", "test"),
		Location: auckland,
		Now:      func() time.Time { return now },
	})

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	serverSession, err := server.Connect(t.Context(), serverTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = serverSession.Close() })

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "test"}, nil)

	session, err := client.Connect(t.Context(), clientTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Close() })

	return session, fake
}

// call runs a tool and decodes its JSON result into out. It fails the test on
// a tool error unless wantErr is set, in which case it returns the error text.
func call(t *testing.T, session *mcp.ClientSession, name string, args map[string]any, out any) string {
	t.Helper()

	res, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: name, Arguments: args})
	require.NoError(t, err)
	require.NotEmpty(t, res.Content)

	text, isText := res.Content[0].(*mcp.TextContent)
	require.True(t, isText)

	if res.IsError {
		require.Nil(t, out, "tool %s failed: %s", name, text.Text)

		return text.Text
	}

	require.NotNil(t, out, "tool %s unexpectedly succeeded: %s", name, text.Text)
	require.NoError(t, json.Unmarshal([]byte(text.Text), out))

	return ""
}

const expenseJSON = `{"id":7,"category_id":2,"category_name":"Food","description":"Groceries",` +
	`"amount":4599,"date":1788220800,"created_at":1790000000,"tags":["food","home"],"note":"weekly shop"}`

func TestToolList(t *testing.T) {
	session, _ := connect(t, nil)

	res, err := session.ListTools(t.Context(), nil)
	require.NoError(t, err)

	names := make([]string, 0, len(res.Tools))
	for _, tool := range res.Tools {
		names = append(names, tool.Name)
		require.NotContains(t, strings.ToLower(tool.Name), "delete", "no tool may delete")
		require.NotEmpty(t, tool.Description, tool.Name)
	}

	require.ElementsMatch(t, []string{
		"search_expenses", "get_expense", "create_expense", "quick_add_expense", "update_expense",
		"list_recurrent_expenses", "get_recurrent_expense", "create_recurrent_expense",
		"update_recurrent_expense", "unarchive_recurrent_expense",
		"list_categories", "list_tags", "get_dashboard", "get_expense_stats",
		"get_budgets", "set_budgets", "get_report_settings", "set_report_settings",
	}, names)
}

func TestExpenseTools(t *testing.T) {
	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_present_amounts_months_and_instants_readably",
			fn: func(t *testing.T) {
				session, _ := connect(t, map[string]func([]byte) (int, string){
					"GET /api/expenses/7": ok(expenseJSON),
				})

				var out tools.Expense
				call(t, session, "get_expense", map[string]any{"id": 7}, &out)

				require.Equal(t, "45.99", out.Amount)
				require.Equal(t, "2026-09", out.BilledMonth)
				require.Equal(t, "2026-09-22T02:13:20+12:00", out.CreatedAt)
				require.NotNil(t, out.Note)
				require.Equal(t, "weekly shop", *out.Note)
			},
		},
		{
			name: "should_search_by_billed_month_range",
			fn: func(t *testing.T) {
				session, fake := connect(t, map[string]func([]byte) (int, string){
					"GET /api/expenses": ok(`{"data":[` + expenseJSON + `],"pagination":{"current_page":1,` +
						`"total_pages":1,"total_count":1,"has_next":false}}`),
				})

				var out tools.ExpenseListOutput
				call(t, session, "search_expenses", map[string]any{
					"query": "groc", "from_month": "2026-08", "to_month": "2026-09", "sort_order": "asc",
				}, &out)

				req := fake.last(http.MethodGet, "/api/expenses")
				require.Equal(t, "groc", req.query.Get("q"))
				require.Equal(t, "1785542400", req.query.Get("start")) // 2026-08-01 UTC
				require.Equal(t, "1790812800", req.query.Get("end"))   // 2026-10-01 UTC
				require.Equal(t, "ASC", req.query.Get("sort_order"))
				require.Equal(t, "created_at", req.query.Get("sort_field"))
				require.Equal(t, "25", req.query.Get("per_page"))
				require.Len(t, out.Expenses, 1)
				require.Nil(t, out.Expenses[0].Note, "list results carry no note")
			},
		},
		{
			name: "should_default_the_sort_order_when_only_a_field_is_given",
			fn: func(t *testing.T) {
				session, fake := connect(t, map[string]func([]byte) (int, string){
					"GET /api/expenses": ok(`{"data":[],"pagination":{}}`),
				})

				var out tools.ExpenseListOutput
				call(t, session, "search_expenses", map[string]any{"sort_field": "date"}, &out)

				// The server rejects a field without an order, so the tool
				// must send the documented default itself.
				req := fake.last(http.MethodGet, "/api/expenses")
				require.Equal(t, "date", req.query.Get("sort_field"))
				require.Equal(t, "DESC", req.query.Get("sort_order"))
			},
		},
		{
			name: "should_search_by_creation_day_in_the_configured_zone",
			fn: func(t *testing.T) {
				session, fake := connect(t, map[string]func([]byte) (int, string){
					"GET /api/expenses": ok(`{"data":[],"pagination":{}}`),
				})

				var out tools.ExpenseListOutput
				call(t, session, "search_expenses", map[string]any{"created_from": "2026-09-26"}, &out)

				req := fake.last(http.MethodGet, "/api/expenses")
				// Auckland midnight is the previous day 12:00 UTC.
				require.Equal(t, "1790337600", req.query.Get("created_start"))
				require.Equal(t, "1790424000", req.query.Get("created_end"))
				require.Empty(t, req.query.Get("start"))
			},
		},
		{
			name: "should_refuse_to_mix_the_two_date_filters",
			fn: func(t *testing.T) {
				session, fake := connect(t, nil)

				msg := call(t, session, "search_expenses", map[string]any{
					"from_month": "2026-09", "created_from": "2026-09-01",
				}, nil)
				require.Contains(t, msg, "not both")
				require.Zero(t, fake.count(), "nothing may be sent for a rejected call")
			},
		},
		{
			name: "should_create_with_cents_and_the_zones_current_month",
			fn: func(t *testing.T) {
				session, fake := connect(t, map[string]func([]byte) (int, string){
					"POST /api/expenses": ok(expenseJSON),
				})

				var out tools.Expense
				call(t, session, "create_expense", map[string]any{
					"category_id": 2, "description": "Groceries", "amount": "45.99",
				}, &out)

				req := fake.last(http.MethodPost, "/api/expenses")
				require.InDelta(t, 4599, req.body["amount"], 0)
				// October in Auckland, although it is still September in UTC.
				require.InDelta(t, 1790812800, req.body["date"], 0)
				require.Equal(t, []any{}, req.body["tags"])
			},
		},
		{
			name: "should_reject_a_malformed_amount_before_calling",
			fn: func(t *testing.T) {
				session, fake := connect(t, nil)

				msg := call(t, session, "create_expense", map[string]any{
					"category_id": 2, "description": "Groceries", "amount": "45,99",
				}, nil)
				require.Contains(t, msg, "amount")
				require.Zero(t, fake.count())
			},
		},
		{
			name: "should_merge_an_update_onto_the_current_expense",
			fn: func(t *testing.T) {
				session, fake := connect(t, map[string]func([]byte) (int, string){
					"GET /api/expenses/7": ok(expenseJSON),
					"PUT /api/expenses/7": ok(expenseJSON),
				})

				var out tools.Expense
				call(t, session, "update_expense", map[string]any{
					"id": 7, "amount": "50", "add_tags": []string{"Weekly"}, "remove_tags": []string{"HOME"},
				}, &out)

				body := fake.last(http.MethodPut, "/api/expenses/7").body
				require.InDelta(t, 5000, body["amount"], 0)
				require.Equal(t, "Groceries", body["description"], "unchanged fields are kept")
				require.Equal(t, "weekly shop", body["note"], "the note survives an update")
				require.InDelta(t, 1788220800, body["date"], 0)
				require.Equal(t, []any{"food", "weekly"}, body["tags"])
			},
		},
		{
			name: "should_refuse_replace_and_adjust_tags_together",
			fn: func(t *testing.T) {
				session, _ := connect(t, map[string]func([]byte) (int, string){
					"GET /api/expenses/7": ok(expenseJSON),
				})

				msg := call(t, session, "update_expense", map[string]any{
					"id": 7, "tags": []string{"a"}, "add_tags": []string{"b"},
				}, nil)
				require.Contains(t, msg, "not both")
			},
		},
		{
			name: "should_send_quick_add_the_zones_offset_and_explain_a_missing_category",
			fn: func(t *testing.T) {
				session, fake := connect(t, map[string]func([]byte) (int, string){
					"POST /api/expenses/quick": func([]byte) (int, string) {
						return http.StatusUnprocessableEntity,
							`{"error":"category required for this description","fields":{"category_id":"required"}}`
					},
				})

				msg := call(t, session, "quick_add_expense", map[string]any{"input": "Coffee, 4.50, current"}, nil)
				require.Contains(t, msg, "category_id")

				body := fake.last(http.MethodPost, "/api/expenses/quick").body
				require.InDelta(t, -780, body["tz_offset"], 0, "Auckland is UTC+13 in October")
				require.Equal(t, "Coffee, 4.50, current", body["quick_input"])
			},
		},
		{
			name: "should_explain_a_rejected_token",
			fn: func(t *testing.T) {
				session, _ := connect(t, map[string]func([]byte) (int, string){
					"GET /api/expenses/7": func([]byte) (int, string) {
						return http.StatusUnauthorized, `{"error":"authentication required"}`
					},
				})

				msg := call(t, session, "get_expense", map[string]any{"id": 7}, nil)
				require.Contains(t, msg, "revoked")
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}

func TestRecurrentAndReferenceTools(t *testing.T) {
	const recurrentJSON = `{"id":3,"category_id":2,"category_name":"Housing","description":"Rent",` +
		`"amount":120000,"period":1,"occurrence_limit":12,"occurrence_count":4,"archived":false,"tags":["home"]}`

	const settingsJSON = `{"selected_tag_ids":[1],"tags":[{"id":1,"name":"food"},{"id":2,"name":"home"}],"tag_limit":20}`

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "should_merge_a_recurrent_expense_update",
			fn: func(t *testing.T) {
				session, fake := connect(t, map[string]func([]byte) (int, string){
					"GET /api/recurrent-expenses/3": ok(recurrentJSON),
					"PUT /api/recurrent-expenses/3": ok(recurrentJSON),
				})

				var out tools.RecurrentExpense
				call(t, session, "update_recurrent_expense", map[string]any{"id": 3, "occurrence_limit": 24}, &out)

				body := fake.last(http.MethodPut, "/api/recurrent-expenses/3").body
				require.InDelta(t, 24, body["occurrence_limit"], 0)
				require.InDelta(t, 120000, body["amount"], 0)
				require.InDelta(t, 1, body["period"], 0)
				require.Equal(t, []any{"home"}, body["tags"])
				require.Equal(t, "1200.00", out.Amount)
			},
		},
		{
			name: "should_resolve_report_tags_by_name",
			fn: func(t *testing.T) {
				session, fake := connect(t, map[string]func([]byte) (int, string){
					"GET /api/report-settings": ok(settingsJSON),
					"PUT /api/report-settings": func([]byte) (int, string) { return http.StatusNoContent, "" },
				})

				var out tools.ReportSettingsOutput
				call(t, session, "set_report_settings", map[string]any{"tags": []string{"Home", "food"}}, &out)

				body := fake.last(http.MethodPut, "/api/report-settings").body
				require.Equal(t, []any{float64(2), float64(1)}, body["tag_ids"])
			},
		},
		{
			name: "should_refuse_an_unknown_report_tag",
			fn: func(t *testing.T) {
				session, fake := connect(t, map[string]func([]byte) (int, string){
					"GET /api/report-settings": ok(settingsJSON),
				})

				msg := call(t, session, "set_report_settings", map[string]any{"tags": []string{"travel"}}, nil)
				require.Contains(t, msg, "travel")
				require.Empty(t, fake.last(http.MethodPut, "/api/report-settings").method, "nothing may be saved")
			},
		},
		{
			name: "should_send_only_the_listed_budgets",
			fn: func(t *testing.T) {
				session, fake := connect(t, map[string]func([]byte) (int, string){
					"PUT /api/expenses/budgets": func([]byte) (int, string) { return http.StatusNoContent, "" },
					"GET /api/expenses/budgets": ok(`{"mode":"month","rows":[],"edit_rows":[` +
						`{"category_id":2,"name":"Food","amount":30000}]}`),
				})

				var out tools.BudgetsOutput
				call(t, session, "set_budgets", map[string]any{
					"budgets": []map[string]any{{"category_id": 2, "amount": "300"}},
				}, &out)

				body := fake.last(http.MethodPut, "/api/expenses/budgets").body
				require.Equal(t, map[string]any{"2": float64(30000)}, body["amounts"])
				require.Equal(t, "300.00", out.Budgets[0].Amount)
			},
		},
		{
			name: "should_default_the_dashboard_to_the_zones_month",
			fn: func(t *testing.T) {
				session, fake := connect(t, map[string]func([]byte) (int, string){
					"GET /api/dashboard": ok(`{"data":{"this_month_total":1000,"last_month_total":2000,` +
						`"month_change_sign":"-","month_change_pct":50,"top_categories":[]}}`),
				})

				var out tools.DashboardOutput
				call(t, session, "get_dashboard", map[string]any{}, &out)

				require.Equal(t, "2026-10", out.Month)
				require.Equal(t, "2026-09", out.PreviousMonth)
				require.Equal(t, "10.00", out.Total)

				req := fake.last(http.MethodGet, "/api/dashboard")
				require.Equal(t, "1790812800", req.query.Get("this_start"))
				require.Equal(t, "1788220800", req.query.Get("last_start"))
			},
		},
		{
			name: "should_list_tags_from_the_report_settings",
			fn: func(t *testing.T) {
				session, _ := connect(t, map[string]func([]byte) (int, string){
					"GET /api/report-settings": ok(settingsJSON),
				})

				var out tools.TagListOutput
				call(t, session, "list_tags", map[string]any{}, &out)
				require.Equal(t, []string{"food", "home"}, out.Tags)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.fn)
	}
}
