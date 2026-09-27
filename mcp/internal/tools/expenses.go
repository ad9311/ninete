package tools

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/ad9311/ninete-mcp/internal/api"
	"github.com/ad9311/ninete-mcp/internal/units"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// perPageChoices are the page sizes the server accepts; anything else it
// silently replaces with 15, so the tool rejects it instead.
var perPageChoices = []int{15, 25, 50, 100} //nolint:gochecknoglobals // static option list

const defaultPerPage = 25

type SearchExpensesInput struct {
	Query      string `json:"query,omitempty" jsonschema:"words to find in the description, at most 50 characters"`
	Tag        string `json:"tag,omitempty" jsonschema:"only expenses carrying this tag"`
	CategoryID int    `json:"category_id,omitempty" jsonschema:"only this category; see list_categories"`
	FromMonth  string `json:"from_month,omitempty" jsonschema:"first billed month to include, YYYY-MM"`
	ToMonth    string `json:"to_month,omitempty" jsonschema:"last billed month to include, YYYY-MM (default from_month)"`
	//nolint:lll // struct tags cannot wrap
	CreatedFrom string `json:"created_from,omitempty" jsonschema:"first day the expense was recorded, YYYY-MM-DD in the configured zone"`
	CreatedTo   string `json:"created_to,omitempty" jsonschema:"last creation day, YYYY-MM-DD (default created_from)"`
	Page        int    `json:"page,omitempty" jsonschema:"page number, from 1"`
	PerPage     int    `json:"per_page,omitempty" jsonschema:"results per page: 15, 25, 50 or 100 (default 25)"`
	SortField   string `json:"sort_field,omitempty" jsonschema:"created_at (default), date, amount or description"`
	SortOrder   string `json:"sort_order,omitempty" jsonschema:"asc or desc (default desc)"`
}

type ExpenseListOutput struct {
	Expenses   []Expense `json:"expenses"`
	Page       int       `json:"page"`
	TotalPages int       `json:"total_pages"`
	TotalCount int       `json:"total_count"`
	HasNext    bool      `json:"has_next"`
}

type ExpenseIDInput struct {
	ID int `json:"id" jsonschema:"the expense id"`
}

type CreateExpenseInput struct {
	CategoryID  int    `json:"category_id" jsonschema:"category id; see list_categories"`
	Description string `json:"description" jsonschema:"3 to 50 characters"`
	Amount      string `json:"amount" jsonschema:"decimal amount greater than zero, e.g. 12.50"`
	//nolint:lll // struct tags cannot wrap
	BilledMonth string   `json:"billed_month,omitempty" jsonschema:"month the expense is billed to, YYYY-MM; defaults to the current month"`
	Note        string   `json:"note,omitempty" jsonschema:"optional note, at most 255 characters"`
	Tags        []string `json:"tags,omitempty" jsonschema:"tag names; new ones are created"`
}

type QuickAddExpenseInput struct {
	//nolint:lll // struct tags cannot wrap
	Input      string `json:"input" jsonschema:"'description, amount, month[, tags]' — month is last, current, next or like 'Jul 2026'; tags are separated by semicolons"`
	CategoryID int    `json:"category_id,omitempty" jsonschema:"needed only when the description has no known category"`
}

type UpdateExpenseInput struct {
	ID          int       `json:"id" jsonschema:"the expense id"`
	CategoryID  *int      `json:"category_id,omitempty" jsonschema:"new category id"`
	Description *string   `json:"description,omitempty" jsonschema:"new description, 3 to 50 characters"`
	Amount      *string   `json:"amount,omitempty" jsonschema:"new decimal amount, e.g. 12.50"`
	BilledMonth *string   `json:"billed_month,omitempty" jsonschema:"new billed month, YYYY-MM"`
	Note        *string   `json:"note,omitempty" jsonschema:"new note; an empty string clears it"`
	Tags        *[]string `json:"tags,omitempty" jsonschema:"replace the whole tag list; an empty list removes every tag"`
	AddTags     []string  `json:"add_tags,omitempty" jsonschema:"tags to add, keeping the others"`
	RemoveTags  []string  `json:"remove_tags,omitempty" jsonschema:"tags to remove, keeping the others"`
}

func registerExpenseTools(server *mcp.Server, d Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "search_expenses",
		Description: "Search and list expenses, newest first by default. With no filters it lists every " +
			"expense, a page at a time. Filter by billed month or by the day it was recorded, not both." +
			userDataNote,
		Annotations: readOnly("Search expenses"),
	}, d.searchExpenses)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_expense",
		Description: "Read one expense, including its note." + userDataNote,
		Annotations: readOnly("Get expense"),
	}, d.getExpense)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_expense",
		Description: "Record a new expense. Amounts are decimals in the account's currency.",
		Annotations: creates("Create expense"),
	}, d.createExpense)

	mcp.AddTool(server, &mcp.Tool{
		Name: "quick_add_expense",
		Description: "Record an expense from one line, the way the app's quick-add box does: " +
			"'Coffee, 4.50, current; food'. The category is remembered from earlier expenses with the " +
			"same description; when there is none, the call fails asking for category_id.",
		Annotations: creates("Quick-add expense"),
	}, d.quickAddExpense)

	mcp.AddTool(server, &mcp.Tool{
		Name: "update_expense",
		Description: "Change an expense. Only the fields you pass change. For tags, pass add_tags/remove_tags " +
			"to adjust the list, or tags to replace it." + userDataNote,
		Annotations: overwrites("Update expense"),
	}, d.updateExpense)
}

func (d Deps) searchExpenses(
	ctx context.Context, _ *mcp.CallToolRequest, in SearchExpensesInput,
) (*mcp.CallToolResult, ExpenseListOutput, error) {
	var out ExpenseListOutput

	query := url.Values{}

	if in.Query != "" {
		query.Set("q", in.Query)
	}

	if in.Tag != "" {
		query.Set("tag", in.Tag)
	}

	if in.CategoryID > 0 {
		query.Set("category_id", strconv.Itoa(in.CategoryID))
	}

	if err := d.setSearchRange(query, in); err != nil {
		return nil, out, err
	}

	if err := setPaging(query, in.Page, in.PerPage); err != nil {
		return nil, out, err
	}

	if err := setSort(query, in.SortField, in.SortOrder,
		[]string{"created_at", "date", "amount", "description"}); err != nil {
		return nil, out, err
	}

	ctx, cancel := callContext(ctx)
	defer cancel()

	var list api.ExpenseList
	if err := d.API.Get(ctx, "/expenses", query, &list); err != nil {
		return nil, out, err
	}

	out.Expenses = make([]Expense, 0, len(list.Data))
	for _, e := range list.Data {
		out.Expenses = append(out.Expenses, toExpense(e, d.Location, nil))
	}

	out.Page = list.Pagination.CurrentPage
	out.TotalPages = list.Pagination.TotalPages
	out.TotalCount = list.Pagination.TotalCount
	out.HasNext = list.Pagination.HasNext

	return nil, out, nil
}

// setSearchRange maps the two kinds of date filter onto the API's two bound
// pairs. They are exclusive because the server drops the billed-month bounds
// whenever creation bounds are present — silently combining them would return
// a result the caller did not ask for.
func (d Deps) setSearchRange(query url.Values, in SearchExpensesInput) error {
	hasBilled := in.FromMonth != "" || in.ToMonth != ""
	hasCreated := in.CreatedFrom != "" || in.CreatedTo != ""

	switch {
	case hasBilled && hasCreated:
		return ErrRangeConflict
	case (in.ToMonth != "" && in.FromMonth == "") || (in.CreatedTo != "" && in.CreatedFrom == ""):
		return ErrToWithoutFrom
	case hasBilled:
		start, end, err := units.MonthRange(in.FromMonth, in.ToMonth)
		if err != nil {
			return err
		}

		query.Set("start", strconv.FormatInt(start, 10))
		query.Set("end", strconv.FormatInt(end, 10))
	case hasCreated:
		start, end, err := units.DayRange(in.CreatedFrom, in.CreatedTo, d.Location)
		if err != nil {
			return err
		}

		query.Set("created_start", strconv.FormatInt(start, 10))
		query.Set("created_end", strconv.FormatInt(end, 10))
	}

	return nil
}

func setPaging(query url.Values, page, perPage int) error {
	if page > 0 {
		query.Set("page", strconv.Itoa(page))
	}

	if perPage == 0 {
		perPage = defaultPerPage
	}

	if !slices.Contains(perPageChoices, perPage) {
		return ErrPerPage
	}

	query.Set("per_page", strconv.Itoa(perPage))

	return nil
}

func setSort(query url.Values, field, order string, allowed []string) error {
	if field != "" {
		if !slices.Contains(allowed, field) {
			return fmt.Errorf("%w %q: use one of %s", ErrSortField, field, strings.Join(allowed, ", "))
		}

		query.Set("sort_field", field)
	}

	if order != "" {
		upper := strings.ToUpper(order)
		if upper != "ASC" && upper != "DESC" {
			return ErrSortOrder
		}

		// The server reads the field and the order as a pair: an order with
		// no field would fall back to its default field and drop the order.
		if field == "" {
			query.Set("sort_field", "created_at")
		}

		query.Set("sort_order", upper)
	}

	return nil
}

func (d Deps) getExpense(
	ctx context.Context, _ *mcp.CallToolRequest, in ExpenseIDInput,
) (*mcp.CallToolResult, Expense, error) {
	ctx, cancel := callContext(ctx)
	defer cancel()

	e, err := d.fetchExpense(ctx, in.ID)
	if err != nil {
		return nil, Expense{}, err
	}

	return nil, toExpense(e.ListedExpense, d.Location, &e.Note), nil
}

func (d Deps) fetchExpense(ctx context.Context, id int) (api.Expense, error) {
	var e api.Expense

	err := d.API.Get(ctx, "/expenses/"+strconv.Itoa(id), nil, &e)
	if errors.Is(err, api.ErrNotFound) {
		return e, fmt.Errorf("expense %d: %w", id, err)
	}

	return e, err
}

func (d Deps) createExpense(
	ctx context.Context, _ *mcp.CallToolRequest, in CreateExpenseInput,
) (*mcp.CallToolResult, Expense, error) {
	amount, err := units.ParseAmount(in.Amount)
	if err != nil {
		return nil, Expense{}, err
	}

	month := in.BilledMonth
	if month == "" {
		month = units.CurrentMonth(d.Now(), d.Location)
	}

	date, err := units.ParseMonth(month)
	if err != nil {
		return nil, Expense{}, err
	}

	ctx, cancel := callContext(ctx)
	defer cancel()

	var created api.Expense

	err = d.API.Post(ctx, "/expenses", api.ExpenseBody{
		CategoryID:  in.CategoryID,
		Description: in.Description,
		Amount:      amount,
		Date:        date,
		Note:        in.Note,
		Tags:        nonNil(in.Tags),
	}, &created)
	if err != nil {
		return nil, Expense{}, err
	}

	return nil, toExpense(created.ListedExpense, d.Location, &created.Note), nil
}

func (d Deps) quickAddExpense(
	ctx context.Context, _ *mcp.CallToolRequest, in QuickAddExpenseInput,
) (*mcp.CallToolResult, Expense, error) {
	ctx, cancel := callContext(ctx)
	defer cancel()

	var created api.Expense

	err := d.API.Post(ctx, "/expenses/quick", api.QuickExpenseBody{
		QuickInput: in.Input,
		CategoryID: in.CategoryID,
		// Quick-add resolves "last"/"current"/"next" against the caller's
		// calendar month, which it takes in getTimezoneOffset() form.
		TZOffset: units.TZOffsetMinutes(d.Now(), d.Location),
	}, &created)
	if err != nil {
		var apiErr *api.Error
		if errors.As(err, &apiErr) && apiErr.Fields["category_id"] == "required" {
			return nil, Expense{}, fmt.Errorf(
				"%w: this description has no remembered category; call again with category_id "+
					"(see list_categories)", err)
		}

		return nil, Expense{}, err
	}

	return nil, toExpense(created.ListedExpense, d.Location, &created.Note), nil
}

func (d Deps) updateExpense(
	ctx context.Context, _ *mcp.CallToolRequest, in UpdateExpenseInput,
) (*mcp.CallToolResult, Expense, error) {
	if in.CategoryID == nil && in.Description == nil && in.Amount == nil && in.BilledMonth == nil &&
		in.Note == nil && in.Tags == nil && len(in.AddTags) == 0 && len(in.RemoveTags) == 0 {
		return nil, Expense{}, ErrNothingToSet
	}

	ctx, cancel := callContext(ctx)
	defer cancel()

	// PUT replaces the whole expense, so the unchanged fields come from the
	// current record. The note is only on the single-expense read, which is
	// why this does not start from a list result.
	current, err := d.fetchExpense(ctx, in.ID)
	if err != nil {
		return nil, Expense{}, err
	}

	body := api.ExpenseBody{
		CategoryID:  current.CategoryID,
		Description: current.Description,
		Amount:      current.Amount,
		Date:        current.Date,
		Note:        current.Note,
	}

	if err := applyExpenseChanges(&body, in); err != nil {
		return nil, Expense{}, err
	}

	body.Tags, err = mergeTags(current.Tags, in.Tags, in.AddTags, in.RemoveTags)
	if err != nil {
		return nil, Expense{}, err
	}

	var updated api.Expense
	if err := d.API.Put(ctx, "/expenses/"+strconv.Itoa(in.ID), body, &updated); err != nil {
		return nil, Expense{}, err
	}

	return nil, toExpense(updated.ListedExpense, d.Location, &updated.Note), nil
}

func applyExpenseChanges(body *api.ExpenseBody, in UpdateExpenseInput) error {
	if in.CategoryID != nil {
		body.CategoryID = *in.CategoryID
	}

	if in.Description != nil {
		body.Description = *in.Description
	}

	if in.Note != nil {
		body.Note = *in.Note
	}

	if in.Amount != nil {
		amount, err := units.ParseAmount(*in.Amount)
		if err != nil {
			return err
		}

		body.Amount = amount
	}

	if in.BilledMonth != nil {
		date, err := units.ParseMonth(*in.BilledMonth)
		if err != nil {
			return err
		}

		body.Date = date
	}

	return nil
}
