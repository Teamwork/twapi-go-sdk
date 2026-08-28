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
// alongside a time report via the API's include mechanism. Only the users,
// projects and tasks sideloads are decoded into typed maps on the response;
// requesting other sideloads is accepted by the API but not surfaced by this
// SDK.
type TimeReportSideload string

// List of valid time report sideloads.
const (
	TimeReportSideloadUsers     TimeReportSideload = "users"
	TimeReportSideloadProjects  TimeReportSideload = "projects"
	TimeReportSideloadCompanies TimeReportSideload = "companies"
	TimeReportSideloadTeams     TimeReportSideload = "teams"
	TimeReportSideloadTasks     TimeReportSideload = "tasks"

	// TimeReportSideloadTasksParentTasks adds the parents of the reported tasks
	// to the tasks sideload and populates TimeReportTask.ParentTask. It is only
	// meaningful on a task-grouped report.
	TimeReportSideloadTasksParentTasks TimeReportSideload = "tasks.parentTasks"
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

	// ParentTask is the task's parent, when it is a subtask. It is only
	// populated when the request includes TimeReportSideloadTasksParentTasks.
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

// TimeReportOrderBy identifies the attributes a time report can be ordered by.
type TimeReportOrderBy string

// Supported time report order-by values.
const (
	TimeReportOrderByName            TimeReportOrderBy = "name"
	TimeReportOrderByLoggedTime      TimeReportOrderBy = "loggedtime"
	TimeReportOrderByBillableTime    TimeReportOrderBy = "billabletime"
	TimeReportOrderByNonBillableTime TimeReportOrderBy = "nonbillabletime"
	TimeReportOrderByBilledTime      TimeReportOrderBy = "billedtime"
	TimeReportOrderByBudget          TimeReportOrderBy = "budget"
)

// TimeReportListRequestFilters contains the filters for loading a time report.
type TimeReportListRequestFilters struct {
	// StartDate is the inclusive start of the report window. This is a required
	// field.
	StartDate twapi.Date

	// EndDate is the inclusive end of the report window. This is a required
	// field.
	EndDate twapi.Date

	// OrderBy is the field to sort the results by. Use the TimeReportOrderBy
	// constants. The endpoint defaults to name.
	OrderBy TimeReportOrderBy

	// OrderMode is the direction to sort the results in. See twapi.OrderMode for
	// the supported values. The endpoint defaults to ascending.
	OrderMode twapi.OrderMode

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

	// Include lists the related entities to sideload. Only the users, projects
	// and tasks sideloads are decoded into the response's typed Included maps.
	Include []TimeReportSideload

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of rows to retrieve per page. Defaults to 50.
	PageSize int64

	// CountMode selects whether the API computes the exact number of time report
	// entries matching the filters, reported in Meta.Page.Count. Defaults to
	// twapi.ListCountModeSkip, matching what this SDK requests for this
	// endpoint.
	CountMode twapi.ListCountMode

	// Fields restricts the attributes returned for each sideloaded entity. Each
	// slot of TimeReportListFields is a separate `fields[entity]=…` selection;
	// populated slots restrict the response, empty slots return the API default.
	// Use the generated UserField, ProjectField and TaskField constants to
	// ensure values match real attributes.
	Fields TimeReportListFields
}

func (f TimeReportListRequestFilters) apply(req *http.Request) {
	query := req.URL.Query()
	if f.OrderBy != "" {
		query.Set("orderBy", string(f.OrderBy))
	}
	if f.OrderMode != "" {
		query.Set("orderMode", string(f.OrderMode))
	}
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

	f.Fields.apply(query)
	f.CountMode.Apply(query)
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
			CountMode: twapi.ListCountModeSkip,
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
	Meta twapi.ListMeta `json:"meta"`

	// TimeReport contains the grouped rows of the report. Exactly one slice is
	// populated, matching the requested dimension.
	TimeReport TimeReport `json:"time"`

	// Included contains the related entities sideloaded with the report.
	Included struct {
		// Users maps a user's string identifier to the sideloaded user.
		Users map[string]User `json:"users,omitempty"`

		// Projects maps a project's string identifier to the sideloaded project.
		Projects map[string]Project `json:"projects,omitempty"`

		// Tasks maps a task's string identifier to the sideloaded task. A
		// task-grouped report resolves its row names from here.
		Tasks map[string]Task `json:"tasks,omitempty"`
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
	r.Meta.ResolveCount(req.Filters.CountMode)
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

var (
	_ twapi.HTTPRequester = (*TimeReportTotalsRequest)(nil)
	_ twapi.HTTPResponser = (*TimeReportTotalsResponse)(nil)
)

// TimeReportGroupBy identifies the period a time report's totals are bucketed
// by. It maps to the `groupBy` parameter of `/time/report/totals.json`.
type TimeReportGroupBy string

// List of possible time report totals periods.
const (
	TimeReportGroupByDay   TimeReportGroupBy = "day"
	TimeReportGroupByWeek  TimeReportGroupBy = "week"
	TimeReportGroupByMonth TimeReportGroupBy = "month"
)

// TimeReportTotalsRequestFilters contains the filters for loading time report
// totals.
type TimeReportTotalsRequestFilters struct {
	// StartDate is the inclusive start of the report window. This is a required
	// field.
	StartDate twapi.Date

	// EndDate is the inclusive end of the report window. This is a required
	// field.
	EndDate twapi.Date

	// GroupBy is the period each entry covers. The endpoint defaults to
	// TimeReportGroupByDay.
	//
	// Buckets are keyed by day of year or month number, with no year, so a
	// multi-year window folds the same day or month of different years into one
	// entry: split it at 1 January, one request per year.
	//
	// Every period is returned in order, empty ones as zeros, with the first
	// and last clipped to the window. Weeks start on the calling user's
	// start-of-week day (Monday by default), so buckets differ per user, and a
	// weekend-only week carrying no time is omitted.
	GroupBy TimeReportGroupBy

	// Type is the dimension entries are counted by, and it turns on RowsCount
	// and the averages. It never splits the totals into rows, but it does
	// narrow them: task and tasklist drop time not logged on a task, and team
	// keeps only time logged by team members.
	Type TimeReportType

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

	// IncludeCompletedTasks includes time on completed tasks when set to true.
	// When nil the API default (true) applies.
	IncludeCompletedTasks *bool

	// ReportType does not change the totals — this endpoint reads it only when
	// authorising the request, where some variants satisfy a plan check that an
	// empty value does not. The numbers are identical either way, so set it
	// only if an otherwise valid request is answered 403.
	ReportType TimeReportReportType
}

func (f TimeReportTotalsRequestFilters) apply(req *http.Request) {
	query := req.URL.Query()
	querySetDate(query, "startDate", &f.StartDate)
	querySetDate(query, "endDate", &f.EndDate)
	querySetString(query, "groupBy", f.GroupBy)
	querySetString(query, "type", f.Type)
	querySetInt64s(query, "projectIds", f.ProjectIDs)
	querySetInt64s(query, "userIds", f.UserIDs)
	querySetInt64s(query, "taskIds", f.TaskIDs)
	querySetInt64s(query, "tasklistIds", f.TasklistIDs)
	querySetInt64s(query, "teamIds", f.TeamIDs)
	querySetInt64s(query, "companyIds", f.CompanyIDs)
	querySetInt64s(query, "timelogTagIds", f.TimelogTagIDs)
	querySetBool(query, "includeArchivedProjects", f.IncludeArchivedProjects)
	querySetBool(query, "includeCompletedTasks", f.IncludeCompletedTasks)
	querySetString(query, "reportType", f.ReportType)
	req.URL.RawQuery = query.Encode()
}

// TimeReportTotalsRequest represents the request for loading the totals of a
// time report bucketed by day, week or month.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/get-projects-api-v3-time-report-totals-json
type TimeReportTotalsRequest struct {
	// Filters contains the filters for loading the totals.
	Filters TimeReportTotalsRequestFilters
}

// NewTimeReportTotalsRequest creates a new TimeReportTotalsRequest bucketed by
// the given period and windowed by the given dates.
func NewTimeReportTotalsRequest(
	groupBy TimeReportGroupBy,
	startDate twapi.Date,
	endDate twapi.Date,
) TimeReportTotalsRequest {
	return TimeReportTotalsRequest{
		Filters: TimeReportTotalsRequestFilters{
			StartDate: startDate,
			EndDate:   endDate,
			GroupBy:   groupBy,
		},
	}
}

// HTTPRequest creates an HTTP request for the TimeReportTotalsRequest.
func (r TimeReportTotalsRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/time/report/totals.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	r.Filters.apply(req)
	return req, nil
}

// TimeReportTotalsPeriod is one period's total. Both dates are inclusive; a
// day has StartDate equal to EndDate.
type TimeReportTotalsPeriod struct {
	TimeReportColumns

	// StartDate is the first day of the period.
	StartDate twapi.Date `json:"startDate"`

	// EndDate is the last day of the period.
	EndDate twapi.Date `json:"endDate"`
}

// TimeReportTotalsResponse contains the totals of a time report over the
// requested window, bucketed by the requested period.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/get-projects-api-v3-time-report-totals-json
type TimeReportTotalsResponse struct {
	// TimeReportColumns holds the totals over the whole window.
	TimeReportColumns

	// Dates holds one entry per period in the window, in order.
	Dates []TimeReportTotalsPeriod `json:"dates"`

	// RowsCount is the number of entries of the requested Type in the window.
	// It and every average below are nil unless the request sets Type.
	RowsCount *int64 `json:"rowsCount,omitempty"`

	// LoggedTimeAverage is LoggedTime divided by RowsCount, in minutes.
	LoggedTimeAverage *int64 `json:"loggedTimeAverage,omitempty"`

	// BilledTimeAverage is BilledTime divided by RowsCount, in minutes.
	BilledTimeAverage *int64 `json:"billedTimeAverage,omitempty"`

	// BillableTimeAverage is BillableTime divided by RowsCount, in minutes.
	BillableTimeAverage *int64 `json:"billableTimeAverage,omitempty"`

	// NonBillableTimeAverage is NonBillableTime divided by RowsCount, in minutes.
	NonBillableTimeAverage *int64 `json:"nonBillableTimeAverage,omitempty"`

	// EstimatedTimeAverage is EstimatedTime divided by RowsCount, in minutes.
	EstimatedTimeAverage *int64 `json:"estimatedTimeAverage,omitempty"`
}

// HandleHTTPResponse handles the HTTP response for the
// TimeReportTotalsResponse. If some unexpected HTTP status code is returned by
// the API, a twapi.HTTPError is returned.
func (r *TimeReportTotalsResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve time report totals")
	}
	if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
		return fmt.Errorf("failed to decode retrieve time report totals response: %w", err)
	}
	return nil
}

// TimeReportTotals retrieves the totals of a time report bucketed by period.
func TimeReportTotals(
	ctx context.Context,
	engine *twapi.Engine,
	req TimeReportTotalsRequest,
) (*TimeReportTotalsResponse, error) {
	return twapi.Execute[TimeReportTotalsRequest, *TimeReportTotalsResponse](ctx, engine, req)
}
