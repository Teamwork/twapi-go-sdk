package projects_test

import (
	"context"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestCalendarList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	req := projects.NewCalendarListRequest()
	resp, err := projects.CalendarList(ctx, engine, req)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if resp == nil {
		t.Error("expected a non-nil response")
	}

	// Note: We can't assert specific calendars exist since this depends on
	// the test environment setup, but we can verify the structure works
	t.Logf("Retrieved %d calendars", len(resp.Calendars))
}

func TestCalendarListPagination(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	// Request with small page size to test pagination
	req := projects.NewCalendarListRequest()
	req.Filters.PageSize = 1

	resp, err := projects.CalendarList(ctx, engine, req)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if resp == nil {
		t.Error("expected a non-nil response")
	}

	// If there are multiple calendars, test iteration
	if resp.Meta.Page.HasMore {
		nextReq := resp.Iterate()
		if nextReq == nil {
			t.Error("expected next request for pagination but got nil")
		} else if nextReq.Filters.Page != 2 {
			t.Errorf("expected page 2, got page %d", nextReq.Filters.Page)
		}
	}
}

func TestCalendarEventList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	t.Cleanup(cancel)

	// First, get the list of calendars to find a calendar ID
	calendarListReq := projects.NewCalendarListRequest()
	calendarListResp, err := projects.CalendarList(ctx, engine, calendarListReq)
	if err != nil {
		t.Fatalf("failed to get calendars: %s", err)
	}

	if len(calendarListResp.Calendars) == 0 {
		t.Skip("No calendars available in test environment")
	}

	// Use the first calendar for testing
	calendarID := calendarListResp.Calendars[0].ID

	// Set up date range for events (e.g., next 7 days)
	now := time.Now()
	startDate := now.Format("2006-01-02")
	endDate := now.Add(7 * 24 * time.Hour).Format("2006-01-02")

	req := projects.NewCalendarEventListRequest(calendarID)
	req.Filters.StartedAfterDate = startDate
	req.Filters.EndedBeforeDate = endDate

	resp, err := projects.CalendarEventList(ctx, engine, req)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if resp == nil {
		t.Error("expected a non-nil response")
	}

	if resp.STATUS != "OK" {
		t.Errorf("expected STATUS to be OK, got %s", resp.STATUS)
	}

	// Note: We can't assert specific events exist since this depends on
	// the test environment setup, but we can verify the structure works
	t.Logf("Retrieved %d calendar events from calendar %d", len(resp.Tasks), calendarID)
}

func TestCalendarEventListWithFilters(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	t.Cleanup(cancel)

	// First, get the list of calendars
	calendarListReq := projects.NewCalendarListRequest()
	calendarListResp, err := projects.CalendarList(ctx, engine, calendarListReq)
	if err != nil {
		t.Fatalf("failed to get calendars: %s", err)
	}

	if len(calendarListResp.Calendars) == 0 {
		t.Skip("No calendars available in test environment")
	}

	calendarID := calendarListResp.Calendars[0].ID

	// Test with various filter combinations
	tests := []struct {
		name    string
		filters func(*projects.CalendarEventListRequestFilters)
	}{
		{
			name: "with includes",
			filters: func(f *projects.CalendarEventListRequestFilters) {
				f.Include = "users,masterInstances"
			},
		},
		{
			name: "skip counts",
			filters: func(f *projects.CalendarEventListRequestFilters) {
				skipCounts := true
				f.SkipCounts = &skipCounts
			},
		},
		{
			name: "include timelogs",
			filters: func(f *projects.CalendarEventListRequestFilters) {
				includeTimelogs := true
				f.IncludeTimelogs = &includeTimelogs
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			now := time.Now()
			req := projects.NewCalendarEventListRequest(calendarID)
			req.Filters.StartedAfterDate = now.Format("2006-01-02")
			req.Filters.EndedBeforeDate = now.Add(7 * 24 * time.Hour).Format("2006-01-02")

			tt.filters(&req.Filters)

			resp, err := projects.CalendarEventList(ctx, engine, req)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			if resp == nil {
				t.Error("expected a non-nil response")
			}
		})
	}
}

func TestCalendarEventListRequestHTTPRequest(t *testing.T) {
	req := projects.NewCalendarEventListRequest(123)
	req.Filters.StartedAfterDate = "2026-01-01"
	req.Filters.EndedBeforeDate = "2026-01-31"
	req.Filters.Include = "users,timelogs"
	skipCounts := true
	req.Filters.SkipCounts = &skipCounts
	includeTimelogs := true
	req.Filters.IncludeTimelogs = &includeTimelogs

	httpReq, err := req.HTTPRequest(context.Background(), "https://test.teamwork.com")
	if err != nil {
		t.Fatalf("unexpected error creating HTTP request: %s", err)
	}

	if httpReq.Method != "GET" {
		t.Errorf("expected GET method, got %s", httpReq.Method)
	}

	expectedPath := "/projects/api/v3/calendars/123/events.json"
	if httpReq.URL.Path != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, httpReq.URL.Path)
	}

	query := httpReq.URL.Query()
	if query.Get("startedAfterDate") != "2026-01-01" {
		t.Errorf("expected startedAfterDate=2026-01-01, got %s", query.Get("startedAfterDate"))
	}
	if query.Get("endedBeforeDate") != "2026-01-31" {
		t.Errorf("expected endedBeforeDate=2026-01-31, got %s", query.Get("endedBeforeDate"))
	}
	if query.Get("include") != "users,timelogs" {
		t.Errorf("expected include=users,timelogs, got %s", query.Get("include"))
	}
	if query.Get("skipCounts") != "true" {
		t.Errorf("expected skipCounts=true, got %s", query.Get("skipCounts"))
	}
	if query.Get("includeTimelogs") != "true" {
		t.Errorf("expected includeTimelogs=true, got %s", query.Get("includeTimelogs"))
	}
}

func TestCalendarListRequestHTTPRequest(t *testing.T) {
	req := projects.NewCalendarListRequest()
	req.Filters.Page = 2
	req.Filters.PageSize = 25

	httpReq, err := req.HTTPRequest(context.Background(), "https://test.teamwork.com")
	if err != nil {
		t.Fatalf("unexpected error creating HTTP request: %s", err)
	}

	if httpReq.Method != "GET" {
		t.Errorf("expected GET method, got %s", httpReq.Method)
	}

	expectedPath := "/projects/api/v3/calendars.json"
	if httpReq.URL.Path != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, httpReq.URL.Path)
	}

	query := httpReq.URL.Query()
	if query.Get("page") != "2" {
		t.Errorf("expected page=2, got %s", query.Get("page"))
	}
	if query.Get("pageSize") != "25" {
		t.Errorf("expected pageSize=25, got %s", query.Get("pageSize"))
	}
}
