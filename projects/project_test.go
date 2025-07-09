package projects_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestProjectCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name          string
		input         projects.ProjectCreateRequest
		expectedError bool
	}{{
		name: "it should create a project with valid input",
		input: projects.ProjectCreateRequest{
			Name: fmt.Sprintf("Test Project %d", time.Now().UnixNano()),
		},
	}, {
		name: "it should fail to create a project with missing name",
		input: projects.ProjectCreateRequest{
			Description: twapi.Ptr("This project has no name"),
		},
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			project, err := projects.ProjectCreate(ctx, engine, tt.input)
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %s", err)
				return
			}
			if project.ID == 0 {
				t.Error("expected a valid project ID but got 0")
			}
		})
	}
}
