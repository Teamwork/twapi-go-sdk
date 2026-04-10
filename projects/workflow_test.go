package projects_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestWorkflowCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name  string
		input projects.WorkflowCreateRequest
	}{{
		name:  "only required fields",
		input: projects.NewWorkflowCreateRequest(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
	}, {
		name: "all fields",
		input: projects.WorkflowCreateRequest{
			Name: fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			workflow, err := projects.WorkflowCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx = context.Background() // t.Context is always canceled in cleanup
				_, err := projects.WorkflowDelete(ctx, engine, projects.NewWorkflowDeleteRequest(workflow.Workflow.ID))
				if err != nil {
					t.Errorf("failed to delete workflow after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if workflow.Workflow.ID == 0 {
				t.Error("expected a valid workflow ID but got 0")
			}
		})
	}
}

func TestWorkflowUpdate(t *testing.T) {
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
		input projects.WorkflowUpdateRequest
	}{{
		name: "all fields",
		input: projects.WorkflowUpdateRequest{
			Path: projects.WorkflowUpdateRequestPath{
				ID: workflowID,
			},
			Name: new(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.WorkflowUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestWorkflowDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	workflowID, _, err := createWorkflow(t)
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.WorkflowDelete(ctx, engine, projects.NewWorkflowDeleteRequest(workflowID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestWorkflowGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	workflowID, workflowCleanup, err := createWorkflow(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(workflowCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.WorkflowGet(ctx, engine, projects.NewWorkflowGetRequest(workflowID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestWorkflowList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	_, workflowCleanup, err := createWorkflow(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(workflowCleanup)

	tests := []struct {
		name  string
		input projects.WorkflowListRequest
	}{{
		name: "all workflows",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.WorkflowList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
