package projects_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestTaskCreate(t *testing.T) {
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
		input         projects.TaskCreateRequest
		expectedError bool
	}{{
		name:  "it should create a task with valid input",
		input: projects.NewTaskCreateRequest(tasklistID, fmt.Sprintf("Test Task %d", time.Now().UnixNano())),
	}, {
		name: "it should fail to create a task with missing name",
		input: projects.TaskCreateRequest{
			Description: twapi.Ptr("This task has no name"),
		},
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			task, err := projects.TaskCreate(ctx, engine, tt.input)
			defer func() {
				if err != nil {
					return
				}
				_, err := projects.TaskDelete(ctx, engine, projects.NewTaskDeleteRequest(task.Task.ID))
				if err != nil {
					t.Errorf("failed to delete task after test: %s", err)
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
			if task.Task.ID == 0 {
				t.Error("expected a valid task ID but got 0")
			}
		})
	}
}

func TestTaskUpdate(t *testing.T) {
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

	taskID, taskCleanup, err := createTask(t, tasklistID)
	if err != nil {
		t.Fatal(err)
	}
	defer taskCleanup()

	tests := []struct {
		name          string
		input         projects.TaskUpdateRequest
		expectedError bool
	}{{
		name: "it should update a task with valid input",
		input: projects.TaskUpdateRequest{
			Path: projects.TaskUpdateRequestPath{
				ID: taskID,
			},
			Description: twapi.Ptr("This is a test task"),
		},
	}, {
		name: "it should fail to update a task with an empty name",
		input: projects.TaskUpdateRequest{
			Path: projects.TaskUpdateRequestPath{
				ID: taskID,
			},
			Name: twapi.Ptr(""),
		},
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.TaskUpdate(ctx, engine, tt.input)
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

func TestTaskDelete(t *testing.T) {
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

	taskID, _, err := createTask(t, tasklistID)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		input         projects.TaskDeleteRequest
		expectedError bool
	}{{
		name:  "it should delete a task with valid input",
		input: projects.NewTaskDeleteRequest(taskID),
	}, {
		name:          "it should fail to delete an unknown task",
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.TaskDelete(ctx, engine, tt.input)
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

func TestTaskGet(t *testing.T) {
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

	taskID, taskCleanup, err := createTask(t, tasklistID)
	if err != nil {
		t.Fatal(err)
	}
	defer taskCleanup()

	tests := []struct {
		name          string
		input         projects.TaskGetRequest
		expectedError bool
	}{{
		name:  "it should retrieve a task with valid input",
		input: projects.NewTaskGetRequest(taskID),
	}, {
		name:          "it should fail to retrieve an unknown task",
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.TaskGet(ctx, engine, tt.input)
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

func TestTaskList(t *testing.T) {
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

	_, taskCleanup, err := createTask(t, tasklistID)
	if err != nil {
		t.Fatal(err)
	}
	defer taskCleanup()

	tests := []struct {
		name          string
		input         projects.TaskListRequest
		expectedError bool
	}{{
		name: "it should list tasks",
	}, {
		name: "it should list tasks for tasklist",
		input: projects.TaskListRequest{
			Path: projects.TaskListRequestPath{
				TasklistID: tasklistID,
			},
		},
	}, {
		name: "it should list tasks for project",
		input: projects.TaskListRequest{
			Path: projects.TaskListRequestPath{
				ProjectID: projectID,
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.TaskList(ctx, engine, tt.input)
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
