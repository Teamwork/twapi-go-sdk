package projects_test

import (
	"context"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestWorkflowProjectLink(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	projectID, projectCleanup, err := createProject(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(projectCleanup)

	tests := []struct {
		name  string
		input projects.WorkflowProjectLinkRequest
	}{{
		name:  "only required fields",
		input: projects.NewWorkflowProjectLinkRequest(testResources.WorkflowID, projectID),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.WorkflowProjectLink(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
				return
			}
		})
	}
}
