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

func TestMilestoneCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name  string
		input projects.MilestoneCreateRequest
	}{{
		name: "only required fields",
		input: projects.NewMilestoneCreateRequest(
			testResources.ProjectID,
			fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
			projects.NewLegacyDate(time.Now()),
			projects.LegacyUserGroups{
				UserIDs: []int64{testResources.UserID},
			},
		),
	}, {
		name: "all fields",
		input: projects.MilestoneCreateRequest{
			Path: projects.MilestoneCreateRequestPath{
				ProjectID: testResources.ProjectID,
			},
			Name:        fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
			Description: twapi.Ptr("This is a test milestone"),
			DueAt:       projects.NewLegacyDate(time.Now().Add(48 * time.Hour)),
			TasklistIDs: []int64{testResources.TasklistID},
			TagIDs:      []int64{testResources.TagID},
			Assignees: projects.LegacyUserGroups{
				UserIDs: []int64{testResources.UserID},
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			milestone, err := projects.MilestoneCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx = context.Background() // t.Context is always canceled in cleanup
				_, err := projects.MilestoneDelete(ctx, engine, projects.NewMilestoneDeleteRequest(int64(milestone.ID)))
				if err != nil {
					t.Errorf("failed to delete milestone after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if milestone.ID == 0 {
				t.Error("expected a valid milestone ID but got 0")
			}
		})
	}
}

func TestMilestoneUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	milestoneID, milestoneCleanup, err := createMilestone(t, testResources.ProjectID, projects.LegacyUserGroups{
		UserIDs: []int64{testResources.UserID},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(milestoneCleanup)

	tests := []struct {
		name  string
		input projects.MilestoneUpdateRequest
	}{{
		name: "all fields",
		input: projects.MilestoneUpdateRequest{
			Path: projects.MilestoneUpdateRequestPath{
				ID: milestoneID,
			},
			Name:        twapi.Ptr(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Description: twapi.Ptr("This is a test milestone"),
			DueAt:       twapi.Ptr(projects.NewLegacyDate(time.Now().Add(48 * time.Hour))),
			TasklistIDs: []int64{testResources.TasklistID},
			TagIDs:      []int64{testResources.TagID},
			Assignees: &projects.LegacyUserGroups{
				UserIDs: []int64{testResources.UserID},
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.MilestoneUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestMilestoneDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	milestoneID, _, err := createMilestone(t, testResources.ProjectID, projects.LegacyUserGroups{
		UserIDs: []int64{testResources.UserID},
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.MilestoneDelete(ctx, engine, projects.NewMilestoneDeleteRequest(milestoneID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestMilestoneGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	milestoneID, milestoneCleanup, err := createMilestone(t, testResources.ProjectID, projects.LegacyUserGroups{
		UserIDs: []int64{testResources.UserID},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(milestoneCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.MilestoneGet(ctx, engine, projects.NewMilestoneGetRequest(milestoneID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestMilestoneList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	_, milestoneCleanup, err := createMilestone(t, testResources.ProjectID, projects.LegacyUserGroups{
		UserIDs: []int64{testResources.UserID},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(milestoneCleanup)

	tests := []struct {
		name  string
		input projects.MilestoneListRequest
	}{{
		name: "all milestones",
	}, {
		name: "milestones for project",
		input: projects.MilestoneListRequest{
			Path: projects.MilestoneListRequestPath{
				ProjectID: testResources.ProjectID,
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.MilestoneList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
