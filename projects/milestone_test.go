package projects_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestMilestoneCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	projectID, projectCleanup, err := createProject(t)
	if err != nil {
		t.Fatal(err)
	}
	defer projectCleanup()

	userID, userCleanup, err := createUser(t)
	if err != nil {
		t.Fatal(err)
	}
	defer userCleanup()

	if err = addProjectMember(t, projectID, userID); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		input         projects.MilestoneCreateRequest
		expectedError bool
	}{{
		name: "it should create a milestone with valid input",
		input: projects.NewMilestoneCreateRequest(
			projectID,
			fmt.Sprintf("Test Milestone %d", time.Now().UnixNano()),
			projects.NewLegacyDate(time.Now()),
			projects.LegacyUserGroups{
				UserIDs: []int64{userID},
			},
		),
	}, {
		name: "it should fail to create a milestone with missing name",
		input: projects.MilestoneCreateRequest{
			Description: twapi.Ptr("This milestone has no name"),
		},
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			milestone, err := projects.MilestoneCreate(ctx, engine, tt.input)
			defer func() {
				if err != nil {
					return
				}
				_, err := projects.MilestoneDelete(ctx, engine, projects.NewMilestoneDeleteRequest(int64(milestone.ID)))
				if err != nil {
					t.Errorf("failed to delete milestone after test: %s", err)
				}
			}()

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %s", err)
				return
			}
			if milestone.ID == 0 {
				t.Error("expected a valid milestone ID but got 0")
			}
		})
	}
}

func TestMilestoneUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	projectID, projectCleanup, err := createProject(t)
	if err != nil {
		t.Fatal(err)
	}
	defer projectCleanup()

	userID, userCleanup, err := createUser(t)
	if err != nil {
		t.Fatal(err)
	}
	defer userCleanup()

	if err = addProjectMember(t, projectID, userID); err != nil {
		t.Fatal(err)
	}

	milestoneID, milestoneCleanup, err := createMilestone(t, projectID, projects.LegacyUserGroups{
		UserIDs: []int64{userID},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer milestoneCleanup()

	tests := []struct {
		name          string
		input         projects.MilestoneUpdateRequest
		expectedError bool
	}{{
		name: "it should update a milestone with valid input",
		input: projects.MilestoneUpdateRequest{
			Path: projects.MilestoneUpdateRequestPath{
				ID: milestoneID,
			},
			Description: twapi.Ptr("This is a test milestone"),
		},
	}, {
		name: "it should fail to update a milestone with missing name",
		input: projects.MilestoneUpdateRequest{
			Path: projects.MilestoneUpdateRequestPath{
				ID: milestoneID,
			},
			Name: twapi.Ptr(""),
		},
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.MilestoneUpdate(ctx, engine, tt.input)
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %s", err)
				return
			}
		})
	}
}

func TestMilestoneDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	projectID, projectCleanup, err := createProject(t)
	if err != nil {
		t.Fatal(err)
	}
	defer projectCleanup()

	userID, userCleanup, err := createUser(t)
	if err != nil {
		t.Fatal(err)
	}
	defer userCleanup()

	if err = addProjectMember(t, projectID, userID); err != nil {
		t.Fatal(err)
	}

	milestoneID, _, err := createMilestone(t, projectID, projects.LegacyUserGroups{
		UserIDs: []int64{userID},
	})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		input         projects.MilestoneDeleteRequest
		expectedError bool
	}{{
		name:  "it should delete a milestone with valid input",
		input: projects.NewMilestoneDeleteRequest(milestoneID),
	}, {
		name:          "it should fail to delete an unknown milestone",
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.MilestoneDelete(ctx, engine, tt.input)
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %s", err)
				return
			}
		})
	}
}

func TestMilestoneGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	projectID, projectCleanup, err := createProject(t)
	if err != nil {
		t.Fatal(err)
	}
	defer projectCleanup()

	userID, userCleanup, err := createUser(t)
	if err != nil {
		t.Fatal(err)
	}
	defer userCleanup()

	if err = addProjectMember(t, projectID, userID); err != nil {
		t.Fatal(err)
	}

	milestoneID, milestoneCleanup, err := createMilestone(t, projectID, projects.LegacyUserGroups{
		UserIDs: []int64{userID},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer milestoneCleanup()

	tests := []struct {
		name          string
		input         projects.MilestoneGetRequest
		expectedError bool
	}{{
		name:  "it should retrieve a milestone with valid input",
		input: projects.NewMilestoneGetRequest(milestoneID),
	}, {
		name:          "it should fail to retrieve an unknown milestone",
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.MilestoneGet(ctx, engine, tt.input)
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %s", err)
				return
			}
		})
	}
}

func TestMilestoneList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	projectID, projectCleanup, err := createProject(t)
	if err != nil {
		t.Fatal(err)
	}
	defer projectCleanup()

	userID, userCleanup, err := createUser(t)
	if err != nil {
		t.Fatal(err)
	}
	defer userCleanup()

	if err = addProjectMember(t, projectID, userID); err != nil {
		t.Fatal(err)
	}

	_, milestoneCleanup, err := createMilestone(t, projectID, projects.LegacyUserGroups{
		UserIDs: []int64{userID},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer milestoneCleanup()

	tests := []struct {
		name          string
		input         projects.MilestoneListRequest
		expectedError bool
	}{{
		name: "it should list milestones",
	}, {
		name: "it should list milestones for project",
		input: projects.MilestoneListRequest{
			Path: projects.MilestoneListRequestPath{
				ProjectID: projectID,
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.MilestoneList(ctx, engine, tt.input)
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %s", err)
				return
			}
		})
	}
}
