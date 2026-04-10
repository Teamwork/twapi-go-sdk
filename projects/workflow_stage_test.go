package projects_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestWorkflowStageCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	workflowID, workflowCleanup, err := createWorkflow(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(workflowCleanup)

	tests := []struct {
		name  string
		input projects.WorkflowStageCreateRequest
	}{{
		name: "only required fields",
		input: projects.NewWorkflowStageCreateRequest(
			workflowID,
			fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
		),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			stage, err := projects.WorkflowStageCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx = context.Background() // t.Context is always canceled in cleanup
				_, err := projects.WorkflowStageDelete(ctx, engine,
					projects.NewWorkflowStageDeleteRequest(workflowID, stage.Stage.ID))
				if err != nil {
					t.Errorf("failed to delete workflow stage after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if stage.Stage.ID == 0 {
				t.Error("expected a valid workflow stage ID but got 0")
			}
		})
	}
}

func TestWorkflowStageUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	workflowID, workflowCleanup, err := createWorkflow(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(workflowCleanup)

	stageID, stageCleanup, err := createWorkflowStage(t, workflowID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(stageCleanup)

	tests := []struct {
		name  string
		input projects.WorkflowStageUpdateRequest
	}{{
		name: "all fields",
		input: projects.WorkflowStageUpdateRequest{
			Path: projects.WorkflowStageUpdateRequestPath{
				WorkflowID: workflowID,
				ID:         stageID,
			},
			Name: new(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.WorkflowStageUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestWorkflowStageDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	workflowID, workflowCleanup, err := createWorkflow(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(workflowCleanup)

	stageID, _, err := createWorkflowStage(t, workflowID)
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.WorkflowStageDelete(ctx, engine,
		projects.NewWorkflowStageDeleteRequest(workflowID, stageID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestWorkflowStageGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	workflowID, workflowCleanup, err := createWorkflow(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(workflowCleanup)

	stageID, stageCleanup, err := createWorkflowStage(t, workflowID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(stageCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.WorkflowStageGet(ctx, engine,
		projects.NewWorkflowStageGetRequest(workflowID, stageID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestWorkflowStageList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	workflowID, workflowCleanup, err := createWorkflow(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(workflowCleanup)

	_, stageCleanup, err := createWorkflowStage(t, workflowID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(stageCleanup)

	tests := []struct {
		name  string
		input projects.WorkflowStageListRequest
	}{{
		name: "all stages",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.input.Path.WorkflowID = workflowID

			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.WorkflowStageList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
