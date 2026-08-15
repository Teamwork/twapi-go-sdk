package projects_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
	"github.com/teamwork/twapi-go-sdk/session"
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
			Options: projects.TaskOptions{
				Notify:            true,
				CheckInvalidUsers: true,
			},
			Name:             fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
			Description:      new("<h1>This is a test task</h1>"),
			Priority:         new("high"),
			Progress:         new(int64(50)),
			StartAt:          new(twapi.Date(time.Now().Add(24 * time.Hour))),
			DueAt:            new(twapi.Date(time.Now().Add(48 * time.Hour))),
			EstimatedMinutes: new(int64(120)),
			ParentTaskID:     &parentTaskID,
			Assignees: &projects.UserGroups{
				UserIDs: []int64{testResources.UserID},
			},
			TagIDs: []int64{testResources.TagID},
			Predecessors: []projects.TaskPredecessor{
				{ID: testResources.TaskID, Type: projects.TaskPredecessorTypeFinish},
			},
			ChangeFollowers: projects.UserGroups{
				UserIDs:    []int64{testResources.UserID},
				CompanyIDs: []int64{testResources.CompanyID},
			},
			CommentFollowers: projects.UserGroups{
				UserIDs:    []int64{testResources.UserID},
				CompanyIDs: []int64{testResources.CompanyID},
			},
			CompleteFollowers: projects.UserGroups{
				UserIDs:    []int64{testResources.UserID},
				CompanyIDs: []int64{testResources.CompanyID},
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
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

// TestTaskAttachments covers attaching a file when creating a task and to a task
// that already exists.
//
// The file is added to the project first so that the test owns an identifier it
// can delete afterwards. Attaching a pending file reference directly works the
// same way, but the API creates the project file itself and returns no way to
// find it, so a test doing that would leave a file behind on every run. The
// encoding of both forms is covered by TestAttachmentsEncoding.
//
// The assertion here is only that the write was accepted: the SDK has no
// endpoint for reading a task's files back.
func TestTaskAttachments(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	fileID, fileCleanup, err := createFile(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(fileCleanup)

	attachments := projects.TaskAttachments{
		Files: []projects.TaskAttachmentFile{{ID: fileID}},
	}

	t.Run("create", func(t *testing.T) {
		ctx := t.Context()
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		t.Cleanup(cancel)

		taskRequest := projects.NewTaskCreateRequest(
			testResources.TasklistID,
			fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
		)
		taskRequest.Attachments = attachments

		task, err := projects.TaskCreate(ctx, engine, taskRequest)
		t.Cleanup(func() {
			if err != nil {
				return
			}
			ctx := context.Background() // t.Context is always canceled in cleanup
			if _, err := projects.TaskDelete(ctx, engine, projects.NewTaskDeleteRequest(task.Task.ID)); err != nil {
				t.Errorf("failed to delete task after test: %s", err)
			}
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		} else if task.Task.ID == 0 {
			t.Error("expected a valid task ID but got 0")
		}
	})

	t.Run("update", func(t *testing.T) {
		taskID, taskCleanup, err := createTask(t, testResources.TasklistID)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(taskCleanup)

		ctx := t.Context()
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		t.Cleanup(cancel)

		taskRequest := projects.NewTaskUpdateRequest(taskID)
		taskRequest.Attachments = attachments

		if _, err := projects.TaskUpdate(ctx, engine, taskRequest); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	})
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
			Options: projects.TaskOptions{
				Notify:            true,
				CheckInvalidUsers: true,
			},
			Name:             new(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Description:      new("<h1>This is a test task</h1>"),
			Priority:         new("high"),
			Progress:         new(int64(50)),
			StartAt:          new(twapi.Date(time.Now().Add(24 * time.Hour))),
			DueAt:            new(twapi.Date(time.Now().Add(48 * time.Hour))),
			EstimatedMinutes: new(int64(120)),
			TasklistID:       &testResources.TasklistID,
			ParentTaskID:     &parentTaskID,
			Assignees: &projects.UserGroups{
				UserIDs: []int64{testResources.UserID},
			},
			TagIDs: []int64{testResources.TagID},
			Predecessors: []projects.TaskPredecessor{
				{ID: testResources.TaskID, Type: projects.TaskPredecessorTypeFinish},
			},
			ChangeFollowers: &projects.UserGroups{
				UserIDs:    []int64{testResources.UserID},
				CompanyIDs: []int64{testResources.CompanyID},
			},
			CommentFollowers: &projects.UserGroups{
				UserIDs:    []int64{testResources.UserID},
				CompanyIDs: []int64{testResources.CompanyID},
			},
			CompleteFollowers: &projects.UserGroups{
				UserIDs:    []int64{testResources.UserID},
				CompanyIDs: []int64{testResources.CompanyID},
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
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
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.TaskDelete(ctx, engine, projects.NewTaskDeleteRequest(taskID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestTaskComplete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	taskID, taskCleanup, err := createTask(t, testResources.TasklistID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(taskCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.TaskComplete(ctx, engine, projects.NewTaskCompleteRequest(taskID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestTaskMoveRequestGeneration(t *testing.T) {
	tests := []struct {
		name             string
		input            projects.TaskMoveRequest
		wantParentTaskID *int64
	}{{
		// omitting the field means detach here, so this is deliberate
		name:  "detaching by default",
		input: projects.NewTaskMoveRequest(12345, 888),
	}, {
		name: "detaching explicitly",
		input: func() projects.TaskMoveRequest {
			req := projects.NewTaskMoveRequest(12345, 888)
			req.ParentTaskID = new(projects.TaskDetachFromParent)
			return req
		}(),
		wantParentTaskID: new(int64(0)),
	}, {
		// The only way to move a subtask and have it stay one.
		name: "keeping the parent link",
		input: func() projects.TaskMoveRequest {
			req := projects.NewTaskMoveRequest(12345, 888)
			req.ParentTaskID = new(int64(999))
			return req
		}(),
		wantParentTaskID: new(int64(999)),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpReq, err := tt.input.HTTPRequest(context.Background(), "https://test.com")
			if err != nil {
				t.Fatalf("unexpected error creating HTTP request: %s", err)
			}

			// the dedicated move route, at the installation root
			if httpReq.URL.Path != "/tasks/12345/move.json" {
				t.Errorf("unexpected request path: %s", httpReq.URL.Path)
			}
			if httpReq.Method != http.MethodPut {
				t.Errorf("expected PUT but got %s", httpReq.Method)
			}

			body, err := io.ReadAll(httpReq.Body)
			if err != nil {
				t.Fatalf("failed to read request body: %s", err)
			}

			var payload struct {
				TasklistID   int64  `json:"taskListId"`
				ParentTaskID *int64 `json:"parentTaskId"`
			}
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("failed to decode request body %q: %s", body, err)
			}

			// top level of the body: an envelope would leave the defaults in place
			if payload.TasklistID != 888 {
				t.Errorf("expected taskListId 888 but got %d (body %q)", payload.TasklistID, body)
			}

			switch {
			case tt.wantParentTaskID == nil && payload.ParentTaskID != nil:
				t.Errorf("expected parentTaskId to be omitted but got %d (body %q)", *payload.ParentTaskID, body)
			case tt.wantParentTaskID != nil && payload.ParentTaskID == nil:
				t.Errorf("expected parentTaskId %d to reach the wire (body %q)", *tt.wantParentTaskID, body)
			case tt.wantParentTaskID != nil && *payload.ParentTaskID != *tt.wantParentTaskID:
				t.Errorf("expected parentTaskId %d but got %d (body %q)",
					*tt.wantParentTaskID, *payload.ParentTaskID, body)
			}
		})
	}
}

// The affected lists are dependency bookkeeping, not a record of what moved, and
// are routinely absent. Decoding must not turn that into an error.
func TestTaskMoveResponseDecoding(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		wantTasks     projects.LegacyNumericList
		wantTasklists projects.LegacyNumericList
		wantProjects  projects.LegacyNumericList
	}{{
		name:          "dependencies and tasklists reported",
		body:          `{"affectedTaskIds":"12346,12347","affectedTaskListIds":"777,888","STATUS":"OK"}`,
		wantTasks:     projects.LegacyNumericList{12346, 12347},
		wantTasklists: projects.LegacyNumericList{777, 888},
	}, {
		name:          "cross project move",
		body:          `{"affectedTaskListIds":"777,888","affectedProjectIds":"1,2","STATUS":"OK"}`,
		wantTasklists: projects.LegacyNumericList{777, 888},
		wantProjects:  projects.LegacyNumericList{1, 2},
	}, {
		// a task with no dependencies: the move happened, the field is empty
		name:          "no affected tasks",
		body:          `{"affectedTaskIds":"","affectedTaskListIds":"777,888","STATUS":"OK"}`,
		wantTasklists: projects.LegacyNumericList{777, 888},
	}, {
		name: "fields absent",
		body: `{"STATUS":"OK"}`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := twapi.NewEngine(
				session.NewBearerToken("token", "https://test.com"),
				twapi.WithHTTPClient(twapi.HTTPClientFunc(func(*http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(tt.body)),
						Header:     http.Header{"Content-Type": []string{"application/json"}},
					}, nil
				})),
			)

			response, err := projects.TaskMove(context.Background(), stub, projects.NewTaskMoveRequest(12345, 888))
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if !slices.Equal(response.AffectedTaskIDs, tt.wantTasks) {
				t.Errorf("affected tasks: got %v, want %v", response.AffectedTaskIDs, tt.wantTasks)
			}
			if !slices.Equal(response.AffectedTasklistIDs, tt.wantTasklists) {
				t.Errorf("affected tasklists: got %v, want %v", response.AffectedTasklistIDs, tt.wantTasklists)
			}
			if !slices.Equal(response.AffectedProjectIDs, tt.wantProjects) {
				t.Errorf("affected projects: got %v, want %v", response.AffectedProjectIDs, tt.wantProjects)
			}
		})
	}
}

func TestTaskMove(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	destinationTasklistID, destinationCleanup, err := createTasklist(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(destinationCleanup)

	parentTaskID, parentTaskCleanup, err := createTask(t, testResources.TasklistID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(parentTaskCleanup)

	subtaskID, subtaskCleanup, err := createSubtask(t, testResources.TasklistID, parentTaskID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(subtaskCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	// moving the parent carries the subtask with it, in one call
	if _, err = projects.TaskMove(ctx, engine,
		projects.NewTaskMoveRequest(parentTaskID, destinationTasklistID)); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	subtaskAfter, err := projects.TaskGet(ctx, engine, projects.NewTaskGetRequest(subtaskID))
	if err != nil {
		t.Fatalf("failed to reload subtask: %s", err)
	}
	if subtaskAfter.Task.Tasklist.ID != destinationTasklistID {
		t.Errorf("expected the subtask to be carried to tasklist %d but got %d",
			destinationTasklistID, subtaskAfter.Task.Tasklist.ID)
	}
	if subtaskAfter.Task.ParentTask == nil || subtaskAfter.Task.ParentTask.ID != parentTaskID {
		t.Errorf("expected the subtask to still hang from task %d, got %+v",
			parentTaskID, subtaskAfter.Task.ParentTask)
	}
}

// Moving a subtask without flattening it. The parent moves first, stranding the
// subtask, which is then repaired.
func TestTaskMoveKeepsParentWhenAsked(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	destinationTasklistID, destinationCleanup, err := createTasklist(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(destinationCleanup)

	parentTaskID, parentTaskCleanup, err := createTask(t, testResources.TasklistID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(parentTaskCleanup)

	subtaskID, subtaskCleanup, err := createSubtask(t, testResources.TasklistID, parentTaskID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(subtaskCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	t.Cleanup(cancel)

	updateRequest := projects.NewTaskUpdateRequest(parentTaskID)
	updateRequest.TasklistID = &destinationTasklistID
	if _, err := projects.TaskUpdate(ctx, engine, updateRequest); err != nil {
		t.Fatalf("failed to move the parent through v3: %s", err)
	}

	moveRequest := projects.NewTaskMoveRequest(subtaskID, destinationTasklistID)
	moveRequest.ParentTaskID = &parentTaskID
	if _, err := projects.TaskMove(ctx, engine, moveRequest); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	subtaskAfter, err := projects.TaskGet(ctx, engine, projects.NewTaskGetRequest(subtaskID))
	if err != nil {
		t.Fatalf("failed to reload subtask: %s", err)
	}
	if subtaskAfter.Task.Tasklist.ID != destinationTasklistID {
		t.Errorf("expected the subtask in tasklist %d but got %d",
			destinationTasklistID, subtaskAfter.Task.Tasklist.ID)
	}
	if subtaskAfter.Task.ParentTask == nil || subtaskAfter.Task.ParentTask.ID != parentTaskID {
		t.Errorf("expected the subtask to still hang from task %d, got %+v",
			parentTaskID, subtaskAfter.Task.ParentTask)
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
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
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
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.TaskList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
