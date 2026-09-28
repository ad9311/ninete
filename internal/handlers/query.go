package handlers

import (
	"net/http"
	"reflect"
)

// This file is the only place in the package that reads a request's query
// string. Every other file declares the keys it reads as `query:"…"` tags on a
// struct and fills it with decodeQuery, so the tags are the single declaration
// of a route's keys: apiContract lists the struct, TestAPIContract records its
// keys in contract/api.json, and a renamed tag changes that file by itself.
// forbidigo in .golangci.yml rejects r.URL.Query() anywhere else in the
// package, since a stray q.Get("…") literal would bypass the contract.
//
// Fields are raw strings on purpose: decoding copies what arrived and nothing
// more, so each handler keeps its own parsing and fallbacks (a malformed page
// number meaning page 1, an unknown mode meaning "month") exactly as before.

// apiBoundsQuery is the explicit [start, end) billed-date pair the
// /api/expenses* routes take (§3.6 of docs/spa-migration.md).
type apiBoundsQuery struct {
	Start string `query:"start"`
	End   string `query:"end"`
}

// apiListQuery is the sorting, pagination and category filter every listing
// reads through userScopedQueryOpts.
type apiListQuery struct {
	SortField  string `query:"sort_field"`
	SortOrder  string `query:"sort_order"`
	Page       string `query:"page"`
	PerPage    string `query:"per_page"`
	CategoryID string `query:"category_id"`
}

// apiExpenseListQuery is GET /api/expenses.
type apiExpenseListQuery struct {
	apiBoundsQuery
	apiListQuery

	CreatedStart string `query:"created_start"`
	CreatedEnd   string `query:"created_end"`
	Q            string `query:"q"`
	Tag          string `query:"tag"`
}

// apiExpenseStatsQuery is GET /api/expenses/stats.
type apiExpenseStatsQuery struct {
	apiBoundsQuery

	SortField string `query:"sort_field"`
	SortOrder string `query:"sort_order"`
}

// apiExpenseBudgetsQuery is GET /api/expenses/budgets.
type apiExpenseBudgetsQuery struct {
	apiBoundsQuery

	Mode string `query:"mode"`
}

// apiDashboardQuery is GET /api/dashboard: the current and prior month's
// bounds.
type apiDashboardQuery struct {
	ThisStart string `query:"this_start"`
	ThisEnd   string `query:"this_end"`
	LastStart string `query:"last_start"`
	LastEnd   string `query:"last_end"`
}

// apiRecurrentExpenseListQuery is GET /api/recurrent-expenses.
type apiRecurrentExpenseListQuery struct {
	apiListQuery

	Archived string `query:"archived"`
}

// reportQuery is GET /reports/expenses.pdf. It is not token-reachable, so it
// is not in the contract; it goes through decodeQuery only because nothing
// else in the package may read the query string.
type reportQuery struct {
	Month string `query:"month"`
}

// decodeQuery copies each `query:"…"` key of r's query string into the
// matching string field of dst, a pointer to a struct. Untagged embedded
// structs are walked, as encoding/json does; fields without a tag, and tagged
// fields that are not strings, are left alone. TestAPIContract fails on a
// tagged non-string field, so that case cannot ship unnoticed.
func decodeQuery(r *http.Request, dst any) {
	q := r.URL.Query()
	fillQuery(reflect.ValueOf(dst).Elem(), func(key string) string { return q.Get(key) })
}

func fillQuery(v reflect.Value, get func(string) string) {
	for i := range v.NumField() {
		field := v.Type().Field(i)

		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			fillQuery(v.Field(i), get)

			continue
		}

		key := field.Tag.Get("query")
		if key == "" || !field.IsExported() || field.Type.Kind() != reflect.String {
			continue
		}

		v.Field(i).SetString(get(key))
	}
}
