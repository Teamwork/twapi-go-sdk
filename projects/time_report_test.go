package projects_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestTimeReportList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	newRequest := func(reportType projects.TimeReportType) projects.TimeReportListRequest {
		req := projects.NewTimeReportListRequest(
			reportType,
			twapi.Date(time.Now().AddDate(0, 0, -7)),
			twapi.Date(time.Now()),
		)
		req.Filters.Include = []projects.TimeReportSideload{
			projects.TimeReportSideloadUsers,
			projects.TimeReportSideloadProjects,
			projects.TimeReportSideloadTasks,
		}
		return req
	}

	tests := []struct {
		name  string
		input projects.TimeReportListRequest
	}{{
		name:  "grouped by user",
		input: newRequest(projects.TimeReportTypeUser),
	}, {
		name:  "grouped by project",
		input: newRequest(projects.TimeReportTypeProject),
	}, {
		name:  "grouped by task",
		input: newRequest(projects.TimeReportTypeTask),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.TimeReportList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestTimeReportTotals(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	newRequest := func(groupBy projects.TimeReportGroupBy) projects.TimeReportTotalsRequest {
		return projects.NewTimeReportTotalsRequest(
			groupBy,
			twapi.Date(time.Now().AddDate(0, -2, 0)),
			twapi.Date(time.Now()),
		)
	}

	tests := []struct {
		name  string
		input projects.TimeReportTotalsRequest
	}{{
		name:  "by day",
		input: newRequest(projects.TimeReportGroupByDay),
	}, {
		name:  "by week",
		input: newRequest(projects.TimeReportGroupByWeek),
	}, {
		name:  "by month",
		input: newRequest(projects.TimeReportGroupByMonth),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.TimeReportTotals(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

// TestTimeReportTotalsRequestGeneration pins the query string, since the
// endpoint ignores an unknown parameter rather than rejecting it.
func TestTimeReportTotalsRequestGeneration(t *testing.T) {
	includeArchived := true
	includeCompleted := false

	req := projects.NewTimeReportTotalsRequest(
		projects.TimeReportGroupByWeek,
		twapi.Date(time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)),
		twapi.Date(time.Date(2026, time.August, 31, 0, 0, 0, 0, time.UTC)),
	)
	req.Filters.ProjectIDs = []int64{777, 778}
	req.Filters.UserIDs = []int64{12345}
	req.Filters.TaskIDs = []int64{1}
	req.Filters.TasklistIDs = []int64{2}
	req.Filters.TeamIDs = []int64{3}
	req.Filters.CompanyIDs = []int64{4}
	req.Filters.TimelogTagIDs = []int64{5, 6}
	req.Filters.IncludeArchivedProjects = &includeArchived
	req.Filters.IncludeCompletedTasks = &includeCompleted

	httpReq, err := req.HTTPRequest(context.Background(), "https://test.com")
	if err != nil {
		t.Fatalf("unexpected error creating HTTP request: %s", err)
	}
	if httpReq.URL.Path != "/projects/api/v3/time/report/totals.json" {
		t.Errorf("unexpected path %q", httpReq.URL.Path)
	}

	query, err := url.ParseQuery(httpReq.URL.RawQuery)
	if err != nil {
		t.Fatalf("failed to parse query string: %s", err)
	}
	want := map[string]string{
		"startDate":               "2026-08-01",
		"endDate":                 "2026-08-31",
		"groupBy":                 "week",
		"projectIds":              "777,778",
		"userIds":                 "12345",
		"taskIds":                 "1",
		"tasklistIds":             "2",
		"teamIds":                 "3",
		"companyIds":              "4",
		"timelogTagIds":           "5,6",
		"includeArchivedProjects": "true",
		"includeCompletedTasks":   "false",
	}
	for key, value := range want {
		if got := query.Get(key); got != value {
			t.Errorf("expected %s=%s, got %q", key, value, got)
		}
	}
	for _, key := range []string{"type", "reportType"} {
		if _, ok := query[key]; ok {
			t.Errorf("expected %s to be absent, got %q", key, query.Get(key))
		}
	}
}

func TestTimeReportTotalsRequestGeneration_type(t *testing.T) {
	req := projects.NewTimeReportTotalsRequest(
		projects.TimeReportGroupByMonth,
		twapi.Date(time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)),
		twapi.Date(time.Date(2026, time.June, 30, 0, 0, 0, 0, time.UTC)),
	)
	req.Filters.Type = projects.TimeReportTypeUser
	req.Filters.ReportType = projects.TimeReportReportTypeLoggedTime

	httpReq, err := req.HTTPRequest(context.Background(), "https://test.com")
	if err != nil {
		t.Fatalf("unexpected error creating HTTP request: %s", err)
	}
	query := httpReq.URL.Query()
	if got := query.Get("type"); got != "user" {
		t.Errorf("expected type=user, got %q", got)
	}
	if got := query.Get("reportType"); got != "loggedtime" {
		t.Errorf("expected reportType=loggedtime, got %q", got)
	}
	if got := query.Get("groupBy"); got != "month" {
		t.Errorf("expected groupBy=month, got %q", got)
	}
}

func TestTimeReportTotalsResponseDecode(t *testing.T) {
	body := `{
		"loggedTime": 900, "billableTime": 600, "nonBillableTime": 300, "billedTime": 120, "estimatedTime": 0,
		"dates": [
			{"startDate": "2026-08-01", "endDate": "2026-08-02",
			 "loggedTime": 300, "billableTime": 300, "nonBillableTime": 0, "billedTime": 120, "estimatedTime": 0},
			{"startDate": "2026-08-03", "endDate": "2026-08-09",
			 "loggedTime": 0, "billableTime": 0, "nonBillableTime": 0, "billedTime": 0, "estimatedTime": 0},
			{"startDate": "2026-08-10", "endDate": "2026-08-16",
			 "loggedTime": 600, "billableTime": 300, "nonBillableTime": 300, "billedTime": 0, "estimatedTime": 0}
		]
	}`

	var response projects.TimeReportTotalsResponse
	err := response.HandleHTTPResponse(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if response.LoggedTime != 900 || response.BilledTime != 120 {
		t.Errorf("unexpected window totals: %+v", response.TimeReportColumns)
	}
	if response.RowsCount != nil {
		t.Errorf("expected rowsCount to be absent without a type, got %d", *response.RowsCount)
	}
	if len(response.Dates) != 3 {
		t.Fatalf("expected 3 periods, got %d", len(response.Dates))
	}
	first := response.Dates[0]
	if first.StartDate.String() != "2026-08-01" || first.EndDate.String() != "2026-08-02" {
		t.Errorf("unexpected first period window: %s to %s", first.StartDate, first.EndDate)
	}
	if first.LoggedTime != 300 || first.BilledTime != 120 {
		t.Errorf("unexpected first period totals: %+v", first.TimeReportColumns)
	}
	if response.Dates[1].LoggedTime != 0 {
		t.Errorf("expected the empty period to decode as zeros, got %+v", response.Dates[1])
	}
}

func TestTimeReportTotalsResponseDecode_forbidden(t *testing.T) {
	var response projects.TimeReportTotalsResponse
	err := response.HandleHTTPResponse(&http.Response{
		StatusCode: http.StatusForbidden,
		Body:       io.NopCloser(strings.NewReader(`{"errors":[{"detail":"forbidden action"}]}`)),
	})

	var httpErr *twapi.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected a twapi.HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", httpErr.StatusCode)
	}
}
