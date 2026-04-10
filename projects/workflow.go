package projects

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var (
	_ twapi.HTTPRequester = (*WorkflowCreateRequest)(nil)
	_ twapi.HTTPResponser = (*WorkflowCreateResponse)(nil)
	_ twapi.HTTPRequester = (*WorkflowUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*WorkflowUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*WorkflowDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*WorkflowDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*WorkflowGetRequest)(nil)
	_ twapi.HTTPResponser = (*WorkflowGetResponse)(nil)
	_ twapi.HTTPRequester = (*WorkflowListRequest)(nil)
	_ twapi.HTTPResponser = (*WorkflowListResponse)(nil)
)

// Workflow represents a configurable process template in Teamwork that defines
// a series of stages through which tasks or projects progress. Workflows help
// teams standardize their processes, automate stage transitions, and maintain
// consistency across projects by providing a structured path from start to
// completion.
//
// More information can be found at:
// https://support.teamwork.com/projects/workflows/create-and-manage-workflows
type Workflow struct {
	// ID is the unique identifier of the workflow.
	ID int64 `json:"id"`

	// Name is the display name of the workflow, providing a short label that
	// identifies the process it represents.
	Name string `json:"name"`

	// Default indicates whether this workflow is the installation-wide default
	// assigned to new projects that do not specify a workflow.
	Default *bool `json:"defaultWorkflow"`
}

// WorkflowCreateRequest represents the request body for creating a new
// workflow.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/workflows/post-projects-api-v3-workflows-json
type WorkflowCreateRequest struct {
	// Name is the display name of the workflow.
	Name string `json:"name"`
}

// NewWorkflowCreateRequest creates a new WorkflowCreateRequest with the
// provided required fields.
func NewWorkflowCreateRequest(name string) WorkflowCreateRequest {
	return WorkflowCreateRequest{
		Name: name,
	}
}

// HTTPRequest creates an HTTP request for the WorkflowCreateRequest.
func (w WorkflowCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/workflows.json"

	payload := struct {
		Workflow WorkflowCreateRequest `json:"workflow"`
	}{Workflow: w}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create workflow request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// WorkflowCreateResponse represents the response body for creating a new
// workflow.
//
// No public docs are available yet.
type WorkflowCreateResponse struct {
	// Workflow is the created workflow.
	Workflow Workflow `json:"workflow"`
}

// HandleHTTPResponse handles the HTTP response for the WorkflowCreateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (w *WorkflowCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create workflow")
	}
	if err := json.NewDecoder(resp.Body).Decode(w); err != nil {
		return fmt.Errorf("failed to decode create workflow response: %w", err)
	}
	if w.Workflow.ID == 0 {
		return fmt.Errorf("create workflow response does not contain a valid identifier")
	}
	return nil
}

// WorkflowCreate creates a new workflow using the provided request and returns
// the response.
func WorkflowCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req WorkflowCreateRequest,
) (*WorkflowCreateResponse, error) {
	return twapi.Execute[WorkflowCreateRequest, *WorkflowCreateResponse](ctx, engine, req)
}

// WorkflowUpdateRequestPath contains the path parameters for updating a
// workflow.
type WorkflowUpdateRequestPath struct {
	// ID is the unique identifier of the workflow to be updated.
	ID int64
}

// WorkflowUpdateRequest represents the request body for updating a workflow.
// Besides the identifier, all other fields are optional. When a field is not
// provided, it will not be modified.
//
// No public docs are available yet.
type WorkflowUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path WorkflowUpdateRequestPath `json:"-"`

	// Name is the display name of the workflow.
	Name *string `json:"name,omitempty"`
}

// NewWorkflowUpdateRequest creates a new WorkflowUpdateRequest with the
// provided workflow ID.
func NewWorkflowUpdateRequest(workflowID int64) WorkflowUpdateRequest {
	return WorkflowUpdateRequest{
		Path: WorkflowUpdateRequestPath{
			ID: workflowID,
		},
	}
}

// HTTPRequest creates an HTTP request for the WorkflowUpdateRequest.
func (w WorkflowUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/workflows/" + strconv.FormatInt(w.Path.ID, 10) + ".json"

	payload := struct {
		Workflow WorkflowUpdateRequest `json:"workflow"`
	}{Workflow: w}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update workflow request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// WorkflowUpdateResponse represents the response body for updating a workflow.
//
// No public docs are available yet.
type WorkflowUpdateResponse struct{}

// HandleHTTPResponse handles the HTTP response for the WorkflowUpdateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (w *WorkflowUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update workflow")
	}
	if err := json.NewDecoder(resp.Body).Decode(w); err != nil {
		return fmt.Errorf("failed to decode update workflow response: %w", err)
	}
	return nil
}

// WorkflowUpdate updates a workflow using the provided request and returns the
// response.
func WorkflowUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req WorkflowUpdateRequest,
) (*WorkflowUpdateResponse, error) {
	return twapi.Execute[WorkflowUpdateRequest, *WorkflowUpdateResponse](ctx, engine, req)
}

// WorkflowDeleteRequestPath contains the path parameters for deleting a
// workflow.
type WorkflowDeleteRequestPath struct {
	// ID is the unique identifier of the workflow to be deleted.
	ID int64
}

// WorkflowDeleteRequest represents the request body for deleting a workflow.
//
// No public docs are available yet.
type WorkflowDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path WorkflowDeleteRequestPath
}

// NewWorkflowDeleteRequest creates a new WorkflowDeleteRequest with the
// provided workflow ID.
func NewWorkflowDeleteRequest(workflowID int64) WorkflowDeleteRequest {
	return WorkflowDeleteRequest{
		Path: WorkflowDeleteRequestPath{
			ID: workflowID,
		},
	}
}

// HTTPRequest creates an HTTP request for the WorkflowDeleteRequest.
func (w WorkflowDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/workflows/" + strconv.FormatInt(w.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// WorkflowDeleteResponse represents the response body for deleting a workflow.
//
// No public docs are available yet.
type WorkflowDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the WorkflowDeleteResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (w *WorkflowDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusNoContent {
		return twapi.NewHTTPError(resp, "failed to delete workflow")
	}
	return nil
}

// WorkflowDelete deletes a workflow using the provided request and returns the
// response.
func WorkflowDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req WorkflowDeleteRequest,
) (*WorkflowDeleteResponse, error) {
	return twapi.Execute[WorkflowDeleteRequest, *WorkflowDeleteResponse](ctx, engine, req)
}

// WorkflowGetRequestPath contains the path parameters for loading a single
// workflow.
type WorkflowGetRequestPath struct {
	// ID is the unique identifier of the workflow to be retrieved.
	ID int64
}

// WorkflowGetRequest represents the request for loading a single workflow.
//
// No public docs are available yet.
type WorkflowGetRequest struct {
	// Path contains the path parameters for the request.
	Path WorkflowGetRequestPath
}

// NewWorkflowGetRequest creates a new WorkflowGetRequest with the provided
// workflow ID.
func NewWorkflowGetRequest(workflowID int64) WorkflowGetRequest {
	return WorkflowGetRequest{
		Path: WorkflowGetRequestPath{
			ID: workflowID,
		},
	}
}

// HTTPRequest creates an HTTP request for the WorkflowGetRequest.
func (w WorkflowGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/workflows/" + strconv.FormatInt(w.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// WorkflowGetResponse contains all the information related to a workflow.
//
// No public docs are available yet.
type WorkflowGetResponse struct {
	// Workflow is the retrieved workflow.
	Workflow Workflow `json:"workflow"`
}

// HandleHTTPResponse handles the HTTP response for the WorkflowGetResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (w *WorkflowGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve workflow")
	}
	if err := json.NewDecoder(resp.Body).Decode(w); err != nil {
		return fmt.Errorf("failed to decode retrieve workflow response: %w", err)
	}
	return nil
}

// WorkflowGet retrieves a single workflow using the provided request and
// returns the response.
func WorkflowGet(
	ctx context.Context,
	engine *twapi.Engine,
	req WorkflowGetRequest,
) (*WorkflowGetResponse, error) {
	return twapi.Execute[WorkflowGetRequest, *WorkflowGetResponse](ctx, engine, req)
}

// WorkflowListRequestFilters contains the filters for loading multiple
// workflows.
type WorkflowListRequestFilters struct {
	// SearchTerm is an optional search term to filter workflows by name.
	SearchTerm string

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of workflows to retrieve per page. Defaults to 50.
	PageSize int64
}

// WorkflowListRequest represents the request for loading multiple workflows.
//
// No public docs are available yet.
type WorkflowListRequest struct {
	// Filters contains the filters for loading multiple workflows.
	Filters WorkflowListRequestFilters
}

// NewWorkflowListRequest creates a new WorkflowListRequest with default values.
func NewWorkflowListRequest() WorkflowListRequest {
	return WorkflowListRequest{
		Filters: WorkflowListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the WorkflowListRequest.
func (w WorkflowListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/workflows.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	if w.Filters.SearchTerm != "" {
		query.Set("searchTerm", w.Filters.SearchTerm)
	}
	if w.Filters.Page > 0 {
		query.Set("page", strconv.FormatInt(w.Filters.Page, 10))
	}
	if w.Filters.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(w.Filters.PageSize, 10))
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// WorkflowListResponse contains information about multiple workflows matching
// the request filters.
//
// No public docs are available yet.
type WorkflowListResponse struct {
	request WorkflowListRequest

	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`

	// Workflows is the list of workflows matching the request filters.
	Workflows []Workflow `json:"workflows"`
}

// HandleHTTPResponse handles the HTTP response for the WorkflowListResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (w *WorkflowListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list workflows")
	}
	if err := json.NewDecoder(resp.Body).Decode(w); err != nil {
		return fmt.Errorf("failed to decode list workflows response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (w *WorkflowListResponse) SetRequest(req WorkflowListRequest) {
	w.request = req
}

// Iterate returns the request set to the next page, if available. If there are
// no more pages, a nil request is returned.
func (w *WorkflowListResponse) Iterate() *WorkflowListRequest {
	if !w.Meta.Page.HasMore {
		return nil
	}
	req := w.request
	req.Filters.Page++
	return &req
}

// WorkflowList retrieves multiple workflows using the provided request and
// returns the response.
func WorkflowList(
	ctx context.Context,
	engine *twapi.Engine,
	req WorkflowListRequest,
) (*WorkflowListResponse, error) {
	return twapi.Execute[WorkflowListRequest, *WorkflowListResponse](ctx, engine, req)
}
