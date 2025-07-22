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

func TestTasklistCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name  string
		input projects.TasklistCreateRequest
	}{{
		name: "it should create a tasklist with valid input",
		input: projects.NewTasklistCreateRequest(
			testResources.ProjectID,
			fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
		),
	}, {
		name: "all fields",
		input: projects.TasklistCreateRequest{
			Path: projects.TasklistCreateRequestPath{
				ProjectID: testResources.ProjectID,
			},
			Name:        fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
			Description: twapi.Ptr("This is a test tasklist"),
			MilestoneID: &testResources.MilestoneID,
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			t.Cleanup(cancel)

			tasklist, err := projects.TasklistCreate(ctx, engine, tt.input)
			defer func() {
				if err != nil {
					return
				}
				ctx = context.Background() // t.Context is always canceled in cleanup
				_, err := projects.TasklistDelete(ctx, engine, projects.NewTasklistDeleteRequest(int64(tasklist.ID)))
				if err != nil {
					t.Errorf("failed to delete tasklist after test: %s", err)
				}
			}()
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if tasklist.ID == 0 {
				t.Error("expected a valid tasklist ID but got 0")
			}
		})
	}
}

func TestTasklistUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tasklistID, tasklistCleanup, err := createTasklist(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(tasklistCleanup)

	tests := []struct {
		name  string
		input projects.TasklistUpdateRequest
	}{{
		name: "all fields",
		input: projects.TasklistUpdateRequest{
			Path: projects.TasklistUpdateRequestPath{
				ID: tasklistID,
			},
			Name:        twapi.Ptr(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Description: twapi.Ptr("This is a test tasklist"),
			MilestoneID: &testResources.MilestoneID,
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.TasklistUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestTasklistDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tasklistID, _, err := createTasklist(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.TasklistDelete(ctx, engine, projects.NewTasklistDeleteRequest(tasklistID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestTasklistGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tasklistID, tasklistCleanup, err := createTasklist(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(tasklistCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.TasklistGet(ctx, engine, projects.NewTasklistGetRequest(tasklistID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestTasklistList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	_, tasklistCleanup, err := createTasklist(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(tasklistCleanup)

	tests := []struct {
		name  string
		input projects.TasklistListRequest
	}{{
		name: "all tasklists",
	}, {
		name: "tasklists for project",
		input: projects.TasklistListRequest{
			Path: projects.TasklistListRequestPath{
				ProjectID: testResources.ProjectID,
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.TasklistList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
