package projects_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestWorkloadGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name          string
		input         projects.WorkloadRequest
		expectedError bool
	}{{
		name: "all users from the workload",
		input: projects.NewWorkloadRequest(
			twapi.Date(time.Now().AddDate(0, 0, -7)),
			twapi.Date(time.Now()),
		),
	}, {
		name: "with every sideload",
		input: func() projects.WorkloadRequest {
			req := projects.NewWorkloadRequest(
				twapi.Date(time.Now().AddDate(0, 0, -7)),
				twapi.Date(time.Now()),
			)
			req.Filters.Include = []projects.WorkloadGetRequestSideload{
				projects.WorkloadGetRequestSideloadUsers,
				projects.WorkloadGetRequestSideloadWorkingHourEntries,
				projects.WorkloadGetRequestSideloadTaskCapacities,
			}
			return req
		}(),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.WorkloadGet(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

// TestWorkloadSideloadsReachTheWire pins the include values, since the endpoint
// ignores one it does not recognise and answers without the sideload.
func TestWorkloadSideloadsReachTheWire(t *testing.T) {
	req := projects.NewWorkloadRequest(
		twapi.Date(time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)),
		twapi.Date(time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC)),
	)
	req.Filters.Include = []projects.WorkloadGetRequestSideload{
		projects.WorkloadGetRequestSideloadUsers,
		projects.WorkloadGetRequestSideloadWorkingHours,
		projects.WorkloadGetRequestSideloadWorkingHourEntries,
		projects.WorkloadGetRequestSideloadTasks,
		projects.WorkloadGetRequestSideloadTaskCapacities,
	}

	httpReq, err := req.HTTPRequest(t.Context(), "http://example.com")
	if err != nil {
		t.Fatalf("failed to build the request: %s", err)
	}
	want := []string{
		"users",
		"users.workingHours",
		"users.workingHours.workingHoursEntry",
		"tasks",
		"tasks.taskCapacities",
	}
	got := httpReq.URL.Query()["include"]
	if len(got) != len(want) {
		t.Fatalf("expected include %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("expected include %v, got %v", want, got)
			break
		}
	}
}

func TestWorkloadResponseDecodesSplit(t *testing.T) {
	body := `{
		"workload": {"users": [{
			"userId": 456,
			"dates": {
				"2026-10-05": {
					"capacity": 104.2,
					"capacityMinutes": 500,
					"unavailableDay": false,
					"isHoliday": false,
					"projects": [{
						"capacity": 104.2,
						"capacityMinutes": 500,
						"project": {"id": 777, "type": "projects"},
						"tasks": [{"id": 12345, "type": "tasks"}],
						"milestones": []
					}]
				},
				"2026-10-06": {"capacity": 0, "capacityMinutes": 0, "unavailableDay": true, "isHoliday": true}
			}
		}]},
		"meta": {"page": {"pageOffset": 0, "pageSize": 50, "count": 1, "hasMore": false}},
		"included": {
			"tasks": {"12345": {"id": 12345, "name": "Example Task", "estimateMinutes": 600,
				"startDate": "2026-10-05", "dueDate": "2026-10-06"}},
			"taskCapacities": {"1": {"id": 1, "taskId": 12345, "userId": 456,
				"date": "2026-10-05", "minutes": 500, "seconds": 0}}
		}
	}`

	var response projects.WorkloadResponse
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		t.Fatalf("failed to decode response: %s", err)
	}
	if len(response.Workload.Users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(response.Workload.Users))
	}
	dates := response.Workload.Users[0].Dates

	day := dates[twapi.Date(time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC))]
	if len(day.Projects) != 1 {
		t.Fatalf("expected 1 project on the day, got %d", len(day.Projects))
	}
	if got := day.Projects[0]; got.Project.ID != 777 || len(got.Tasks) != 1 || got.Tasks[0].ID != 12345 {
		t.Errorf("unexpected project breakdown: %+v", got)
	}
	holiday := dates[twapi.Date(time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC))]
	if !holiday.IsHoliday {
		t.Error("expected the holiday flag to decode")
	}

	if got := response.Included.Tasks["12345"].EstimatedMinutes; got != 600 {
		t.Errorf("expected the task sideload to decode, got estimate %d", got)
	}
	if got := response.Included.TaskCapacities["1"]; got.TaskID != 12345 || got.Minutes != 500 {
		t.Errorf("expected the task capacity sideload to decode, got %+v", got)
	}
}
