package projects_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestTeamCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	parentTeamID, teamCleanup, err := createTeam(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(teamCleanup)

	tests := []struct {
		name  string
		input projects.TeamCreateRequest
	}{{
		name:  "only required fields",
		input: projects.NewTeamCreateRequest(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
	}, {
		name: "all fields for company",
		input: projects.TeamCreateRequest{
			Name:         fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
			Handle:       new(fmt.Sprintf("testhandle%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Description:  new("This is a test team."),
			ParentTeamID: &parentTeamID,
			CompanyID:    &testResources.CompanyID,
			UserIDs:      []int64{testResources.UserID},
		},
	}, {
		name: "all fields for project",
		input: projects.TeamCreateRequest{
			Name:         fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
			Handle:       new(fmt.Sprintf("testhandle%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Description:  new("This is a test team."),
			ParentTeamID: &parentTeamID,
			ProjectID:    &testResources.ProjectID,
			UserIDs:      []int64{testResources.UserID},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			team, err := projects.TeamCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx = context.Background() // t.Context is always canceled in cleanup
				_, err := projects.TeamDelete(ctx, engine, projects.NewTeamDeleteRequest(int64(team.ID)))
				if err != nil {
					t.Errorf("failed to delete team after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if team.ID == 0 {
				t.Error("expected a valid team ID but got 0")
			}
		})
	}
}

func TestTeamUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	teamID, teamCleanup, err := createTeam(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(teamCleanup)

	tests := []struct {
		name  string
		input projects.TeamUpdateRequest
	}{{
		name: "all fields for company",
		input: projects.TeamUpdateRequest{
			Path: projects.TeamUpdateRequestPath{
				ID: teamID,
			},
			Name:        new(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Handle:      new(fmt.Sprintf("testhandle%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Description: new("This is a test team."),
			CompanyID:   &testResources.CompanyID,
			UserIDs:     []int64{testResources.UserID},
		},
	}, {
		name: "all fields for project",
		input: projects.TeamUpdateRequest{
			Path: projects.TeamUpdateRequestPath{
				ID: teamID,
			},
			Name:        new(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Handle:      new(fmt.Sprintf("testhandle%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Description: new("This is a test team."),
			ProjectID:   &testResources.ProjectID,
			UserIDs:     []int64{testResources.UserID},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.TeamUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestTeamDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	teamID, _, err := createTeam(t)
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.TeamDelete(ctx, engine, projects.NewTeamDeleteRequest(teamID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestTeamGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	teamID, teamCleanup, err := createTeam(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(teamCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.TeamGet(ctx, engine, projects.NewTeamGetRequest(teamID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestTeamList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	_, teamCleanup, err := createTeam(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(teamCleanup)

	tests := []struct {
		name          string
		input         projects.TeamListRequest
		expectedError bool
	}{{
		name: "all teams",
	}, {
		name: "teams for company",
		input: projects.TeamListRequest{
			Path: projects.TeamListRequestPath{
				CompanyID: testResources.CompanyID,
			},
		},
	}, {
		name: "teams for project",
		input: projects.TeamListRequest{
			Path: projects.TeamListRequestPath{
				ProjectID: testResources.ProjectID,
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.TeamList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

// TestTeamDeletedAtEncoding guards the JSON shape of Team.DeletedAt, the SDK's
// only twapi.OptionalDateTime field. Consumers that derive a JSON Schema from
// these models by reflection — the MCP server does — declare the field once and
// then validate every response against it, so a value that re-encodes as a
// timestamp where the model says "unset" makes the whole response unusable.
//
// The empty string is the case to watch: the API sends it for every live team,
// and encoding/json allocates the pointer before OptionalDateTime.UnmarshalJSON
// runs, so the field survives the round trip as a non-nil pointer to the zero
// time rather than as nil.
func TestTeamDeletedAtEncoding(t *testing.T) {
	tests := []struct {
		name     string
		payload  string
		wantNil  bool
		wantJSON string
	}{{
		name:     "empty string for a live team",
		payload:  `{"team":{"id":"1","name":"Live","deleted":false,"deletedDate":""}}`,
		wantNil:  false,
		wantJSON: "null",
	}, {
		name:     "null",
		payload:  `{"team":{"id":"2","name":"Live","deleted":false,"deletedDate":null}}`,
		wantNil:  true,
		wantJSON: "null",
	}, {
		name:     "absent",
		payload:  `{"team":{"id":"3","name":"Live","deleted":false}}`,
		wantNil:  true,
		wantJSON: "null",
	}, {
		name:     "timestamp for a deleted team",
		payload:  `{"team":{"id":"4","name":"Gone","deleted":true,"deletedDate":"2026-01-02T03:04:05Z"}}`,
		wantNil:  false,
		wantJSON: `"2026-01-02T03:04:05Z"`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response projects.TeamGetResponse
			if err := json.Unmarshal([]byte(tt.payload), &response); err != nil {
				t.Fatalf("unexpected error decoding %s: %s", tt.payload, err)
			}

			if isNil := response.Team.DeletedAt == nil; isNil != tt.wantNil {
				t.Errorf("expected DeletedAt nil to be %t but got %t", tt.wantNil, isNil)
			}

			encoded, err := json.Marshal(response.Team.DeletedAt)
			if err != nil {
				t.Fatalf("unexpected error encoding: %s", err)
			}
			if string(encoded) != tt.wantJSON {
				t.Errorf("expected deletedDate to encode as %s but got %s", tt.wantJSON, encoded)
			}
		})
	}
}

// TestTeamListDeletedAtEncoding covers the same ground as
// TestTeamDeletedAtEncoding for the list envelope, which is where the failure
// was first reported: one live team in a page was enough to reject the response.
func TestTeamListDeletedAtEncoding(t *testing.T) {
	payload := `{"teams":[
		{"id":"1","name":"Live","deleted":false,"deletedDate":""},
		{"id":"2","name":"Gone","deleted":true,"deletedDate":"2026-01-02T03:04:05Z"}
	]}`

	var response projects.TeamListResponse
	if err := json.Unmarshal([]byte(payload), &response); err != nil {
		t.Fatalf("unexpected error decoding: %s", err)
	}
	if len(response.Teams) != 2 {
		t.Fatalf("expected 2 teams but got %d", len(response.Teams))
	}

	encoded, err := json.Marshal(response.Teams)
	if err != nil {
		t.Fatalf("unexpected error encoding: %s", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unexpected error re-decoding: %s", err)
	}

	if got := decoded[0]["deletedDate"]; got != nil {
		t.Errorf("expected the live team to encode deletedDate as null but got %v", got)
	}
	if got := decoded[1]["deletedDate"]; got != "2026-01-02T03:04:05Z" {
		t.Errorf("expected the deleted team to keep its timestamp but got %v", got)
	}
}
