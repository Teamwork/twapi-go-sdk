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
			Handle:       twapi.Ptr(fmt.Sprintf("testhandle%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Description:  twapi.Ptr("This is a test team."),
			ParentTeamID: &parentTeamID,
			CompanyID:    &testResources.CompanyID,
			UserIDs:      []int64{testResources.UserID},
		},
	}, {
		name: "all fields for project",
		input: projects.TeamCreateRequest{
			Name:         fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
			Handle:       twapi.Ptr(fmt.Sprintf("testhandle%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Description:  twapi.Ptr("This is a test team."),
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
			Name:        twapi.Ptr(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Handle:      twapi.Ptr(fmt.Sprintf("testhandle%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Description: twapi.Ptr("This is a test team."),
			CompanyID:   &testResources.CompanyID,
			UserIDs:     []int64{testResources.UserID},
		},
	}, {
		name: "all fields for project",
		input: projects.TeamUpdateRequest{
			Path: projects.TeamUpdateRequestPath{
				ID: teamID,
			},
			Name:        twapi.Ptr(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Handle:      twapi.Ptr(fmt.Sprintf("testhandle%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Description: twapi.Ptr("This is a test team."),
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
