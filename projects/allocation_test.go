package projects_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

// allocationTitle builds a unique title, since a site accumulates allocations
// across runs and a fixed title makes a failure ambiguous.
func allocationTitle() string {
	return fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))
}

func TestAllocationCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	start := twapi.Date(time.Now())
	end := twapi.Date(time.Now().Add(48 * time.Hour))

	tests := []struct {
		name  string
		input projects.AllocationCreateRequest
	}{{
		name: "only required fields",
		input: projects.NewAllocationCreateRequest(
			testResources.ProjectID,
			testResources.UserID,
			allocationTitle(),
			start,
			end,
			4,
			"#3c8f7c",
		),
	}, {
		name: "all fields",
		input: projects.AllocationCreateRequest{
			Allocation: projects.AllocationUpsert{
				ProjectID:              new(testResources.ProjectID),
				AssignedUserID:         new(testResources.UserID),
				Title:                  new(allocationTitle()),
				Description:            new("This is a test allocation"),
				StartDate:              &start,
				EndDate:                &end,
				HoursPerDay:            new(2.5),
				Color:                  new("#3c8f7c"),
				DistributeType:         new(projects.AllocationDistributeTypeDistributed),
				IsBillable:             new(true),
				IgnoreCollisions:       new(true),
				InformOfOverAllocation: new(true),
				LinkedTaskIDs:          []int64{testResources.TaskID},
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			allocation, err := projects.AllocationCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx = context.Background() // t.Context is always canceled in cleanup
				deleteRequest := projects.NewAllocationDeleteRequest(allocation.Allocation.ID)
				deleteRequest.HardDelete = true
				if _, err := projects.AllocationDelete(ctx, engine, deleteRequest); err != nil {
					t.Errorf("failed to delete allocation after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if allocation.Allocation.ID == 0 {
				t.Error("expected a valid allocation ID but got 0")
			}
		})
	}
}

func TestAllocationUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	allocationID, allocationCleanup, err := createAllocation(t, testResources.ProjectID, testResources.UserID)
	if err != nil {
		t.Fatalf("failed to create allocation for test: %s", err)
	}
	t.Cleanup(allocationCleanup)

	tests := []struct {
		name  string
		input projects.AllocationUpdateRequest
	}{{
		name: "all fields",
		input: projects.AllocationUpdateRequest{
			Path: projects.AllocationUpdateRequestPath{
				ID: allocationID,
			},
			Allocation: projects.AllocationUpsert{
				Title:       new(allocationTitle()),
				Description: new("This is an updated test allocation"),
				HoursPerDay: new(3.0),
				IsBillable:  new(false),
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.AllocationUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestAllocationDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	allocationID, _, err := createAllocation(t, testResources.ProjectID, testResources.UserID)
	if err != nil {
		t.Fatalf("failed to create allocation for test: %s", err)
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err := projects.AllocationDelete(ctx, engine, projects.NewAllocationDeleteRequest(allocationID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

// TestAllocationRestore covers the soft-delete pairing: a plain delete leaves
// the allocation restorable, so the two are exercised together.
func TestAllocationRestore(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	allocationID, allocationCleanup, err := createAllocation(t, testResources.ProjectID, testResources.UserID)
	if err != nil {
		t.Fatalf("failed to create allocation for test: %s", err)
	}
	t.Cleanup(allocationCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	t.Cleanup(cancel)

	if _, err := projects.AllocationDelete(ctx, engine, projects.NewAllocationDeleteRequest(allocationID)); err != nil {
		t.Fatalf("failed to delete allocation before restoring it: %s", err)
	}

	restored, err := projects.AllocationRestore(ctx, engine, projects.NewAllocationRestoreRequest(allocationID))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if restored.Allocation.Status != projects.AllocationStatusActive {
		t.Errorf("expected a restored allocation to be active, got %q", restored.Allocation.Status)
	}
}

func TestAllocationTaskLink(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	allocationID, allocationCleanup, err := createAllocation(t, testResources.ProjectID, testResources.UserID)
	if err != nil {
		t.Fatalf("failed to create allocation for test: %s", err)
	}
	t.Cleanup(allocationCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	t.Cleanup(cancel)

	linkRequest := projects.NewAllocationTaskLinkRequest(allocationID, testResources.TaskID)
	if _, err := projects.AllocationTaskLink(ctx, engine, linkRequest); err != nil {
		t.Fatalf("unexpected error linking task: %s", err)
	}

	unlinkRequest := projects.NewAllocationTaskUnlinkRequest(allocationID, testResources.TaskID)
	if _, err := projects.AllocationTaskUnlink(ctx, engine, unlinkRequest); err != nil {
		t.Errorf("unexpected error unlinking task: %s", err)
	}
}

func TestAllocationGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	allocationID, allocationCleanup, err := createAllocation(t, testResources.ProjectID, testResources.UserID)
	if err != nil {
		t.Fatalf("failed to create allocation for test: %s", err)
	}
	t.Cleanup(allocationCleanup)

	tests := []struct {
		name  string
		input projects.AllocationGetRequest
	}{{
		name:  "all fields",
		input: projects.NewAllocationGetRequest(allocationID),
	}, {
		name: "sparse fields",
		input: func() projects.AllocationGetRequest {
			req := projects.NewAllocationGetRequest(allocationID)
			req.Fields.Allocation = []projects.AllocationField{
				projects.AllocationFieldID,
				projects.AllocationFieldTitle,
			}
			return req
		}(),
	}, {
		name: "with sideloads",
		input: func() projects.AllocationGetRequest {
			req := projects.NewAllocationGetRequest(allocationID)
			req.Include = []projects.AllocationSideload{
				projects.AllocationSideloadProjects,
				projects.AllocationSideloadAssignee,
			}
			return req
		}(),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			allocation, err := projects.AllocationGet(ctx, engine, tt.input)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if allocation.Allocation.ID != allocationID {
				t.Errorf("expected allocation %d, got %d", allocationID, allocation.Allocation.ID)
			}
		})
	}
}

func TestAllocationList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	_, allocationCleanup, err := createAllocation(t, testResources.ProjectID, testResources.UserID)
	if err != nil {
		t.Fatalf("failed to create allocation for test: %s", err)
	}
	t.Cleanup(allocationCleanup)

	start := twapi.Date(time.Now().Add(-24 * time.Hour))
	end := twapi.Date(time.Now().Add(30 * 24 * time.Hour))

	tests := []struct {
		name  string
		input projects.AllocationListRequest
	}{{
		name:  "all allocations",
		input: projects.NewAllocationListRequest(),
	}, {
		name: "explicit window",
		input: func() projects.AllocationListRequest {
			req := projects.NewAllocationListRequest()
			req.Filters.StartDate = &start
			req.Filters.EndDate = &end
			return req
		}(),
	}, {
		name: "filtered by project and user",
		input: func() projects.AllocationListRequest {
			req := projects.NewAllocationListRequest()
			req.Filters.StartDate = &start
			req.Filters.EndDate = &end
			req.Filters.ProjectIDs = []int64{testResources.ProjectID}
			req.Filters.AssignedUserIDs = []int64{testResources.UserID}
			return req
		}(),
	}, {
		name: "ordered with an exact count",
		input: func() projects.AllocationListRequest {
			req := projects.NewAllocationListRequest()
			req.Filters.StartDate = &start
			req.Filters.EndDate = &end
			req.Filters.OrderBy = projects.AllocationOrderByStartDate
			req.Filters.OrderMode = twapi.OrderModeDescending
			req.Filters.CountMode = twapi.ListCountModeExact
			return req
		}(),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			allocations, err := projects.AllocationList(ctx, engine, tt.input)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if len(allocations.Allocations) == 0 {
				t.Error("expected at least one allocation but got 0")
			}
		})
	}
}
