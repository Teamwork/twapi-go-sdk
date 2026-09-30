package projects_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestTaskCapacity(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	t.Cleanup(cancel)

	// a split needs an estimate, a due date and a user assignee
	start := time.Now().AddDate(0, 0, 7).Truncate(24 * time.Hour)
	due := start.AddDate(0, 0, 1)
	taskResponse, err := projects.TaskCreate(ctx, engine, projects.TaskCreateRequest{
		Path:             projects.TaskCreateRequestPath{TasklistID: testResources.TasklistID},
		Name:             "task capacity " + time.Now().Format(time.RFC3339Nano),
		StartAt:          new(twapi.Date(start)),
		DueAt:            new(twapi.Date(due)),
		EstimatedMinutes: new(int64(120)),
		Assignees:        &projects.UserGroups{UserIDs: []int64{testResources.UserID}},
	})
	if err != nil {
		t.Fatalf("failed to create task: %s", err)
	}
	taskID := taskResponse.Task.ID
	t.Cleanup(func() {
		if _, err := projects.TaskDelete(context.Background(), engine, projects.NewTaskDeleteRequest(taskID)); err != nil {
			t.Errorf("failed to delete task after test: %s", err)
		}
	})

	listSplit := func() []projects.TaskCapacity {
		t.Helper()
		response, err := projects.TaskCapacityList(ctx, engine,
			projects.NewTaskCapacityListRequest(taskID, testResources.UserID))
		if err != nil {
			t.Fatalf("failed to list task capacity: %s", err)
		}
		return response.Capacities
	}

	created, err := projects.TaskCapacityCreate(ctx, engine, projects.NewTaskCapacityCreateRequest(
		taskID, testResources.UserID, []projects.TaskCapacityDate{
			{Date: twapi.Date(start), Minutes: 90},
			{Date: twapi.Date(due), Minutes: 30},
		},
	))
	if err != nil {
		t.Fatalf("failed to create task capacity: %s", err)
	}
	if len(created.Capacities) != 2 {
		t.Fatalf("expected 2 stored rows, got %d", len(created.Capacities))
	}
	if got := listSplit(); len(got) != 2 {
		t.Fatalf("expected 2 listed rows, got %d", len(got))
	}

	replaced, err := projects.TaskCapacityReplace(ctx, engine, projects.NewTaskCapacityReplaceRequest(
		taskID, testResources.UserID, []projects.TaskCapacityDate{
			{Date: twapi.Date(start), Minutes: 120},
			{Date: twapi.Date(due), Minutes: 0},
		},
	))
	if err != nil {
		t.Fatalf("failed to replace task capacity: %s", err)
	}
	for _, row := range replaced.Capacities {
		if time.Time(row.Date).Equal(start) && row.Minutes != 120 {
			t.Errorf("expected 120 minutes on the start date, got %d", row.Minutes)
		}
	}

	if _, err := projects.TaskCapacityDelete(ctx, engine,
		projects.NewTaskCapacityDeleteRequest(taskID, testResources.UserID)); err != nil {
		t.Fatalf("failed to delete task capacity: %s", err)
	}
	if got := listSplit(); len(got) != 0 {
		t.Errorf("expected no rows after the delete, got %d", len(got))
	}
}

// TestTaskCapacityRequestsReachTheWire pins the verb, path and payload of every
// split request, since the write routes share one path and differ only in the
// verb.
func TestTaskCapacityRequestsReachTheWire(t *testing.T) {
	date := twapi.Date(time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC))
	dates := []projects.TaskCapacityDate{{Date: date, Minutes: 90, Seconds: new(int64(5400))}}
	const wantWrite = `{"taskCapacity":{"userId":456,"dates":[{"date":"2026-10-05","minutes":90,"seconds":5400}]}}`

	tests := []struct {
		name      string
		request   twapi.HTTPRequester
		method    string
		path      string
		query     string
		body      string
		emptyBody bool
	}{{
		name:    "list",
		request: projects.NewTaskCapacityListRequest(12345, 456, 789),
		method:  http.MethodGet,
		path:    "/projects/api/v3/tasks/capacity.json",
		query:   "page=1&pageSize=50&taskId=12345&userIds=456%2C789",
	}, {
		name:    "create",
		request: projects.NewTaskCapacityCreateRequest(12345, 456, dates),
		method:  http.MethodPost,
		path:    "/projects/api/v3/tasks/12345/capacity.json",
		body:    wantWrite,
	}, {
		name:    "replace",
		request: projects.NewTaskCapacityReplaceRequest(12345, 456, dates),
		method:  http.MethodPut,
		path:    "/projects/api/v3/tasks/12345/capacity.json",
		body:    wantWrite,
	}, {
		name: "create omits unset seconds",
		request: projects.NewTaskCapacityCreateRequest(12345, 456,
			[]projects.TaskCapacityDate{{Date: date, Minutes: 90}}),
		method: http.MethodPost,
		path:   "/projects/api/v3/tasks/12345/capacity.json",
		body:   `{"taskCapacity":{"userId":456,"dates":[{"date":"2026-10-05","minutes":90}]}}`,
	}, {
		name:    "delete for some users",
		request: projects.NewTaskCapacityDeleteRequest(12345, 456),
		method:  http.MethodDelete,
		path:    "/projects/api/v3/tasks/12345/capacity.json",
		body:    `{"userIds":[456]}`,
	}, {
		name:    "delete for every user",
		request: projects.NewTaskCapacityDeleteRequest(12345),
		method:  http.MethodDelete,
		path:    "/projects/api/v3/tasks/12345/capacity.json",
		body:    `{"userIds":[]}`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := tt.request.HTTPRequest(t.Context(), "http://example.com")
			if err != nil {
				t.Fatalf("failed to build the request: %s", err)
			}
			if req.Method != tt.method {
				t.Errorf("expected method %s, got %s", tt.method, req.Method)
			}
			if req.URL.Path != tt.path {
				t.Errorf("expected path %s, got %s", tt.path, req.URL.Path)
			}
			if req.URL.RawQuery != tt.query {
				t.Errorf("expected query %q, got %q", tt.query, req.URL.RawQuery)
			}
			var body string
			if req.Body != nil {
				raw, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatalf("failed to read the body: %s", err)
				}
				body = strings.TrimSpace(string(raw))
			}
			if body != tt.body {
				t.Errorf("expected body %s, got %s", tt.body, body)
			}
		})
	}
}

// TestTaskCapacityCreateResponseNoContent pins that an even split, which the
// endpoint answers with 204 and does not store, is not reported as a failure.
func TestTaskCapacityCreateResponseNoContent(t *testing.T) {
	var response projects.TaskCapacityCreateResponse
	err := response.HandleHTTPResponse(&http.Response{
		StatusCode: http.StatusNoContent,
		Body:       io.NopCloser(strings.NewReader("")),
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(response.Capacities) != 0 {
		t.Errorf("expected no rows, got %d", len(response.Capacities))
	}
}

func TestTaskCapacityListResponseDecodes(t *testing.T) {
	body := `{
		"capacities": [
			{"id": 1, "taskId": 12345, "userId": 456, "date": "2026-10-05", "minutes": 90, "seconds": 5400}
		],
		"meta": {"page": {"pageOffset": 0, "pageSize": 50, "count": 1, "hasMore": false}},
		"included": {}
	}`

	var response projects.TaskCapacityListResponse
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		t.Fatalf("failed to decode response: %s", err)
	}
	if len(response.Capacities) != 1 {
		t.Fatalf("expected 1 row, got %d", len(response.Capacities))
	}
	row := response.Capacities[0]
	if row.TaskID != 12345 || row.UserID != 456 || row.Minutes != 90 || row.Seconds != 5400 {
		t.Errorf("unexpected row: %+v", row)
	}
	if got := time.Time(row.Date).Format("2006-01-02"); got != "2026-10-05" {
		t.Errorf("expected date 2026-10-05, got %s", got)
	}
	if count := response.Meta.Page.Count; count == nil || *count != 1 {
		t.Errorf("expected count 1, got %v", count)
	}
}
