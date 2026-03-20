package projects_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestTasklistBudgetListRequestGeneration(t *testing.T) {
	req := projects.NewTasklistBudgetListRequest(123)
	req.Filters.OrderMode = twapi.OrderModeDescending
	req.Filters.OrderBy = projects.TasklistBudgetListRequestOrderByDateCreated
	req.Filters.ProjectBudgetID = 123
	req.Filters.Page = 2
	req.Filters.PageSize = 25
	req.Filters.Include = []projects.TasklistBudgetListRequestSideload{
		projects.TasklistBudgetListRequestSideloadTasklists,
		projects.TasklistBudgetListRequestSideloadProjectBudgets,
	}
	req.Filters.Fields.Tasklists = []string{"id", "name"}
	req.Filters.Fields.TasklistBudgetNotifications = []string{"id", "projectId"}
	req.Filters.Fields.ProjectBudgets = []string{"id", "status", "dateCreated"}

	httpReq, err := req.HTTPRequest(context.Background(), "https://test.com")
	if err != nil {
		t.Fatalf("unexpected error creating HTTP request: %s", err)
	}

	if httpReq.URL.Path != "/projects/api/v3/projects/budgets/123/tasklists/budgets.json" {
		t.Fatalf("unexpected request path: %s", httpReq.URL.Path)
	}

	query, err := url.ParseQuery(httpReq.URL.RawQuery)
	if err != nil {
		t.Fatalf("failed to parse query string: %s", err)
	}

	if query.Get("orderMode") != "desc" {
		t.Errorf("expected orderMode=desc but got %q", query.Get("orderMode"))
	}
	if query.Get("orderBy") != "dateCreated" {
		t.Errorf("expected orderBy=dateCreated but got %q", query.Get("orderBy"))
	}
	if query.Get("projectBudgetId") != "123" {
		t.Errorf("expected projectBudgetId=123 but got %q", query.Get("projectBudgetId"))
	}
	if query.Get("page") != "2" {
		t.Errorf("expected page=2 but got %q", query.Get("page"))
	}
	if query.Get("pageSize") != "25" {
		t.Errorf("expected pageSize=25 but got %q", query.Get("pageSize"))
	}
	if query.Get("include") != "tasklists,projectBudgets" {
		t.Errorf("expected include=tasklists,projectBudgets but got %q", query.Get("include"))
	}
	if query.Get("fields[tasklists]") != "id,name" {
		t.Errorf("expected fields[tasklists]=id,name but got %q", query.Get("fields[tasklists]"))
	}
	if query.Get("fields[tasklistBudgetNotifications]") != "id,projectId" {
		t.Errorf(
			"expected fields[tasklistBudgetNotifications]=id,projectId but got %q",
			query.Get("fields[tasklistBudgetNotifications]"),
		)
	}
	if query.Get("fields[projectBudgets]") != "id,status,dateCreated" {
		t.Errorf("expected fields[projectBudgets]=id,status,dateCreated but got %q", query.Get("fields[projectBudgets]"))
	}
}

func TestTasklistBudgetListIterate(t *testing.T) {
	resp := &projects.TasklistBudgetListResponse{}
	req := projects.NewTasklistBudgetListRequest(987)
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
		t.Fatal("expected no next request when hasMore=false")
	}
}
