package projects_test

import (
	"context"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestWorkflowStageTaskMove(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

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
