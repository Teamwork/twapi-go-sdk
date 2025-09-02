package projects

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var (
	_ twapi.HTTPRequester = (*ActivityListRequest)(nil)
	_ twapi.HTTPResponser = (*ActivityListResponse)(nil)
)

// Activity is a record of actions and updates that occur across your projects,
// tasks, and communications, giving you a clear view of what’s happening within
// your workspace. Activities capture changes such as task completions, activities
// added, files uploaded, or milestones updated, and present them in a
// chronological feed so teams can stay aligned without needing to check each
// individual project or task. This stream of information helps improve
// transparency, ensures accountability, and keeps everyone aware of progress
// and decisions as they happen.
//
// More information can be found at:
// https://support.teamwork.com/projects/using-teamwork/activity
type Activity struct {
	// ID is the unique identifier of the activity.
	ID int64 `json:"id"`

	// Action is the type of activity that occurred.
	Action Action `json:"activityType"`

	// LatestAction is the most recent activity that occurred.
	LatestAction Action `json:"latestActivityType"`

	// At is the timestamp when the activity occurred.
	At time.Time `json:"dateTime"`

	// Description is a brief summary of the activity.
	Description *string `json:"description"`

	// ExtraDescription provides additional context about the activity.
	ExtraDescription *string `json:"extraDescription"`

	// PublicInfo provides information about the activity that is visible to all
	// users.
	PublicInfo *string `json:"publicInfo"`

	// DueAt is the deadline for the activity.
	DueAt *time.Time `json:"dueDate"`

	// ForUserName is the name of the user for whom the activity is intended.
	ForUserName *string `json:"forUserName"`

	// ItemLink provides a link to the item associated with the activity.
	ItemLink *string `json:"itemLink"`

	// Link provides a link to the activity itself.
	Link *string `json:"link"`

	// User is the relationship to the user that this activity is associated with.
	User twapi.Relationship `json:"user"`

	// ForUser is the relationship to the user for whom the activity is intended.
	ForUser *twapi.Relationship `json:"forUser"`

	// Project is the relationship to the project that this activity belongs to.
	Project twapi.Relationship `json:"project"`

	// Company is the relationship to the client/company that this activity is
	// associated with.
	Company twapi.Relationship `json:"company"`

	// Item is the relationship to the item that this activity is associated with.
	Item twapi.Relationship `json:"item"`
}

// Action contains all possible activity types.
type Action string

// List of activity types.
const (
	LogTypeNew       Action = "new"
	LogTypeEdited    Action = "edited"
	LogTypeCompleted Action = "completed"
	LogTypeReopened  Action = "reopened"
	LogTypeDeleted   Action = "deleted"
	LogTypeUndeleted Action = "undeleted"
	LogTypeLiked     Action = "liked"
	LogTypeReacted   Action = "reacted"
	LogTypeViewed    Action = "viewed"
)

// LogItemType contains all possible activity item types.
type LogItemType string

// List of activity item types.
const (
	LogItemTypeMessage          LogItemType = "message"
	LogItemTypeComment          LogItemType = "comment"
	LogItemTypeTask             LogItemType = "task"
	LogItemTypeTasklist         LogItemType = "tasklist"
	LogItemTypeTaskgroup        LogItemType = "taskgroup"
	LogItemTypeMilestone        LogItemType = "milestone"
	LogItemTypeFile             LogItemType = "file"
	LogItemTypeForm             LogItemType = "form"
	LogItemTypeNotebook         LogItemType = "notebook"
	LogItemTypeTimelog          LogItemType = "timelog"
	LogItemTypeTaskComment      LogItemType = "task_comment"
	LogItemTypeNotebookComment  LogItemType = "notebook_comment"
	LogItemTypeFileComment      LogItemType = "file_comment"
	LogItemTypeLinkComment      LogItemType = "link_comment"
	LogItemTypeMilestoneComment LogItemType = "milestone_comment"
	LogItemTypeProject          LogItemType = "project"
	LogItemTypeLink             LogItemType = "link"
	LogItemTypeBillingInvoice   LogItemType = "billingInvoice"
	LogItemTypeRisk             LogItemType = "risk"
	LogItemTypeProjectUpdate    LogItemType = "projectUpdate"
	LogItemTypeReacted          LogItemType = "reacted"
	LogItemTypeBudget           LogItemType = "budget"
)

// UnmarshalText decodes the text into a LogItemType.
func (l *LogItemType) UnmarshalText(text []byte) error {
	if l == nil {
		panic("unmarshal LogItemType: nil pointer")
	}
	logItemType := LogItemType(strings.ToLower(string(text)))
	switch logItemType {
	case LogItemTypeMessage,
		LogItemTypeComment,
		LogItemTypeTask,
		LogItemTypeTasklist,
		LogItemTypeTaskgroup,
		LogItemTypeMilestone,
		LogItemTypeFile,
		LogItemTypeForm,
		LogItemTypeNotebook,
		LogItemTypeTimelog,
		LogItemTypeTaskComment,
		LogItemTypeNotebookComment,
		LogItemTypeFileComment,
		LogItemTypeLinkComment,
		LogItemTypeMilestoneComment,
		LogItemTypeProject,
		LogItemTypeLink,
		LogItemTypeBillingInvoice,
		LogItemTypeRisk,
		LogItemTypeProjectUpdate,
		LogItemTypeReacted,
		LogItemTypeBudget:
		*l = logItemType
	default:
		return fmt.Errorf("invalid log item type: %q", text)
	}
	return nil
}

// ActivityListRequestPath contains the path parameters for loading multiple
// activities.
type ActivityListRequestPath struct {
	// ProjectID is the unique identifier of the project containing the
	// activities.
	ProjectID int64
}

// ActivityListRequestFilters contains the filters for loading multiple
// activities.
type ActivityListRequestFilters struct {
	// StartDate is the start date for filtering activities.
	StartDate time.Time

	// EndDate is the end date for filtering activities.
	EndDate time.Time

	// LogItemTypes is the list of log item types to filter activities.
	LogItemTypes []LogItemType

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of activities to retrieve per page. Defaults to 50.
	PageSize int64
}

// ActivityListRequest represents the request body for loading multiple activities.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/activity/get-projects-api-v3-latestactivity-json
// https://apidocs.teamwork.com/docs/teamwork/v3/activity/get-projects-api-v3-projects-project-id-latestactivity
type ActivityListRequest struct {
	// Path contains the path parameters for the request.
	Path ActivityListRequestPath

	// Filters contains the filters for loading multiple activities.
	Filters ActivityListRequestFilters
}

// NewActivityListRequest creates a new ActivityListRequest with default values.
func NewActivityListRequest() ActivityListRequest {
	return ActivityListRequest{
		Filters: ActivityListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the ActivityListRequest.
func (a ActivityListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	var uri string
	switch {
	case a.Path.ProjectID > 0:
		uri = fmt.Sprintf("%s/projects/api/v3/projects/%d/latestactivity.json", server, a.Path.ProjectID)
	default:
		uri = server + "/projects/api/v3/latestactivity.json"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	if !a.Filters.StartDate.IsZero() {
		query.Set("startDate", a.Filters.StartDate.Format(time.RFC3339))
	}
	if !a.Filters.EndDate.IsZero() {
		query.Set("endDate", a.Filters.EndDate.Format(time.RFC3339))
	}
	if len(a.Filters.LogItemTypes) > 0 {
		logItemTypes := make([]string, len(a.Filters.LogItemTypes))
		for i, logType := range a.Filters.LogItemTypes {
			logItemTypes[i] = string(logType)
		}
		query.Set("activityTypes", strings.Join(logItemTypes, ","))
	}
	if a.Filters.Page > 0 {
		query.Set("page", strconv.FormatInt(a.Filters.Page, 10))
	}
	if a.Filters.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(a.Filters.PageSize, 10))
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// ActivityListResponse contains information by multiple activities matching the
// request filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/activity/get-projects-api-v3-latestactivity-json
// https://apidocs.teamwork.com/docs/teamwork/v3/activity/get-projects-api-v3-projects-project-id-latestactivity
type ActivityListResponse struct {
	request ActivityListRequest

	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`
	Activities []Activity `json:"activities"`
}

// HandleHTTPResponse handles the HTTP response for the ActivityListResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (a *ActivityListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list activities")
	}

	if err := json.NewDecoder(resp.Body).Decode(a); err != nil {
		return fmt.Errorf("failed to decode list activities response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (a *ActivityListResponse) SetRequest(req ActivityListRequest) {
	a.request = req
}

// Iterate returns the request set to the next page, if available. If there
// are no more pages, a nil request is returned.
func (a *ActivityListResponse) Iterate() *ActivityListRequest {
	if !a.Meta.Page.HasMore {
		return nil
	}
	req := a.request
	req.Filters.Page++
	return &req
}

// ActivityList retrieves multiple activities using the provided request and
// returns the response.
func ActivityList(
	ctx context.Context,
	engine *twapi.Engine,
	req ActivityListRequest,
) (*ActivityListResponse, error) {
	return twapi.Execute[ActivityListRequest, *ActivityListResponse](ctx, engine, req)
}
