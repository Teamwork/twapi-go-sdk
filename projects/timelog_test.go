package projects_test

import (
	"context"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestTimelogCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name  string
		input projects.TimelogCreateRequest
	}{{
		name: "only required fields for task",
		input: projects.NewTimelogCreateRequestInTask(
			testResources.TaskID,
			time.Now(),
			30*time.Minute,
		),
	}, {
		name: "only required fields for project",
		input: projects.NewTimelogCreateRequestInProject(
			testResources.ProjectID,
			time.Now(),
			30*time.Minute,
		),
	}, {
		name: "all fields for task",
		input: projects.TimelogCreateRequest{
			Path: projects.TimelogCreateRequestPath{
				TaskID: testResources.TaskID,
			},
			Description: new("This is a test timelog"),
			Date:        twapi.Date(time.Now().UTC()),
			Time:        twapi.Time(time.Now().UTC()),
			IsUTC:       true,
			Hours:       2,
			Minutes:     30,
			Billable:    new(true),
			UserID:      &testResources.UserID,
			TagIDs:      []int64{testResources.TagID},
		},
	}, {
		name: "all fields for project",
		input: projects.TimelogCreateRequest{
			Path: projects.TimelogCreateRequestPath{
				ProjectID: testResources.ProjectID,
			},
			Description: new("This is a test timelog"),
			Date:        twapi.Date(time.Now().UTC()),
			Time:        twapi.Time(time.Now().UTC()),
			IsUTC:       true,
			Hours:       2,
			Minutes:     30,
			Billable:    new(true),
			UserID:      &testResources.UserID,
			TagIDs:      []int64{testResources.TagID},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			timelogResponse, err := projects.TimelogCreate(ctx, engine, tt.input)
			defer func() {
				if err != nil {
					return
				}
				ctx = context.Background() // t.Context is always canceled in cleanup
				_, err := projects.TimelogDelete(ctx, engine, projects.NewTimelogDeleteRequest(timelogResponse.Timelog.ID))
				if err != nil {
					t.Errorf("failed to delete timelog after test: %s", err)
				}
			}()
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if timelogResponse.Timelog.ID == 0 {
				t.Error("expected a valid timelog ID but got 0")
			}
		})
	}
}

func TestTimelogUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	timelogID, timelogCleanup, err := createTimelogInTask(t, testResources.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(timelogCleanup)

	tests := []struct {
		name  string
		input projects.TimelogUpdateRequest
	}{{
		name: "all fields",
		input: projects.TimelogUpdateRequest{
			Path: projects.TimelogUpdateRequestPath{
				ID: timelogID,
			},
			Description: new("This is a test timelog"),
			Date:        new(twapi.Date(time.Now().UTC())),
			Time:        new(twapi.Time(time.Now().UTC())),
			IsUTC:       new(true),
			Hours:       new(int64(2)),
			Minutes:     new(int64(30)),
			Billable:    new(true),
			UserID:      &testResources.UserID,
			TagIDs:      []int64{testResources.TagID},
		},
	}, {
		name: "detached from its task",
		input: projects.TimelogUpdateRequest{
			Path: projects.TimelogUpdateRequestPath{
				ID: timelogID,
			},
			ProjectID: &testResources.ProjectID,
			TaskID:    twapi.NullInt64(),
		},
	}, {
		name: "moved back to a task",
		input: projects.TimelogUpdateRequest{
			Path: projects.TimelogUpdateRequestPath{
				ID: timelogID,
			},
			TaskID: twapi.NewNullableInt64(testResources.TaskID),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.TimelogUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestTimelogDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	timelogID, timelogCleanup, err := createTimelogInTask(t, testResources.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(timelogCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.TimelogDelete(ctx, engine, projects.NewTimelogDeleteRequest(timelogID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestTimelogGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	timelogID, timelogCleanup, err := createTimelogInTask(t, testResources.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(timelogCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.TimelogGet(ctx, engine, projects.NewTimelogGetRequest(timelogID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestTimelogList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	_, taskTimelogCleanup, err := createTimelogInTask(t, testResources.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(taskTimelogCleanup)

	_, projectTimelogCleanup, err := createTimelogInProject(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(projectTimelogCleanup)

	tests := []struct {
		name  string
		input projects.TimelogListRequest
	}{{
		name: "all timelogs",
	}, {
		name: "timelogs for task",
		input: projects.TimelogListRequest{
			Path: projects.TimelogListRequestPath{
				TaskID: testResources.TaskID,
			},
		},
	}, {
		name: "timelogs for project",
		input: projects.TimelogListRequest{
			Path: projects.TimelogListRequestPath{
				ProjectID: testResources.ProjectID,
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.TimelogList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

// TestTimelogListFiltersApplied pins the whole query string the timelog list
// builds when every filter is populated, against the parameter names the
// endpoint documents:
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/get-projects-api-v3-time-json
//
// Comparing the complete map rather than a subset is deliberate: an unrecognised
// query key is silently ignored by the API, so a misspelled parameter looks
// exactly like a working one from the caller's side.
func TestTimelogListFiltersApplied(t *testing.T) {
	startDate := time.Date(2026, 3, 4, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 5, 6, 0, 0, 0, 0, time.UTC)

	req := projects.TimelogListRequest{
		Filters: projects.TimelogListRequestFilters{
			OrderBy:   projects.TimelogOrderByDate,
			OrderMode: twapi.OrderModeDescending,

			TagIDs:       []int64{777, 888},
			MatchAllTags: new(true),

			StartDate: &startDate,
			EndDate:   &endDate,

			AssignedToUserIDs:    []int64{11, 22},
			AssignedToCompanyIDs: []int64{33},
			AssignedToTeamIDs:    []int64{44},
			DeskTicketIDs:        []int64{55},

			BillableType: projects.TimelogBillableTypeBillable,
			InvoicedType: projects.TimelogInvoicedTypeNonInvoiced,

			Page:      2,
			PageSize:  25,
			CountMode: twapi.ListCountModeExact,

			Fields: projects.TimelogListFields{
				Timelogs: []projects.TimelogField{projects.TimelogFieldDescription},
			},
		},
	}

	want := map[string]string{
		"orderBy":   "date",
		"orderMode": "desc",

		"tagIds":       "777,888",
		"matchAllTags": "true",

		"startDate": "2026-03-04T00:00:00Z",
		"endDate":   "2026-05-06T00:00:00Z",

		"assignedToUserIds":    "11,22",
		"assignedToCompanyIds": "33",
		"assignedToTeamIds":    "44",
		"deskTicketIds":        "55",

		"billableType": "billable",
		"invoicedType": "noninvoiced",

		"page":       "2",
		"pageSize":   "25",
		"skipCounts": "false",

		"fields[timelogs]": "description",
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

// TestTimelogListFiltersUnset checks the zero-value filters send nothing, so the
// endpoint applies its own defaults rather than being pinned to "all".
func TestTimelogListFiltersUnset(t *testing.T) {
	query := listQuery(t, projects.TimelogListRequest{})
	if len(query) != 0 {
		t.Errorf("expected no query parameters but got %v", query)
	}
}
