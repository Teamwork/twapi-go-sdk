package projects_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestTasklistCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	projectID, projectCleanup, err := createProject(t)
	if err != nil {
		t.Fatal(err)
	}
	defer projectCleanup()

	tests := []struct {
		name          string
		input         projects.TasklistCreateRequest
		expectedError bool
	}{{
		name:  "it should create a tasklist with valid input",
		input: projects.NewTasklistCreateRequest(projectID, fmt.Sprintf("Test Tasklist %d", time.Now().UnixNano())),
	}, {
		name: "it should fail to create a tasklist with missing name",
		input: projects.TasklistCreateRequest{
			Description: twapi.Ptr("This tasklist has no name"),
		},
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			project, err := projects.TasklistCreate(ctx, engine, tt.input)
			defer func() {
				if err != nil {
					return
				}
				_, err := projects.TasklistDelete(ctx, engine, projects.NewTasklistDeleteRequest(int64(project.ID)))
				if err != nil {
					t.Errorf("failed to delete tasklist after test: %s", err)
				}
			}()

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
				t.Error("expected a valid tasklist ID but got 0")
			}
		})
	}
}

func TestTasklistUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	projectID, projectCleanup, err := createProject(t)
	if err != nil {
		t.Fatal(err)
	}
	defer projectCleanup()

	tasklistID, tasklistCleanup, err := createTasklist(t, projectID)
	if err != nil {
		t.Fatal(err)
	}
	defer tasklistCleanup()

	tests := []struct {
		name          string
		input         projects.TasklistUpdateRequest
		expectedError bool
	}{{
		name: "it should update a tasklist with valid input",
		input: projects.TasklistUpdateRequest{
			Path: projects.TasklistUpdateRequestPath{
				ID: tasklistID,
			},
			Description: twapi.Ptr("This is a test tasklist"),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.TasklistUpdate(ctx, engine, tt.input)
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

func TestTasklistDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	projectID, projectCleanup, err := createProject(t)
	if err != nil {
		t.Fatal(err)
	}
	defer projectCleanup()

	tasklistID, _, err := createTasklist(t, projectID)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		input         projects.TasklistDeleteRequest
		expectedError bool
	}{{
		name:  "it should delete a tasklist with valid input",
		input: projects.NewTasklistDeleteRequest(tasklistID),
	}, {
		name:          "it should fail to delete an unknown tasklist",
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.TasklistDelete(ctx, engine, tt.input)
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

func TestTasklistGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	projectID, projectCleanup, err := createProject(t)
	if err != nil {
		t.Fatal(err)
	}
	defer projectCleanup()

	tasklistID, tasklistCleanup, err := createTasklist(t, projectID)
	if err != nil {
		t.Fatal(err)
	}
	defer tasklistCleanup()

	tests := []struct {
		name          string
		input         projects.TasklistGetRequest
		expectedError bool
	}{{
		name:  "it should retrieve a tasklist with valid input",
		input: projects.NewTasklistGetRequest(tasklistID),
	}, {
		name:          "it should fail to retrieve an unknown tasklist",
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.TasklistGet(ctx, engine, tt.input)
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

func TestTasklistList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	projectID, projectCleanup, err := createProject(t)
	if err != nil {
		t.Fatal(err)
	}
	defer projectCleanup()

	_, tasklistCleanup, err := createTasklist(t, projectID)
	if err != nil {
		t.Fatal(err)
	}
	defer tasklistCleanup()

	tests := []struct {
		name          string
		input         projects.TasklistListRequest
		expectedError bool
	}{{
		name: "it should list tasklists",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.TasklistList(ctx, engine, tt.input)
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
