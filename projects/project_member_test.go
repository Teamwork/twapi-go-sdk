package projects_test

import (
	"context"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestProjectMemberAdd(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	projectID, projectCleanup, err := createProject(t)
	if err != nil {
		t.Fatal(err)
	}
	defer projectCleanup()

	userID, userCleanup, err := createUser(t)
	if err != nil {
		t.Fatal(err)
	}
	defer userCleanup()

	tests := []struct {
		name          string
		input         projects.ProjectMemberAddRequest
		expectedError bool
	}{{
		name:  "it should add a user to a project",
		input: projects.NewProjectMemberAddRequest(projectID, userID),
	}, {
		name:          "it should fail to an unknown user to a project",
		input:         projects.NewProjectMemberAddRequest(projectID, 0),
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.ProjectMemberAdd(ctx, engine, tt.input)
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
		})
	}
}
