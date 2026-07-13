package projects

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var (
	_ twapi.HTTPRequester = (*TimeReportListRequest)(nil)
	_ twapi.HTTPResponser = (*TimeReportListResponse)(nil)
)

// TimeReportType identifies the dimension a time report is grouped by. It maps
// to the `type` selector of the `/time/report/{type}` endpoint and determines
// which slice of the response is populated. Every request targets exactly one
// dimension.
type TimeReportType string

// List of possible time report dimensions.
const (
	TimeReportTypeUser     TimeReportType = "user"
	TimeReportTypeProject  TimeReportType = "project"
	TimeReportTypeCompany  TimeReportType = "company"
	TimeReportTypeTasklist TimeReportType = "tasklist"
	TimeReportTypeTask     TimeReportType = "task"
	TimeReportTypeTeam     TimeReportType = "team"
)

// TimeReportReportType selects the report variant, controlling which timelogs
// are aggregated into each row. It maps to the endpoint's `reportType` query
// parameter (distinct from TimeReportType, which selects the grouping
// dimension). When left empty the API defaults to TimeReportReportTypeTime.
type TimeReportReportType string

// List of possible time report variants.
const (
	// TimeReportReportTypeTime aggregates all tracked time.
	TimeReportReportTypeTime TimeReportReportType = "time"
	// TimeReportReportTypeLoggedTime aggregates logged time only.
	TimeReportReportTypeLoggedTime TimeReportReportType = "loggedtime"
	// TimeReportReportTypeUserLoggedTime aggregates logged time for the "logged
	// time per user" report. It is the variant the precanned user report uses.
	TimeReportReportTypeUserLoggedTime TimeReportReportType = "userloggedtime"
	// TimeReportReportTypeProjectLoggedTime aggregates logged time for the
	// "logged time per project" report. It is the variant the precanned project
	// report uses.
	TimeReportReportTypeProjectLoggedTime TimeReportReportType = "projecttime"
)

// TimeReportSideload identifies the related entities that can be requested
// alongside a time report via the API's include mechanism. Only the users and
// projects sideloads are decoded into typed maps on the response; requesting
// other sideloads is accepted by the API but not surfaced by this SDK.
type TimeReportSideload string

// List of valid time report sideloads.
const (
	TimeReportSideloadUsers     TimeReportSideload = "users"
	TimeReportSideloadProjects  TimeReportSideload = "projects"
	TimeReportSideloadCompanies TimeReportSideload = "companies"
	TimeReportSideloadTeams     TimeReportSideload = "teams"
	TimeReportSideloadTasks     TimeReportSideload = "tasks"
)

// TimeReportColumns contains the time totals shared by every time report row.
// All values are expressed in minutes.
type TimeReportColumns struct {
	// LoggedTime is the total time logged, in minutes.
	LoggedTime int64 `json:"loggedTime"`

	// BilledTime is the total time already invoiced, in minutes.
	BilledTime int64 `json:"billedTime"`

	// BillableTime is the total billable time, in minutes.
	BillableTime int64 `json:"billableTime"`

	// NonBillableTime is the total non-billable time, in minutes.
	NonBillableTime int64 `json:"nonBillableTime"`

	// EstimatedTime is the total estimated time, in minutes. It rides a
	// per-project permission gate and may be silently zeroed for callers without
	// the relevant permission.
	EstimatedTime int64 `json:"estimatedTime"`
}

// TimeReportCompany is a single row of a company-grouped time report.
type TimeReportCompany struct {
	TimeReportColumns

	// Company is the company the row aggregates time for.
	Company twapi.Relationship `json:"company"`
}

// TimeReportProject is a single row of a project-grouped time report.
type TimeReportProject struct {
	TimeReportColumns

	// Project is the project the row aggregates time for.
	Project twapi.Relationship `json:"project"`

	// TimeBudget is the project's time budget, when one is set.
	TimeBudget *twapi.Relationship `json:"timeBudget"`

	// FinancialBudget is the project's financial budget, when one is set.
	FinancialBudget *twapi.Relationship `json:"financialBudget"`
}

// TimeReportTasklist is a single row of a tasklist-grouped time report.
type TimeReportTasklist struct {
	TimeReportColumns

	// Tasklist is the tasklist the row aggregates time for.
	Tasklist twapi.Relationship `json:"tasklist"`

	// TimeBudget is the tasklist's time budget, when one is set.
	TimeBudget *twapi.Relationship `json:"timeBudget,omitempty"`

	// FinancialBudget is the tasklist's financial budget, when one is set.
	FinancialBudget *twapi.Relationship `json:"financialBudget,omitempty"`
}

// TimeReportTask is a single row of a task-grouped time report.
type TimeReportTask struct {
	TimeReportColumns

	// Task is the task the row aggregates time for.
	Task twapi.Relationship `json:"task"`

	// ParentTask is the task's parent, when it is a subtask.
	ParentTask *twapi.Relationship `json:"parentTask"`
}

// TimeReportTeam is a single row of a team-grouped time report.
type TimeReportTeam struct {
	TimeReportColumns

	// Team is the team the row aggregates time for.
	Team twapi.Relationship `json:"team"`
}

// TimeReportUser is a single row of a user-grouped time report.
type TimeReportUser struct {
	TimeReportColumns

	// User is the user the row aggregates time for. Time is attributed to whose
	// time it is, not who keyed the entry.
	User twapi.Relationship `json:"user"`

	// UtilizationTarget is the user's utilization target. It is plan-gated and
	// may be zero for installations without the feature.
	UtilizationTarget int64 `json:"utilizationTarget"`
}

// TimeReport contains the grouped rows of a time report. Exactly one slice is
// populated per request, matching the requested TimeReportType.
type TimeReport struct {
	// Companies holds the rows of a company-grouped report.
	Companies []TimeReportCompany `json:"companies,omitempty"`

	// Projects holds the rows of a project-grouped report.
	Projects []TimeReportProject `json:"projects,omitempty"`

	// Tasklists holds the rows of a tasklist-grouped report.
	Tasklists []TimeReportTasklist `json:"tasklists,omitempty"`

	// Tasks holds the rows of a task-grouped report.
	Tasks []TimeReportTask `json:"tasks,omitempty"`

	// Teams holds the rows of a team-grouped report.
	Teams []TimeReportTeam `json:"teams,omitempty"`

	// Users holds the rows of a user-grouped report.
	Users []TimeReportUser `json:"users,omitempty"`
}

// TimeReportListRequestPath contains the path parameters for loading a time
// report.
type TimeReportListRequestPath struct {
	// Type is the dimension the report is grouped by. It selects both the
	// endpoint path segment and the populated response slice. This is a required
	// field.
	Type TimeReportType
}

// TimeReportListRequestFilters contains the filters for loading a time report.
type TimeReportListRequestFilters struct {
	// StartDate is the inclusive start of the report window. This is a required
	// field.
	StartDate twapi.Date

	// EndDate is the inclusive end of the report window. This is a required
	// field.
	EndDate twapi.Date

	// ProjectIDs filters the report to the given projects.
	ProjectIDs []int64

	// UserIDs filters the report to the given users.
	UserIDs []int64

	// TaskIDs filters the report to the given tasks.
	TaskIDs []int64

	// TasklistIDs filters the report to the given tasklists.
	TasklistIDs []int64

	// TeamIDs filters the report to the given teams.
	TeamIDs []int64

	// CompanyIDs filters the report to the given companies.
	CompanyIDs []int64

	// TimelogTagIDs filters the report to timelogs carrying the given tags.
	TimelogTagIDs []int64

	// IncludeArchivedProjects includes time from archived projects when set to
	// true. When nil the API default (false) applies.
	IncludeArchivedProjects *bool

	// ReportType selects the report variant. When empty the API defaults to
	// TimeReportReportTypeTime.
	ReportType TimeReportReportType

	// Include lists the related entities to sideload. Only the users and projects
	// sideloads are decoded into the response's typed Included maps.
	Include []TimeReportSideload

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of rows to retrieve per page. Defaults to 50.
	PageSize int64

	// Fields restricts the attributes returned for each sideloaded entity. Each
	// slot of TimeReportListFields is a separate `fields[entity]=…` selection;
	// populated slots restrict the response, empty slots return the API default.
	// Use the generated UserField and ProjectField constants to ensure values
	// match real attributes.
	Fields TimeReportListFields
}

func (f TimeReportListRequestFilters) apply(req *http.Request) {
	query := req.URL.Query()
	if !f.StartDate.IsZero() {
		query.Set("startDate", f.StartDate.String())
	}
	if !f.EndDate.IsZero() {
		query.Set("endDate", f.EndDate.String())
	}
	if len(f.ProjectIDs) > 0 {
		ids := make([]string, len(f.ProjectIDs))
		for i, id := range f.ProjectIDs {
			ids[i] = strconv.FormatInt(id, 10)
		}
		query.Set("projectIds", strings.Join(ids, ","))
	}
	if len(f.UserIDs) > 0 {
		ids := make([]string, len(f.UserIDs))
		for i, id := range f.UserIDs {
			ids[i] = strconv.FormatInt(id, 10)
		}
		query.Set("userIds", strings.Join(ids, ","))
	}
	if len(f.TaskIDs) > 0 {
		ids := make([]string, len(f.TaskIDs))
		for i, id := range f.TaskIDs {
			ids[i] = strconv.FormatInt(id, 10)
		}
		query.Set("taskIds", strings.Join(ids, ","))
	}
	if len(f.TasklistIDs) > 0 {
		ids := make([]string, len(f.TasklistIDs))
		for i, id := range f.TasklistIDs {
			ids[i] = strconv.FormatInt(id, 10)
		}
		query.Set("tasklistIds", strings.Join(ids, ","))
	}
	if len(f.TeamIDs) > 0 {
		ids := make([]string, len(f.TeamIDs))
		for i, id := range f.TeamIDs {
			ids[i] = strconv.FormatInt(id, 10)
		}
		query.Set("teamIds", strings.Join(ids, ","))
	}
	if len(f.CompanyIDs) > 0 {
		ids := make([]string, len(f.CompanyIDs))
		for i, id := range f.CompanyIDs {
			ids[i] = strconv.FormatInt(id, 10)
		}
		query.Set("companyIds", strings.Join(ids, ","))
	}
	if len(f.TimelogTagIDs) > 0 {
		ids := make([]string, len(f.TimelogTagIDs))
		for i, id := range f.TimelogTagIDs {
			ids[i] = strconv.FormatInt(id, 10)
		}
		query.Set("timelogTagIds", strings.Join(ids, ","))
	}
	if f.IncludeArchivedProjects != nil {
		query.Set("includeArchivedProjects", strconv.FormatBool(*f.IncludeArchivedProjects))
	}
	if f.ReportType != "" {
		query.Set("reportType", string(f.ReportType))
	}
	if len(f.Include) > 0 {
		includes := make([]string, len(f.Include))
		for i, include := range f.Include {
			includes[i] = string(include)
		}
		query.Set("include", strings.Join(includes, ","))
	}
	if f.Page > 0 {
		query.Set("page", strconv.FormatInt(f.Page, 10))
	}
	if f.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(f.PageSize, 10))
	}

	// Total counts are derivable from the page window and hasMore, so the slower
	// counting path is always skipped.
	query.Set("skipCounts", "true")

	f.Fields.apply(query)
	req.URL.RawQuery = query.Encode()
}

// TimeReportListRequest represents the request for loading a grouped time
// report.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/get-projects-api-v3-time-report-type-json
type TimeReportListRequest struct {
	// Path contains the path parameters for the request.
	Path TimeReportListRequestPath

	// Filters contains the filters for loading the time report.
	Filters TimeReportListRequestFilters
}

// NewTimeReportListRequest creates a new TimeReportListRequest grouped by the
// given dimension and windowed by the given dates. The dimension and the date
// window are all required to load a time report.
func NewTimeReportListRequest(
	reportType TimeReportType,
	startDate twapi.Date,
	endDate twapi.Date,
) TimeReportListRequest {
	return TimeReportListRequest{
		Path: TimeReportListRequestPath{
			Type: reportType,
		},
		Filters: TimeReportListRequestFilters{
			StartDate: startDate,
			EndDate:   endDate,
			Page:      1,
			PageSize:  50,
		},
	}
}

// HTTPRequest creates an HTTP request for the TimeReportListRequest.
func (r TimeReportListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/time/report/%s.json", server, r.Path.Type)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	// The handler reads the grouping dimension from the `type` query parameter;
	// the path segment mirrors the public endpoint shape but is not read
	// server-side.
	query := req.URL.Query()
	query.Set("type", string(r.Path.Type))
	req.URL.RawQuery = query.Encode()

	r.Filters.apply(req)
	return req, nil
}

// TimeReportListResponse contains the grouped rows of a time report matching the
// request filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/get-projects-api-v3-time-report-type-json
//
// sparsefields:list
type TimeReportListResponse struct {
	request TimeReportListRequest

	// Meta contains metadata about the response, including pagination details.
	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`

	// TimeReport contains the grouped rows of the report. Exactly one slice is
	// populated, matching the requested dimension.
	TimeReport TimeReport `json:"time"`

	// Included contains the related entities sideloaded with the report.
	Included struct {
		// Users maps a user's string identifier to the sideloaded user.
		Users map[string]User `json:"users,omitempty"`

		// Projects maps a project's string identifier to the sideloaded project.
		Projects map[string]Project `json:"projects,omitempty"`
	} `json:"included"`
}

// HandleHTTPResponse handles the HTTP response for the TimeReportListResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (r *TimeReportListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve time report")
	}
	if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
		return fmt.Errorf("failed to decode retrieve time report response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (r *TimeReportListResponse) SetRequest(req TimeReportListRequest) {
	r.request = req
}

// Iterate returns the request set to the next page, if available. If there are
// no more pages, a nil request is returned.
func (r *TimeReportListResponse) Iterate() *TimeReportListRequest {
	if !r.Meta.Page.HasMore {
		return nil
	}
	req := r.request
	req.Filters.Page++
	return &req
}

// TimeReportList retrieves a grouped time report using the provided request and
// returns the response.
func TimeReportList(
	ctx context.Context,
	engine *twapi.Engine,
	req TimeReportListRequest,
) (*TimeReportListResponse, error) {
	return twapi.Execute[TimeReportListRequest, *TimeReportListResponse](ctx, engine, req)
}
