package projects

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var (
	_ twapi.HTTPRequester = (*ProjectBudgetListRequest)(nil)
	_ twapi.HTTPResponser = (*ProjectBudgetListResponse)(nil)
)

type ProjectBudgetExpenseType string

const (
	ProjectBudgetExpenseTypeAll         ProjectBudgetExpenseType = "ALL"
	ProjectBudgetExpenseTypeBillable    ProjectBudgetExpenseType = "BILLABLE"
	ProjectBudgetExpenseTypeNonBillable ProjectBudgetExpenseType = "NON-BILLABLE"
)

type ProjectBudgetRepeatUnit string

const (
	ProjectBudgetRepeatUnitNone    ProjectBudgetRepeatUnit = ""
	ProjectBudgetRepeatUnitDay     ProjectBudgetRepeatUnit = "DAY"
	ProjectBudgetRepeatUnitWeek    ProjectBudgetRepeatUnit = "WEEK"
	ProjectBudgetRepeatUnitMonth   ProjectBudgetRepeatUnit = "MONTH"
	ProjectBudgetRepeatUnitQuarter ProjectBudgetRepeatUnit = "QUARTER"
	ProjectBudgetRepeatUnitYear    ProjectBudgetRepeatUnit = "YEAR"
)

type ProjectBudgetTimelogType string

const (
	ProjectBudgetTimelogTypeAll         ProjectBudgetTimelogType = "ALL"
	ProjectBudgetTimelogTypeBilled      ProjectBudgetTimelogType = "BILLED"
	ProjectBudgetTimelogTypeUnbilled    ProjectBudgetTimelogType = "UNBILLED"
	ProjectBudgetTimelogTypeBillable    ProjectBudgetTimelogType = "BILLABLE"
	ProjectBudgetTimelogTypeNonBillable ProjectBudgetTimelogType = "NON-BILLABLE"
)

// ProjectBudget contains project budget data exposed in included sideloads.
type ProjectBudget struct {
	// ID is the unique identifier of the project budget.
	ID int64 `json:"id"`

	// ProjectID is the identifier of the parent project.
	ProjectID int64 `json:"projectId"`

	// Type is the project budget type.
	Type BudgetType `json:"type"`

	// Status is the current project budget status.
	Status ProjectBudgetStatus `json:"status"`

	// Capacity is the total budget capacity.
	Capacity int64 `json:"capacity"`

	// CapacityUsed is the consumed amount of budget capacity.
	CapacityUsed int64 `json:"capacityUsed"`

	// OriginatorBudgetID points to the originating budget in repeating sequences.
	OriginatorBudgetID *int64 `json:"originatorBudgetId"`

	// IsRepeating indicates whether this budget repeats.
	IsRepeating bool `json:"isRepeating"`

	// RepeatPeriod defines how often the budget repeats.
	RepeatPeriod *int64 `json:"repeatPeriod"`

	// RepeatUnit defines the repeat unit (for example "month").
	RepeatUnit *ProjectBudgetRepeatUnit `json:"repeatUnit"`

	// RepeatsRemaining is the remaining repeat count.
	RepeatsRemaining *int64 `json:"repeatsRemaining"`

	// SequenceNumber is the position in a repeated sequence.
	SequenceNumber *int64 `json:"sequenceNumber"`

	// StartDateTime is the budget period start date and time.
	StartDateTime *time.Time `json:"startDateTime"`

	// EndDateTime is the budget period end date and time.
	EndDateTime *time.Time `json:"endDateTime"`

	// CurrencyCode is the currency used by this budget.
	CurrencyCode *string `json:"currencyCode"`

	// TimelogType is the timelog calculation mode used by this budget.
	TimelogType *ProjectBudgetTimelogType `json:"timelogType"`

	// ExpenseType is the expense calculation mode used by this budget.
	ExpenseType *ProjectBudgetExpenseType `json:"expenseType"`

	// DefaultRate is the default rate applied by this budget.
	DefaultRate *float64 `json:"defaultRate"`

	// NotificationIDs are identifiers of notifications associated with this budget.
	NotificationIDs []int64 `json:"notificationIds"`

	// CreatedBy is the identifier of the user who created this budget.
	CreatedBy *int64 `json:"createdByUserId"`

	// CreatedAt is the date and time when this budget was created.
	CreatedAt *time.Time `json:"dateCreated"`

	// UpdatedBy is the identifier of the user who last updated this budget.
	UpdatedBy *int64 `json:"updatedUserId"`

	// UpdatedAt is the date and time when this budget was last updated.
	UpdatedAt *time.Time `json:"dateUpdated"`

	// CompletedBy is the identifier of the user who completed this budget.
	CompletedBy *int64 `json:"completedByUserId"`

	// CompletedAt is the date and time when this budget was completed.
	CompletedAt *time.Time `json:"dateCompleted"`

	// DeletedBy is the identifier of the user who deleted this budget.
	DeletedBy *int64 `json:"deletedByUserId"`

	// DeletedAt is the date and time when this budget was deleted.
	DeletedAt *time.Time `json:"dateDeleted"`
}

// ProjectBudgetStatus is the status filter value for project budgets.
type ProjectBudgetStatus string

// Project budget status values.
const (
	ProjectBudgetStatusUpcoming ProjectBudgetStatus = "upcoming"
	ProjectBudgetStatusActive   ProjectBudgetStatus = "active"
	ProjectBudgetStatusComplete ProjectBudgetStatus = "complete"
)

// ProjectBudgetListRequestFilters contains filters for listing project budgets.
type ProjectBudgetListRequestFilters struct {
	// ProjectIDs filters budgets to one or more projects.
	ProjectIDs []int64

	// Status filters budgets by status.
	Status ProjectBudgetStatus

	// Limit limits the number of items returned by the endpoint.
	Limit int64

	// PageSize sets the number of items per page.
	PageSize int64

	// Cursor is the pagination cursor used by the endpoint.
	Cursor string
}

// ProjectBudgetListRequest represents the request for listing project budgets.
//
// projects/api/v3/projects/budgets.json
type ProjectBudgetListRequest struct {
	// Filters contains optional query string filters for the request.
	Filters ProjectBudgetListRequestFilters
}

// NewProjectBudgetListRequest creates a new ProjectBudgetListRequest with no
// filters. All query parameters are optional.
func NewProjectBudgetListRequest() ProjectBudgetListRequest {
	return ProjectBudgetListRequest{}
}

// HTTPRequest creates an HTTP request for the ProjectBudgetListRequest.
func (p ProjectBudgetListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/projects/budgets.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	if len(p.Filters.ProjectIDs) > 0 {
		ids := make([]string, 0, len(p.Filters.ProjectIDs))
		for _, id := range p.Filters.ProjectIDs {
			ids = append(ids, strconv.FormatInt(id, 10))
		}
		query.Set("projectIds", strings.Join(ids, ","))
	}
	if p.Filters.Status != "" {
		query.Set("status", string(p.Filters.Status))
	}
	if p.Filters.Limit > 0 {
		query.Set("limit", strconv.FormatInt(p.Filters.Limit, 10))
	}
	if p.Filters.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(p.Filters.PageSize, 10))
	}
	if p.Filters.Cursor != "" {
		query.Set("cursor", p.Filters.Cursor)
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// ProjectBudgetListResponse contains the list of project budgets matching the
// request filters.
type ProjectBudgetListResponse struct {
	Budgets []ProjectBudget `json:"budgets"`

	Meta struct {
		Page struct {
			PageOffset int64 `json:"pageOffset"`
			PageSize   int64 `json:"pageSize"`
			Count      int64 `json:"count"`
			HasMore    bool  `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`

	// Included contains sideloaded entities. The shape depends on selected API
	// options and can vary, so values are kept as raw JSON blobs.
	Included map[string]json.RawMessage `json:"included"`
}

// HandleHTTPResponse handles the HTTP response for the
// ProjectBudgetListResponse.
func (p *ProjectBudgetListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list project budgets")
	}

	if err := json.NewDecoder(resp.Body).Decode(p); err != nil {
		return fmt.Errorf("failed to decode list project budgets response: %w", err)
	}
	return nil
}

// ProjectBudgetList retrieves project budgets using the provided request and
// returns the response.
func ProjectBudgetList(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectBudgetListRequest,
) (*ProjectBudgetListResponse, error) {
	return twapi.Execute[ProjectBudgetListRequest, *ProjectBudgetListResponse](ctx, engine, req)
}
