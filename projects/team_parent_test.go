package projects_test

import (
	"testing"

	"github.com/teamwork/twapi-go-sdk/projects"
)

// TestTeamUpdateParentEncoding pins how an update encodes the parent team. The
// endpoint reads a missing parentTeamId as "leave it alone" and a zero as "move
// to the top level", so an omitted field encoded as a zero would flatten the
// hierarchy on any unrelated update.
func TestTeamUpdateParentEncoding(t *testing.T) {
	tests := []struct {
		name    string
		request func() projects.TeamUpdateRequest
		want    any
		absent  bool
	}{{
		name: "not named",
		request: func() projects.TeamUpdateRequest {
			return projects.NewTeamUpdateRequest(12345)
		},
		absent: true,
	}, {
		name: "moved under another team",
		request: func() projects.TeamUpdateRequest {
			req := projects.NewTeamUpdateRequest(12345)
			req.ParentTeamID = new(int64(777))
			return req
		},
		want: float64(777),
	}, {
		name: "moved to the top level",
		request: func() projects.TeamUpdateRequest {
			req := projects.NewTeamUpdateRequest(12345)
			req.ParentTeamID = new(int64(0))
			return req
		},
		want: float64(0),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := encodeRequestBody(t, tt.request())
			team, ok := body["team"].(map[string]any)
			if !ok {
				t.Fatalf("expected a team object in the request body, got %v", body)
			}
			got, ok := team["parentTeamId"]
			switch {
			case tt.absent && ok:
				t.Errorf("expected no parentTeamId key in the request body, got %v", got)
			case !tt.absent && !ok:
				t.Errorf("expected parentTeamId in the request body, got %v", team)
			case !tt.absent && got != tt.want:
				t.Errorf("expected parentTeamId to be %v, got %v", tt.want, got)
			}
		})
	}
}
