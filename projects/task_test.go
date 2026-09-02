package projects_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
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

func TestTaskListCountMode(t *testing.T) {
	tests := []struct {
		name              string
		mode              twapi.ListCountMode
		expectedSkipCount string
	}{{
		name:              "default leaves the decision to the API",
		mode:              twapi.ListCountModeDefault,
		expectedSkipCount: "",
	}, {
		name:              "exact asks for the count query",
		mode:              twapi.ListCountModeExact,
		expectedSkipCount: "false",
	}, {
		name:              "skip opts out of the count query",
		mode:              twapi.ListCountModeSkip,
		expectedSkipCount: "true",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := projects.NewTaskListRequest()
			req.Filters.CountMode = tt.mode

			httpReq, err := req.HTTPRequest(context.Background(), "https://test.com")
			if err != nil {
				t.Fatalf("unexpected error creating HTTP request: %s", err)
			}

			query, err := url.ParseQuery(httpReq.URL.RawQuery)
			if err != nil {
				t.Fatalf("failed to parse query string: %s", err)
			}

			if got := query.Get("skipCounts"); got != tt.expectedSkipCount {
				t.Errorf("expected skipCounts=%q but got %q", tt.expectedSkipCount, got)
			}
		})
	}
}

func TestTaskListResponseCount(t *testing.T) {
	// the API answers a skipped count with a lower bound derived from the page, so
	// only a requested count may reach the caller as a total.
	const payload = `{"meta":{"page":{"count":51,"hasMore":true}},"tasks":[{"id":12345}]}`

	count := int64(51)

	tests := []struct {
		name          string
		mode          twapi.ListCountMode
		expectedCount *int64
	}{{
		name:          "default keeps the count",
		mode:          twapi.ListCountModeDefault,
		expectedCount: &count,
	}, {
		name:          "exact keeps the count",
		mode:          twapi.ListCountModeExact,
		expectedCount: &count,
	}, {
		name:          "skip drops the lower bound",
		mode:          twapi.ListCountModeSkip,
		expectedCount: nil,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response projects.TaskListResponse
			httpResp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(payload)),
			}
			if err := response.HandleHTTPResponse(httpResp); err != nil {
				t.Fatalf("unexpected error handling response: %s", err)
			}

			req := projects.NewTaskListRequest()
			req.Filters.CountMode = tt.mode
			response.SetRequest(req)

			switch got := response.Meta.Page.Count; {
			case got == nil && tt.expectedCount != nil:
				t.Errorf("expected count %d but got nil", *tt.expectedCount)
			case got != nil && tt.expectedCount == nil:
				t.Errorf("expected no count but got %d", *got)
			case got != nil && *got != *tt.expectedCount:
				t.Errorf("expected count %d but got %d", *tt.expectedCount, *got)
			}
		})
	}
}

func TestTaskSubTaskIDsAreDecoded(t *testing.T) {
	// The attribute rides on includeRelatedTasks and is absent from the response
	// otherwise, so the model has to carry it for a get to report subtasks at all.
	const payload = `{"task":{"id":777,"name":"parent","subTaskIds":[1,2,3]}}`

	var resp projects.TaskGetResponse
	if err := json.Unmarshal([]byte(payload), &resp); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if got := len(resp.Task.SubTaskIDs); got != 3 {
		t.Fatalf("expected 3 subtask IDs but got %d", got)
	}
	if resp.Task.SubTaskIDs[0] != 1 || resp.Task.SubTaskIDs[2] != 3 {
		t.Errorf("unexpected subtask IDs: %v", resp.Task.SubTaskIDs)
	}
}

// TestTaskWorkflowsRequestGeneration pins where the workflows block lands and
// that it disappears when unset. It rides beside "task" rather than inside it,
// and an empty object is not inert on every endpoint, so sending one on an
// unrelated update would be a change of meaning rather than a wasted field.
func TestTaskWorkflowsRequestGeneration(t *testing.T) {
	placement := projects.TaskWorkflows{
		WorkflowID:          new(int64(123)),
		StageID:             new(int64(456)),
		PositionAfterTaskID: new(int64(789)),
	}

	tests := []struct {
		name    string
		request twapi.HTTPRequester
		want    *projects.TaskWorkflows
	}{{
		name:    "create without a placement",
		request: projects.NewTaskCreateRequest(888, "test"),
	}, {
		name: "create with a placement",
		request: func() projects.TaskCreateRequest {
			req := projects.NewTaskCreateRequest(888, "test")
			req.Workflows = placement
			return req
		}(),
		want: &placement,
	}, {
		name:    "update without a placement",
		request: projects.NewTaskUpdateRequest(12345),
	}, {
		name: "update with a placement",
		request: func() projects.TaskUpdateRequest {
			req := projects.NewTaskUpdateRequest(12345)
			req.Workflows = placement
			return req
		}(),
		want: &placement,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpReq, err := tt.request.HTTPRequest(context.Background(), "https://test.teamwork.com")
			if err != nil {
				t.Fatalf("unexpected error creating HTTP request: %s", err)
			}
			body, err := io.ReadAll(httpReq.Body)
			if err != nil {
				t.Fatalf("failed to read request body: %s", err)
			}

			var payload struct {
				Task      map[string]any          `json:"task"`
				Workflows *projects.TaskWorkflows `json:"workflows"`
			}
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("failed to decode request body %q: %s", body, err)
			}

			// A "workflows" key inside "task" is ignored by the endpoint, so the
			// sibling position is the whole point.
			if _, ok := payload.Task["workflows"]; ok {
				t.Errorf("workflows must sit beside task, not inside it (body %q)", body)
			}

			if tt.want == nil {
				if payload.Workflows != nil {
					t.Errorf("expected workflows to be omitted but got %+v (body %q)", *payload.Workflows, body)
				}
				return
			}
			if payload.Workflows == nil {
				t.Fatalf("expected workflows to reach the wire (body %q)", body)
			}
			for _, field := range []struct {
				name      string
				got, want *int64
			}{
				{"workflowId", payload.Workflows.WorkflowID, tt.want.WorkflowID},
				{"stageId", payload.Workflows.StageID, tt.want.StageID},
				{"positionAfterTask", payload.Workflows.PositionAfterTaskID, tt.want.PositionAfterTaskID},
			} {
				if field.got == nil || *field.got != *field.want {
					t.Errorf("expected %s %d but got %v (body %q)", field.name, *field.want, field.got, body)
				}
			}
		})
	}
}

// TestTaskListFiltersApplied pins the whole query string the task list builds
// when every filter is populated, against the parameter names the endpoint
// documents:
//
// https://apidocs.teamwork.com/docs/teamwork/v3/tasks/get-projects-api-v3-tasks-json
//
// Comparing the complete map rather than a subset is deliberate. An
// unrecognised query key is silently ignored by the API, so a misspelled
// parameter looks exactly like a working one from the caller's side — only an
// exact comparison catches it, and only an exact comparison catches a filter
// that stopped reaching the wire.
//
// StartAfter is deliberately absent, so that TestTaskListStartAfterAlone can
// pin it as the only date bound reaching the wire.
func TestTaskListFiltersApplied(t *testing.T) {
	createdAfter := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	createdBefore := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
	updatedAfter := time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)
	updatedBefore := time.Date(2026, 4, 5, 6, 7, 8, 0, time.UTC)
	completedAfter := time.Date(2026, 5, 6, 7, 8, 9, 0, time.UTC)
	completedBefore := time.Date(2026, 6, 7, 8, 9, 10, 0, time.UTC)
	dueAfter := twapi.Date(time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC))
	dueBefore := twapi.Date(time.Date(2026, 8, 9, 0, 0, 0, 0, time.UTC))

	req := projects.TaskListRequest{
		Filters: projects.TaskListRequestFilters{
			TaskRequestFilters: projects.TaskRequestFilters{
				IncludeRelatedTasks:          true,
				IncludeCompletedPredecessors: true,
				IncludeTasksWithoutDueDates:  new(false),
				Include: []projects.TaskRequestSideload{
					projects.TaskRequestSideloadCustomFields,
				},
			},

			SearchTerm: "acme",

			AssigneeUserIDs:        []int64{777, 888},
			ExcludeAssigneeUserIDs: []int64{999},

			CreatedAfter:     &createdAfter,
			CreatedBefore:    &createdBefore,
			CreatedByUserIDs: []int64{12345},
			UpdatedAfter:     &updatedAfter,
			UpdatedBefore:    &updatedBefore,
			CompletedAfter:   &completedAfter,
			CompletedBefore:  &completedBefore,

			IncludeCompleted:          new(true),
			IncludeCompletedTasklists: new(true),

			DateFilter: projects.TaskDateFilterStarted,
			DueAfter:   &dueAfter,
			DueBefore:  &dueBefore,

			TagIDs:       []int64{111, 222},
			MatchAllTags: new(true),

			OnlyUnassigned: new(false),
			OnlyUnplanned:  new(true),

			OrderBy:              projects.TaskOrderByCustomField,
			OrderMode:            twapi.OrderModeDescending,
			OrderByCustomFieldID: 42,

			Page:      2,
			PageSize:  25,
			CountMode: twapi.ListCountModeExact,

			Fields: projects.TaskListFields{
				Tasks: []projects.TaskField{projects.TaskFieldName},
			},
		},
	}

	want := map[string]string{
		"includeRelatedTasks":          "true",
		"includeCompletedPredecessors": "true",
		"includeTasksWithoutDueDates":  "false",
		"include":                      "customfields",

		"searchTerm": "acme",

		"responsiblePartyIds":        "777,888",
		"excludeResponsiblePartyIds": "999",

		"createdAfter":     "2026-01-02T03:04:05Z",
		"createdBefore":    "2026-02-03T04:05:06Z",
		"createdByUserIds": "12345",
		"updatedAfter":     "2026-03-04T05:06:07Z",
		"updatedBefore":    "2026-04-05T06:07:08Z",
		"completedAfter":   "2026-05-06T07:08:09Z",
		"completedBefore":  "2026-06-07T08:09:10Z",

		"includeCompletedTasks": "true",
		"showCompletedLists":    "true",

		"taskFilter": "started",
		"dueAfter":   "2026-07-08",
		"dueBefore":  "2026-08-09",

		"tagIds":       "111,222",
		"matchAllTags": "true",

		"onlyUnassignedTasks": "false",
		"onlyUnplanned":       "true",

		"orderBy":              "customfield",
		"orderMode":            "desc",
		"orderByCustomFieldId": "42",

		"page":       "2",
		"pageSize":   "25",
		"skipCounts": "false",

		"fields[tasks]": "name",
	}

	query := listQuery(t, req)

	for key, expected := range want {
		if got := query.Get(key); got != expected {
			t.Errorf("expected %s=%q but got %q", key, expected, got)
		}
	}
	for key := range query {
		if _, ok := want[key]; !ok {
			t.Errorf("unexpected query parameter %s=%q", key, query.Get(key))
		}
	}
}

// TestTaskListFiltersUnset checks the zero-value filters send nothing, so the
// endpoint applies its own defaults — TaskDateFilterAnytime among them —
// rather than receiving a wall of false.
func TestTaskListFiltersUnset(t *testing.T) {
	query := listQuery(t, projects.TaskListRequest{})
	if len(query) != 0 {
		t.Errorf("expected no query parameters but got %v", query)
	}
}

// TestTaskListStartAfterAlone pins the start-date bound on the query string,
// and that nothing sends the endpoint's companion `endDate`: with both set the
// endpoint stops reading `startDate` as a start-date bound at all.
func TestTaskListStartAfterAlone(t *testing.T) {
	startAfter := twapi.Date(time.Date(2026, 3, 4, 0, 0, 0, 0, time.UTC))

	query := listQuery(t, projects.TaskListRequest{
		Filters: projects.TaskListRequestFilters{
			StartAfter: &startAfter,
		},
	})

	if got := query.Get("startDate"); got != "2026-03-04" {
		t.Errorf("expected startDate=%q but got %q", "2026-03-04", got)
	}
	if got := query.Get("endDate"); got != "" {
		t.Errorf("expected no endDate but got %q", got)
	}
}

// TestTaskListDateFilterValuesReachTheWire drives every published date filter
// through the query string. The endpoint rejects an unknown value with 400, and
// answers a value it does understand with an ordinary task list, so a constant
// carrying a typo is only visible here.
func TestTaskListDateFilterValuesReachTheWire(t *testing.T) {
	for _, dateFilter := range []projects.TaskDateFilter{
		projects.TaskDateFilterAnytime,
		projects.TaskDateFilterOverdue,
		projects.TaskDateFilterToday,
		projects.TaskDateFilterTomorrow,
		projects.TaskDateFilterYesterday,
		projects.TaskDateFilterThisWeek,
		projects.TaskDateFilterUpcoming,
		projects.TaskDateFilterStarted,
		projects.TaskDateFilterWithin7,
		projects.TaskDateFilterWithin14,
		projects.TaskDateFilterWithin30,
		projects.TaskDateFilterWithin365,
		projects.TaskDateFilterNoDate,
		projects.TaskDateFilterNoDueDate,
		projects.TaskDateFilterNoStartDate,
		projects.TaskDateFilterHasDate,
	} {
		t.Run(string(dateFilter), func(t *testing.T) {
			query := listQuery(t, projects.TaskListRequest{
				Filters: projects.TaskListRequestFilters{
					DateFilter: dateFilter,
				},
			})
			if got := query.Get("taskFilter"); got != string(dateFilter) {
				t.Errorf("expected taskFilter=%q but got %q", dateFilter, got)
			}
		})
	}
}
