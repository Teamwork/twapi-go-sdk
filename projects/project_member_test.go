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

	userID, userCleanup, err := createUser(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(userCleanup)

	tests := []struct {
		name  string
		input projects.ProjectMemberAddRequest
	}{{
		name:  "only required fields",
		input: projects.NewProjectMemberAddRequest(testResources.ProjectID, userID),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.ProjectMemberAdd(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
				return
			}
		})
	}
}
