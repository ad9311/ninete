package serve_test

import (
	"bytes"
	"html/template"
	"testing"
	texttemplate "text/template"
	"time"

	"github.com/ad9311/ninete/internal/handlers"
	"github.com/ad9311/ninete/internal/serve"
	"github.com/stretchr/testify/require"
)

func newTestTemplate(t *testing.T, tmplText string) *template.Template {
	t.Helper()

	tmpl, err := template.New("test").Funcs(serve.TemplateFuncMap()).Parse(tmplText)
	require.NoError(t, err)

	return tmpl
}

func renderTemplate(t *testing.T, tmpl *template.Template, data any) string {
	t.Helper()

	var buf bytes.Buffer
	require.NoError(t, tmpl.Execute(&buf, data))

	return buf.String()
}

func newTextTemplate(t *testing.T, tmplText string) *texttemplate.Template {
	t.Helper()

	funcs := texttemplate.FuncMap{}
	for k, v := range serve.TemplateFuncMap() {
		funcs[k] = v
	}

	tmpl, err := texttemplate.New("test").Funcs(funcs).Parse(tmplText)
	require.NoError(t, err)

	return tmpl
}

func renderTextTemplate(t *testing.T, tmpl *texttemplate.Template, data any) string {
	t.Helper()

	var buf bytes.Buffer
	require.NoError(t, tmpl.Execute(&buf, data))

	return buf.String()
}

func TestCurrency(t *testing.T) {
	tmpl := newTestTemplate(t, `{{ currency . }}`)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"zero", func(t *testing.T) {
			require.Equal(t, "$0.00", renderTemplate(t, tmpl, uint64(0)))
		}},
		{"small value no separator", func(t *testing.T) {
			require.Equal(t, "$9.99", renderTemplate(t, tmpl, uint64(999)))
		}},
		{"hundreds no separator", func(t *testing.T) {
			require.Equal(t, "$99.99", renderTemplate(t, tmpl, uint64(9999)))
		}},
		{"thousands with separator", func(t *testing.T) {
			require.Equal(t, "$1,000.00", renderTemplate(t, tmpl, uint64(100000)))
		}},
		{"tens of thousands", func(t *testing.T) {
			require.Equal(t, "$12,345.67", renderTemplate(t, tmpl, uint64(1234567)))
		}},
		{"millions", func(t *testing.T) {
			require.Equal(t, "$1,234,567.89", renderTemplate(t, tmpl, uint64(123456789)))
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestSumAmount(t *testing.T) {
	type row struct{ Amount uint64 }

	tmpl := newTestTemplate(t, `{{ . | sumAmount | currency }}`)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"sums and formats multiple rows", func(t *testing.T) {
			data := []row{{Amount: 100000}, {Amount: 50000}, {Amount: 25050}}
			require.Equal(t, "$1,750.50", renderTemplate(t, tmpl, data))
		}},
		{"empty slice", func(t *testing.T) {
			require.Equal(t, "$0.00", renderTemplate(t, tmpl, []row{}))
		}},
		{"nil input", func(t *testing.T) {
			require.Equal(t, "$0.00", renderTemplate(t, tmpl, nil))
		}},
		{"non-slice input", func(t *testing.T) {
			require.Equal(t, "$0.00", renderTemplate(t, tmpl, "not a slice"))
		}},
		{"pointer rows with nil", func(t *testing.T) {
			r := &row{Amount: 5000}
			data := []*row{r, nil, {Amount: 3000}}
			require.Equal(t, "$80.00", renderTemplate(t, tmpl, data))
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestTimeStamp(t *testing.T) {
	tmpl := newTestTemplate(t, `{{ timeStamp . }}`)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"formats unix timestamp", func(t *testing.T) {
			ts := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC).Unix()
			result := renderTemplate(t, tmpl, ts)
			require.Contains(t, result, "2025-06-15")
		}},
		{"epoch timestamp", func(t *testing.T) {
			result := renderTemplate(t, tmpl, int64(0))
			expected := time.Unix(0, 0).UTC().Format(time.DateOnly)
			require.Equal(t, expected, result)
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestSortURL(t *testing.T) {
	tmpl := newTextTemplate(t, `{{ sortURL .basePath .field .pg }}`)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"defaults to ASC for new field", func(t *testing.T) {
			data := map[string]any{
				"basePath": "/expenses",
				"field":    "amount",
				"pg":       handlers.PaginationData{SortField: "date", SortOrder: "ASC", PerPage: 10},
			}
			result := renderTextTemplate(t, tmpl, data)
			require.Equal(t, "/expenses?sort_field=amount&sort_order=ASC&per_page=10&page=1", result)
		}},
		{"toggles to DESC when same field is ASC", func(t *testing.T) {
			data := map[string]any{
				"basePath": "/expenses",
				"field":    "date",
				"pg":       handlers.PaginationData{SortField: "date", SortOrder: "ASC", PerPage: 10},
			}
			result := renderTextTemplate(t, tmpl, data)
			require.Equal(t, "/expenses?sort_field=date&sort_order=DESC&per_page=10&page=1", result)
		}},
		{"toggles to ASC when same field is DESC", func(t *testing.T) {
			data := map[string]any{
				"basePath": "/expenses",
				"field":    "date",
				"pg":       handlers.PaginationData{SortField: "date", SortOrder: "DESC", PerPage: 10},
			}
			result := renderTextTemplate(t, tmpl, data)
			require.Equal(t, "/expenses?sort_field=date&sort_order=ASC&per_page=10&page=1", result)
		}},
		{"preserves filter params", func(t *testing.T) {
			data := map[string]any{
				"basePath": "/expenses",
				"field":    "amount",
				"pg": handlers.PaginationData{
					SortField:  "date",
					SortOrder:  "ASC",
					PerPage:    10,
					CategoryID: 5,
					DateRange:  "this_month",
					Done:       "true",
					Priority:   2,
				},
			}
			result := renderTextTemplate(t, tmpl, data)
			require.Contains(t, result, "&category_id=5")
			require.Contains(t, result, "&date_range=this_month")
			require.Contains(t, result, "&done=true")
			require.Contains(t, result, "&priority=2")
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestPageURL(t *testing.T) {
	tmpl := newTextTemplate(t, `{{ pageURL .basePath .page .pg }}`)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"builds page URL with sort params", func(t *testing.T) {
			data := map[string]any{
				"basePath": "/expenses",
				"page":     3,
				"pg":       handlers.PaginationData{SortField: "date", SortOrder: "DESC", PerPage: 10},
			}
			result := renderTextTemplate(t, tmpl, data)
			require.Equal(t, "/expenses?sort_field=date&sort_order=DESC&per_page=10&page=3", result)
		}},
		{"preserves filter params", func(t *testing.T) {
			data := map[string]any{
				"basePath": "/lists/1",
				"page":     2,
				"pg": handlers.PaginationData{
					SortField: "created_at",
					SortOrder: "ASC",
					PerPage:   10,
					Done:      "false",
					Priority:  3,
				},
			}
			result := renderTextTemplate(t, tmpl, data)
			require.Contains(t, result, "page=2")
			require.Contains(t, result, "&done=false")
			require.Contains(t, result, "&priority=3")
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestPageRange(t *testing.T) {
	tmpl := newTestTemplate(t, `{{ range pageRange .total .current }}{{ . }} {{ end }}`)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"zero pages returns empty", func(t *testing.T) {
			data := map[string]any{"total": 0, "current": 1}
			require.Equal(t, "", renderTemplate(t, tmpl, data))
		}},
		{"single page", func(t *testing.T) {
			data := map[string]any{"total": 1, "current": 1}
			require.Equal(t, "1 ", renderTemplate(t, tmpl, data))
		}},
		{"five pages at start", func(t *testing.T) {
			data := map[string]any{"total": 10, "current": 1}
			require.Equal(t, "1 2 3 4 5 ", renderTemplate(t, tmpl, data))
		}},
		{"five pages in middle", func(t *testing.T) {
			data := map[string]any{"total": 10, "current": 5}
			require.Equal(t, "3 4 5 6 7 ", renderTemplate(t, tmpl, data))
		}},
		{"five pages at end", func(t *testing.T) {
			data := map[string]any{"total": 10, "current": 10}
			require.Equal(t, "6 7 8 9 10 ", renderTemplate(t, tmpl, data))
		}},
		{"fewer than five total pages", func(t *testing.T) {
			data := map[string]any{"total": 3, "current": 2}
			require.Equal(t, "1 2 3 ", renderTemplate(t, tmpl, data))
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestFilterURL(t *testing.T) {
	tmpl := newTextTemplate(t, `{{ filterURL .basePath .pg .key .value }}`)

	basePg := handlers.PaginationData{
		SortField:  "date",
		SortOrder:  "DESC",
		PerPage:    10,
		CategoryID: 2,
		DateRange:  "this_month",
		Done:       "true",
		Priority:   1,
	}

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"overrides category_id", func(t *testing.T) {
			data := map[string]any{
				"basePath": "/expenses", "pg": basePg,
				"key": "category_id", "value": "7",
			}
			result := renderTextTemplate(t, tmpl, data)
			require.Contains(t, result, "&category_id=7")
			require.NotContains(t, result, "&category_id=2")
		}},
		{"overrides date_range", func(t *testing.T) {
			data := map[string]any{
				"basePath": "/expenses", "pg": basePg,
				"key": "date_range", "value": "this_year",
			}
			result := renderTextTemplate(t, tmpl, data)
			require.Contains(t, result, "&date_range=this_year")
			require.NotContains(t, result, "&date_range=this_month")
		}},
		{"overrides done", func(t *testing.T) {
			data := map[string]any{
				"basePath": "/lists/1", "pg": basePg,
				"key": "done", "value": "false",
			}
			result := renderTextTemplate(t, tmpl, data)
			require.Contains(t, result, "&done=false")
			require.NotContains(t, result, "&done=true")
		}},
		{"overrides priority", func(t *testing.T) {
			data := map[string]any{
				"basePath": "/lists/1", "pg": basePg,
				"key": "priority", "value": "3",
			}
			result := renderTextTemplate(t, tmpl, data)
			require.Contains(t, result, "&priority=3")
			require.NotContains(t, result, "&priority=1")
		}},
		{"clears category_id when zero", func(t *testing.T) {
			data := map[string]any{
				"basePath": "/expenses", "pg": basePg,
				"key": "category_id", "value": "0",
			}
			result := renderTextTemplate(t, tmpl, data)
			require.NotContains(t, result, "category_id")
		}},
		{"resets page to 1", func(t *testing.T) {
			data := map[string]any{
				"basePath": "/expenses", "pg": basePg,
				"key": "category_id", "value": "3",
			}
			result := renderTextTemplate(t, tmpl, data)
			require.Contains(t, result, "page=1")
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestDateRangeOptions(t *testing.T) {
	tmpl := newTestTemplate(t, `{{ range dateRangeOptions }}{{ .Value }} {{ end }}`)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"returns all options", func(t *testing.T) {
			result := renderTemplate(t, tmpl, nil)
			require.Contains(t, result, "this_month")
			require.Contains(t, result, "next_month")
			require.Contains(t, result, "last_month")
			require.Contains(t, result, "this_week")
			require.Contains(t, result, "six_months")
			require.Contains(t, result, "this_year")
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestAdd(t *testing.T) {
	tmpl := newTestTemplate(t, `{{ add .a .b }}`)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"positive numbers", func(t *testing.T) {
			require.Equal(t, "5", renderTemplate(t, tmpl, map[string]any{"a": 2, "b": 3}))
		}},
		{"with zero", func(t *testing.T) {
			require.Equal(t, "7", renderTemplate(t, tmpl, map[string]any{"a": 7, "b": 0}))
		}},
		{"negative numbers", func(t *testing.T) {
			require.Equal(t, "-1", renderTemplate(t, tmpl, map[string]any{"a": 2, "b": -3}))
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}

func TestSub(t *testing.T) {
	tmpl := newTestTemplate(t, `{{ sub .a .b }}`)

	cases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"positive result", func(t *testing.T) {
			require.Equal(t, "2", renderTemplate(t, tmpl, map[string]any{"a": 5, "b": 3}))
		}},
		{"negative result", func(t *testing.T) {
			require.Equal(t, "-3", renderTemplate(t, tmpl, map[string]any{"a": 2, "b": 5}))
		}},
		{"zero result", func(t *testing.T) {
			require.Equal(t, "0", renderTemplate(t, tmpl, map[string]any{"a": 4, "b": 4}))
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.fn)
	}
}
