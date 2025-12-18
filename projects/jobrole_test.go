package projects_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestJobRoleCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name  string
		input projects.JobRoleCreateRequest
	}{{
		name: "only required fields",
		input: projects.NewJobRoleCreateRequest(
			fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
		),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			t.Cleanup(cancel)

			jobRoleResponse, err := projects.JobRoleCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx = context.Background() // t.Context is always canceled in cleanup
				_, err := projects.JobRoleDelete(ctx, engine, projects.NewJobRoleDeleteRequest(jobRoleResponse.JobRole.ID))
				if err != nil {
					t.Errorf("failed to delete job role after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if jobRoleResponse.JobRole.ID == 0 {
				t.Error("expected a valid job role ID but got 0")
			}
		})
	}
}

func TestJobRoleUpdate(t *testing.T) {
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
		input projects.JobRoleUpdateRequest
	}{{
		name: "all fields",
		input: projects.JobRoleUpdateRequest{
			Path: projects.JobRoleUpdateRequestPath{
				ID: jobRoleID,
			},
			Name: twapi.Ptr(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.JobRoleUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestJobRoleDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	jobRoleID, _, err := createJobRole(t)
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.JobRoleDelete(ctx, engine, projects.NewJobRoleDeleteRequest(jobRoleID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestJobRoleGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	jobRoleID, jobRoleCleanup, err := createJobRole(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(jobRoleCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.JobRoleGet(ctx, engine, projects.NewJobRoleGetRequest(jobRoleID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestJobRoleList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	_, jobRoleCleanup, err := createJobRole(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(jobRoleCleanup)

	tests := []struct {
		name          string
		input         projects.JobRoleListRequest
		expectedError bool
	}{{
		name: "all jobRoles",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.JobRoleList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
