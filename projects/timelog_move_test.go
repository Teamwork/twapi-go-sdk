package projects_test

import (
	"testing"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

// TestTimelogUpdateMoveEncoding pins how an update encodes the project and the
// task a timelog is logged against. The endpoint answers 2xx whether or not it
// understood the move, and taskId is tri-state, so an omitted field encoded as
// a null would detach every timelog an unrelated update touches.
func TestTimelogUpdateMoveEncoding(t *testing.T) {
	tests := []struct {
		name    string
		request func() projects.TimelogUpdateRequest
		want    map[string]any
		absent  []string
	}{{
		name: "neither named",
		request: func() projects.TimelogUpdateRequest {
			return projects.NewTimelogUpdateRequest(12345)
		},
		absent: []string{"projectId", "taskId"},
	}, {
		name: "moved to a project",
		request: func() projects.TimelogUpdateRequest {
			req := projects.NewTimelogUpdateRequest(12345)
			req.ProjectID = new(int64(777))
			return req
		},
		want:   map[string]any{"projectId": float64(777)},
		absent: []string{"taskId"},
	}, {
		name: "moved to a task",
		request: func() projects.TimelogUpdateRequest {
			req := projects.NewTimelogUpdateRequest(12345)
			req.TaskID = twapi.NewNullableInt64(67890)
			return req
		},
		want:   map[string]any{"taskId": float64(67890)},
		absent: []string{"projectId"},
	}, {
		name: "detached from its task",
		request: func() projects.TimelogUpdateRequest {
			req := projects.NewTimelogUpdateRequest(12345)
			req.ProjectID = new(int64(777))
			req.TaskID = twapi.NullInt64()
			return req
		},
		want: map[string]any{"projectId": float64(777), "taskId": nil},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := encodeRequestBody(t, tt.request())
			timelog, ok := body["timelog"].(map[string]any)
			if !ok {
				t.Fatalf("expected a timelog object in the request body, got %v", body)
			}
			for key, want := range tt.want {
				got, ok := timelog[key]
				if !ok {
					t.Errorf("expected %q in the request body, got %v", key, timelog)
					continue
				}
				if got != want {
					t.Errorf("expected %q to be %v, got %v", key, want, got)
				}
			}
			for _, key := range tt.absent {
				if value, ok := timelog[key]; ok {
					t.Errorf("expected no %q key in the request body, got %v", key, value)
				}
			}
		})
	}
}
