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

func TestTaskCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	parentTaskID, parentTaskCleanup, err := createTask(t, testResources.TasklistID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(parentTaskCleanup)

	tests := []struct {
		name  string
		input projects.TaskCreateRequest
	}{{
		name: "only required fields",
		input: projects.NewTaskCreateRequest(
			testResources.TasklistID,
			fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
		),
	}, {
		name: "all fields",
		input: projects.TaskCreateRequest{
			Path: projects.TaskCreateRequestPath{
				TasklistID: testResources.TasklistID,
			},
			Name:             fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
			Description:      twapi.Ptr("This is a test task"),
			Priority:         twapi.Ptr("high"),
			Progress:         twapi.Ptr(int64(50)),
			StartAt:          twapi.Ptr(twapi.Date(time.Now().Add(24 * time.Hour))),
			DueAt:            twapi.Ptr(twapi.Date(time.Now().Add(48 * time.Hour))),
			EstimatedMinutes: twapi.Ptr(int64(120)),
			ParentTaskID:     &parentTaskID,
			Assignees: &projects.UserGroups{
				UserIDs: []int64{testResources.UserID},
			},
			TagIDs: []int64{testResources.TagID},
			Predecessors: []projects.TaskPredecessor{
				{ID: testResources.TaskID, Type: projects.TaskPredecessorTypeFinish},
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			t.Cleanup(cancel)

			task, err := projects.TaskCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx = context.Background() // t.Context is always canceled in cleanup
				_, err := projects.TaskDelete(ctx, engine, projects.NewTaskDeleteRequest(task.Task.ID))
				if err != nil {
					t.Errorf("failed to delete task after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if task.Task.ID == 0 {
				t.Error("expected a valid task ID but got 0")
			}
		})
	}
}

func TestTaskUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	taskID, taskCleanup, err := createTask(t, testResources.TasklistID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(taskCleanup)

	parentTaskID, parentTaskCleanup, err := createTask(t, testResources.TasklistID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(parentTaskCleanup)

	tests := []struct {
		name  string
		input projects.TaskUpdateRequest
	}{{
		name: "all fields",
		input: projects.TaskUpdateRequest{
			Path: projects.TaskUpdateRequestPath{
				ID: taskID,
			},
			Name:             twapi.Ptr(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Description:      twapi.Ptr("This is a test task"),
			Priority:         twapi.Ptr("high"),
			Progress:         twapi.Ptr(int64(50)),
			StartAt:          twapi.Ptr(twapi.Date(time.Now().Add(24 * time.Hour))),
			DueAt:            twapi.Ptr(twapi.Date(time.Now().Add(48 * time.Hour))),
			EstimatedMinutes: twapi.Ptr(int64(120)),
			TasklistID:       &testResources.TasklistID,
			ParentTaskID:     &parentTaskID,
			Assignees: &projects.UserGroups{
				UserIDs: []int64{testResources.UserID},
			},
			TagIDs: []int64{testResources.TagID},
			Predecessors: []projects.TaskPredecessor{
				{ID: testResources.TaskID, Type: projects.TaskPredecessorTypeFinish},
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.TaskUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestTaskDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	taskID, _, err := createTask(t, testResources.TasklistID)
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.TaskDelete(ctx, engine, projects.NewTaskDeleteRequest(taskID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestTaskGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	taskID, taskCleanup, err := createTask(t, testResources.TasklistID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(taskCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.TaskGet(ctx, engine, projects.NewTaskGetRequest(taskID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestTaskList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	_, taskCleanup, err := createTask(t, testResources.TasklistID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(taskCleanup)

	tests := []struct {
		name  string
		input projects.TaskListRequest
	}{{
		name: "all tasks",
	}, {
		name: "tasks for tasklist",
		input: projects.TaskListRequest{
			Path: projects.TaskListRequestPath{
				TasklistID: testResources.TasklistID,
			},
		},
	}, {
		name: "tasks for project",
		input: projects.TaskListRequest{
			Path: projects.TaskListRequestPath{
				ProjectID: testResources.ProjectID,
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.TaskList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
