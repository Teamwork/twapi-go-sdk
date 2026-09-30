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
	_ twapi.HTTPRequester = (*WorkloadRequest)(nil)
	_ twapi.HTTPResponser = (*WorkloadResponse)(nil)
)

// Workload is a visual representation of how tasks are distributed across team
// members, helping you understand who is overloaded, who has capacity, and how
// work is balanced within a project or across multiple projects. It takes into
// account assigned tasks, due dates, estimated time, and working hours to give
// managers and teams a clear picture of availability and resource allocation.
// By providing this insight, workload makes it easier to plan effectively,
// prevent burnout, and ensure that deadlines are met without placing too much
// pressure on any single person.
//
// A task counts towards a user's day when the user is assigned to it directly
// (team, company and job-role assignments are not counted), it has an estimate
// and it has a due date, or its milestone does. Completed tasks are left out.
// Its estimate is divided between the user assignees and then spread evenly
// over the user's working days from start to due date, both included; a task
// with no start date puts it all on the due date, and one with a start date but
// no due date of its own or from its milestone is not counted at all. A custom split (see TaskCapacity) replaces the
// even spread for the user it belongs to. Parent tasks and their subtasks are
// counted independently.
//
// More information can be found at:
// https://support.teamwork.com/projects/workload/using-the-workload-planner
type Workload struct {
	// Users is a list of users in the workload response.
	Users []WorkloadUser `json:"users"`
}

// WorkloadUser represents a user in the workload response. It contains the
// user's ID and a map of dates with their corresponding workload information.
type WorkloadUser struct {
	// ID is the unique identifier for the user.
	ID int64 `json:"userId"`

	// Dates is a map of dates to their corresponding workload information for the
	// user.
	Dates map[twapi.Date]WorkloadUserDate `json:"dates"`
}

// WorkloadUserDate represents the workload information for a specific user on a
// specific date. Both figures are the day's planned load, not the time left.
type WorkloadUserDate struct {
	// Capacity is CapacityMinutes as a percentage of the user's working minutes
	// for the day, rounded to one decimal. Above 100 the user is over capacity.
	Capacity float64 `json:"capacity"`

	// CapacityMinutes is the planned load in minutes: the user's share of the
	// estimates of the tasks falling on the day, plus the day's unavailable
	// time.
	CapacityMinutes int64 `json:"capacityMinutes"`

	// UnavailableDay indicates whether the day's unavailable time covers all of
	// the user's working hours.
	UnavailableDay bool `json:"unavailableDay"`

	// IsHoliday indicates whether the day is a holiday for the user.
	IsHoliday bool `json:"isHoliday"`

	// Projects breaks the day down by project, naming the tasks that fall on
	// it. The day's unavailable time is included in every project's figures,
	// so they do not add up to the day's.
	Projects []WorkloadUserDateProject `json:"projects,omitempty"`
}

// WorkloadUserDateProject is one project's part of a user's day.
type WorkloadUserDateProject struct {
	// Capacity is CapacityMinutes as a percentage of the user's working minutes
	// for the day.
	Capacity float64 `json:"capacity"`

	// CapacityMinutes is the user's share of the day's estimates for this
	// project's tasks, plus the day's unavailable time.
	CapacityMinutes int64 `json:"capacityMinutes"`

	// Project is the project.
	Project twapi.Relationship `json:"project"`

	// Tasks are the project's tasks that fall on the day. Sideload
	// WorkloadGetRequestSideloadTasks to read their dates and estimates.
	Tasks []twapi.Relationship `json:"tasks"`
}

// WorkloadGetRequestSideload represents the related objects that can be
// included in the workload response to provide additional context.
type WorkloadGetRequestSideload string

// List of valid sideload options for the workload response.
const (
	WorkloadGetRequestSideloadUsers WorkloadGetRequestSideload = "users"

	// WorkloadGetRequestSideloadWorkingHours returns each user's working hours
	// under Included.WorkingHours.
	WorkloadGetRequestSideloadWorkingHours WorkloadGetRequestSideload = "users.workingHours"

	// WorkloadGetRequestSideloadWorkingHourEntries returns each user's working
	// hours together with their per-weekday entries.
	WorkloadGetRequestSideloadWorkingHourEntries WorkloadGetRequestSideload = "users.workingHours.workingHoursEntry"

	// WorkloadGetRequestSideloadTasks returns the tasks counted in the
	// response under Included.Tasks.
	WorkloadGetRequestSideloadTasks WorkloadGetRequestSideload = "tasks"

	// WorkloadGetRequestSideloadTaskCapacities returns the tasks and their
	// custom splits, under Included.Tasks and Included.TaskCapacities.
	WorkloadGetRequestSideloadTaskCapacities WorkloadGetRequestSideload = "tasks.taskCapacities"
)

// WorkloadRequestFilters contains the filters for loading the workload.
type WorkloadRequestFilters struct {
	// StartDate is the start date for the workload. This is a required field.
	// The boundary day is included.
	StartDate twapi.Date

	// EndDate is the end date for the workload. This is a required field. The
	// boundary day is included.
	EndDate twapi.Date

	// UserIDs is a list of user IDs to filter the workload by.
	UserIDs []int64

	// UserCompanyIDs is a list of users' client/company IDs to filter the
	// workload by.
	UserCompanyIDs []int64

	// UserTeamIDs is a list of users' team IDs to filter the workload by.
	UserTeamIDs []int64

	// UserJobRoleIDs is a list of users' job role IDs to filter the workload by.
	UserJobRoleIDs []int64

	// ProjectIDs is a list of project IDs to filter the workload by.
	ProjectIDs []int64

	// Include is a list of related objects to include in the response.
	Include []WorkloadGetRequestSideload

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of users to retrieve per page. Defaults to 50.
	PageSize int64
}

func (w WorkloadRequestFilters) apply(req *http.Request) {
	query := req.URL.Query()
	if !time.Time(w.StartDate).IsZero() {
		query.Set("startDate", w.StartDate.String())
	}
	if !time.Time(w.EndDate).IsZero() {
		query.Set("endDate", w.EndDate.String())
	}
	if len(w.UserIDs) > 0 {
		var ids []string
		for _, id := range w.UserIDs {
			ids = append(ids, strconv.FormatInt(id, 10))
		}
		query.Set("userIds", strings.Join(ids, ","))
	}
	if len(w.UserCompanyIDs) > 0 {
		var ids []string
		for _, id := range w.UserCompanyIDs {
			ids = append(ids, strconv.FormatInt(id, 10))
		}
		query.Set("companyIds", strings.Join(ids, ","))
	}
	if len(w.UserTeamIDs) > 0 {
		var ids []string
		for _, id := range w.UserTeamIDs {
			ids = append(ids, strconv.FormatInt(id, 10))
		}
		query.Set("teamIds", strings.Join(ids, ","))
	}
	if len(w.UserJobRoleIDs) > 0 {
		var ids []string
		for _, id := range w.UserJobRoleIDs {
			ids = append(ids, strconv.FormatInt(id, 10))
		}
		query.Set("jobRoleIds", strings.Join(ids, ","))
	}
	if len(w.ProjectIDs) > 0 {
		var ids []string
		for _, id := range w.ProjectIDs {
			ids = append(ids, strconv.FormatInt(id, 10))
		}
		query.Set("projectIds", strings.Join(ids, ","))
	}
	if w.Page > 0 {
		query.Set("page", strconv.FormatInt(w.Page, 10))
	}
	if w.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(w.PageSize, 10))
	}
	if len(w.Include) > 0 {
		for _, include := range w.Include {
			query.Add("include", string(include))
		}
	}

	// to reduce the size of the response, we omit empty date entries where the
	// user has no capacity and is not unavailable.
	query.Set("omitEmptyDateEntries", "true")

	req.URL.RawQuery = query.Encode()
}

// WorkloadRequest represents the request body for loading workload data.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/workload/get-projects-api-v3-workload-json
type WorkloadRequest struct {
	// Filters contains the filters for loading the workload.
	Filters WorkloadRequestFilters
}

// NewWorkloadRequest creates a new WorkloadRequest with the provided
// start and end dates. These dates are required to load a workload.
func NewWorkloadRequest(startDate, endDate twapi.Date) WorkloadRequest {
	return WorkloadRequest{
		Filters: WorkloadRequestFilters{
			StartDate: startDate,
			EndDate:   endDate,
		},
	}
}

// HTTPRequest creates an HTTP request for the WorkloadRequest.
func (w WorkloadRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/workload.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	w.Filters.apply(req)
	return req, nil
}

// WorkloadResponse contains all the information related to a workload.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/workload/get-projects-api-v3-workload-json
type WorkloadResponse struct {
	// Meta contains metadata about the response, including pagination details.
	Meta twapi.ListMeta `json:"meta"`

	// Workload contains the workload data.
	Workload Workload `json:"workload"`

	// Included contains related objects included in the response.
	Included struct {
		// Users is a map of user IDs to User objects.
		//
		// The key is the string representation of the user ID.
		Users map[string]User `json:"users,omitempty"`

		// WorkingHours is a map of working hour IDs to their corresponding
		// working hour information.
		//
		// The key is the string representation of the working hour ID.
		WorkingHours map[string]struct {
			// ID is the unique identifier for the working hours entry.
			ID int64 `json:"id"`

			// Object is a relationship object that links to the user associated
			// with these working hours.
			//
			// This field helps identify which user's working hours are being
			// represented.
			Object twapi.Relationship `json:"object"`

			// Entries is a list of relationships to the working hour entries
			// associated with these working hours, one per weekday.
			Entries []twapi.Relationship `json:"entries"`
		} `json:"workingHours,omitempty"`

		// WorkingHoursEntries is a map of working hour entry IDs to their
		// corresponding working hour entry information.
		//
		// The key is the string representation of the working hour entry ID.
		//
		// Each entry is one weekday of a user's working hours. The
		// "workingHour" field links back to the parent working hours object,
		// whose "object" names the user.
		WorkingHoursEntries map[string]struct {
			// ID is the unique identifier for the working hour entry.
			ID int64 `json:"id"`

			// WorkingHour is a relationship object that links back to the
			// parent working hours object.
			//
			// This field helps identify which working hours entry this
			// particular day's working hours belong to.
			WorkingHour twapi.Relationship `json:"workingHour"`

			// Weekday indicates the day of the week (e.g., "Monday", "Tuesday") for
			// which these working hours apply.
			Weekday string `json:"weekday"`

			// TaskHours is the number of hours the user works on this weekday,
			// which is the denominator of a day's Capacity. It is not the time
			// already planned.
			TaskHours float64 `json:"taskHours"`
		} `json:"workingHourEntries,omitempty"`

		// Tasks is a map of task IDs to the tasks counted in the response.
		//
		// The key is the string representation of the task ID.
		Tasks map[string]Task `json:"tasks,omitempty"`

		// TaskCapacities is a map of row IDs to the custom splits of the tasks
		// counted in the response. A task with no rows is on the even spread.
		//
		// The key is the string representation of the row ID.
		TaskCapacities map[string]TaskCapacity `json:"taskCapacities,omitempty"`
	} `json:"included"`
}

// HandleHTTPResponse handles the HTTP response for the WorkloadResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (w *WorkloadResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve workload")
	}

	if err := json.NewDecoder(resp.Body).Decode(w); err != nil {
		return fmt.Errorf("failed to decode retrieve workload response: %w", err)
	}
	return nil
}

// WorkloadGet retrieves a workload using the provided request and returns the
// response.
func WorkloadGet(
	ctx context.Context,
	engine *twapi.Engine,
	req WorkloadRequest,
) (*WorkloadResponse, error) {
	return twapi.Execute[WorkloadRequest, *WorkloadResponse](ctx, engine, req)
}
