package projects

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var (
	_ twapi.HTTPRequester = (*ProjectBudgetListRequest)(nil)
	_ twapi.HTTPResponser = (*ProjectBudgetListResponse)(nil)
)

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
