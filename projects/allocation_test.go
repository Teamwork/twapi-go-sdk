package projects_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"strings"
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
			4*60*60,
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
				SecondsPerDay:          new(int64(2*60*60 + 30*60)),
				Color:                  new("#3c8f7c"),
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
				Title:         new(allocationTitle()),
				Description:   new("This is an updated test allocation"),
				SecondsPerDay: new(int64(3 * 60 * 60)),
				IsBillable:    new(false),
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

// TestAllocationDelete covers the hard delete, which leaves nothing behind. The
// soft delete is covered by TestAllocationRestore, which needs a recoverable
// row; doing it here as well would strand one on the site every run, since a
// soft-deleted allocation is still there afterwards.
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

	request := projects.NewAllocationDeleteRequest(allocationID)
	request.HardDelete = true
	if _, err := projects.AllocationDelete(ctx, engine, request); err != nil {
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

// TestAllocationGetResponseDecodesSideloads pins that the sideloads a caller
// asked for actually survive decoding. The response type carried no Included
// struct at first, so `include=projects,assignee` was honoured by the API and
// then silently discarded here — the request looked right and the related
// objects simply never arrived.
//
// It also covers the jobRoles key, which the response spells in camelCase while
// the endpoint reads its sparse fieldset from fields[jobroles].
func TestAllocationGetResponseDecodesSideloads(t *testing.T) {
	body := `{
		"allocation": {
			"id": 12345,
			"title": "Design phase",
			"project": {"id": 777, "type": "projects"},
			"assignedUser": {"id": 456, "type": "users"},
			"startedAt": "2026-08-27",
			"endedAt": "2026-08-28",
			"secondsPerDay": 28800,
			"canViewFinancialDetails": true,
			"financialDetails": {
				"forecastedRevenue": 960,
				"forecastedCost": 640,
				"billableRates": [{
					"userEffectiveRate": {"id": 888, "type": "userEffectiveRate"},
					"billableRate": 60,
					"allocatedMinutes": 960,
					"forecastedRevenue": 960,
					"source": "userprojectrate",
					"effectiveDate": "2023-08-17",
					"currency": {"id": 3, "type": "currencies"}
				}],
				"costRates": [{
					"userCostRate": {"id": 999, "type": "costRates"},
					"costRate": 40,
					"allocatedMinutes": 960,
					"forecastedCost": 640,
					"currency": {"id": 3, "type": "currencies"},
					"source": "userCostRate"
				}]
			}
		},
		"included": {
			"companies": {"111": {"id": 111, "name": "Example Company"}},
			"projects": {"777": {"id": 777, "name": "Example Project", "companyId": 111}},
			"users": {"456": {"id": 456, "firstName": "John", "lastName": "Doe"}},
			"jobRoles": {"222": {"id": 222, "name": "Creative Director"}},
			"tasks": {"333": {"id": 333, "name": "Example Task"}}
		}
	}`

	var response projects.AllocationGetResponse
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		t.Fatalf("failed to decode response: %s", err)
	}

	if got := response.Included.Projects["777"].Name; got != "Example Project" {
		t.Errorf("expected the project sideload to decode, got %q", got)
	}
	if got := response.Included.Companies["111"].Name; got != "Example Company" {
		t.Errorf("expected the company sideload to decode, got %q", got)
	}
	if got := response.Included.Users["456"].FirstName; got != "John" {
		t.Errorf("expected the user sideload to decode, got %q", got)
	}
	if got := response.Included.JobRoles["222"].Name; got != "Creative Director" {
		t.Errorf("expected the jobRoles sideload to decode, got %q", got)
	}
	if got := response.Included.Tasks["333"].Name; got != "Example Task" {
		t.Errorf("expected the task sideload to decode, got %q", got)
	}

	// the inline financial figures are what make the rate sideloads unnecessary,
	// so they have to survive the same round trip
	financial := response.Allocation.FinancialDetails
	if financial.ForecastedRevenue == nil || *financial.ForecastedRevenue != 960 {
		t.Errorf("expected forecasted revenue to decode, got %v", financial.ForecastedRevenue)
	}
	if len(financial.BillableRates) != 1 {
		t.Fatalf("expected one billable rate period, got %d", len(financial.BillableRates))
	}
	if got := financial.BillableRates[0].Source; got == nil ||
		*got != projects.EffectiveRateSourceUserProjectRate {
		t.Errorf("expected the billable rate source to decode as a typed value, got %v", got)
	}
	if len(financial.CostRates) != 1 {
		t.Fatalf("expected one cost rate period, got %d", len(financial.CostRates))
	}
	if got := financial.CostRates[0].Source; got == nil || *got != projects.CostRateSourceUserCostRate {
		t.Errorf("expected the cost rate source to decode as a typed value, got %v", got)
	}
}

// TestAllocationDeleteRequestBody pins the body of a delete, which is where
// this endpoint is unusual: hardDelete travels in a request body rather than a
// query parameter, and it decides whether the allocation can be restored
// afterwards. Nothing else exercises it offline — the example server ignores
// the body, and the live tests skip without an engine.
func TestAllocationDeleteRequestBody(t *testing.T) {
	tests := []struct {
		name string
		hard bool
		want string
	}{
		{name: "soft delete", hard: false, want: `{"hardDelete":false}`},
		{name: "hard delete", hard: true, want: `{"hardDelete":true}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := projects.NewAllocationDeleteRequest(12345)
			request.HardDelete = tt.hard

			httpRequest, err := request.HTTPRequest(t.Context(), "https://test.teamwork.com")
			if err != nil {
				t.Fatalf("unexpected error building the request: %s", err)
			}
			if want := "/projects/api/v3/allocations/12345.json"; httpRequest.URL.Path != want {
				t.Errorf("expected path %q, got %q", want, httpRequest.URL.Path)
			}

			body, err := io.ReadAll(httpRequest.Body)
			if err != nil {
				t.Fatalf("failed to read the request body: %s", err)
			}
			if got := strings.TrimSpace(string(body)); got != tt.want {
				t.Errorf("expected body %s, got %s", tt.want, got)
			}
		})
	}
}

// TestAllocationRequestsRequireAnID pins the client-side guard on every request
// that addresses one allocation. Without it a zero identifier is formatted into
// the path and comes back as a 404 from the server, which reads as "no such
// allocation" rather than "you did not pass one".
func TestAllocationRequestsRequireAnID(t *testing.T) {
	requests := map[string]twapi.HTTPRequester{
		"get":     projects.NewAllocationGetRequest(0),
		"update":  projects.NewAllocationUpdateRequest(0),
		"delete":  projects.NewAllocationDeleteRequest(0),
		"restore": projects.NewAllocationRestoreRequest(0),
		"link":    projects.NewAllocationTaskLinkRequest(0, 999),
		"unlink":  projects.NewAllocationTaskUnlinkRequest(0, 999),
	}

	for name, request := range requests {
		t.Run(name, func(t *testing.T) {
			if _, err := request.HTTPRequest(t.Context(), "https://test.teamwork.com"); err == nil {
				t.Error("expected a missing allocation ID to be rejected before the request is sent")
			}
		})
	}
}

// TestAllocationIncludesAreOneCommaSeparatedParam pins how the sideload list is
// encoded. The endpoint reads only the first `include` parameter it receives and
// silently drops repeated ones, so a request adding them one at a time comes
// back with the first sideload and nothing else — successfully, which is what
// makes it easy to miss. Asserting the raw query rather than the parsed values
// is deliberate: url.Values renders both encodings as a slice, so a parsed
// comparison passes either way.
func TestAllocationIncludesAreOneCommaSeparatedParam(t *testing.T) {
	sideloads := []projects.AllocationSideload{
		projects.AllocationSideloadProjects,
		projects.AllocationSideloadAssignee,
		projects.AllocationSideloadAssigneeJobRoles,
	}

	listRequest := projects.NewAllocationListRequest()
	listRequest.Filters.Include = sideloads

	getRequest := projects.NewAllocationGetRequest(12345)
	getRequest.Include = sideloads

	for name, request := range map[string]twapi.HTTPRequester{
		"list": listRequest,
		"get":  getRequest,
	} {
		t.Run(name, func(t *testing.T) {
			httpRequest, err := request.HTTPRequest(t.Context(), "https://test.teamwork.com")
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			raw := httpRequest.URL.RawQuery
			if got := strings.Count(raw, "include="); got != 1 {
				t.Errorf("expected exactly one include parameter, got %d in %q", got, raw)
			}
			if want := "include=projects%2Cassignee%2Cassignee.jobRoles"; !strings.Contains(raw, want) {
				t.Errorf("expected %s in the query, got %q", want, raw)
			}
		})
	}
}
