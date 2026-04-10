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
	_ twapi.HTTPRequester = (*WorkflowStageCreateRequest)(nil)
	_ twapi.HTTPResponser = (*WorkflowStageCreateResponse)(nil)
	_ twapi.HTTPRequester = (*WorkflowStageUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*WorkflowStageUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*WorkflowStageDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*WorkflowStageDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*WorkflowStageGetRequest)(nil)
	_ twapi.HTTPResponser = (*WorkflowStageGetResponse)(nil)
	_ twapi.HTTPRequester = (*WorkflowStageListRequest)(nil)
	_ twapi.HTTPResponser = (*WorkflowStageListResponse)(nil)
)

// WorkflowStage represents a single step within a Workflow. Stages are
// ordered and define the progression path for tasks or projects as they move
// through the workflow from start to completion.
//
// More information can be found at:
// https://support.teamwork.com/projects/using-views/workflow-board
type WorkflowStage struct {
	// ID is the unique identifier of the workflow stage.
	ID int64 `json:"id"`

	// Name is the display name of the stage as it appears on the workflow
	// board.
	Name string `json:"name"`

	// Workflow is the workflow that owns this stage.
	Workflow twapi.Relationship `json:"workflow"`
}

// WorkflowStageCreateRequestPath contains the path parameters for creating a
// workflow stage.
type WorkflowStageCreateRequestPath struct {
	// WorkflowID is the unique identifier of the workflow that will own the
	// new stage.
	WorkflowID int64
}

// WorkflowStageCreateRequest represents the request body for creating a new
// workflow stage.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/workflows/post-projects-api-v3-workflows-workflow-id-stages-json
type WorkflowStageCreateRequest struct {
	// Path contains the path parameters for the request.
	Path WorkflowStageCreateRequestPath `json:"-"`

	// Name is the display name of the stage.
	Name string `json:"name"`
}

// NewWorkflowStageCreateRequest creates a new WorkflowStageCreateRequest with
// the provided required fields.
func NewWorkflowStageCreateRequest(workflowID int64, name string) WorkflowStageCreateRequest {
	return WorkflowStageCreateRequest{
		Path: WorkflowStageCreateRequestPath{
			WorkflowID: workflowID,
		},
		Name: name,
	}
}

// HTTPRequest creates an HTTP request for the WorkflowStageCreateRequest.
func (w WorkflowStageCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/workflows/%d/stages.json", server, w.Path.WorkflowID)

	payload := struct {
		Stage WorkflowStageCreateRequest `json:"stage"`
	}{Stage: w}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create workflow stage request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// WorkflowStageCreateResponse represents the response body for creating a new
// workflow stage.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/workflows/post-projects-api-v3-workflows-workflow-id-stages-json
type WorkflowStageCreateResponse struct {
	// Stage is the created workflow stage.
	Stage WorkflowStage `json:"stage"`
}

// HandleHTTPResponse handles the HTTP response for the
// WorkflowStageCreateResponse. If some unexpected HTTP status code is returned
// by the API, a twapi.HTTPError is returned.
func (w *WorkflowStageCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create workflow stage")
	}
	if err := json.NewDecoder(resp.Body).Decode(w); err != nil {
		return fmt.Errorf("failed to decode create workflow stage response: %w", err)
	}
	if w.Stage.ID == 0 {
		return fmt.Errorf("create workflow stage response does not contain a valid identifier")
	}
	return nil
}

// WorkflowStageCreate creates a new workflow stage using the provided request
// and returns the response.
func WorkflowStageCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req WorkflowStageCreateRequest,
) (*WorkflowStageCreateResponse, error) {
	return twapi.Execute[WorkflowStageCreateRequest, *WorkflowStageCreateResponse](ctx, engine, req)
}

// WorkflowStageUpdateRequestPath contains the path parameters for updating a
// workflow stage.
type WorkflowStageUpdateRequestPath struct {
	// WorkflowID is the unique identifier of the workflow that owns the stage.
	WorkflowID int64

	// ID is the unique identifier of the workflow stage to be updated.
	ID int64
}

// WorkflowStageUpdateRequest represents the request body for updating a
// workflow stage. Besides the identifiers, all other fields are optional. When
// a field is not provided, it will not be modified.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/workflows/patch-projects-api-v3-workflows-workflow-id-stages-stage-id-json
//
//nolint:lll
type WorkflowStageUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path WorkflowStageUpdateRequestPath `json:"-"`

	// Name is the display name of the stage.
	Name *string `json:"name,omitempty"`
}

// NewWorkflowStageUpdateRequest creates a new WorkflowStageUpdateRequest with
// the provided workflow and stage IDs.
func NewWorkflowStageUpdateRequest(workflowID, stageID int64) WorkflowStageUpdateRequest {
	return WorkflowStageUpdateRequest{
		Path: WorkflowStageUpdateRequestPath{
			WorkflowID: workflowID,
			ID:         stageID,
		},
	}
}

// HTTPRequest creates an HTTP request for the WorkflowStageUpdateRequest.
func (w WorkflowStageUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/workflows/%d/stages/%d.json", server, w.Path.WorkflowID, w.Path.ID)

	payload := struct {
		Stage WorkflowStageUpdateRequest `json:"stage"`
	}{Stage: w}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update workflow stage request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// WorkflowStageUpdateResponse represents the response body for updating a
// workflow stage.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/workflows/patch-projects-api-v3-workflows-workflow-id-stages-stage-id-json
//
//nolint:lll
type WorkflowStageUpdateResponse struct{}

// HandleHTTPResponse handles the HTTP response for the
// WorkflowStageUpdateResponse. If some unexpected HTTP status code is returned
// by the API, a twapi.HTTPError is returned.
func (w *WorkflowStageUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update workflow stage")
	}
	if err := json.NewDecoder(resp.Body).Decode(w); err != nil {
		return fmt.Errorf("failed to decode update workflow stage response: %w", err)
	}
	return nil
}

// WorkflowStageUpdate updates a workflow stage using the provided request and
// returns the response.
func WorkflowStageUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req WorkflowStageUpdateRequest,
) (*WorkflowStageUpdateResponse, error) {
	return twapi.Execute[WorkflowStageUpdateRequest, *WorkflowStageUpdateResponse](ctx, engine, req)
}

// WorkflowStageDeleteRequestPath contains the path parameters for deleting a
// workflow stage.
type WorkflowStageDeleteRequestPath struct {
	// WorkflowID is the unique identifier of the workflow that owns the stage.
	WorkflowID int64

	// ID is the unique identifier of the workflow stage to be deleted.
	ID int64
}

// WorkflowStageDeleteRequest represents the request for deleting a workflow
// stage.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/workflows/delete-projects-api-v3-workflows-workflow-id-stages-stage-id-json
//
//nolint:lll
type WorkflowStageDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path WorkflowStageDeleteRequestPath `json:"-"`

	// MapTasksToStageID is the unique identifier of the stage to which the tasks
	// in the deleted stage will be moved. If zeroed, the tasks will go back to
	// the backlog.
	MapTasksToStageID int64 `json:"mapTasksToStageId"`
}

// NewWorkflowStageDeleteRequest creates a new WorkflowStageDeleteRequest with
// the provided workflow and stage IDs.
func NewWorkflowStageDeleteRequest(workflowID, stageID int64) WorkflowStageDeleteRequest {
	return WorkflowStageDeleteRequest{
		Path: WorkflowStageDeleteRequestPath{
			WorkflowID: workflowID,
			ID:         stageID,
		},
	}
}

// HTTPRequest creates an HTTP request for the WorkflowStageDeleteRequest.
func (w WorkflowStageDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/workflows/%d/stages/%d.json", server, w.Path.WorkflowID, w.Path.ID)

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(w); err != nil {
		return nil, fmt.Errorf("failed to encode update workflow stage request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, &body)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// WorkflowStageDeleteResponse represents the response body for deleting a
// workflow stage.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/workflows/delete-projects-api-v3-workflows-workflow-id-stages-stage-id-json
//
//nolint:lll
type WorkflowStageDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the
// WorkflowStageDeleteResponse. If some unexpected HTTP status code is returned
// by the API, a twapi.HTTPError is returned.
func (w *WorkflowStageDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusNoContent {
		return twapi.NewHTTPError(resp, "failed to delete workflow stage")
	}
	return nil
}

// WorkflowStageDelete deletes a workflow stage using the provided request and
// returns the response.
func WorkflowStageDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req WorkflowStageDeleteRequest,
) (*WorkflowStageDeleteResponse, error) {
	return twapi.Execute[WorkflowStageDeleteRequest, *WorkflowStageDeleteResponse](ctx, engine, req)
}

// WorkflowStageGetRequestPath contains the path parameters for loading a
// single workflow stage.
type WorkflowStageGetRequestPath struct {
	// WorkflowID is the unique identifier of the workflow that owns the stage.
	WorkflowID int64

	// ID is the unique identifier of the workflow stage to be retrieved.
	ID int64
}

// WorkflowStageGetRequest represents the request for loading a single workflow
// stage.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/workflows/get-projects-api-v3-workflows-workflow-id-stages-stage-id-json
//
//nolint:lll
type WorkflowStageGetRequest struct {
	// Path contains the path parameters for the request.
	Path WorkflowStageGetRequestPath
}

// NewWorkflowStageGetRequest creates a new WorkflowStageGetRequest with the
// provided workflow and stage IDs.
func NewWorkflowStageGetRequest(workflowID, stageID int64) WorkflowStageGetRequest {
	return WorkflowStageGetRequest{
		Path: WorkflowStageGetRequestPath{
			WorkflowID: workflowID,
			ID:         stageID,
		},
	}
}

// HTTPRequest creates an HTTP request for the WorkflowStageGetRequest.
func (w WorkflowStageGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/workflows/%d/stages/%d.json", server, w.Path.WorkflowID, w.Path.ID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// WorkflowStageGetResponse contains all the information related to a workflow
// stage.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/workflows/get-projects-api-v3-workflows-workflow-id-stages-stage-id-json
//
//nolint:lll
type WorkflowStageGetResponse struct {
	// Stage is the retrieved workflow stage.
	Stage WorkflowStage `json:"stage"`
}

// HandleHTTPResponse handles the HTTP response for the WorkflowStageGetResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (w *WorkflowStageGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve workflow stage")
	}
	if err := json.NewDecoder(resp.Body).Decode(w); err != nil {
		return fmt.Errorf("failed to decode retrieve workflow stage response: %w", err)
	}
	return nil
}

// WorkflowStageGet retrieves a single workflow stage using the provided request
// and returns the response.
func WorkflowStageGet(
	ctx context.Context,
	engine *twapi.Engine,
	req WorkflowStageGetRequest,
) (*WorkflowStageGetResponse, error) {
	return twapi.Execute[WorkflowStageGetRequest, *WorkflowStageGetResponse](ctx, engine, req)
}

// WorkflowStageListRequestPath contains the path parameters for loading
// multiple workflow stages.
type WorkflowStageListRequestPath struct {
	// WorkflowID is the unique identifier of the workflow whose stages are to
	// be listed.
	WorkflowID int64
}

// WorkflowStageListRequestFilters contains the filters for loading multiple
// workflow stages.
type WorkflowStageListRequestFilters struct {
	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of stages to retrieve per page. Defaults to 50.
	PageSize int64
}

// WorkflowStageListRequest represents the request for loading multiple
// workflow stages belonging to a workflow.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/workflows/get-projects-api-v3-workflows-workflow-id-stages-json
type WorkflowStageListRequest struct {
	// Path contains the path parameters for the request.
	Path WorkflowStageListRequestPath

	// Filters contains the filters for loading multiple workflow stages.
	Filters WorkflowStageListRequestFilters
}

// NewWorkflowStageListRequest creates a new WorkflowStageListRequest with the
// provided workflow ID and default filter values.
func NewWorkflowStageListRequest(workflowID int64) WorkflowStageListRequest {
	return WorkflowStageListRequest{
		Path: WorkflowStageListRequestPath{
			WorkflowID: workflowID,
		},
		Filters: WorkflowStageListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the WorkflowStageListRequest.
func (w WorkflowStageListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/workflows/%d/stages.json", server, w.Path.WorkflowID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	if w.Filters.Page > 0 {
		query.Set("page", strconv.FormatInt(w.Filters.Page, 10))
	}
	if w.Filters.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(w.Filters.PageSize, 10))
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// WorkflowStageListResponse contains information about multiple workflow stages
// matching the request filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/workflows/get-projects-api-v3-workflows-workflow-id-stages-json
type WorkflowStageListResponse struct {
	request WorkflowStageListRequest

	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`

	// Stages is the list of workflow stages matching the request filters.
	Stages []WorkflowStage `json:"stages"`
}

// HandleHTTPResponse handles the HTTP response for the
// WorkflowStageListResponse. If some unexpected HTTP status code is returned by
// the API, a twapi.HTTPError is returned.
func (w *WorkflowStageListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list workflow stages")
	}
	if err := json.NewDecoder(resp.Body).Decode(w); err != nil {
		return fmt.Errorf("failed to decode list workflow stages response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (w *WorkflowStageListResponse) SetRequest(req WorkflowStageListRequest) {
	w.request = req
}

// Iterate returns the request set to the next page, if available. If there are
// no more pages, a nil request is returned.
func (w *WorkflowStageListResponse) Iterate() *WorkflowStageListRequest {
	if !w.Meta.Page.HasMore {
		return nil
	}
	req := w.request
	req.Filters.Page++
	return &req
}

// WorkflowStageList retrieves multiple workflow stages using the provided
// request and returns the response.
func WorkflowStageList(
	ctx context.Context,
	engine *twapi.Engine,
	req WorkflowStageListRequest,
) (*WorkflowStageListResponse, error) {
	return twapi.Execute[WorkflowStageListRequest, *WorkflowStageListResponse](ctx, engine, req)
}
