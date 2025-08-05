package projects

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var (
	_ twapi.HTTPRequester = (*TasklistCreateRequest)(nil)
	_ twapi.HTTPResponser = (*TasklistCreateResponse)(nil)
	_ twapi.HTTPRequester = (*TasklistUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*TasklistUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*TasklistDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*TasklistDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*TasklistGetRequest)(nil)
	_ twapi.HTTPResponser = (*TasklistGetResponse)(nil)
	_ twapi.HTTPRequester = (*TasklistListRequest)(nil)
	_ twapi.HTTPResponser = (*TasklistListResponse)(nil)
)

// Tasklist is a way to group related tasks within a project, helping teams
// organize their work into meaningful sections such as phases, categories, or
// deliverables. Each task list belongs to a specific project and can include
// multiple tasks that are typically aligned with a common goal. Task lists can
// be associated with milestones, and they support privacy settings that control
// who can view or interact with the tasks they contain. This structure helps
// teams manage progress, assign responsibilities, and maintain clarity across
// complex projects.
//
// More information can be found at:
// https://support.teamwork.com/projects/getting-started/task-lists-overview
type Tasklist struct {
	// ID is the unique identifier of the tasklist.
	ID int64 `json:"id"`

	// Name is the name of the tasklist.
	Name string `json:"name"`

	// Description is the description of the tasklist.
	Description string `json:"description"`

	// Project is the project associated with the tasklist.
	Project twapi.Relationship `json:"project"`

	// Milestone is the milestone associated with the tasklist.
	Milestone *twapi.Relationship `json:"milestone"`

	// CreatedAt is the date and time when the tasklist was created.
	CreatedAt *time.Time `json:"createdAt"`

	// UpdatedAt is the date and time when the tasklist was last updated.
	UpdatedAt *time.Time `json:"updatedAt"`

	// Status is the status of the tasklist. It can be "new", "reopened",
	// "completed" or "deleted".
	Status string `json:"status"`
}

// TasklistUpdateRequestPath contains the path parameters for creating a
// tasklist.
type TasklistCreateRequestPath struct {
	// ProjectID is the unique identifier of the project that will contain the
	// tasklist.
	ProjectID int64
}

// TasklistCreateRequest represents the request body for creating a new
// tasklist.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/task-lists/post-projects-id-tasklists-json
type TasklistCreateRequest struct {
	// Path contains the path parameters for the request.
	Path TasklistCreateRequestPath `json:"-"`

	// Name is the name of the tasklist
	Name string `json:"name"`

	// Description is an optional description of the tasklist.
	Description *string `json:"description,omitempty"`

	// MilestoneID is an optional ID of the milestone associated with the
	// tasklist.
	MilestoneID *int64 `json:"milestone-Id,omitempty"`
}

// NewTasklistCreateRequest creates a new TasklistCreateRequest with the
// provided name in a specific project.
func NewTasklistCreateRequest(projectID int64, name string) TasklistCreateRequest {
	return TasklistCreateRequest{
		Path: TasklistCreateRequestPath{
			ProjectID: projectID,
		},
		Name: name,
	}
}

// HTTPRequest creates an HTTP request for the TasklistCreateRequest.
func (t TasklistCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/%d/tasklists.json", server, t.Path.ProjectID)

	payload := struct {
		Tasklist TasklistCreateRequest `json:"todo-list"`
	}{Tasklist: t}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create tasklist request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// TasklistCreateResponse represents the response body for creating a new
// tasklist.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/task-lists/post-projects-id-tasklists-json
type TasklistCreateResponse struct {
	// ID is the unique identifier of the created tasklist.
	ID LegacyNumber `json:"tasklistId"`
}

// HandleHTTPResponse handles the HTTP response for the TasklistCreateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (t *TasklistCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create tasklist")
	}
	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode create tasklist response: %w", err)
	}
	if t.ID == 0 {
		return fmt.Errorf("create tasklist response does not contain a valid identifier")
	}
	return nil
}

// TasklistCreate creates a new tasklist using the provided request and returns
// the response.
func TasklistCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req TasklistCreateRequest,
) (*TasklistCreateResponse, error) {
	return twapi.Execute[TasklistCreateRequest, *TasklistCreateResponse](ctx, engine, req)
}

// TasklistUpdateRequestPath contains the path parameters for updating a
// tasklist.
type TasklistUpdateRequestPath struct {
	// ID is the unique identifier of the tasklist to be updated.
	ID int64
}

// TasklistUpdateRequest represents the request body for updating a tasklist.
// Besides the identifier, all other fields are optional. When a field is not
// provided, it will not be modified.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/task-lists/put-tasklists-id-json
type TasklistUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path TasklistUpdateRequestPath `json:"-"`

	// Name is the name of the tasklist.
	Name *string `json:"name,omitempty"`

	// Description is the tasklist description.
	Description *string `json:"description,omitempty"`

	// MilestoneID is the ID of the milestone associated with the tasklist.
	MilestoneID *int64 `json:"milestone-Id,omitempty"`
}

// NewTasklistUpdateRequest creates a new TasklistUpdateRequest with the
// provided tasklist ID. The ID is required to update a tasklist.
func NewTasklistUpdateRequest(tasklistID int64) TasklistUpdateRequest {
	return TasklistUpdateRequest{
		Path: TasklistUpdateRequestPath{
			ID: tasklistID,
		},
	}
}

// HTTPRequest creates an HTTP request for the TasklistUpdateRequest.
func (t TasklistUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/tasklists/" + strconv.FormatInt(t.Path.ID, 10) + ".json"

	payload := struct {
		Tasklist TasklistUpdateRequest `json:"todo-list"`
	}{Tasklist: t}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update tasklist request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// TasklistUpdateResponse represents the response body for updating a tasklist.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/task-lists/put-tasklists-id-json
type TasklistUpdateResponse struct{}

// HandleHTTPResponse handles the HTTP response for the TasklistUpdateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (t *TasklistUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update tasklist")
	}
	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode update tasklist response: %w", err)
	}
	return nil
}

// TasklistUpdate updates a tasklist using the provided request and returns the
// response.
func TasklistUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req TasklistUpdateRequest,
) (*TasklistUpdateResponse, error) {
	return twapi.Execute[TasklistUpdateRequest, *TasklistUpdateResponse](ctx, engine, req)
}

// TasklistDeleteRequestPath contains the path parameters for deleting a
// tasklist.
type TasklistDeleteRequestPath struct {
	// ID is the unique identifier of the tasklist to be deleted.
	ID int64
}

// TasklistDeleteRequest represents the request body for deleting a tasklist.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/task-lists/delete-tasklists-id-json
type TasklistDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path TasklistDeleteRequestPath
}

// NewTasklistDeleteRequest creates a new TasklistDeleteRequest with the
// provided tasklist ID.
func NewTasklistDeleteRequest(tasklistID int64) TasklistDeleteRequest {
	return TasklistDeleteRequest{
		Path: TasklistDeleteRequestPath{
			ID: tasklistID,
		},
	}
}

// HTTPRequest creates an HTTP request for the TasklistDeleteRequest.
func (t TasklistDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/tasklists/" + strconv.FormatInt(t.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// TasklistDeleteResponse represents the response body for deleting a tasklist.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/task-lists/delete-tasklists-id-json
type TasklistDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the TasklistDeleteResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (t *TasklistDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to delete tasklist")
	}
	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode delete tasklist response: %w", err)
	}
	return nil
}

// TasklistDelete deletes a tasklist using the provided request and returns the
// response.
func TasklistDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req TasklistDeleteRequest,
) (*TasklistDeleteResponse, error) {
	return twapi.Execute[TasklistDeleteRequest, *TasklistDeleteResponse](ctx, engine, req)
}

// TasklistGetRequestPath contains the path parameters for loading a single
// tasklist.
type TasklistGetRequestPath struct {
	// ID is the unique identifier of the tasklist to be retrieved.
	ID int64 `json:"id"`
}

// TasklistGetRequest represents the request body for loading a single tasklist.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/task-lists/get-projects-api-v3-tasklists-tasklist-id
type TasklistGetRequest struct {
	// Path contains the path parameters for the request.
	Path TasklistGetRequestPath
}

// NewTasklistGetRequest creates a new TasklistGetRequest with the provided
// tasklist ID. The ID is required to load a tasklist.
func NewTasklistGetRequest(tasklistID int64) TasklistGetRequest {
	return TasklistGetRequest{
		Path: TasklistGetRequestPath{
			ID: tasklistID,
		},
	}
}

// HTTPRequest creates an HTTP request for the TasklistGetRequest.
func (t TasklistGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/tasklists/" + strconv.FormatInt(t.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// TasklistGetResponse contains all the information related to a tasklist.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/task-lists/get-projects-api-v3-tasklists-tasklist-id
type TasklistGetResponse struct {
	Tasklist Tasklist `json:"tasklist"`
}

// HandleHTTPResponse handles the HTTP response for the TasklistGetResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (t *TasklistGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve tasklist")
	}

	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode retrieve tasklist response: %w", err)
	}
	return nil
}

// TasklistGet retrieves a single tasklist using the provided request and
// returns the response.
func TasklistGet(
	ctx context.Context,
	engine *twapi.Engine,
	req TasklistGetRequest,
) (*TasklistGetResponse, error) {
	return twapi.Execute[TasklistGetRequest, *TasklistGetResponse](ctx, engine, req)
}

// TasklistListRequestPath contains the path parameters for loading multiple
// tasklists.
type TasklistListRequestPath struct {
	// ProjectID is the unique identifier of the project whose tasklists are to be
	// retrieved.
	ProjectID int64
}

// TasklistListRequestFilters contains the filters for loading multiple
// tasklists.
type TasklistListRequestFilters struct {
	// SearchTerm is an optional search term to filter tasklists by name.
	SearchTerm string

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of tasklists to retrieve per page. Defaults to 50.
	PageSize int64
}

// TasklistListRequest represents the request body for loading multiple tasklists.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/task-lists/get-projects-api-v3-tasklists
// https://apidocs.teamwork.com/docs/teamwork/v3/task-lists/get-projects-api-v3-projects-project-id-tasklists
type TasklistListRequest struct {
	// Path contains the path parameters for the request.
	Path TasklistListRequestPath

	// Filters contains the filters for loading multiple tasklists.
	Filters TasklistListRequestFilters
}

// NewTasklistListRequest creates a new TasklistListRequest with default values.
func NewTasklistListRequest() TasklistListRequest {
	return TasklistListRequest{
		Filters: TasklistListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the TasklistListRequest.
func (t TasklistListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	var uri string
	switch {
	case t.Path.ProjectID > 0:
		uri = fmt.Sprintf("%s/projects/api/v3/projects/%d/tasklists.json", server, t.Path.ProjectID)
	default:
		uri = server + "/projects/api/v3/tasklists.json"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	if t.Filters.SearchTerm != "" {
		query.Set("searchTerm", t.Filters.SearchTerm)
	}
	if t.Filters.Page > 0 {
		query.Set("page", strconv.FormatInt(t.Filters.Page, 10))
	}
	if t.Filters.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(t.Filters.PageSize, 10))
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// TasklistListResponse contains information by multiple tasklists matching the
// request filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/task-lists/get-projects-api-v3-tasklists
// https://apidocs.teamwork.com/docs/teamwork/v3/task-lists/get-projects-api-v3-projects-project-id-tasklists
type TasklistListResponse struct {
	request TasklistListRequest

	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`
	Tasklists []Tasklist `json:"tasklists"`
}

// HandleHTTPResponse handles the HTTP response for the TasklistListResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (t *TasklistListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list tasklists")
	}

	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode list tasklists response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (t *TasklistListResponse) SetRequest(req TasklistListRequest) {
	t.request = req
}

// Iterate returns the request set to the next page, if available. If there
// are no more pages, a nil request is returned.
func (t *TasklistListResponse) Iterate() *TasklistListRequest {
	if !t.Meta.Page.HasMore {
		return nil
	}
	req := t.request
	req.Filters.Page++
	return &req
}

// TasklistList retrieves multiple tasklists using the provided request and
// returns the response.
func TasklistList(
	ctx context.Context,
	engine *twapi.Engine,
	req TasklistListRequest,
) (*TasklistListResponse, error) {
	return twapi.Execute[TasklistListRequest, *TasklistListResponse](ctx, engine, req)
}
