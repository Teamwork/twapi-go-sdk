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
	_ twapi.HTTPRequester = (*WorkflowStageTaskMoveRequest)(nil)
	_ twapi.HTTPResponser = (*WorkflowStageTaskMoveResponse)(nil)
)

// WorkflowStageTaskMoveRequestPath contains the path parameters for moving
// tasks to a workflow stage.
type WorkflowStageTaskMoveRequestPath struct {
	// WorkflowID is the identifier of the workflow that contains the stage to
	// which the task will be moved.
	WorkflowID int64 `json:"workflowId"`

	// TaskID is the identifier of the task to be moved to the workflow stage.
	TaskID int64 `json:"taskId"`
}

// WorkflowStageTaskMoveRequest represents the request body for moving tasks to
// a workflow stage.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/workflows/patch-projects-api-v3-tasks-task-id-workflows-workflow-id-json
//
//nolint:lll
type WorkflowStageTaskMoveRequest struct {
	// Path contains the path parameters for the request.
	Path WorkflowStageTaskMoveRequestPath `json:"-"`

	// StageID is the identifier of the stage to which the task will be moved to.
	StageID int64 `json:"stageId"`

	// PositionAfterTaskID is the identifier of the task after which the current
	// task will be positioned within the stage. If not provided or set as '-1',
	// the task will be moved to the end of the stage.
	PositionAfterTaskID int64 `json:"positionAfterTask,omitempty"`
}

// NewWorkflowStageTaskMoveRequest creates a new WorkflowStageTaskMoveRequest
// with the provided workflow and task IDs.
func NewWorkflowStageTaskMoveRequest(workflowID, stageID, taskID int64) WorkflowStageTaskMoveRequest {
	return WorkflowStageTaskMoveRequest{
		Path: WorkflowStageTaskMoveRequestPath{
			WorkflowID: workflowID,
			TaskID:     taskID,
		},
		StageID: stageID,
	}
}

// HTTPRequest creates an HTTP request for the WorkflowStageTaskMoveRequest.
func (w WorkflowStageTaskMoveRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if w.PositionAfterTaskID == 0 || w.PositionAfterTaskID < -1 {
		// default is to move the task to the end of the stage
		w.PositionAfterTaskID = -1
	}

	uri := fmt.Sprintf("%s/projects/api/v3/tasks/%d/workflows/%d.json", server, w.Path.TaskID, w.Path.WorkflowID)

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(w); err != nil {
		return nil, fmt.Errorf("failed to encode workflow stage task request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// WorkflowStageTaskMoveResponse represents the response body for moving tasks
// to a workflow stage.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/workflows/patch-projects-api-v3-tasks-task-id-workflows-workflow-id-json
//
//nolint:lll
type WorkflowStageTaskMoveResponse struct{}

// HandleHTTPResponse handles the HTTP response for the
// WorkflowStageTaskMoveResponse. If some unexpected HTTP status code is
// returned by the API, a twapi.HTTPError is returned.
func (u *WorkflowStageTaskMoveResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusNoContent {
		return twapi.NewHTTPError(resp, "failed to move task to workflow stage")
	}
	return nil
}

// WorkflowStageTaskMove moves a task to a workflow stage.
func WorkflowStageTaskMove(
	ctx context.Context,
	engine *twapi.Engine,
	req WorkflowStageTaskMoveRequest,
) (*WorkflowStageTaskMoveResponse, error) {
	return twapi.Execute[WorkflowStageTaskMoveRequest, *WorkflowStageTaskMoveResponse](ctx, engine, req)
}
