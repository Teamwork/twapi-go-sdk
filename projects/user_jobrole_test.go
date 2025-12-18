package projects_test

import (
	"context"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestUserAssignJobRole(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	jobRoleID, jobRoleCleanup, err := createJobRole(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(jobRoleCleanup)

	tests := []struct {
		name  string
		input projects.UserAssignJobRoleRequest
	}{{
		name: "all fields",
		input: projects.UserAssignJobRoleRequest{
			Path: projects.UserAssignJobRoleRequestPath{
				ID: jobRoleID,
			},
			IDs:       []int64{testResources.UserID},
			IsPrimary: true,
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.UserAssignJobRole(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestUserUnassignJobRole(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	jobRoleID, jobRoleCleanup, err := createJobRole(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(jobRoleCleanup)

	_, err = projects.UserAssignJobRole(t.Context(), engine, projects.UserAssignJobRoleRequest{
		Path:      projects.UserAssignJobRoleRequestPath{ID: jobRoleID},
		IDs:       []int64{testResources.UserID},
		IsPrimary: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	tests := []struct {
		name  string
		input projects.UserUnassignJobRoleRequest
	}{{
		name: "all fields",
		input: projects.UserUnassignJobRoleRequest{
			Path: projects.UserUnassignJobRoleRequestPath{
				ID: jobRoleID,
			},
			IDs:       []int64{testResources.UserID},
			IsPrimary: true,
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.UserUnassignJobRole(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
