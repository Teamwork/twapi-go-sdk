//nolint:lll
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
	_ twapi.HTTPRequester = (*ProjectBudgetTasklistBudgetListRequest)(nil)
	_ twapi.HTTPResponser = (*ProjectBudgetTasklistBudgetListResponse)(nil)
)

type TasklistBudgetType string

const (
	TasklistBudgetTypeAll       TasklistBudgetType = "ALL"
	TasklistBudgetTypeFinancial TasklistBudgetType = "FINANCIAL"
	TasklistBudgetTypeTime      TasklistBudgetType = "TIME"
)

// TasklistBudget represents a budget item attached to a tasklist.
type TasklistBudget struct {
	// ID is the unique identifier of the tasklist budget.
	ID int64 `json:"id"`

	// Type is the budget type.
	Type TasklistBudgetType `json:"type"`

	// Capacity is the budget capacity in the smallest unit supported by the API.
	Capacity int64 `json:"capacity"`

	// CapacityUsed is the consumed budget capacity.
	CapacityUsed int64 `json:"capacityUsed"`

	// ProjectID is the parent project identifier.
	ProjectID int64 `json:"projectId"`

	// ProjectBudget is the relationship to the parent project budget.
	ProjectBudget twapi.Relationship `json:"projectbudget"`

	// Tasklist is the relationship to the tasklist associated with this budget.
	Tasklist twapi.Relationship `json:"tasklist"`

	// Milestone is the relationship to the milestone associated with this budget,
	// when available.
	Milestone *twapi.Relationship `json:"milestone"`

	// Notifications are relationships to the budget notifications.
	Notifications []twapi.Relationship `json:"notifications"`

	// CreatedAt is the date and time when the budget was created.
	CreatedAt *time.Time `json:"createdAt"`

	// CreatedBy is the identifier of the user who created the budget.
	CreatedBy *int64 `json:"createdBy"`

	// UpdatedAt is the date and time when the budget was last updated.
	UpdatedAt *time.Time `json:"updatedAt"`

	// UpdatedBy is the identifier of the user who last updated the budget.
	UpdatedBy *int64 `json:"updatedBy"`

	// DeletedAt is the date and time when the budget was deleted.
	DeletedAt *time.Time `json:"deletedAt"`

	// DeletedBy is the identifier of the user who deleted the budget.
	DeletedBy *int64 `json:"deletedBy"`
}

type ProjectBudgetExpenseType string

const (
	ProjectBudgetExpenseTypeAll         ProjectBudgetExpenseType = "ALL"
	ProjectBudgetExpenseTypeBillable    ProjectBudgetExpenseType = "BILLABLE"
	ProjectBudgetExpenseTypeNonBillable ProjectBudgetExpenseType = "NON-BILLABLE"
)

// ProjectBudget contains project budget data exposed in included sideloads.
type ProjectBudget struct {
	// ID is the unique identifier of the project budget.
	ID int64 `json:"id"`

	// ProjectID is the identifier of the parent project.
	ProjectID int64 `json:"projectId"`

	// Type is the project budget type.
	Type TasklistBudgetType `json:"type"`

	// Status is the current project budget status.
	Status string `json:"status"`

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
	RepeatUnit *string `json:"repeatUnit"`

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
	TimelogType *string `json:"timelogType"`

	// ExpenseType is the expense calculation mode used by this budget.
	ExpenseType *ProjectBudgetExpenseType `json:"expenseType"`

	// DefaultRate is the default rate applied by this budget.
	DefaultRate *float64 `json:"defaultRate"`

	// NotificationIDs are identifiers of notifications associated with this budget.
	NotificationIDs []int64 `json:"notificationIds"`

	// CreatedByUserID is the identifier of the user who created this budget.
	CreatedByUserID *int64 `json:"createdByUserId"`

	// DateCreated is the date and time when this budget was created.
	DateCreated *time.Time `json:"dateCreated"`

	// UpdatedUserID is the identifier of the user who last updated this budget.
	UpdatedUserID *int64 `json:"updatedUserId"`

	// DateUpdated is the date and time when this budget was last updated.
	DateUpdated *time.Time `json:"dateUpdated"`

	// CompletedByUserID is the identifier of the user who completed this budget.
	CompletedByUserID *int64 `json:"completedByUserId"`

	// DateCompleted is the date and time when this budget was completed.
	DateCompleted *time.Time `json:"dateCompleted"`

	// DeletedByUserID is the identifier of the user who deleted this budget.
	DeletedByUserID *int64 `json:"deletedByUserId"`

	// DateDeleted is the date and time when this budget was deleted.
	DateDeleted *time.Time `json:"dateDeleted"`
}

// TasklistBudgetNotification contains notification details for a tasklist
// budget.
type TasklistBudgetNotification struct {
	// ID is the unique identifier of the notification.
	ID int64 `json:"id"`

	// BudgetID is the associated budget identifier.
	BudgetID int64 `json:"budgetId"`

	// CapacityThreshold is the threshold that triggers this notification.
	CapacityThreshold float64 `json:"capacityThreshold"`

	// NotificationMedium is the medium used for the notification.
	NotificationMedium string `json:"notificationMedium"`

	// UserID is the user identifier for user-targeted notifications.
	UserID *int64 `json:"userId"`

	// TeamID is the team identifier for team-targeted notifications.
	TeamID *int64 `json:"teamId"`

	// CompanyID is the company identifier for company-targeted notifications.
	CompanyID *int64 `json:"companyId"`
}

// ProjectBudgetTasklistBudgetListRequestPath contains the path parameters for
// listing tasklist budgets in a project budget.
type ProjectBudgetTasklistBudgetListRequestPath struct {
	// ProjectBudgetID is the unique identifier of the parent project budget.
	ProjectBudgetID int64
}

// ProjectBudgetTasklistBudgetListRequestSideload represents related objects
// that can be included in the response.
type ProjectBudgetTasklistBudgetListRequestSideload string

const (
	ProjectBudgetTasklistBudgetListRequestSideloadTasklists                   ProjectBudgetTasklistBudgetListRequestSideload = "tasklists"
	ProjectBudgetTasklistBudgetListRequestSideloadProjectBudgets              ProjectBudgetTasklistBudgetListRequestSideload = "projectBudgets"
	ProjectBudgetTasklistBudgetListRequestSideloadTasklistBudgetNotifications ProjectBudgetTasklistBudgetListRequestSideload = "tasklistBudgetNotifications"
)

// ProjectBudgetTasklistBudgetListRequestOrderBy defines sortable fields for
// tasklist budget listings.
type ProjectBudgetTasklistBudgetListRequestOrderBy string

const (
	ProjectBudgetTasklistBudgetListRequestOrderByDateCreated ProjectBudgetTasklistBudgetListRequestOrderBy = "dateCreated"
)

// ProjectBudgetTasklistBudgetListRequestFields contains field selectors for
// sideloaded entities.
type ProjectBudgetTasklistBudgetListRequestFields struct {
	// Tasklists limits fields returned for sideloaded tasklists.
	Tasklists []string

	// TasklistBudgetNotifications limits fields returned for sideloaded
	// notifications.
	TasklistBudgetNotifications []string

	// ProjectBudgets limits fields returned for sideloaded project budgets.
	ProjectBudgets []string
}

// ProjectBudgetTasklistBudgetListRequestFilters contains filters for listing
// tasklist budgets in a project budget.
type ProjectBudgetTasklistBudgetListRequestFilters struct {
	// OrderMode specifies sort direction. Allowed values are "asc" and "desc".
	OrderMode twapi.OrderMode

	// OrderBy specifies the field used for sorting.
	OrderBy ProjectBudgetTasklistBudgetListRequestOrderBy

	// ProjectBudgetID is an optional explicit project budget filter. It usually
	// matches the project budget identifier in the request path.
	ProjectBudgetID int64

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of entries to retrieve per page. Defaults to 50.
	PageSize int64

	// Include specifies sideloaded entities to include in the response.
	Include []ProjectBudgetTasklistBudgetListRequestSideload

	// Fields specifies field filtering for sideloaded entities.
	Fields ProjectBudgetTasklistBudgetListRequestFields
}

// ProjectBudgetTasklistBudgetListRequest represents the request for listing
// tasklist budgets under a project budget.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/budgets/get-projects-api-v3-projects-budgets-id-tasklists-budgets-json
type ProjectBudgetTasklistBudgetListRequest struct {
	// Path contains path parameters for the request.
	Path ProjectBudgetTasklistBudgetListRequestPath

	// Filters contains query string filters for the request.
	Filters ProjectBudgetTasklistBudgetListRequestFilters
}

// NewProjectBudgetTasklistBudgetListRequest creates a new
// ProjectBudgetTasklistBudgetListRequest with default values.
func NewProjectBudgetTasklistBudgetListRequest(projectBudgetID int64) ProjectBudgetTasklistBudgetListRequest {
	return ProjectBudgetTasklistBudgetListRequest{
		Path: ProjectBudgetTasklistBudgetListRequestPath{ProjectBudgetID: projectBudgetID},
		Filters: ProjectBudgetTasklistBudgetListRequestFilters{
			OrderMode: twapi.OrderModeAscending,
			OrderBy:   ProjectBudgetTasklistBudgetListRequestOrderByDateCreated,
			Page:      1,
			PageSize:  50,
		},
	}
}

// HTTPRequest creates an HTTP request for the
// ProjectBudgetTasklistBudgetListRequest.
func (p ProjectBudgetTasklistBudgetListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/projects/budgets/%d/tasklists/budgets.json", server, p.Path.ProjectBudgetID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	if p.Filters.OrderMode != "" {
		query.Set("orderMode", string(p.Filters.OrderMode))
	}
	if p.Filters.OrderBy != "" {
		query.Set("orderBy", string(p.Filters.OrderBy))
	}
	if p.Filters.ProjectBudgetID > 0 {
		query.Set("projectBudgetId", strconv.FormatInt(p.Filters.ProjectBudgetID, 10))
	}
	if p.Filters.Page > 0 {
		query.Set("page", strconv.FormatInt(p.Filters.Page, 10))
	}
	if p.Filters.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(p.Filters.PageSize, 10))
	}
	if len(p.Filters.Include) > 0 {
		include := make([]string, 0, len(p.Filters.Include))
		for _, sideload := range p.Filters.Include {
			include = append(include, string(sideload))
		}
		query.Set("include", strings.Join(include, ","))
	}
	if len(p.Filters.Fields.Tasklists) > 0 {
		query.Set("fields[tasklists]", strings.Join(p.Filters.Fields.Tasklists, ","))
	}
	if len(p.Filters.Fields.TasklistBudgetNotifications) > 0 {
		query.Set("fields[tasklistBudgetNotifications]", strings.Join(p.Filters.Fields.TasklistBudgetNotifications, ","))
	}
	if len(p.Filters.Fields.ProjectBudgets) > 0 {
		query.Set("fields[projectBudgets]", strings.Join(p.Filters.Fields.ProjectBudgets, ","))
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// ProjectBudgetTasklistBudgetListResponse contains a collection of tasklist
// budgets for a project budget.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/budgets/get-projects-api-v3-projects-budgets-id-tasklists-budgets-json
type ProjectBudgetTasklistBudgetListResponse struct {
	request ProjectBudgetTasklistBudgetListRequest

	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`

	TasklistBudgets []TasklistBudget `json:"tasklistBudgets"`

	Included struct {
		Companies      map[string]Company                    `json:"companies,omitempty"`
		Notifications  map[string]TasklistBudgetNotification `json:"notifications,omitempty"`
		ProjectBudgets map[string]ProjectBudget              `json:"projectBudgets,omitempty"`
		Tasklists      map[string]Tasklist                   `json:"tasklists,omitempty"`
		Teams          map[string]Team                       `json:"teams,omitempty"`
		Users          map[string]User                       `json:"users,omitempty"`
	} `json:"included"`
}

// HandleHTTPResponse handles the HTTP response for the
// ProjectBudgetTasklistBudgetListResponse.
func (p *ProjectBudgetTasklistBudgetListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list project budget tasklist budgets")
	}

	if err := json.NewDecoder(resp.Body).Decode(p); err != nil {
		return fmt.Errorf("failed to decode list project budget tasklist budgets response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (p *ProjectBudgetTasklistBudgetListResponse) SetRequest(req ProjectBudgetTasklistBudgetListRequest) {
	p.request = req
}

// Iterate returns the request set to the next page, if available. If there
// are no more pages, a nil request is returned.
func (p *ProjectBudgetTasklistBudgetListResponse) Iterate() *ProjectBudgetTasklistBudgetListRequest {
	if !p.Meta.Page.HasMore {
		return nil
	}
	req := p.request
	req.Filters.Page++
	return &req
}

// ProjectBudgetTasklistBudgetList retrieves tasklist budgets for a project
// budget using the provided request and returns the response.
func ProjectBudgetTasklistBudgetList(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectBudgetTasklistBudgetListRequest,
) (*ProjectBudgetTasklistBudgetListResponse, error) {
	return twapi.Execute[ProjectBudgetTasklistBudgetListRequest, *ProjectBudgetTasklistBudgetListResponse](ctx, engine, req)
}
