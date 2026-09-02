package projects_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

// orderingQuery builds the HTTP request for a list request and returns its
// parsed query string. Every list request implements twapi.HTTPRequester, so the
// whole package's ordering wiring can be exercised through one table.
func orderingQuery(t *testing.T, req twapi.HTTPRequester) url.Values {
	t.Helper()

	httpReq, err := req.HTTPRequest(context.Background(), "https://test.com")
	if err != nil {
		t.Fatalf("unexpected error creating HTTP request: %s", err)
	}
	query, err := url.ParseQuery(httpReq.URL.RawQuery)
	if err != nil {
		t.Fatalf("failed to parse query string: %s", err)
	}
	return query
}

// TestListOrderingApplied checks that each list request that supports ordering
// writes the caller's selection to the parameters the endpoint reads. The
// expected keys are part of the table because the v1 team routes spell them
// sortBy/sortOrder instead of orderBy/orderMode.
func TestListOrderingApplied(t *testing.T) {
	tests := []struct {
		name string
		req  twapi.HTTPRequester
		want map[string]string
	}{{
		name: "activity",
		req: func() twapi.HTTPRequester {
			req := projects.NewActivityListRequest()
			req.Filters.OrderBy = projects.ActivityOrderByDate
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		}(),
		want: map[string]string{"orderBy": "date", "orderMode": "desc"},
	}, {
		name: "allocation",
		req: func() twapi.HTTPRequester {
			req := projects.NewAllocationListRequest()
			req.Filters.OrderBy = projects.AllocationOrderByStartDate
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		}(),
		want: map[string]string{"orderBy": "startdate", "orderMode": "desc"},
	}, {
		name: "calendar",
		req: func() twapi.HTTPRequester {
			req := projects.NewCalendarListRequest()
			req.Filters.OrderBy = projects.CalendarOrderByName
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		}(),
		want: map[string]string{"orderBy": "name", "orderMode": "desc"},
	}, {
		name: "calendar event",
		req: func() twapi.HTTPRequester {
			req := projects.NewCalendarEventListRequest(123)
			req.Filters.OrderBy = projects.CalendarEventOrderByStartTime
			req.Filters.OrderMode = twapi.OrderModeAscending
			return req
		}(),
		want: map[string]string{"orderBy": "startTime", "orderMode": "asc"},
	}, {
		name: "comment",
		req: func() twapi.HTTPRequester {
			req := projects.NewCommentListRequest()
			req.Filters.OrderBy = projects.CommentOrderByDate
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		}(),
		want: map[string]string{"orderBy": "date", "orderMode": "desc"},
	}, {
		name: "company",
		req: func() twapi.HTTPRequester {
			req := projects.NewCompanyListRequest()
			req.Filters.OrderBy = projects.CompanyOrderByCustomField
			req.Filters.OrderMode = twapi.OrderModeDescending
			req.Filters.OrderByCustomFieldID = 777
			return req
		}(),
		want: map[string]string{
			"orderBy":              "customfield",
			"orderMode":            "desc",
			"orderByCustomFieldId": "777",
		},
	}, {
		name: "custom field",
		req: func() twapi.HTTPRequester {
			req := projects.NewCustomFieldListRequest()
			req.Filters.OrderBy = projects.CustomFieldOrderByID
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		}(),
		want: map[string]string{"orderBy": "id", "orderMode": "desc"},
	}, {
		name: "custom item",
		req: func() twapi.HTTPRequester {
			req := projects.NewCustomItemListRequest(123)
			req.Filters.OrderBy = projects.CustomItemOrderByID
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		}(),
		want: map[string]string{"orderBy": "id", "orderMode": "desc"},
	}, {
		name: "custom item field",
		req: func() twapi.HTTPRequester {
			req := projects.NewCustomItemFieldListRequest(123)
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		}(),
		want: map[string]string{"orderMode": "desc"},
	}, {
		name: "custom item record",
		req: func() twapi.HTTPRequester {
			req := projects.NewCustomItemRecordListRequest(123)
			req.Filters.OrderBy = projects.CustomItemRecordOrderByCustomItemField
			req.Filters.OrderMode = twapi.OrderModeDescending
			req.Filters.OrderByFieldID = 777
			return req
		}(),
		want: map[string]string{
			"orderBy":        "customitemfield",
			"orderMode":      "desc",
			"orderByFieldId": "777",
		},
	}, {
		name: "job role",
		req: func() twapi.HTTPRequester {
			req := projects.NewJobRoleListRequest()
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		}(),
		want: map[string]string{"orderMode": "desc"},
	}, {
		name: "message",
		req: func() twapi.HTTPRequester {
			req := projects.NewMessageListRequest()
			req.Filters.OrderBy = projects.MessageOrderByCreatedAt
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		}(),
		want: map[string]string{"orderBy": "createdat", "orderMode": "desc"},
	}, {
		name: "message reply",
		req: func() twapi.HTTPRequester {
			req := projects.NewMessageReplyListRequest()
			req.Filters.OrderBy = projects.MessageReplyOrderByCreatedAt
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		}(),
		want: map[string]string{"orderBy": "createdat", "orderMode": "desc"},
	}, {
		name: "milestone",
		req: func() twapi.HTTPRequester {
			req := projects.NewMilestoneListRequest()
			req.Filters.OrderBy = projects.MilestoneOrderByDateCreated
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		}(),
		want: map[string]string{"orderBy": "dateCreated", "orderMode": "desc"},
	}, {
		name: "notebook",
		req: func() twapi.HTTPRequester {
			req := projects.NewNotebookListRequest()
			req.Filters.OrderBy = projects.NotebookOrderByDateUpdated
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		}(),
		want: map[string]string{"orderBy": "dateUpdated", "orderMode": "desc"},
	}, {
		name: "project",
		req: func() twapi.HTTPRequester {
			req := projects.NewProjectListRequest()
			req.Filters.OrderBy = projects.ProjectOrderByCustomField
			req.Filters.OrderMode = twapi.OrderModeDescending
			req.Filters.OrderByCustomFieldID = 777
			return req
		}(),
		want: map[string]string{
			"orderBy":              "customfield",
			"orderMode":            "desc",
			"orderByCustomFieldId": "777",
		},
	}, {
		name: "project template",
		req: func() twapi.HTTPRequester {
			req := projects.NewProjectTemplateListRequest()
			req.Filters.OrderBy = projects.ProjectOrderByName
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		}(),
		want: map[string]string{"orderBy": "name", "orderMode": "desc"},
	}, {
		name: "project status update",
		req: func() twapi.HTTPRequester {
			req := projects.NewProjectStatusUpdateListRequest()
			req.Filters.OrderBy = projects.ProjectStatusUpdateOrderByHealth
			req.Filters.OrderMode = twapi.OrderModeAscending
			return req
		}(),
		want: map[string]string{"orderBy": "health", "orderMode": "asc"},
	}, {
		name: "rate project user",
		req: func() twapi.HTTPRequester {
			req := projects.NewRateProjectUserListRequest(123)
			req.Filters.OrderBy = projects.RateProjectUserOrderByName
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		}(),
		want: map[string]string{"orderBy": "name", "orderMode": "desc"},
	}, {
		name: "rate project user history",
		req: func() twapi.HTTPRequester {
			req := projects.NewRateProjectUserHistoryGetRequest(123, 456)
			req.Filters.OrderBy = projects.RateProjectUserHistoryOrderByName
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		}(),
		want: map[string]string{"orderBy": "name", "orderMode": "desc"},
	}, {
		name: "skill",
		req: func() twapi.HTTPRequester {
			req := projects.NewSkillListRequest()
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		}(),
		want: map[string]string{"orderMode": "desc"},
	}, {
		name: "tag",
		req: func() twapi.HTTPRequester {
			req := projects.NewTagListRequest()
			req.Filters.OrderBy = projects.TagOrderByProjectDateLastUsed
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		}(),
		want: map[string]string{"orderBy": "projectdatelastused", "orderMode": "desc"},
	}, {
		name: "task",
		req: func() twapi.HTTPRequester {
			req := projects.NewTaskListRequest()
			req.Filters.OrderBy = projects.TaskOrderByCustomField
			req.Filters.OrderMode = twapi.OrderModeDescending
			req.Filters.OrderByCustomFieldID = 777
			return req
		}(),
		want: map[string]string{
			"orderBy":              "customfield",
			"orderMode":            "desc",
			"orderByCustomFieldId": "777",
		},
	}, {
		name: "tasklist",
		req: func() twapi.HTTPRequester {
			req := projects.NewTasklistListRequest()
			req.Filters.OrderBy = projects.TasklistOrderByDisplayOrder
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		}(),
		want: map[string]string{"orderBy": "displayorder", "orderMode": "desc"},
	}, {
		name: "tasklist budget",
		req: func() twapi.HTTPRequester {
			req := projects.NewTasklistBudgetListRequest(123)
			req.Filters.OrderBy = projects.TasklistBudgetOrderByDisplayOrder
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		}(),
		want: map[string]string{"orderBy": "displayOrder", "orderMode": "desc"},
	}, {
		name: "team maps ordering onto the v1 sort parameters",
		req: func() twapi.HTTPRequester {
			req := projects.NewTeamListRequest()
			req.Filters.OrderBy = projects.TeamOrderByDateAdded
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		}(),
		want: map[string]string{"sortBy": "dateAdded", "sortOrder": "desc"},
	}, {
		name: "time report",
		req: func() twapi.HTTPRequester {
			req := projects.NewTimeReportListRequest(projects.TimeReportTypeUser,
				twapi.Date(time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)),
				twapi.Date(time.Date(2026, time.August, 31, 0, 0, 0, 0, time.UTC)),
			)
			req.Filters.OrderBy = projects.TimeReportOrderByLoggedTime
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		}(),
		want: map[string]string{"orderBy": "loggedtime", "orderMode": "desc"},
	}, {
		name: "timelog",
		req: func() twapi.HTTPRequester {
			req := projects.NewTimelogListRequest()
			req.Filters.OrderBy = projects.TimelogOrderByTimeSpent
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		}(),
		want: map[string]string{"orderBy": "timespent", "orderMode": "desc"},
	}, {
		name: "user",
		req: func() twapi.HTTPRequester {
			req := projects.NewUserListRequest()
			req.Filters.OrderBy = projects.UserOrderByNameCaseInsensitive
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		}(),
		want: map[string]string{"orderBy": "namecaseinsensitive", "orderMode": "desc"},
	}, {
		name: "workflow stage",
		req: func() twapi.HTTPRequester {
			req := projects.NewWorkflowStageListRequest(123)
			req.Filters.OrderBy = projects.WorkflowStageOrderByDisplayOrder
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		}(),
		want: map[string]string{"orderBy": "displayorder", "orderMode": "desc"},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := orderingQuery(t, tt.req)
			for key, want := range tt.want {
				if got := query.Get(key); got != want {
					t.Errorf("%s = %q, want %q", key, got, want)
				}
			}
		})
	}
}

// TestListOrderingOmittedWhenUnset checks that an unset ordering filter sends
// nothing, leaving the endpoint's own default ordering in place. Requests whose
// constructors deliberately seed an ordering are built as zero values here.
func TestListOrderingOmittedWhenUnset(t *testing.T) {
	tests := []struct {
		name string
		req  twapi.HTTPRequester
		keys []string
	}{
		{name: "activity", req: projects.ActivityListRequest{}, keys: []string{"orderBy", "orderMode"}},
		{name: "allocation", req: projects.AllocationListRequest{}, keys: []string{"orderBy", "orderMode"}},
		{name: "calendar", req: projects.CalendarListRequest{}, keys: []string{"orderBy", "orderMode"}},
		{name: "calendar event", req: projects.CalendarEventListRequest{}, keys: []string{"orderBy", "orderMode"}},
		{name: "comment", req: projects.CommentListRequest{}, keys: []string{"orderBy", "orderMode"}},
		{name: "company", req: projects.CompanyListRequest{}, keys: []string{
			"orderBy", "orderMode", "orderByCustomFieldId",
		}},
		{name: "custom field", req: projects.CustomFieldListRequest{}, keys: []string{"orderBy", "orderMode"}},
		{
			// The custom item routes reject a missing path identifier before they
			// reach the query string, so these carry a path but no filters.
			name: "custom item",
			req:  projects.CustomItemListRequest{Path: projects.CustomItemListRequestPath{ProjectID: 123}},
			keys: []string{"orderBy", "orderMode"},
		},
		{
			name: "custom item field",
			req:  projects.CustomItemFieldListRequest{Path: projects.CustomItemFieldListRequestPath{CustomItemID: 123}},
			keys: []string{"orderMode"},
		},
		{
			name: "custom item record",
			req:  projects.CustomItemRecordListRequest{Path: projects.CustomItemRecordListRequestPath{CustomItemID: 123}},
			keys: []string{"orderBy", "orderMode", "orderByFieldId"},
		},
		{name: "job role", req: projects.JobRoleListRequest{}, keys: []string{"orderMode"}},
		{name: "message", req: projects.MessageListRequest{}, keys: []string{"orderBy", "orderMode"}},
		{name: "message reply", req: projects.MessageReplyListRequest{}, keys: []string{"orderBy", "orderMode"}},
		{name: "milestone", req: projects.MilestoneListRequest{}, keys: []string{"orderBy", "orderMode"}},
		{name: "notebook", req: projects.NotebookListRequest{}, keys: []string{"orderBy", "orderMode"}},
		{name: "project", req: projects.ProjectListRequest{}, keys: []string{
			"orderBy", "orderMode", "orderByCustomFieldId",
		}},
		{name: "project status update", req: projects.ProjectStatusUpdateListRequest{}, keys: []string{
			"orderBy", "orderMode",
		}},
		{name: "rate project user", req: projects.RateProjectUserListRequest{}, keys: []string{"orderBy", "orderMode"}},
		{
			name: "rate project user history",
			req:  projects.RateProjectUserHistoryGetRequest{},
			keys: []string{"orderBy", "orderMode"},
		},
		{name: "skill", req: projects.SkillListRequest{}, keys: []string{"orderMode"}},
		{name: "tag", req: projects.TagListRequest{}, keys: []string{"orderBy", "orderMode"}},
		{name: "task", req: projects.TaskListRequest{}, keys: []string{"orderBy", "orderMode", "orderByCustomFieldId"}},
		{name: "tasklist", req: projects.TasklistListRequest{}, keys: []string{"orderBy", "orderMode"}},
		{name: "tasklist budget", req: projects.TasklistBudgetListRequest{}, keys: []string{"orderBy", "orderMode"}},
		{name: "team", req: projects.TeamListRequest{}, keys: []string{"sortBy", "sortOrder"}},
		{name: "time report", req: projects.TimeReportListRequest{}, keys: []string{"orderBy", "orderMode"}},
		{name: "timelog", req: projects.TimelogListRequest{}, keys: []string{"orderBy", "orderMode"}},
		{name: "user", req: projects.UserListRequest{}, keys: []string{"orderBy", "orderMode"}},
		{name: "workflow stage", req: projects.WorkflowStageListRequest{}, keys: []string{"orderBy", "orderMode"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := orderingQuery(t, tt.req)
			for _, key := range tt.keys {
				if _, ok := query[key]; ok {
					t.Errorf("expected %s to be absent, got %q", key, query.Get(key))
				}
			}
		})
	}
}
