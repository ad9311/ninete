package tools

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"github.com/ad9311/ninete-mcp/internal/api"
	"github.com/ad9311/ninete-mcp/internal/units"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RecurrentExpense is a recurrent expense as the tools present it.
type RecurrentExpense struct {
	ID              int    `json:"id"`
	CategoryID      int    `json:"category_id"`
	Category        string `json:"category"`
	Description     string `json:"description"`
	Amount          string `json:"amount" jsonschema:"decimal amount, e.g. 12.50"`
	PeriodMonths    uint   `json:"period_months" jsonschema:"an expense is generated every this many months"`
	OccurrenceLimit uint   `json:"occurrence_limit" jsonschema:"copies before it archives itself; 0 is unlimited"`
	OccurrenceCount uint   `json:"occurrence_count" jsonschema:"copies generated so far"`
	Archived        bool   `json:"archived"`
	// Tags is never nil, so the output always matches its schema's array type.
	Tags []string `json:"tags"`
}

func toRecurrentExpense(r api.RecurrentExpense) RecurrentExpense {
	return RecurrentExpense{
		ID:              r.ID,
		CategoryID:      r.CategoryID,
		Category:        r.CategoryName,
		Description:     r.Description,
		Amount:          units.FormatAmount(r.Amount),
		PeriodMonths:    r.Period,
		OccurrenceLimit: r.OccurrenceLimit,
		OccurrenceCount: r.OccurrenceCount,
		Archived:        r.Archived,
		Tags:            nonNil(r.Tags),
	}
}

type ListRecurrentExpensesInput struct {
	Archived   bool   `json:"archived,omitempty" jsonschema:"list the archived ones instead of the active ones"`
	CategoryID int    `json:"category_id,omitempty" jsonschema:"only this category"`
	Page       int    `json:"page,omitempty" jsonschema:"page number, from 1"`
	PerPage    int    `json:"per_page,omitempty" jsonschema:"results per page: 15, 25, 50 or 100 (default 25)"`
	SortField  string `json:"sort_field,omitempty" jsonschema:"created_at (default), description, amount or period"`
	SortOrder  string `json:"sort_order,omitempty" jsonschema:"asc or desc (default desc)"`
}

type RecurrentExpenseListOutput struct {
	RecurrentExpenses []RecurrentExpense `json:"recurrent_expenses"`
	Page              int                `json:"page"`
	TotalPages        int                `json:"total_pages"`
	TotalCount        int                `json:"total_count"`
	HasNext           bool               `json:"has_next"`
}

type RecurrentExpenseIDInput struct {
	ID int `json:"id" jsonschema:"the recurrent expense id"`
}

type CreateRecurrentExpenseInput struct {
	CategoryID      int      `json:"category_id" jsonschema:"category id; see list_categories"`
	Description     string   `json:"description" jsonschema:"3 to 50 characters"`
	Amount          string   `json:"amount" jsonschema:"decimal amount greater than zero, e.g. 12.50"`
	PeriodMonths    uint     `json:"period_months" jsonschema:"generate an expense every this many months, at least 1"`
	OccurrenceLimit uint     `json:"occurrence_limit,omitempty" jsonschema:"archive after this many copies; 0 = no limit"`
	Tags            []string `json:"tags,omitempty" jsonschema:"tag names copied onto every generated expense"`
}

type UpdateRecurrentExpenseInput struct {
	ID              int       `json:"id" jsonschema:"the recurrent expense id"`
	CategoryID      *int      `json:"category_id,omitempty" jsonschema:"new category id"`
	Description     *string   `json:"description,omitempty" jsonschema:"new description, 3 to 50 characters"`
	Amount          *string   `json:"amount,omitempty" jsonschema:"new decimal amount, e.g. 12.50"`
	PeriodMonths    *uint     `json:"period_months,omitempty" jsonschema:"new period in months, at least 1"`
	OccurrenceLimit *uint     `json:"occurrence_limit,omitempty" jsonschema:"new copy limit; 0 is unlimited"`
	Tags            *[]string `json:"tags,omitempty" jsonschema:"replace the whole tag list; empty removes all"`
	AddTags         []string  `json:"add_tags,omitempty" jsonschema:"tags to add, keeping the others"`
	RemoveTags      []string  `json:"remove_tags,omitempty" jsonschema:"tags to remove, keeping the others"`
}

func registerRecurrentExpenseTools(server *mcp.Server, d Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "list_recurrent_expenses",
		Description: "List recurrent expenses — the templates that generate an expense every few months. " +
			"Active ones by default; archived ones with archived=true." + userDataNote,
		Annotations: readOnly("List recurrent expenses"),
	}, d.listRecurrentExpenses)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_recurrent_expense",
		Description: "Read one recurrent expense." + userDataNote,
		Annotations: readOnly("Get recurrent expense"),
	}, d.getRecurrentExpense)

	mcp.AddTool(server, &mcp.Tool{
		Name: "create_recurrent_expense",
		Description: "Create a recurrent expense. It generates a copy of itself as an expense every " +
			"period_months months, carrying its tags.",
		Annotations: creates("Create recurrent expense"),
	}, d.createRecurrentExpense)

	mcp.AddTool(server, &mcp.Tool{
		Name: "update_recurrent_expense",
		Description: "Change a recurrent expense. Only the fields you pass change. Expenses it already " +
			"generated are not touched. Editing does not unarchive it; use unarchive_recurrent_expense.",
		Annotations: overwrites("Update recurrent expense"),
	}, d.updateRecurrentExpense)

	mcp.AddTool(server, &mcp.Tool{
		Name: "unarchive_recurrent_expense",
		Description: "Put an archived recurrent expense back in rotation. One archives itself after " +
			"generating occurrence_limit copies; raise the limit first or it will archive again.",
		Annotations: &mcp.ToolAnnotations{
			Title: "Unarchive recurrent expense", DestructiveHint: new(false),
			IdempotentHint: true, OpenWorldHint: new(false),
		},
	}, d.unarchiveRecurrentExpense)
}

func (d Deps) listRecurrentExpenses(
	ctx context.Context, _ *mcp.CallToolRequest, in ListRecurrentExpensesInput,
) (*mcp.CallToolResult, RecurrentExpenseListOutput, error) {
	var out RecurrentExpenseListOutput

	query := url.Values{}
	query.Set("archived", strconv.FormatBool(in.Archived))

	if in.CategoryID > 0 {
		query.Set("category_id", strconv.Itoa(in.CategoryID))
	}

	if err := setPaging(query, in.Page, in.PerPage); err != nil {
		return nil, out, err
	}

	if err := setSort(query, in.SortField, in.SortOrder,
		[]string{"created_at", "description", "amount", "period"}); err != nil {
		return nil, out, err
	}

	ctx, cancel := callContext(ctx)
	defer cancel()

	var list api.RecurrentExpenseList
	if err := d.API.Get(ctx, "/recurrent-expenses", query, &list); err != nil {
		return nil, out, err
	}

	out.RecurrentExpenses = make([]RecurrentExpense, 0, len(list.Data))
	for _, r := range list.Data {
		out.RecurrentExpenses = append(out.RecurrentExpenses, toRecurrentExpense(r))
	}

	out.Page = list.Pagination.CurrentPage
	out.TotalPages = list.Pagination.TotalPages
	out.TotalCount = list.Pagination.TotalCount
	out.HasNext = list.Pagination.HasNext

	return nil, out, nil
}

func (d Deps) getRecurrentExpense(
	ctx context.Context, _ *mcp.CallToolRequest, in RecurrentExpenseIDInput,
) (*mcp.CallToolResult, RecurrentExpense, error) {
	ctx, cancel := callContext(ctx)
	defer cancel()

	r, err := d.fetchRecurrentExpense(ctx, in.ID)
	if err != nil {
		return nil, RecurrentExpense{}, err
	}

	return nil, toRecurrentExpense(r), nil
}

func (d Deps) fetchRecurrentExpense(ctx context.Context, id int) (api.RecurrentExpense, error) {
	var r api.RecurrentExpense

	err := d.API.Get(ctx, recurrentExpensePath(id), nil, &r)
	if errors.Is(err, api.ErrNotFound) {
		return r, fmt.Errorf("recurrent expense %d: %w", id, err)
	}

	return r, err
}

func recurrentExpensePath(id int) string {
	return "/recurrent-expenses/" + strconv.Itoa(id)
}

func (d Deps) createRecurrentExpense(
	ctx context.Context, _ *mcp.CallToolRequest, in CreateRecurrentExpenseInput,
) (*mcp.CallToolResult, RecurrentExpense, error) {
	amount, err := units.ParseAmount(in.Amount)
	if err != nil {
		return nil, RecurrentExpense{}, err
	}

	ctx, cancel := callContext(ctx)
	defer cancel()

	var created api.RecurrentExpense

	err = d.API.Post(ctx, "/recurrent-expenses", api.RecurrentExpenseBody{
		CategoryID:      in.CategoryID,
		Description:     in.Description,
		Amount:          amount,
		Period:          in.PeriodMonths,
		OccurrenceLimit: in.OccurrenceLimit,
		Tags:            nonNil(in.Tags),
	}, &created)
	if err != nil {
		return nil, RecurrentExpense{}, err
	}

	return nil, toRecurrentExpense(created), nil
}

func (d Deps) updateRecurrentExpense(
	ctx context.Context, _ *mcp.CallToolRequest, in UpdateRecurrentExpenseInput,
) (*mcp.CallToolResult, RecurrentExpense, error) {
	if in.CategoryID == nil && in.Description == nil && in.Amount == nil && in.PeriodMonths == nil &&
		in.OccurrenceLimit == nil && in.Tags == nil && len(in.AddTags) == 0 && len(in.RemoveTags) == 0 {
		return nil, RecurrentExpense{}, ErrNothingToSet
	}

	ctx, cancel := callContext(ctx)
	defer cancel()

	current, err := d.fetchRecurrentExpense(ctx, in.ID)
	if err != nil {
		return nil, RecurrentExpense{}, err
	}

	body := api.RecurrentExpenseBody{
		CategoryID:      current.CategoryID,
		Description:     current.Description,
		Amount:          current.Amount,
		Period:          current.Period,
		OccurrenceLimit: current.OccurrenceLimit,
	}

	if err := applyRecurrentExpenseChanges(&body, in); err != nil {
		return nil, RecurrentExpense{}, err
	}

	body.Tags, err = mergeTags(current.Tags, in.Tags, in.AddTags, in.RemoveTags)
	if err != nil {
		return nil, RecurrentExpense{}, err
	}

	var updated api.RecurrentExpense
	if err := d.API.Put(ctx, recurrentExpensePath(in.ID), body, &updated); err != nil {
		return nil, RecurrentExpense{}, err
	}

	return nil, toRecurrentExpense(updated), nil
}

func applyRecurrentExpenseChanges(body *api.RecurrentExpenseBody, in UpdateRecurrentExpenseInput) error {
	if in.CategoryID != nil {
		body.CategoryID = *in.CategoryID
	}

	if in.Description != nil {
		body.Description = *in.Description
	}

	if in.PeriodMonths != nil {
		body.Period = *in.PeriodMonths
	}

	if in.OccurrenceLimit != nil {
		body.OccurrenceLimit = *in.OccurrenceLimit
	}

	if in.Amount != nil {
		amount, err := units.ParseAmount(*in.Amount)
		if err != nil {
			return err
		}

		body.Amount = amount
	}

	return nil
}

func (d Deps) unarchiveRecurrentExpense(
	ctx context.Context, _ *mcp.CallToolRequest, in RecurrentExpenseIDInput,
) (*mcp.CallToolResult, RecurrentExpense, error) {
	ctx, cancel := callContext(ctx)
	defer cancel()

	var updated api.RecurrentExpense

	err := d.API.Post(ctx, recurrentExpensePath(in.ID)+"/unarchive", nil, &updated)
	if errors.Is(err, api.ErrNotFound) {
		return nil, RecurrentExpense{}, fmt.Errorf("recurrent expense %d: %w", in.ID, err)
	}

	if err != nil {
		return nil, RecurrentExpense{}, err
	}

	return nil, toRecurrentExpense(updated), nil
}
