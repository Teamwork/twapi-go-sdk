package projects_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestWorkflowStageTaskMove(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	taskID, taskCleanup, err := createTask(t, testResources.TasklistID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(taskCleanup)

	tests := []struct {
		name  string
		input projects.WorkflowStageTaskMoveRequest
	}{{
		name: "only required fields",
		input: projects.NewWorkflowStageTaskMoveRequest(
			testResources.WorkflowID,
			testResources.WorkflowStageID,
			testResources.TaskID,
		),
	}, {
		name: "all fields",
		input: projects.WorkflowStageTaskMoveRequest{
			Path: projects.WorkflowStageTaskMoveRequestPath{
				WorkflowID: testResources.WorkflowID,
				TaskID:     taskID,
			},
			StageID:             testResources.WorkflowStageID,
			PositionAfterTaskID: testResources.TaskID,
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.WorkflowStageTaskMove(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
				return
			}
		})
	}
}

func TestWorkflowStageTasksMoveRequestGeneration(t *testing.T) {
	req := projects.NewWorkflowStageTasksMoveRequest(123, 456, 789, 790, 791)

	httpReq, err := req.HTTPRequest(context.Background(), "https://test.com")
	if err != nil {
		t.Fatalf("unexpected error creating HTTP request: %s", err)
	}

	if httpReq.URL.Path != "/projects/api/v3/workflows/123/stages/456/tasks.json" {
		t.Errorf("unexpected request path: %s", httpReq.URL.Path)
	}
	if httpReq.Method != http.MethodPost {
		t.Errorf("expected POST but got %s", httpReq.Method)
	}

	body, err := io.ReadAll(httpReq.Body)
	if err != nil {
		t.Fatalf("failed to read request body: %s", err)
	}

	var payload struct {
		TaskIDs []int64 `json:"taskIds"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("failed to decode request body %q: %s", body, err)
	}

	// the whole set travels in the body, not the path
	if want := []int64{789, 790, 791}; !slices.Equal(payload.TaskIDs, want) {
		t.Errorf("expected taskIds %v but got %v (body %q)", want, payload.TaskIDs, body)
	}
}

func TestWorkflowStageTasksMove(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	firstTaskID, firstTaskCleanup, err := createTask(t, testResources.TasklistID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(firstTaskCleanup)

	secondTaskID, secondTaskCleanup, err := createTask(t, testResources.TasklistID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(secondTaskCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	_, err = projects.WorkflowStageTasksMove(ctx, engine, projects.NewWorkflowStageTasksMoveRequest(
		testResources.WorkflowID,
		testResources.WorkflowStageID,
		firstTaskID,
		secondTaskID,
	))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}
