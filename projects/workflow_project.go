package projects

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var (
	_ twapi.HTTPRequester = (*WorkflowProjectLinkRequest)(nil)
	_ twapi.HTTPResponser = (*WorkflowProjectLinkResponse)(nil)
)

// WorkflowProjectLinkRequestPath contains the path parameters for linking a
// project to a workflow.
type WorkflowProjectLinkRequestPath struct {
	// ProjectID is the identifier of the project to be linked to the workflow.
	ProjectID int64 `json:"projectId"`
}

// WorkflowProjectLinkRequest represents the request body for linking a project
// to a workflow.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/workflows/post-projects-api-v3-projects-project-id-workflows-json
type WorkflowProjectLinkRequest struct {
	// Path contains the path parameters for the request.
	Path WorkflowProjectLinkRequestPath `json:"-"`

	// WorkflowID is the identifier of the workflow to which the project will be
	// linked.
	WorkflowID int64 `json:"-"`
}

// NewWorkflowProjectLinkRequest creates a new WorkflowProjectLinkRequest with
// the provided workflow and project IDs.
func NewWorkflowProjectLinkRequest(workflowID, projectID int64) WorkflowProjectLinkRequest {
	return WorkflowProjectLinkRequest{
		Path: WorkflowProjectLinkRequestPath{
			ProjectID: projectID,
		},
		WorkflowID: workflowID,
	}
}

// HTTPRequest creates an HTTP request for the WorkflowProjectLinkRequest.
func (w WorkflowProjectLinkRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/projects/%d/workflows.json", server, w.Path.ProjectID)

	var payload struct {
		Workflow struct {
			ID int64 `json:"id"`
		} `json:"workflow"`
	}
	payload.Workflow.ID = w.WorkflowID

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode workflow project link request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// WorkflowProjectLinkResponse represents the response body for linking a project
// to a workflow.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/workflows/post-projects-api-v3-projects-project-id-workflows-json
type WorkflowProjectLinkResponse struct{}

// HandleHTTPResponse handles the HTTP response for the
// WorkflowProjectLinkResponse. If some unexpected HTTP status code is
// returned by the API, a twapi.HTTPError is returned.
func (w *WorkflowProjectLinkResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update workflow project link")
	}
	if err := json.NewDecoder(resp.Body).Decode(w); err != nil {
		return fmt.Errorf("failed to decode update workflow project link response: %w", err)
	}
	return nil
}

// WorkflowProjectLink links a project to a workflow.
func WorkflowProjectLink(
	ctx context.Context,
	engine *twapi.Engine,
	req WorkflowProjectLinkRequest,
) (*WorkflowProjectLinkResponse, error) {
	return twapi.Execute[WorkflowProjectLinkRequest, *WorkflowProjectLinkResponse](ctx, engine, req)
}
