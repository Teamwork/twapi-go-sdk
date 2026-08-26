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
			Description: new("This is a test milestone"),
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
			Name:        new(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Description: new("This is a test milestone"),
			DueAt:       new(projects.NewLegacyDate(time.Now().Add(48 * time.Hour))),
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

	dueAfter := twapi.Date(time.Now().AddDate(0, 0, -30))
	dueBefore := twapi.Date(time.Now().AddDate(0, 0, 30))

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
	}, {
		name: "incomplete milestones in a deadline range",
		input: projects.MilestoneListRequest{
			Filters: projects.MilestoneListRequestFilters{
				Statuses:  []projects.MilestoneStatus{projects.MilestoneStatusIncomplete},
				DueAfter:  &dueAfter,
				DueBefore: &dueBefore,
			},
		},
	}, {
		name: "late milestones including completed ones",
		input: projects.MilestoneListRequest{
			Filters: projects.MilestoneListRequestFilters{
				Statuses:         []projects.MilestoneStatus{projects.MilestoneStatusLate},
				IncludeCompleted: new(true),
			},
		},
	}, {
		name: "deleted milestones",
		input: projects.MilestoneListRequest{
			Filters: projects.MilestoneListRequestFilters{
				Statuses: []projects.MilestoneStatus{projects.MilestoneStatusDeleted},
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

// TestMilestoneListFiltersApplied pins the whole query string the milestone list
// builds when every filter is populated, against the parameter names the
// endpoint documents:
//
// https://apidocs.teamwork.com/docs/teamwork/v3/milestones/get-projects-api-v3-milestones-json
//
// Comparing the complete map rather than a subset is deliberate: an unrecognised
// query key is silently ignored by the API, so a misspelled parameter looks
// exactly like a working one from the caller's side.
func TestMilestoneListFiltersApplied(t *testing.T) {
	dueAfter := twapi.Date(time.Date(2026, 3, 4, 0, 0, 0, 0, time.UTC))
	dueBefore := twapi.Date(time.Date(2026, 5, 6, 0, 0, 0, 0, time.UTC))

	req := projects.MilestoneListRequest{
		Filters: projects.MilestoneListRequestFilters{
			SearchTerm: "acme",

			OrderBy:   projects.MilestoneOrderByName,
			OrderMode: twapi.OrderModeDescending,

			Statuses: []projects.MilestoneStatus{
				projects.MilestoneStatusLate,
				projects.MilestoneStatusUpcoming,
			},
			IncludeCompleted: new(true),
			ShowDeleted:      new(false),
			DueAfter:         &dueAfter,
			DueBefore:        &dueBefore,

			TagIDs:       []int64{777, 888},
			MatchAllTags: new(true),

			Page:      2,
			PageSize:  25,
			CountMode: twapi.ListCountModeExact,

			Fields: projects.MilestoneListFields{
				Milestones: []projects.MilestoneField{projects.MilestoneFieldName},
			},
		},
	}

	want := map[string]string{
		"searchTerm": "acme",

		"orderBy":   "name",
		"orderMode": "desc",

		"status":           "late,upcoming",
		"includeCompleted": "true",
		"showDeleted":      "false",
		"dueAfter":         "2026-03-04",
		"dueBefore":        "2026-05-06",

		"tagIds":       "777,888",
		"matchAllTags": "true",

		"page":       "2",
		"pageSize":   "25",
		"skipCounts": "false",

		"fields[milestones]": "name",
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

// TestMilestoneListFiltersUnset checks the zero-value filters send nothing, so
// the endpoint applies its own defaults rather than receiving a wall of false.
func TestMilestoneListFiltersUnset(t *testing.T) {
	query := listQuery(t, projects.MilestoneListRequest{})
	if len(query) != 0 {
		t.Errorf("expected no query parameters but got %v", query)
	}
}

// TestMilestoneListShowDeletedImplied covers the one filter that is not sent
// verbatim: the endpoint drops deleted milestones before it applies the status
// filter, so asking for MilestoneStatusDeleted without showDeleted would match
// nothing. The SDK turns the flag on, unless the caller set it themselves.
func TestMilestoneListShowDeletedImplied(t *testing.T) {
	tests := []struct {
		name    string
		filters projects.MilestoneListRequestFilters
		want    string
	}{{
		name: "deleted status implies the flag",
		filters: projects.MilestoneListRequestFilters{
			Statuses: []projects.MilestoneStatus{projects.MilestoneStatusDeleted},
		},
		want: "true",
	}, {
		name: "deleted status alongside another one implies the flag",
		filters: projects.MilestoneListRequestFilters{
			Statuses: []projects.MilestoneStatus{
				projects.MilestoneStatusLate,
				projects.MilestoneStatusDeleted,
			},
		},
		want: "true",
	}, {
		name: "an explicit false is respected",
		filters: projects.MilestoneListRequestFilters{
			Statuses:    []projects.MilestoneStatus{projects.MilestoneStatusDeleted},
			ShowDeleted: new(false),
		},
		want: "false",
	}, {
		name: "an explicit true is kept",
		filters: projects.MilestoneListRequestFilters{
			ShowDeleted: new(true),
		},
		want: "true",
	}, {
		name: "another status implies nothing",
		filters: projects.MilestoneListRequestFilters{
			Statuses: []projects.MilestoneStatus{projects.MilestoneStatusLate},
		},
		want: "",
	}, {
		name:    "no statuses imply nothing",
		filters: projects.MilestoneListRequestFilters{},
		want:    "",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := listQuery(t, projects.MilestoneListRequest{Filters: tt.filters})
			if got := query.Get("showDeleted"); got != tt.want {
				t.Errorf("expected showDeleted=%q but got %q", tt.want, got)
			}
		})
	}
}
