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
	_ twapi.HTTPRequester = (*TasklistBudgetListRequest)(nil)
	_ twapi.HTTPResponser = (*TasklistBudgetListResponse)(nil)
)

// TasklistBudget represents a budget item attached to a tasklist.
type TasklistBudget struct {
	// ID is the unique identifier of the tasklist budget.
	ID int64 `json:"id"`

	// Type is the budget type.
	Type BudgetType `json:"type"`

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

// TasklistBudgetListRequestPath contains the path parameters for listing
// tasklist budgets in a project budget.
type TasklistBudgetListRequestPath struct {
	// ProjectBudgetID is the unique identifier of the parent project budget.
	ProjectBudgetID int64
}

// TasklistBudgetListRequestSideload represents related objects that can be
// included in the response.
type TasklistBudgetListRequestSideload string

// List of valid sideload values for TasklistBudgetListRequest.
const (
	TasklistBudgetListRequestSideloadTasklists                   TasklistBudgetListRequestSideload = "tasklists"
	TasklistBudgetListRequestSideloadProjectBudgets              TasklistBudgetListRequestSideload = "projectBudgets"
	TasklistBudgetListRequestSideloadTasklistBudgetNotifications TasklistBudgetListRequestSideload = "tasklistBudgetNotifications"
)

// TasklistBudgetListRequestOrderBy defines sortable fields for tasklist budget
// listings.
type TasklistBudgetListRequestOrderBy string

// List of valid order by values for TasklistBudgetListRequest.
const (
	TasklistBudgetListRequestOrderByDateCreated TasklistBudgetListRequestOrderBy = "dateCreated"
)

// TasklistBudgetListRequestFields contains field selectors for sideloaded
// entities.
type TasklistBudgetListRequestFields struct {
	// Tasklists limits fields returned for sideloaded tasklists.
	Tasklists []string

	// TasklistBudgetNotifications limits fields returned for sideloaded
	// notifications.
	TasklistBudgetNotifications []string

	// ProjectBudgets limits fields returned for sideloaded project budgets.
	ProjectBudgets []string
}

// TasklistBudgetListRequestFilters contains filters for listing tasklist
// budgets in a project budget.
type TasklistBudgetListRequestFilters struct {
	// OrderMode specifies sort direction. Allowed values are "asc" and "desc".
	OrderMode twapi.OrderMode

	// OrderBy specifies the field used for sorting.
	OrderBy TasklistBudgetListRequestOrderBy

	// ProjectBudgetID is an optional explicit project budget filter. It usually
	// matches the project budget identifier in the request path.
	ProjectBudgetID int64

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of entries to retrieve per page. Defaults to 50.
	PageSize int64

	// Include specifies sideloaded entities to include in the response.
	Include []TasklistBudgetListRequestSideload

	// Fields specifies field filtering for sideloaded entities.
	Fields TasklistBudgetListRequestFields
}

func (p TasklistBudgetListRequestFilters) apply(req *http.Request) {
	query := req.URL.Query()
	if p.OrderMode != "" {
		query.Set("orderMode", string(p.OrderMode))
	}
	if p.OrderBy != "" {
		query.Set("orderBy", string(p.OrderBy))
	}
	if p.ProjectBudgetID > 0 {
		query.Set("projectBudgetId", strconv.FormatInt(p.ProjectBudgetID, 10))
	}
	if p.Page > 0 {
		query.Set("page", strconv.FormatInt(p.Page, 10))
	}
	if p.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(p.PageSize, 10))
	}
	if len(p.Include) > 0 {
		include := make([]string, 0, len(p.Include))
		for _, sideload := range p.Include {
			include = append(include, string(sideload))
		}
		query.Set("include", strings.Join(include, ","))
	}
	if len(p.Fields.Tasklists) > 0 {
		query.Set("fields[tasklists]", strings.Join(p.Fields.Tasklists, ","))
	}
	if len(p.Fields.TasklistBudgetNotifications) > 0 {
		query.Set("fields[tasklistBudgetNotifications]", strings.Join(p.Fields.TasklistBudgetNotifications, ","))
	}
	if len(p.Fields.ProjectBudgets) > 0 {
		query.Set("fields[projectBudgets]", strings.Join(p.Fields.ProjectBudgets, ","))
	}
	req.URL.RawQuery = query.Encode()
}

// TasklistBudgetListRequest represents the request for listing tasklist budgets
// under a project budget.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/budgets/get-projects-api-v3-projects-budgets-id-tasklists-budgets-json
type TasklistBudgetListRequest struct {
	// Path contains path parameters for the request.
	Path TasklistBudgetListRequestPath

	// Filters contains query string filters for the request.
	Filters TasklistBudgetListRequestFilters
}

// NewTasklistBudgetListRequest creates a new TasklistBudgetListRequest with
// default values.
func NewTasklistBudgetListRequest(projectBudgetID int64) TasklistBudgetListRequest {
	return TasklistBudgetListRequest{
		Path: TasklistBudgetListRequestPath{ProjectBudgetID: projectBudgetID},
		Filters: TasklistBudgetListRequestFilters{
			OrderMode: twapi.OrderModeAscending,
			OrderBy:   TasklistBudgetListRequestOrderByDateCreated,
			Page:      1,
			PageSize:  50,
		},
	}
}

// HTTPRequest creates an HTTP request for the TasklistBudgetListRequest.
func (p TasklistBudgetListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/projects/budgets/%d/tasklists/budgets.json", server, p.Path.ProjectBudgetID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	p.Filters.apply(req)

	return req, nil
}

// TasklistBudgetListResponse contains a collection of tasklist budgets for a
// project budget.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/budgets/get-projects-api-v3-projects-budgets-id-tasklists-budgets-json
type TasklistBudgetListResponse struct {
	request TasklistBudgetListRequest

	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`

	TasklistBudgets []TasklistBudget `json:"tasklistBudgets"`

	Included struct {
		Notifications  map[string]TasklistBudgetNotification `json:"notifications,omitempty"`
		ProjectBudgets map[string]ProjectBudget              `json:"projectBudgets,omitempty"`
		Tasklists      map[string]Tasklist                   `json:"tasklists,omitempty"`
	} `json:"included"`
}

// HandleHTTPResponse handles the HTTP response for the
// TasklistBudgetListResponse.
func (p *TasklistBudgetListResponse) HandleHTTPResponse(resp *http.Response) error {
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
func (p *TasklistBudgetListResponse) SetRequest(req TasklistBudgetListRequest) {
	p.request = req
}

// Iterate returns the request set to the next page, if available. If there
// are no more pages, a nil request is returned.
func (p *TasklistBudgetListResponse) Iterate() *TasklistBudgetListRequest {
	if !p.Meta.Page.HasMore {
		return nil
	}
	req := p.request
	req.Filters.Page++
	return &req
}

// TasklistBudgetList retrieves tasklist budgets for a project budget using the
// provided request and returns the response.
func TasklistBudgetList(
	ctx context.Context,
	engine *twapi.Engine,
	req TasklistBudgetListRequest,
) (*TasklistBudgetListResponse, error) {
	return twapi.Execute[TasklistBudgetListRequest, *TasklistBudgetListResponse](ctx, engine, req)
}
