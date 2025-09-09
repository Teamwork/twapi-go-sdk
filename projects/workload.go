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
// specific date. It includes the user's capacity, capacity in minutes, and
// whether the user is unavailable on that date.
type WorkloadUserDate struct {
	// Capacity is the user's capacity percentage for the day.
	Capacity float64 `json:"capacity"`

	// CapacityMinutes is the user's capacity in minutes for the day.
	CapacityMinutes int64 `json:"capacityMinutes"`

	// UnavailableDay indicates whether the user is unavailable on that date.
	UnavailableDay bool `json:"unavailableDay"`
}

// WorkloadGetRequestSideload represents the related objects that can be
// included in the workload response to provide additional context.
type WorkloadGetRequestSideload string

// List of valid sideload options for the workload response.
const (
	WorkloadGetRequestSideloadUsers              WorkloadGetRequestSideload = "users"
	WorkloadGetRequestSideloadWorkingHours       WorkloadGetRequestSideload = "workingHours"
	WorkloadGetRequestSideloadWorkingHourEntries WorkloadGetRequestSideload = "workingHourEntries"
)

// WorkloadRequestFilters contains the filters for loading the workload.
type WorkloadRequestFilters struct {
	// StartDate is the start date for the workload. This is a required field.
	StartDate twapi.Date

	// EndDate is the end date for the workload. This is a required field.
	EndDate twapi.Date

	// UserIDs is a list of user IDs to filter the workload by.
	UserIDs []int64

	// UserCompanyIDs is a list of users' client/company IDs to filter the
	// workload by.
	UserCompanyIDs []int64

	// UserTeamIDs is a list of users' team IDs to filter the workload by.
	UserTeamIDs []int64

	// ProjectIDs is a list of project IDs to filter the workload by.
	ProjectIDs []int64

	// Include is a list of related objects to include in the response.
	Include []WorkloadGetRequestSideload

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of users to retrieve per page. Defaults to 50.
	PageSize int64
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

	query := req.URL.Query()
	if !time.Time(w.Filters.StartDate).IsZero() {
		query.Set("startDate", w.Filters.StartDate.String())
	}
	if !time.Time(w.Filters.EndDate).IsZero() {
		query.Set("endDate", w.Filters.EndDate.String())
	}
	if len(w.Filters.UserIDs) > 0 {
		var ids []string
		for _, id := range w.Filters.UserIDs {
			ids = append(ids, strconv.FormatInt(id, 10))
		}
		query.Set("userIds", strings.Join(ids, ","))
	}
	if len(w.Filters.UserCompanyIDs) > 0 {
		var ids []string
		for _, id := range w.Filters.UserCompanyIDs {
			ids = append(ids, strconv.FormatInt(id, 10))
		}
		query.Set("companyIds", strings.Join(ids, ","))
	}
	if len(w.Filters.UserTeamIDs) > 0 {
		var ids []string
		for _, id := range w.Filters.UserTeamIDs {
			ids = append(ids, strconv.FormatInt(id, 10))
		}
		query.Set("teamIds", strings.Join(ids, ","))
	}
	if len(w.Filters.ProjectIDs) > 0 {
		var ids []string
		for _, id := range w.Filters.ProjectIDs {
			ids = append(ids, strconv.FormatInt(id, 10))
		}
		query.Set("projectIds", strings.Join(ids, ","))
	}
	if w.Filters.Page > 0 {
		query.Set("page", strconv.FormatInt(w.Filters.Page, 10))
	}
	if w.Filters.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(w.Filters.PageSize, 10))
	}
	if len(w.Filters.Include) > 0 {
		for _, include := range w.Filters.Include {
			query.Add("include", string(include))
		}
	}

	// to reduce the size of the response, we omit empty date entries where the
	// user has no capacity and is not unavailable.
	query.Set("omitEmptyDateEntries", "true")

	req.URL.RawQuery = query.Encode()
	return req, nil
}

// WorkloadResponse contains all the information related to a workload.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/workload/get-projects-api-v3-workload-json
type WorkloadResponse struct {
	// Meta contains metadata about the response, including pagination details.
	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`

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
			// associated with these working hours.
			//
			// Each entry in this list represents a specific day's working hours
			// for the user, including the number of task hours assigned for
			// that day.
			Entries []twapi.Relationship `json:"entries"`
		} `json:"workingHours,omitempty"`

		// WorkingHoursEntries is a map of working hour entry IDs to their
		// corresponding working hour entry information.
		//
		// The key is the string representation of the working hour entry ID.
		//
		// Note: Each working hour entry represents a specific day's working hours
		// for a user, including the number of task hours assigned for that day.
		// The "workingHour" field links back to the parent working hours object.
		// The "weekday" field indicates the day of the week (e.g., "Monday",
		// "Tuesday") for which these working hours apply.
		//
		// This structure allows you to see not only the overall working hours
		// for a user but also how those hours are distributed across different
		// days of the week, along with the specific task hours assigned for each
		// day.
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

			// TaskHours represents the number of task hours assigned for this
			// particular day.
			TaskHours float64 `json:"taskHours"`
		} `json:"workingHourEntries,omitempty"`
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
