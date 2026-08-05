package projects_test

import (
	"context"
	"net/url"
	"strings"
	"testing"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

// sparseFieldsGetRequests lists every single-entity endpoint that exposes typed
// sparse fieldsets. The generated tests exercise the `<Resource>GetFields`
// containers in isolation; these cases cover the wiring through HTTPRequest,
// which the generator can only check statically.
var sparseFieldsGetRequests = []struct {
	name string

	// populated carries a Fields selection, zero is the same request without
	// one.
	populated twapi.HTTPRequester
	zero      twapi.HTTPRequester

	// want maps each expected fields[...] parameter to its expected value.
	want map[string]string
}{{
	name: "comment",
	populated: func() projects.CommentGetRequest {
		req := projects.NewCommentGetRequest(1)
		req.Fields.Comment = []projects.CommentField{projects.CommentFieldID}
		return req
	}(),
	zero: projects.NewCommentGetRequest(1),
	want: map[string]string{"fields[comments]": "id"},
}, {
	name: "company",
	populated: func() projects.CompanyGetRequest {
		req := projects.NewCompanyGetRequest(1)
		req.Fields.Company = []projects.CompanyField{projects.CompanyFieldID}
		req.Fields.CustomFields = []projects.CustomFieldField{projects.CustomFieldFieldID}
		req.Fields.CustomFieldValues = []projects.CustomFieldValueField{projects.CustomFieldValueFieldID}
		return req
	}(),
	zero: projects.NewCompanyGetRequest(1),
	want: map[string]string{
		"fields[companies]":            "id",
		"fields[customfields]":         "id",
		"fields[customfieldCompanies]": "id",
	},
}, {
	name: "jobrole",
	populated: func() projects.JobRoleGetRequest {
		req := projects.NewJobRoleGetRequest(1)
		req.Fields.JobRole = []projects.JobRoleField{projects.JobRoleFieldID}
		return req
	}(),
	zero: projects.NewJobRoleGetRequest(1),
	want: map[string]string{"fields[jobRoles]": "id"},
}, {
	name: "message",
	populated: func() projects.MessageGetRequest {
		req := projects.NewMessageGetRequest(1)
		req.Fields.Message = []projects.MessageField{projects.MessageFieldID}
		return req
	}(),
	zero: projects.NewMessageGetRequest(1),
	want: map[string]string{"fields[messages]": "id"},
}, {
	name: "message reply",
	populated: func() projects.MessageReplyGetRequest {
		req := projects.NewMessageReplyGetRequest(1)
		req.Fields.MessageReply = []projects.MessageReplyField{projects.MessageReplyFieldID}
		return req
	}(),
	zero: projects.NewMessageReplyGetRequest(1),
	want: map[string]string{"fields[messageReplies]": "id"},
}, {
	name: "milestone",
	populated: func() projects.MilestoneGetRequest {
		req := projects.NewMilestoneGetRequest(1)
		req.Fields.Milestone = []projects.MilestoneField{projects.MilestoneFieldID}
		return req
	}(),
	zero: projects.NewMilestoneGetRequest(1),
	want: map[string]string{"fields[milestones]": "id"},
}, {
	name: "notebook",
	populated: func() projects.NotebookGetRequest {
		req := projects.NewNotebookGetRequest(1)
		req.Fields.Notebook = []projects.NotebookField{projects.NotebookFieldID}
		return req
	}(),
	zero: projects.NewNotebookGetRequest(1),
	want: map[string]string{"fields[notebooks]": "id"},
}, {
	name: "project",
	populated: func() projects.ProjectGetRequest {
		req := projects.NewProjectGetRequest(1)
		req.Fields.Project = []projects.ProjectField{projects.ProjectFieldID}
		req.Fields.ProjectCategories = []projects.ProjectCategoryField{projects.ProjectCategoryFieldID}
		req.Fields.CustomFields = []projects.CustomFieldField{projects.CustomFieldFieldID}
		req.Fields.CustomFieldValues = []projects.CustomFieldValueField{projects.CustomFieldValueFieldID}
		return req
	}(),
	zero: projects.NewProjectGetRequest(1),
	want: map[string]string{
		"fields[projects]":            "id",
		"fields[projectCategories]":   "id",
		"fields[customfields]":        "id",
		"fields[customfieldProjects]": "id",
	},
}, {
	name: "project category",
	populated: func() projects.ProjectCategoryGetRequest {
		req := projects.NewProjectCategoryGetRequest(1)
		req.Fields.ProjectCategory = []projects.ProjectCategoryField{projects.ProjectCategoryFieldID}
		return req
	}(),
	zero: projects.NewProjectCategoryGetRequest(1),
	want: map[string]string{"fields[projectCategories]": "id"},
}, {
	name: "task",
	populated: func() projects.TaskGetRequest {
		req := projects.NewTaskGetRequest(1)
		req.Fields.Task = []projects.TaskField{projects.TaskFieldID}
		req.Fields.CustomFields = []projects.CustomFieldField{projects.CustomFieldFieldID}
		req.Fields.CustomFieldValues = []projects.CustomFieldValueField{projects.CustomFieldValueFieldID}
		return req
	}(),
	zero: projects.NewTaskGetRequest(1),
	want: map[string]string{
		"fields[tasks]":            "id",
		"fields[customfields]":     "id",
		"fields[customfieldTasks]": "id",
	},
}, {
	name: "tasklist",
	populated: func() projects.TasklistGetRequest {
		req := projects.NewTasklistGetRequest(1)
		req.Fields.Tasklist = []projects.TasklistField{projects.TasklistFieldID}
		return req
	}(),
	zero: projects.NewTasklistGetRequest(1),
	want: map[string]string{"fields[tasklists]": "id"},
}, {
	name: "timelog",
	populated: func() projects.TimelogGetRequest {
		req := projects.NewTimelogGetRequest(1)
		req.Fields.Timelog = []projects.TimelogField{projects.TimelogFieldID}
		return req
	}(),
	zero: projects.NewTimelogGetRequest(1),
	want: map[string]string{"fields[timelogs]": "id"},
}, {
	name: "timer",
	populated: func() projects.TimerGetRequest {
		req := projects.NewTimerGetRequest(1)
		req.Fields.Timer = []projects.TimerField{projects.TimerFieldID}
		return req
	}(),
	zero: projects.NewTimerGetRequest(1),
	want: map[string]string{"fields[timers]": "id"},
}, {
	// The plural key filters the list envelope only, so this endpoint expects
	// the singular fields[person].
	name: "user",
	populated: func() projects.UserGetRequest {
		req := projects.NewUserGetRequest(1)
		req.Fields.User = []projects.UserField{projects.UserFieldID}
		return req
	}(),
	zero: projects.NewUserGetRequest(1),
	want: map[string]string{"fields[person]": "id"},
}, {
	name: "workflow",
	populated: func() projects.WorkflowGetRequest {
		req := projects.NewWorkflowGetRequest(1)
		req.Fields.Workflow = []projects.WorkflowField{projects.WorkflowFieldID}
		return req
	}(),
	zero: projects.NewWorkflowGetRequest(1),
	want: map[string]string{"fields[workflows]": "id"},
}, {
	name: "workflow stage",
	populated: func() projects.WorkflowStageGetRequest {
		req := projects.NewWorkflowStageGetRequest(1, 2)
		req.Fields.Stage = []projects.WorkflowStageField{projects.WorkflowStageFieldID}
		return req
	}(),
	zero: projects.NewWorkflowStageGetRequest(1, 2),
	want: map[string]string{"fields[stages]": "id"},
}}

// TestGetRequestSparseFields verifies that a populated Fields slot reaches the
// generated HTTP request. A missing `req.URL.RawQuery = query.Encode()` would
// drop the selection silently, since req.URL.Query() hands back a copy.
func TestGetRequestSparseFields(t *testing.T) {
	for _, tt := range sparseFieldsGetRequests {
		t.Run(tt.name, func(t *testing.T) {
			httpReq, err := tt.populated.HTTPRequest(context.Background(), "https://test.com")
			if err != nil {
				t.Fatalf("unexpected error creating HTTP request: %s", err)
			}

			query, err := url.ParseQuery(httpReq.URL.RawQuery)
			if err != nil {
				t.Fatalf("failed to parse query string: %s", err)
			}

			for key, want := range tt.want {
				if got := query.Get(key); got != want {
					t.Errorf("expected %s=%q but got %q", key, want, got)
				}
			}
		})
	}
}

// TestGetRequestSparseFieldsUnset verifies that a request without a Fields
// selection sends no fields[...] parameter at all, so the API keeps returning
// its default attributes.
func TestGetRequestSparseFieldsUnset(t *testing.T) {
	for _, tt := range sparseFieldsGetRequests {
		t.Run(tt.name, func(t *testing.T) {
			httpReq, err := tt.zero.HTTPRequest(context.Background(), "https://test.com")
			if err != nil {
				t.Fatalf("unexpected error creating HTTP request: %s", err)
			}

			query, err := url.ParseQuery(httpReq.URL.RawQuery)
			if err != nil {
				t.Fatalf("failed to parse query string: %s", err)
			}

			for key := range query {
				if strings.HasPrefix(key, "fields[") {
					t.Errorf("unexpected sparse-fields parameter %q", key)
				}
			}
		})
	}
}
