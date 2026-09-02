package projects_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestProjectStatusUpdateListRequestGeneration(t *testing.T) {
	createdAfter := time.Date(2026, 8, 3, 14, 30, 0, 0, time.UTC)
	updatedAfter := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)
	yes, no := true, false

	req := projects.NewProjectStatusUpdateListRequest()
	req.Filters.ProjectIDs = []int64{777, 12345}
	req.Filters.ProjectOwnerIDs = []int64{11}
	req.Filters.ProjectCategoryIDs = []int64{22}
	req.Filters.ProjectTagIDs = []int64{33}
	req.Filters.MatchAllProjectTags = &yes
	req.Filters.ProjectStatuses = []projects.ProjectListStatus{projects.ProjectListStatusLate}
	req.Filters.ProjectHealths = []projects.ProjectHealth{
		projects.ProjectHealthBad,
		projects.ProjectHealthNotSet,
	}
	req.Filters.OnlyStarredProjects = &yes
	req.Filters.ProjectCompanyIDs = []int64{44}
	req.Filters.IncludeArchivedProjects = &yes
	req.Filters.ShowDeleted = &yes
	req.Filters.ActiveOnly = &no
	req.Filters.CreatedAfter = &createdAfter
	req.Filters.UpdatedAfter = &updatedAfter
	req.Filters.Emoji = &no
	req.Filters.Reactions = &yes
	req.Filters.Page = 2
	req.Filters.PageSize = 25
	req.Filters.Include = []projects.ProjectStatusUpdateListRequestSideload{
		projects.ProjectStatusUpdateListRequestSideloadProjects,
		projects.ProjectStatusUpdateListRequestSideloadCreatedBy,
	}
	req.Filters.Fields.ProjectUpdates = []projects.ProjectStatusUpdateField{"id", "text", "health"}
	req.Filters.Fields.Users = []projects.UserField{"id", "firstName"}
	req.Filters.Fields.Projects = []projects.ProjectField{"id", "name"}

	httpReq, err := req.HTTPRequest(context.Background(), "https://test.com")
	if err != nil {
		t.Fatalf("unexpected error creating HTTP request: %s", err)
	}

	if httpReq.URL.Path != "/projects/api/v3/projects/updates.json" {
		t.Fatalf("unexpected request path: %s", httpReq.URL.Path)
	}

	query, err := url.ParseQuery(httpReq.URL.RawQuery)
	if err != nil {
		t.Fatalf("failed to parse query string: %s", err)
	}

	for key, want := range map[string]string{
		"projectIds":              "777,12345",
		"projectOwnerIds":         "11",
		"projectCategoryIds":      "22",
		"projectTagIds":           "33",
		"matchAllProjectTags":     "true",
		"projectStatuses":         "late",
		"projectHealths":          "1,0",
		"onlyStarredProjects":     "true",
		"projectCompanyIds":       "44",
		"includeArchivedProjects": "true",
		"showDeleted":             "true",
		"activeOnly":              "false",
		"createdAfter":            "2026-08-03T14:30:00Z",
		"updatedAfter":            "2026-08-10T00:00:00Z",
		"emoji":                   "false",
		"reactions":               "true",
		"page":                    "2",
		"pageSize":                "25",
		"include":                 "projects,createdBy",
		"fields[projectUpdates]":  "id,text,health",
		"fields[users]":           "id,firstName",
		"fields[projects]":        "id,name",
	} {
		if got := query.Get(key); got != want {
			t.Errorf("expected %s=%s but got %q", key, want, got)
		}
	}
}

// TestProjectStatusUpdateListRequestProjectPath checks the project-scoped route,
// which differs from the site-wide one only by the path segment.
func TestProjectStatusUpdateListRequestProjectPath(t *testing.T) {
	req := projects.NewProjectStatusUpdateListRequest()
	req.Path.ProjectID = 777

	httpReq, err := req.HTTPRequest(context.Background(), "https://test.com")
	if err != nil {
		t.Fatalf("unexpected error creating HTTP request: %s", err)
	}
	if httpReq.URL.Path != "/projects/api/v3/projects/777/updates.json" {
		t.Fatalf("unexpected request path: %s", httpReq.URL.Path)
	}
}

// TestProjectStatusUpdateListRequestPinsNewestFirst covers the one list request
// that does not leave the ordering to the endpoint. The reference documents the
// default orderMode as ascending while the endpoint answers newest-first, so the
// constructor pins the order rather than inheriting either.
func TestProjectStatusUpdateListRequestPinsNewestFirst(t *testing.T) {
	query := orderingQuery(t, projects.NewProjectStatusUpdateListRequest())
	if got := query.Get("orderBy"); got != "date" {
		t.Errorf("expected orderBy=date but got %q", got)
	}
	if got := query.Get("orderMode"); got != "desc" {
		t.Errorf("expected orderMode=desc but got %q", got)
	}

	req := projects.NewProjectStatusUpdateListRequest()
	req.Filters.OrderBy = projects.ProjectStatusUpdateOrderByID
	req.Filters.OrderMode = twapi.OrderModeAscending

	query = orderingQuery(t, req)
	if got := query.Get("orderBy"); got != "id" {
		t.Errorf("expected the caller's orderBy=id but got %q", got)
	}
	if got := query.Get("orderMode"); got != "asc" {
		t.Errorf("expected the caller's orderMode=asc but got %q", got)
	}
}

func TestProjectStatusUpdateListIterate(t *testing.T) {
	resp := &projects.ProjectStatusUpdateListResponse{}
	req := projects.NewProjectStatusUpdateListRequest()
	req.Filters.Page = 3

	resp.SetRequest(req)
	resp.Meta.Page.HasMore = true

	nextReq := resp.Iterate()
	if nextReq == nil {
		t.Fatal("expected next request but got nil")
	}
	if nextReq.Filters.Page != 4 {
		t.Fatalf("expected next page to be 4 but got %d", nextReq.Filters.Page)
	}

	resp.Meta.Page.HasMore = false
	if resp.Iterate() != nil {
		t.Error("expected no next request when there are no more pages")
	}
}
