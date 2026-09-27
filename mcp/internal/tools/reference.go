package tools

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/ad9311/ninete-mcp/internal/api"
	"github.com/ad9311/ninete-mcp/internal/units"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type NoInput struct{}

type CategoryListOutput struct {
	Categories []api.Category `json:"categories"`
}

type TagListOutput struct {
	Tags []string `json:"tags"`
}

type CategoryTotal struct {
	Category string `json:"category"`
	Total    string `json:"total" jsonschema:"decimal amount"`
}

func toCategoryTotals(rows []api.CategoryTotal) []CategoryTotal {
	out := make([]CategoryTotal, 0, len(rows))
	for _, row := range rows {
		out = append(out, CategoryTotal{Category: row.Name, Total: units.FormatAmount(row.Total)})
	}

	return out
}

type DashboardInput struct {
	Month string `json:"month,omitempty" jsonschema:"billed month to summarize, YYYY-MM; defaults to the current month"`
}

type DashboardOutput struct {
	Month         string          `json:"month"`
	Total         string          `json:"total"`
	PreviousMonth string          `json:"previous_month"`
	PreviousTotal string          `json:"previous_total"`
	ChangeSign    string          `json:"change_sign" jsonschema:"direction of the change from the previous month"`
	ChangePct     int             `json:"change_pct" jsonschema:"size of the change from the previous month, in percent"`
	TopCategories []CategoryTotal `json:"top_categories"`
}

type StatsInput struct {
	FromMonth string `json:"from_month,omitempty" jsonschema:"first billed month, YYYY-MM; omit both months for all time"`
	ToMonth   string `json:"to_month,omitempty" jsonschema:"last billed month, YYYY-MM; defaults to from_month"`
}

type StatsOutput struct {
	Categories []CategoryTotal `json:"categories" jsonschema:"spending per category, largest first"`
	Total      string          `json:"total"`
}

type BudgetsInput struct {
	FromMonth string `json:"from_month" jsonschema:"first billed month, YYYY-MM"`
	ToMonth   string `json:"to_month,omitempty" jsonschema:"last billed month, YYYY-MM; defaults to from_month"`
	//nolint:lll // struct tags cannot wrap
	Mode string `json:"mode,omitempty" jsonschema:"month (default) compares the whole range to one month's budget; months breaks it down month by month"`
}

type BudgetMonth struct {
	Month      string `json:"month"`
	Spent      string `json:"spent"`
	PercentUse int    `json:"percent_used"`
	Over       bool   `json:"over"`
}

type BudgetRow struct {
	Category    string        `json:"category"`
	Spent       string        `json:"spent"`
	Budget      *string       `json:"budget,omitempty" jsonschema:"absent when the category has no budget"`
	Left        *string       `json:"left,omitempty" jsonschema:"negative when over budget"`
	PercentUsed int           `json:"percent_used"`
	Over        bool          `json:"over"`
	Months      []BudgetMonth `json:"months,omitempty"`
	MonthsOver  int           `json:"months_over,omitempty"`
	AvgPerMonth string        `json:"avg_per_month,omitempty"`
}

type CategoryBudget struct {
	CategoryID int    `json:"category_id"`
	Category   string `json:"category"`
	Amount     string `json:"amount" jsonschema:"monthly budget; 0.00 means none"`
}

type BudgetsOutput struct {
	Mode    string           `json:"mode"`
	Rows    []BudgetRow      `json:"rows"`
	Budgets []CategoryBudget `json:"budgets" jsonschema:"the monthly budget set for every category"`
}

type SetBudgetsInput struct {
	Budgets []CategoryBudgetInput `json:"budgets" jsonschema:"categories to set; the others are left as they are"`
}

type CategoryBudgetInput struct {
	CategoryID int    `json:"category_id"`
	Amount     string `json:"amount" jsonschema:"monthly budget as a decimal, e.g. 300.00; 0 removes the budget"`
}

type ReportSettingsOutput struct {
	GroupingTags []string `json:"grouping_tags" jsonschema:"the tags the monthly report groups expenses by"`
	TagLimit     int      `json:"tag_limit"`
}

type SetReportSettingsInput struct {
	Tags []string `json:"tags" jsonschema:"tag names to group the monthly report by; an empty list means one flat list"`
}

func registerReferenceTools(server *mcp.Server, d Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_categories",
		Description: "List the expense categories and their ids. Categories are shared and fixed.",
		Annotations: readOnly("List categories"),
	}, d.listCategories)

	mcp.AddTool(server, &mcp.Tool{
		Name: "list_tags",
		Description: "List every tag the account has. Tags are created by adding them to an expense." +
			userDataNote,
		Annotations: readOnly("List tags"),
	}, d.listTags)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_dashboard",
		Description: "Summarize a billed month: its total, the previous month's, and the top categories.",
		Annotations: readOnly("Get dashboard"),
	}, d.getDashboard)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_expense_stats",
		Description: "Total spending per category over a range of billed months, or all time.",
		Annotations: readOnly("Get expense stats"),
	}, d.getExpenseStats)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_budgets",
		Description: "Compare spending with each category's monthly budget over a range of billed months.",
		Annotations: readOnly("Get budgets"),
	}, d.getBudgets)

	mcp.AddTool(server, &mcp.Tool{
		Name: "set_budgets",
		Description: "Set the monthly budget of one or more categories. Categories not listed keep " +
			"their budget; an amount of 0 removes one.",
		Annotations: overwrites("Set budgets"),
	}, d.setBudgets)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_report_settings",
		Description: "Read which tags the monthly expense report groups by.",
		Annotations: readOnly("Get report settings"),
	}, d.getReportSettings)

	mcp.AddTool(server, &mcp.Tool{
		Name: "set_report_settings",
		Description: "Choose the tags the monthly expense report groups by, replacing the current choice. " +
			"Only existing tags can be chosen.",
		Annotations: overwrites("Set report settings"),
	}, d.setReportSettings)
}

func (d Deps) listCategories(
	ctx context.Context, _ *mcp.CallToolRequest, _ NoInput,
) (*mcp.CallToolResult, CategoryListOutput, error) {
	ctx, cancel := callContext(ctx)
	defer cancel()

	var list api.CategoryList
	if err := d.API.Get(ctx, "/categories", nil, &list); err != nil {
		return nil, CategoryListOutput{}, err
	}

	categories := list.Data
	if categories == nil {
		categories = []api.Category{}
	}

	return nil, CategoryListOutput{Categories: categories}, nil
}

// listTags reads the tag list off /api/report-settings: there is no /api/tags
// endpoint, and the settings response is the one place the whole list is sent.
func (d Deps) listTags(
	ctx context.Context, _ *mcp.CallToolRequest, _ NoInput,
) (*mcp.CallToolResult, TagListOutput, error) {
	ctx, cancel := callContext(ctx)
	defer cancel()

	settings, err := d.fetchReportSettings(ctx)
	if err != nil {
		return nil, TagListOutput{}, err
	}

	names := make([]string, 0, len(settings.Tags))
	for _, tag := range settings.Tags {
		names = append(names, tag.Name)
	}

	return nil, TagListOutput{Tags: names}, nil
}

func (d Deps) getDashboard(
	ctx context.Context, _ *mcp.CallToolRequest, in DashboardInput,
) (*mcp.CallToolResult, DashboardOutput, error) {
	var out DashboardOutput

	month := in.Month
	if month == "" {
		month = units.CurrentMonth(d.Now(), d.Location)
	}

	previous, err := units.PreviousMonth(month)
	if err != nil {
		return nil, out, err
	}

	thisStart, thisEnd, err := units.MonthRange(month, "")
	if err != nil {
		return nil, out, err
	}

	lastStart, lastEnd, err := units.MonthRange(previous, "")
	if err != nil {
		return nil, out, err
	}

	query := url.Values{}
	query.Set("this_start", strconv.FormatInt(thisStart, 10))
	query.Set("this_end", strconv.FormatInt(thisEnd, 10))
	query.Set("last_start", strconv.FormatInt(lastStart, 10))
	query.Set("last_end", strconv.FormatInt(lastEnd, 10))

	ctx, cancel := callContext(ctx)
	defer cancel()

	var dashboard api.Dashboard
	if err := d.API.Get(ctx, "/dashboard", query, &dashboard); err != nil {
		return nil, out, err
	}

	data := dashboard.Data

	return nil, DashboardOutput{
		Month:         month,
		Total:         units.FormatAmount(data.ThisMonthTotal),
		PreviousMonth: previous,
		PreviousTotal: units.FormatAmount(data.LastMonthTotal),
		ChangeSign:    data.MonthChangeSign,
		ChangePct:     data.MonthChangePct,
		TopCategories: toCategoryTotals(data.TopCategories),
	}, nil
}

func (d Deps) getExpenseStats(
	ctx context.Context, _ *mcp.CallToolRequest, in StatsInput,
) (*mcp.CallToolResult, StatsOutput, error) {
	var out StatsOutput

	query := url.Values{}

	if in.ToMonth != "" && in.FromMonth == "" {
		return nil, out, ErrToWithoutFrom
	}

	if in.FromMonth != "" {
		start, end, err := units.MonthRange(in.FromMonth, in.ToMonth)
		if err != nil {
			return nil, out, err
		}

		query.Set("start", strconv.FormatInt(start, 10))
		query.Set("end", strconv.FormatInt(end, 10))
	}

	ctx, cancel := callContext(ctx)
	defer cancel()

	var stats api.ExpenseStats
	if err := d.API.Get(ctx, "/expenses/stats", query, &stats); err != nil {
		return nil, out, err
	}

	var total uint64
	for _, row := range stats.Data {
		total += row.Total
	}

	return nil, StatsOutput{Categories: toCategoryTotals(stats.Data), Total: units.FormatAmount(total)}, nil
}

func (d Deps) getBudgets(
	ctx context.Context, _ *mcp.CallToolRequest, in BudgetsInput,
) (*mcp.CallToolResult, BudgetsOutput, error) {
	var out BudgetsOutput

	start, end, err := units.MonthRange(in.FromMonth, in.ToMonth)
	if err != nil {
		return nil, out, err
	}

	mode := in.Mode
	if mode == "" {
		mode = "month"
	}

	if mode != "month" && mode != "months" {
		return nil, out, fmt.Errorf("%w, not %q", ErrBudgetMode, mode)
	}

	query := url.Values{}
	query.Set("start", strconv.FormatInt(start, 10))
	query.Set("end", strconv.FormatInt(end, 10))
	query.Set("mode", mode)

	ctx, cancel := callContext(ctx)
	defer cancel()

	var budgets api.Budgets
	if err := d.API.Get(ctx, "/expenses/budgets", query, &budgets); err != nil {
		return nil, out, err
	}

	out.Mode = budgets.Mode
	out.Rows = make([]BudgetRow, 0, len(budgets.Rows))

	for _, row := range budgets.Rows {
		out.Rows = append(out.Rows, toBudgetRow(row))
	}

	out.Budgets = make([]CategoryBudget, 0, len(budgets.EditRows))
	for _, row := range budgets.EditRows {
		out.Budgets = append(out.Budgets, CategoryBudget{
			CategoryID: row.CategoryID,
			Category:   row.Name,
			Amount:     units.FormatAmount(row.Amount),
		})
	}

	return nil, out, nil
}

func toBudgetRow(row api.BudgetRow) BudgetRow {
	out := BudgetRow{
		Category:    row.CategoryName,
		Spent:       units.FormatAmount(row.Total),
		PercentUsed: row.Pct,
		Over:        row.Over,
		MonthsOver:  row.MonthsOver,
	}

	if row.HasBudget {
		budget := units.FormatAmount(row.Budget)
		left := units.FormatSignedAmount(row.Left)
		out.Budget, out.Left = &budget, &left
	}

	if len(row.Months) > 0 {
		out.AvgPerMonth = units.FormatAmount(row.AvgPerMonth)

		for _, m := range row.Months {
			out.Months = append(out.Months, BudgetMonth{
				Month: m.Month, Spent: units.FormatAmount(m.Total), PercentUse: m.Pct, Over: m.Over,
			})
		}
	}

	return out
}

func (d Deps) setBudgets(
	ctx context.Context, _ *mcp.CallToolRequest, in SetBudgetsInput,
) (*mcp.CallToolResult, BudgetsOutput, error) {
	if len(in.Budgets) == 0 {
		return nil, BudgetsOutput{}, ErrNoBudgets
	}

	// Only the categories listed are sent: the server leaves an omitted
	// category untouched, which is what "the others keep their budget" means.
	amounts := make(map[string]uint64, len(in.Budgets))

	for _, b := range in.Budgets {
		amount, err := units.ParseAmount(b.Amount)
		if err != nil {
			return nil, BudgetsOutput{}, fmt.Errorf("category %d: %w", b.CategoryID, err)
		}

		amounts[strconv.Itoa(b.CategoryID)] = amount
	}

	ctx, cancel := callContext(ctx)
	defer cancel()

	if err := d.API.Put(ctx, "/expenses/budgets", api.BudgetsBody{Amounts: amounts}, nil); err != nil {
		return nil, BudgetsOutput{}, err
	}

	// Read back this month so the caller sees what is now set.
	month := units.CurrentMonth(d.Now(), d.Location)

	return d.getBudgets(ctx, nil, BudgetsInput{FromMonth: month})
}

func (d Deps) fetchReportSettings(ctx context.Context) (api.ReportSettings, error) {
	var settings api.ReportSettings
	err := d.API.Get(ctx, "/report-settings", nil, &settings)

	return settings, err
}

func (d Deps) getReportSettings(
	ctx context.Context, _ *mcp.CallToolRequest, _ NoInput,
) (*mcp.CallToolResult, ReportSettingsOutput, error) {
	ctx, cancel := callContext(ctx)
	defer cancel()

	settings, err := d.fetchReportSettings(ctx)
	if err != nil {
		return nil, ReportSettingsOutput{}, err
	}

	return nil, toReportSettingsOutput(settings), nil
}

func toReportSettingsOutput(settings api.ReportSettings) ReportSettingsOutput {
	nameByID := make(map[int]string, len(settings.Tags))
	for _, tag := range settings.Tags {
		nameByID[tag.ID] = tag.Name
	}

	names := make([]string, 0, len(settings.SelectedTagIDs))
	for _, id := range settings.SelectedTagIDs {
		if name, ok := nameByID[id]; ok {
			names = append(names, name)
		}
	}

	return ReportSettingsOutput{GroupingTags: names, TagLimit: settings.TagLimit}
}

// setReportSettings takes tag names, which is what a person says, and resolves
// them to the ids the API stores. An unknown name is an error rather than
// being dropped: the server would silently ignore it, and the caller would
// believe the report now groups by a tag it does not.
func (d Deps) setReportSettings(
	ctx context.Context, _ *mcp.CallToolRequest, in SetReportSettingsInput,
) (*mcp.CallToolResult, ReportSettingsOutput, error) {
	ctx, cancel := callContext(ctx)
	defer cancel()

	settings, err := d.fetchReportSettings(ctx)
	if err != nil {
		return nil, ReportSettingsOutput{}, err
	}

	idByName := make(map[string]int, len(settings.Tags))
	for _, tag := range settings.Tags {
		idByName[normalizeTag(tag.Name)] = tag.ID
	}

	ids := make([]int, 0, len(in.Tags))

	var unknown []string

	for _, name := range in.Tags {
		id, ok := idByName[normalizeTag(name)]
		if !ok {
			unknown = append(unknown, name)

			continue
		}

		ids = append(ids, id)
	}

	if len(unknown) > 0 {
		return nil, ReportSettingsOutput{}, fmt.Errorf("%w: %s", ErrUnknownTags, strings.Join(unknown, ", "))
	}

	if err := d.API.Put(ctx, "/report-settings", api.ReportSettingsBody{TagIDs: ids}, nil); err != nil {
		return nil, ReportSettingsOutput{}, err
	}

	updated, err := d.fetchReportSettings(ctx)
	if err != nil {
		return nil, ReportSettingsOutput{}, err
	}

	return nil, toReportSettingsOutput(updated), nil
}
