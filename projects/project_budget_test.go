package projects_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestProjectBudgetListRequestGeneration(t *testing.T) {
	req := projects.NewProjectBudgetListRequest()
	req.Filters.ProjectIDs = []int64{1215814}
	req.Filters.Status = projects.ProjectBudgetStatusUpcoming
	req.Filters.Limit = 1
	req.Filters.PageSize = 1

	httpReq, err := req.HTTPRequest(context.Background(), "https://test.com")
	if err != nil {
		t.Fatalf("unexpected error creating HTTP request: %s", err)
	}

	if httpReq.URL.Path != "/projects/api/v3/projects/budgets.json" {
		t.Fatalf("unexpected request path: %s", httpReq.URL.Path)
	}

	query, err := url.ParseQuery(httpReq.URL.RawQuery)
	if err != nil {
		t.Fatalf("failed to parse query string: %s", err)
	}

	if query.Get("projectIds") != "1215814" {
		t.Errorf("expected projectIds=1215814 but got %q", query.Get("projectIds"))
	}
	if query.Get("status") != "upcoming" {
		t.Errorf("expected status=upcoming but got %q", query.Get("status"))
	}
	if query.Get("limit") != "1" {
		t.Errorf("expected limit=1 but got %q", query.Get("limit"))
	}
	if query.Get("pageSize") != "1" {
		t.Errorf("expected pageSize=1 but got %q", query.Get("pageSize"))
	}
	if query.Get("cursor") != "" {
		t.Errorf("expected empty cursor when not set but got %q", query.Get("cursor"))
	}
}

func TestProjectBudgetListRequestGeneration_AllOptional(t *testing.T) {
	req := projects.NewProjectBudgetListRequest()

	httpReq, err := req.HTTPRequest(context.Background(), "https://test.com")
	if err != nil {
		t.Fatalf("unexpected error creating HTTP request: %s", err)
	}

	if httpReq.URL.RawQuery != "" {
		t.Fatalf("expected empty query string but got %q", httpReq.URL.RawQuery)
	}
}
