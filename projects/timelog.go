package projects

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var (
	_ twapi.HTTPRequester = (*TimelogCreateRequest)(nil)
	_ twapi.HTTPResponser = (*TimelogCreateResponse)(nil)
	_ twapi.HTTPRequester = (*TimelogUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*TimelogUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*TimelogBulkMoveRequest)(nil)
	_ twapi.HTTPResponser = (*TimelogBulkMoveResponse)(nil)
	_ twapi.HTTPRequester = (*TimelogBulkBillableUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*TimelogBulkBillableUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*TimelogDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*TimelogDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*TimelogGetRequest)(nil)
	_ twapi.HTTPResponser = (*TimelogGetResponse)(nil)
	_ twapi.HTTPRequester = (*TimelogListRequest)(nil)
	_ twapi.HTTPResponser = (*TimelogListResponse)(nil)
)

// Timelog refers to a recorded entry that tracks the amount of time a person
// has spent working on a specific task, project, or piece of work. These
// entries typically include details such as the duration of time worked, the
// date and time it was logged, who logged it, and any optional notes describing
// what was done during that period. Timelogs are essential for understanding
// how time is being allocated across projects, enabling teams to manage
// resources more effectively, invoice clients accurately, and assess
// productivity. They can be created manually or with timers, and are often used
// for reporting and billing purposes.
//
// sparsefields:gen
type Timelog struct {
	// ID is the unique identifier of the timelog.
	ID int64 `json:"id"`

	// Description is the description of the timelog.
	Description string `json:"description"`

	// Billable indicates whether the timelog is billable or not.
	Billable bool `json:"billable"`

	// Minutes is the number of minutes logged in the timelog.
	Minutes int64 `json:"minutes"`

	// LoggedAt is the date and time when the timelog was logged.
	LoggedAt time.Time `json:"timeLogged"`

	// User is the user that this timelog belongs to.
	User twapi.Relationship `json:"user"`

	// Task is the task associated with the timelog. It can be nil if the timelog
	// is not associated with a task.
	Task *twapi.Relationship `json:"task"`

	// Project is the project associated with the timelog.
	Project twapi.Relationship `json:"project"`

	// Tags are the tags associated with the timelog. They can be used to
	// categorize or label the timelog for easier filtering and searching.
	Tags []twapi.Relationship `json:"tags,omitempty"`

	// DeskTicketID is the desk ticket associated with the timelog. It can be
	// nil if the timelog is not associated with a desk ticket.
	DeskTicketID *int64 `json:"deskTicketId,omitempty"`

	// CreatedAt is the date and time when the timelog was created.
	CreatedAt time.Time `json:"createdAt"`

	// LoggedBy is the unique identifier of the user who logged the timelog.
	LoggedBy int64 `json:"loggedBy"`

	// UpdatedAt is the date and time when the timelog was last updated.
	UpdatedAt *time.Time `json:"updatedAt"`

	// UpdatedBy is the unique identifier of the user who last updated the
	// timelog.
	UpdatedBy *int64 `json:"updatedBy"`

	// DeletedAt is the date and time when the timelog was deleted, if it has been
	// deleted.
	DeletedAt *time.Time `json:"deletedAt"`

	// DeletedBy is the unique identifier of the user who deleted the timelog, if
	// it has been deleted.
	DeletedBy *int64 `json:"deletedBy"`

	// Deleted indicates whether the timelog has been deleted or not.
	Deleted bool `json:"deleted"`
}

// TimelogUpdateRequestPath contains the path parameters for creating a timelog.
type TimelogCreateRequestPath struct {
	// TaskID is an optional ID of the task associated with the timelog. If
	// provided, the timelog will be associated with this task. At least one of
	// TaskID or ProjectID must be provided. If both are provided, the timelog
	// will be associated with the task.
	TaskID int64
	// ProjectID is the unique identifier of the project where the timelog will be
	// created. At least one of TaskID or ProjectID must be provided. If both are
	// provided, the timelog will be associated with the task.
	ProjectID int64
}

// TimelogCreateRequest represents the request body for creating a new
// timelog.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/post-projects-api-v3-tasks-task-id-time-json
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/post-projects-api-v3-projects-project-id-time-json
type TimelogCreateRequest struct {
	// Path contains the path parameters for the request.
	Path TimelogCreateRequestPath `json:"-"`

	// Description is an optional description of the timelog.
	Description *string `json:"description"`

	// Date is the date when the timelog was logged. Only the date part is used,
	// the time part is ignored.
	Date twapi.Date `json:"date"`

	// Time is the time when the timelog was logged. It can be in local time or
	// UTC.
	Time twapi.Time `json:"time"`

	// IsUTC indicates whether the time is in UTC. When false, it will consider
	// the timezone configured for the logged user.
	IsUTC bool `json:"isUTC"`

	// Hours is the number of hours logged in the timelog. This is optional and
	// can be used instead of Minutes. If both Hours and Minutes are provided,
	// they will be summed up to calculate the total time logged.
	Hours int64 `json:"hours"`

	// Minutes is the number of minutes logged in the timelog. This is optional
	// and can be used instead of Hours. If both Hours and Minutes are provided,
	// they will be summed up to calculate the total time logged.
	Minutes int64 `json:"minutes"`

	// Billable indicates whether the timelog is billable or not. When not
	// provided, it will default to the tasklist or project settings.
	Billable *bool `json:"isBillable"`

	// UserID is an optional ID of the user who logged the timelog. If not
	// provided, the timelog will be logged by the user making the request.
	UserID *int64 `json:"userId"`

	// TagIDs is an optional list of tag IDs to associate with the timelog.
	TagIDs []int64 `json:"tagIds"`
}

// NewTimelogCreateRequestInTask creates a new TimelogCreateRequest with the
// provided datetime and duration, associating it with a specific task.
func NewTimelogCreateRequestInTask(taskID int64, datetime time.Time, duration time.Duration) TimelogCreateRequest {
	return TimelogCreateRequest{
		Path: TimelogCreateRequestPath{
			TaskID: taskID,
		},
		Date:    twapi.Date(datetime.UTC()),
		Time:    twapi.Time(datetime.UTC()),
		IsUTC:   true,
		Minutes: int64(duration.Minutes()),
	}
}

// NewTimelogCreateRequestInProject creates a new TimelogCreateRequest with the
// provided datetime and duration in a specific project.
func NewTimelogCreateRequestInProject(
	projectID int64,
	datetime time.Time,
	duration time.Duration,
) TimelogCreateRequest {
	return TimelogCreateRequest{
		Path: TimelogCreateRequestPath{
			ProjectID: projectID,
		},
		Date:    twapi.Date(datetime.UTC()),
		Time:    twapi.Time(datetime.UTC()),
		IsUTC:   true,
		Minutes: int64(duration.Minutes()),
	}
}

// HTTPRequest creates an HTTP request for the TimelogCreateRequest.
func (t TimelogCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	var uri string
	switch {
	case t.Path.TaskID > 0:
		uri = fmt.Sprintf("%s/projects/api/v3/tasks/%d/time.json", server, t.Path.TaskID)
	case t.Path.ProjectID > 0:
		uri = fmt.Sprintf("%s/projects/api/v3/projects/%d/time.json", server, t.Path.ProjectID)
	default:
		return nil, fmt.Errorf("either the task or project must be provided to create a timelog")
	}

	payload := struct {
		Timelog TimelogCreateRequest `json:"timelog"`
	}{Timelog: t}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create timelog request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// TimelogCreateResponse represents the response body for creating a new
// timelog.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/post-projects-api-v3-tasks-task-id-time-json
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/post-projects-api-v3-projects-project-id-time-json
type TimelogCreateResponse struct {
	// Timelog contains the created timelog information.
	Timelog Timelog `json:"timelog"`
}

// HandleHTTPResponse handles the HTTP response for the TimelogCreateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (t *TimelogCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create timelog")
	}
	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode create timelog response: %w", err)
	}
	if t.Timelog.ID == 0 {
		return fmt.Errorf("create timelog response does not contain a valid identifier")
	}
	return nil
}

// TimelogCreate creates a new timelog using the provided request and returns
// the response.
func TimelogCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req TimelogCreateRequest,
) (*TimelogCreateResponse, error) {
	return twapi.Execute[TimelogCreateRequest, *TimelogCreateResponse](ctx, engine, req)
}

// TimelogUpdateRequestPath contains the path parameters for updating a
// timelog.
type TimelogUpdateRequestPath struct {
	// ID is the unique identifier of the timelog to be updated.
	ID int64
}

// TimelogUpdateRequest represents the request body for updating a timelog.
// Besides the identifier, all other fields are optional. When a field is not
// provided, it will not be modified.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/patch-projects-api-v3-time-timelog-id-json
type TimelogUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path TimelogUpdateRequestPath `json:"-"`

	// ProjectID moves the timelog to another project. The API derives the
	// project from the task whenever the timelog ends up with one, so this only
	// takes effect on a timelog logged against a project, or alongside a TaskID
	// set to null.
	ProjectID *int64 `json:"projectId,omitempty"`

	// TaskID moves the timelog to another task, and to that task's project.
	// Setting it to null detaches the timelog from its task, leaving it logged
	// against the project.
	TaskID twapi.NullableInt64 `json:"taskId,omitzero"`

	// Description is an optional description of the timelog.
	Description *string `json:"description,omitempty"`

	// Date is the date when the timelog was logged. Only the date part is used,
	// the time part is ignored.
	Date *twapi.Date `json:"date,omitempty"`

	// Time is the time when the timelog was logged. It can be in local time or
	// UTC.
	Time *twapi.Time `json:"time,omitempty"`

	// IsUTC indicates whether the time is in UTC. When false, it will consider
	// the timezone configured for the logged user.
	IsUTC *bool `json:"isUTC,omitempty"`

	// Hours is the number of hours logged in the timelog. This is optional and
	// can be used instead of Minutes. If both Hours and Minutes are provided,
	// they will be summed up to calculate the total time logged.
	Hours *int64 `json:"hours,omitempty"`

	// Minutes is the number of minutes logged in the timelog. This is optional
	// and can be used instead of Hours. If both Hours and Minutes are provided,
	// they will be summed up to calculate the total time logged.
	Minutes *int64 `json:"minutes,omitempty"`

	// Billable indicates whether the timelog is billable or not.
	Billable *bool `json:"isBillable,omitempty"`

	// UserID is an optional ID of the user who logged the timelog. If not
	// provided, the timelog will be logged by the user making the request.
	UserID *int64 `json:"userId,omitempty"`

	// TagIDs is an optional list of tag IDs to associate with the timelog.
	TagIDs []int64 `json:"tagIds,omitempty"`
}

// NewTimelogUpdateRequest creates a new TimelogUpdateRequest with the
// provided timelog ID. The ID is required to update a timelog.
func NewTimelogUpdateRequest(timelogID int64) TimelogUpdateRequest {
	return TimelogUpdateRequest{
		Path: TimelogUpdateRequestPath{
			ID: timelogID,
		},
	}
}

// HTTPRequest creates an HTTP request for the TimelogUpdateRequest.
func (t TimelogUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/time/" + strconv.FormatInt(t.Path.ID, 10) + ".json"

	payload := struct {
		Timelog TimelogUpdateRequest `json:"timelog"`
	}{Timelog: t}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update timelog request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// TimelogUpdateResponse represents the response body for updating a timelog.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/patch-projects-api-v3-time-timelog-id-json
type TimelogUpdateResponse struct{}

// HandleHTTPResponse handles the HTTP response for the TimelogUpdateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (t *TimelogUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update timelog")
	}
	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode update timelog response: %w", err)
	}
	return nil
}

// TimelogUpdate updates a timelog using the provided request and returns the
// response.
func TimelogUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req TimelogUpdateRequest,
) (*TimelogUpdateResponse, error) {
	return twapi.Execute[TimelogUpdateRequest, *TimelogUpdateResponse](ctx, engine, req)
}

// TimelogBulkMoveRequestPath contains the path parameters for moving several
// timelogs to another project.
type TimelogBulkMoveRequestPath struct {
	// ProjectID is the project the timelogs are currently logged against. Only
	// timelogs in this project are moved.
	ProjectID int64
}

// TimelogBulkMoveRequest represents the request body for moving several
// timelogs from one project to another in a single request, through
// PUT /projects/{id}/time_entries/move.json.
//
// The endpoint only moves timelogs that are in the source project, are not
// deleted and have not been added to an invoice; any other listed ID is skipped
// and the response count shows how many moved. A moved timelog loses its task
// and its project-level tags, since neither belongs to the destination. To move
// a single timelog onto a task in the destination, use TimelogUpdateRequest
// with TaskID instead. The caller needs permission to log time in both
// projects.
type TimelogBulkMoveRequest struct {
	// Path contains the path parameters for the request.
	Path TimelogBulkMoveRequestPath `json:"-"`

	// TargetProjectID is the project that will receive the timelogs.
	TargetProjectID int64 `json:"projectId"`

	// TimelogIDs are the timelogs to move. The endpoint reads them as a
	// comma-separated list.
	TimelogIDs LegacyNumericList `json:"timelogIdList"`
}

// NewTimelogBulkMoveRequest creates a new TimelogBulkMoveRequest moving the
// given timelogs from the source project to the target project.
func NewTimelogBulkMoveRequest(sourceProjectID, targetProjectID int64, timelogIDs []int64) TimelogBulkMoveRequest {
	return TimelogBulkMoveRequest{
		Path: TimelogBulkMoveRequestPath{
			ProjectID: sourceProjectID,
		},
		TargetProjectID: targetProjectID,
		TimelogIDs:      timelogIDs,
	}
}

// HTTPRequest creates an HTTP request for the TimelogBulkMoveRequest.
func (t TimelogBulkMoveRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/" + strconv.FormatInt(t.Path.ProjectID, 10) + "/time_entries/move.json"

	// no envelope on this endpoint, the parameters are read from the top level
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(t); err != nil {
		return nil, fmt.Errorf("failed to encode bulk move timelogs request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// TimelogBulkMoveResponse represents the response body for moving several
// timelogs to another project.
type TimelogBulkMoveResponse struct {
	// Status is "OK" when the move was applied.
	Status string `json:"STATUS"`

	// MovedCount is the number of timelogs that were moved. It can be lower
	// than the number of IDs sent, since timelogs outside the source project,
	// deleted or already invoiced are skipped. The endpoint does not say which
	// ones were skipped.
	MovedCount LegacyNumber `json:"numberOfMovedTimelogs"`
}

// HandleHTTPResponse handles the HTTP response for the
// TimelogBulkMoveResponse. If some unexpected HTTP status code is returned by
// the API, a twapi.HTTPError is returned.
//
// The endpoint can also answer 200 with an `errMsg` body and no `STATUS`, for
// example when none of the listed timelogs could be moved or the caller cannot
// log time in one of the projects. That is reported as a twapi.HTTPError
// carrying the 200 status and the body as its details, so nothing is mistaken
// for a successful move.
func (t *TimelogBulkMoveResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to bulk move timelogs")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read bulk move timelogs response: %w", err)
	}
	var envelope struct {
		Status string `json:"STATUS"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil || !strings.EqualFold(envelope.Status, "OK") {
		details := string(body)
		if details == "" {
			details = "no response body"
		}
		return &twapi.HTTPError{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Message:    "failed to bulk move timelogs",
			Details:    details,
		}
	}
	if err := json.Unmarshal(body, t); err != nil {
		return fmt.Errorf("failed to decode bulk move timelogs response: %w", err)
	}
	return nil
}

// TimelogBulkMove moves several timelogs from one project to another in a
// single request, and returns how many were moved.
func TimelogBulkMove(
	ctx context.Context,
	engine *twapi.Engine,
	req TimelogBulkMoveRequest,
) (*TimelogBulkMoveResponse, error) {
	return twapi.Execute[TimelogBulkMoveRequest, *TimelogBulkMoveResponse](ctx, engine, req)
}

// TimelogBulkBillableUpdateFilters selects the timelogs a bulk billable update
// changes. Every filter narrows the selection; a filter left empty does not
// restrict it. With every filter empty the update reaches every timelog in the
// installation that the caller can see, so callers should always set at least
// ProjectIDs.
type TimelogBulkBillableUpdateFilters struct {
	// ProjectIDs restricts the update to timelogs in these projects.
	ProjectIDs LegacyNumericList `json:"projectId,omitempty"`

	// UserIDs restricts the update to timelogs logged for these users. It is
	// ignored when UserCompanyIDs is set.
	UserIDs LegacyNumericList `json:"userId,omitempty"`

	// UserCompanyIDs restricts the update to timelogs logged for users who
	// belong to these companies.
	UserCompanyIDs LegacyNumericList `json:"companyId,omitempty"`

	// ProjectCompanyIDs restricts the update to timelogs in projects owned by
	// these companies.
	ProjectCompanyIDs LegacyNumericList `json:"projectsFromCompanyId,omitempty"`

	// FromDate restricts the update to timelogs logged on or after the start of
	// this day, in the caller's timezone.
	FromDate *LegacyDate `json:"fromDate,omitempty"`

	// ToDate restricts the update to timelogs logged on or before the end of
	// this day, in the caller's timezone.
	ToDate *LegacyDate `json:"toDate,omitempty"`

	// BillableType restricts the update to timelogs that are currently billable
	// or non-billable. Use the TimelogBillableType constants. The endpoint
	// defaults to all. Setting it to the opposite of the target status makes the
	// response count the timelogs that actually changed.
	BillableType TimelogBillableType `json:"billableType,omitempty"`

	// InvoicedType restricts the update to timelogs that have or have not been
	// added to an invoice. Use the TimelogInvoicedType constants. The endpoint
	// defaults to all, so invoiced timelogs are changed too unless this is set
	// to TimelogInvoicedTypeNonInvoiced.
	InvoicedType TimelogInvoicedType `json:"invoicedType,omitempty"`

	// TagIDs restricts the update to timelogs carrying these tags.
	TagIDs LegacyNumericList `json:"tagIds,omitempty"`

	// TaskTagIDs restricts the update to timelogs whose task carries these
	// tags.
	TaskTagIDs LegacyNumericList `json:"taskTagIds,omitempty"`

	// MatchAllTags requires a timelog to carry every tag in TagIDs rather than
	// any one of them. The endpoint defaults to false.
	MatchAllTags *bool `json:"matchAllTags,omitempty"`
}

// TimelogBulkBillableUpdateRequest represents the request body for setting the
// billable status of every timelog matching a filter in a single request,
// through PUT /time_entries.json.
//
// The endpoint selects timelogs by filter, not by ID, and billable status is
// the only attribute it changes. Only site administrators who belong to the
// site owner's company may call it; anyone else is answered 403. When no
// timelog matches the filters it answers 422 and changes nothing.
type TimelogBulkBillableUpdateRequest struct {
	// Billable is the status every matching timelog is given.
	Billable bool `json:"isBillable"`

	// Filters selects the timelogs to change.
	TimelogBulkBillableUpdateFilters
}

// NewTimelogBulkBillableUpdateRequest creates a new
// TimelogBulkBillableUpdateRequest that sets the billable status of the
// timelogs in the given projects. Narrow it further with the other filters.
func NewTimelogBulkBillableUpdateRequest(billable bool, projectIDs ...int64) TimelogBulkBillableUpdateRequest {
	return TimelogBulkBillableUpdateRequest{
		Billable: billable,
		TimelogBulkBillableUpdateFilters: TimelogBulkBillableUpdateFilters{
			ProjectIDs: projectIDs,
		},
	}
}

// HTTPRequest creates an HTTP request for the TimelogBulkBillableUpdateRequest.
func (t TimelogBulkBillableUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/time_entries.json"

	// no envelope on this endpoint, the parameters are read from the top level
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(t); err != nil {
		return nil, fmt.Errorf("failed to encode bulk billable update request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// TimelogBulkBillableUpdateResponse represents the response for setting the
// billable status of the timelogs matching a filter.
type TimelogBulkBillableUpdateResponse struct {
	// Count is the number of timelogs the filters matched, all of which now
	// have the requested status. The endpoint reports it in the `count`
	// response header; it is nil when the header is missing.
	Count *int64 `json:"-"`
}

// HandleHTTPResponse handles the HTTP response for the
// TimelogBulkBillableUpdateResponse. If some unexpected HTTP status code is
// returned by the API, a twapi.HTTPError is returned.
func (t *TimelogBulkBillableUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to bulk update timelogs billable status")
	}
	if header := resp.Header.Get("count"); header != "" {
		count, err := strconv.ParseInt(strings.TrimSpace(header), 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse count header of bulk billable update response: %w", err)
		}
		t.Count = &count
	}
	return nil
}

// TimelogBulkBillableUpdate sets the billable status of every timelog matching
// the request's filters in a single request, and returns how many matched.
func TimelogBulkBillableUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req TimelogBulkBillableUpdateRequest,
) (*TimelogBulkBillableUpdateResponse, error) {
	return twapi.Execute[TimelogBulkBillableUpdateRequest, *TimelogBulkBillableUpdateResponse](ctx, engine, req)
}

// TimelogDeleteRequestPath contains the path parameters for deleting a
// timelog.
type TimelogDeleteRequestPath struct {
	// ID is the unique identifier of the timelog to be deleted.
	ID int64
}

// TimelogDeleteRequest represents the request body for deleting a timelog.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/delete-projects-api-v3-time-timelog-id-json
type TimelogDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path TimelogDeleteRequestPath
}

// NewTimelogDeleteRequest creates a new TimelogDeleteRequest with the provided
// timelog ID.
func NewTimelogDeleteRequest(timelogID int64) TimelogDeleteRequest {
	return TimelogDeleteRequest{
		Path: TimelogDeleteRequestPath{
			ID: timelogID,
		},
	}
}

// HTTPRequest creates an HTTP request for the TimelogDeleteRequest.
func (t TimelogDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/time/" + strconv.FormatInt(t.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// TimelogDeleteResponse represents the response body for deleting a timelog.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/delete-projects-api-v3-time-timelog-id-json
type TimelogDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the TimelogDeleteResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (t *TimelogDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusNoContent {
		return twapi.NewHTTPError(resp, "failed to delete timelog")
	}
	return nil
}

// TimelogDelete deletes a timelog using the provided request and returns the
// response.
func TimelogDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req TimelogDeleteRequest,
) (*TimelogDeleteResponse, error) {
	return twapi.Execute[TimelogDeleteRequest, *TimelogDeleteResponse](ctx, engine, req)
}

// TimelogGetRequestPath contains the path parameters for loading a single
// timelog.
type TimelogGetRequestPath struct {
	// ID is the unique identifier of the timelog to be retrieved.
	ID int64 `json:"id"`
}

// TimelogGetRequest represents the request body for loading a single timelog.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/get-projects-api-v3-time-timelog-id-json
type TimelogGetRequest struct {
	// Path contains the path parameters for the request.
	Path TimelogGetRequestPath

	// Fields restricts the attributes returned for the timelog. Each slot of
	// TimelogGetFields is a separate `fields[entity]=…` selection; populated
	// slots restrict the response, empty slots return the API default. Use the
	// generated TimelogField constants to ensure values match real attributes.
	Fields TimelogGetFields
}

// NewTimelogGetRequest creates a new TimelogGetRequest with the provided
// timelog ID. The ID is required to load a timelog.
func NewTimelogGetRequest(timelogID int64) TimelogGetRequest {
	return TimelogGetRequest{
		Path: TimelogGetRequestPath{
			ID: timelogID,
		},
	}
}

// HTTPRequest creates an HTTP request for the TimelogGetRequest.
func (t TimelogGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/time/" + strconv.FormatInt(t.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	t.Fields.apply(query)
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// TimelogGetResponse contains all the information related to a timelog.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/get-projects-api-v3-time-timelog-id-json
//
// sparsefields:get
type TimelogGetResponse struct {
	Timelog Timelog `json:"timelog"`
}

// HandleHTTPResponse handles the HTTP response for the TimelogGetResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (t *TimelogGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve timelog")
	}

	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode retrieve timelog response: %w", err)
	}
	return nil
}

// TimelogGet retrieves a single timelog using the provided request and returns
// the response.
func TimelogGet(
	ctx context.Context,
	engine *twapi.Engine,
	req TimelogGetRequest,
) (*TimelogGetResponse, error) {
	return twapi.Execute[TimelogGetRequest, *TimelogGetResponse](ctx, engine, req)
}

// TimelogListRequestPath contains the path parameters for loading multiple
// timelogs.
type TimelogListRequestPath struct {
	// TaskID is an optional ID of the task whose timelogs are to be retrieved. If
	// provided, the timelogs will be filtered by this task. When both TaskID and
	// ProjectID are provided, the timelogs will be filtered by the task.
	TaskID int64
	// ProjectID is the unique identifier of the project whose timelogs are to be
	// retrieved. If provided, the timelogs will be filtered by this project. When
	// both TaskID and ProjectID are provided, the timelogs will be filtered by
	// the task.
	ProjectID int64
}

// TimelogOrderBy identifies the attributes a timelog list can be ordered by.
type TimelogOrderBy string

// Supported timelog order-by values.
const (
	TimelogOrderByCompany     TimelogOrderBy = "company"
	TimelogOrderByDate        TimelogOrderBy = "date"
	TimelogOrderByDateUpdated TimelogOrderBy = "dateupdated"
	TimelogOrderByProject     TimelogOrderBy = "project"
	TimelogOrderByTask        TimelogOrderBy = "task"
	TimelogOrderByTasklist    TimelogOrderBy = "tasklist"
	TimelogOrderByUser        TimelogOrderBy = "user"
	TimelogOrderByDescription TimelogOrderBy = "description"
	TimelogOrderByBilled      TimelogOrderBy = "billed"
	TimelogOrderByBillable    TimelogOrderBy = "billable"
	TimelogOrderByTimeSpent   TimelogOrderBy = "timespent"
	TimelogOrderByID          TimelogOrderBy = "id"
)

// TimelogBillableType identifies whether a timelog list is restricted to
// billable or non-billable entries.
type TimelogBillableType string

// Supported timelog billable-type values.
const (
	TimelogBillableTypeAll         TimelogBillableType = "all"
	TimelogBillableTypeBillable    TimelogBillableType = "billable"
	TimelogBillableTypeNonBillable TimelogBillableType = "nonbillable"
)

// TimelogInvoicedType identifies whether a timelog list is restricted to
// entries that have been added to an invoice.
type TimelogInvoicedType string

// Supported timelog invoiced-type values.
const (
	TimelogInvoicedTypeAll         TimelogInvoicedType = "all"
	TimelogInvoicedTypeInvoiced    TimelogInvoicedType = "invoiced"
	TimelogInvoicedTypeNonInvoiced TimelogInvoicedType = "noninvoiced"
)

// TimelogListRequestFilters contains the filters for loading multiple
// timelogs.
type TimelogListRequestFilters struct {
	// OrderBy is the field to sort the results by. Use the TimelogOrderBy
	// constants. The endpoint defaults to date.
	OrderBy TimelogOrderBy

	// OrderMode is the direction to sort the results in. See twapi.OrderMode for
	// the supported values. The endpoint defaults to ascending.
	OrderMode twapi.OrderMode

	// TagIDs is an optional list of tag IDs to filter the timelogs by. If provided,
	// only timelogs associated with these tags will be returned.
	TagIDs []int64

	// MatchAllTags indicates whether to match all tags or any tag. If true, only
	// timelogs that have all the specified tags will be returned. If false,
	// timelogs that have at least one of the specified tags will be returned.
	MatchAllTags *bool

	// StartDate is an optional start date to filter the timelogs by. If provided,
	// only timelogs that started on or after this date will be returned.
	StartDate *time.Time

	// EndDate is an optional end date to filter the timelogs by. If provided,
	// only timelogs that ended on or before this date will be returned.
	EndDate *time.Time

	// AssignedToUserIDs is an optional list of user IDs to filter the timelogs
	// by. If provided, only timelogs assigned to these users will be returned.
	AssignedToUserIDs []int64

	// AssignedToCompanyIDs is an optional list of client/company IDs to filter
	// the timelogs by users in the specific companies. If provided, only timelogs
	// assigned to these companies will be returned.
	AssignedToCompanyIDs []int64

	// AssignedToTeamIDs is an optional list of team IDs to filter the timelogs by
	// users in the specific teams. If provided, only timelogs assigned to these
	// teams will be returned.
	AssignedToTeamIDs []int64

	// DeskTicketIDs is an optional list of desk ticket IDs to filter the
	// timelogs by. If provided, only timelogs associated with these desk
	// tickets will be returned.
	DeskTicketIDs []int64

	// BillableType restricts the results to billable or non-billable timelogs.
	// Use the TimelogBillableType constants. The endpoint defaults to all.
	BillableType TimelogBillableType

	// InvoicedType restricts the results to timelogs that have or have not been
	// added to an invoice. Use the TimelogInvoicedType constants. The endpoint
	// defaults to all.
	InvoicedType TimelogInvoicedType

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of timelogs to retrieve per page. Defaults to 50.
	PageSize int64

	// CountMode selects whether the API computes the exact number of timelogs
	// matching the filters, reported in Meta.Page.Count. Defaults to
	// twapi.ListCountModeDefault, which leaves the decision to the API.
	CountMode twapi.ListCountMode

	// Fields restricts the attributes returned for the timelog and each of its
	// sideloads. Each slot of TimelogListFields is a separate `fields[entity]=…`
	// selection; populated slots restrict the response, empty slots return the
	// API default. Use the generated TimelogField constants to ensure values
	// match real attributes.
	Fields TimelogListFields
}

func (t TimelogListRequestFilters) apply(req *http.Request) {
	query := req.URL.Query()
	if t.OrderBy != "" {
		query.Set("orderBy", string(t.OrderBy))
	}
	if t.OrderMode != "" {
		query.Set("orderMode", string(t.OrderMode))
	}
	if len(t.TagIDs) > 0 {
		tagIDs := make([]string, len(t.TagIDs))
		for i, id := range t.TagIDs {
			tagIDs[i] = strconv.FormatInt(id, 10)
		}
		query.Set("tagIds", strings.Join(tagIDs, ","))
	}
	if t.MatchAllTags != nil {
		query.Set("matchAllTags", strconv.FormatBool(*t.MatchAllTags))
	}
	if t.StartDate != nil && !t.StartDate.IsZero() {
		query.Set("startDate", t.StartDate.Format(time.RFC3339))
	}
	if t.EndDate != nil && !t.EndDate.IsZero() {
		query.Set("endDate", t.EndDate.Format(time.RFC3339))
	}
	if len(t.AssignedToUserIDs) > 0 {
		assignedToUserIDs := make([]string, len(t.AssignedToUserIDs))
		for i, id := range t.AssignedToUserIDs {
			assignedToUserIDs[i] = strconv.FormatInt(id, 10)
		}
		query.Set("assignedToUserIds", strings.Join(assignedToUserIDs, ","))
	}
	if len(t.AssignedToCompanyIDs) > 0 {
		assignedToCompanyIDs := make([]string, len(t.AssignedToCompanyIDs))
		for i, id := range t.AssignedToCompanyIDs {
			assignedToCompanyIDs[i] = strconv.FormatInt(id, 10)
		}
		query.Set("assignedToCompanyIds", strings.Join(assignedToCompanyIDs, ","))
	}
	if len(t.AssignedToTeamIDs) > 0 {
		assignedToTeamIDs := make([]string, len(t.AssignedToTeamIDs))
		for i, id := range t.AssignedToTeamIDs {
			assignedToTeamIDs[i] = strconv.FormatInt(id, 10)
		}
		query.Set("assignedToTeamIds", strings.Join(assignedToTeamIDs, ","))
	}
	if len(t.DeskTicketIDs) > 0 {
		deskTicketIDs := make([]string, len(t.DeskTicketIDs))
		for i, id := range t.DeskTicketIDs {
			deskTicketIDs[i] = strconv.FormatInt(id, 10)
		}
		query.Set("deskTicketIds", strings.Join(deskTicketIDs, ","))
	}
	if t.BillableType != "" {
		query.Set("billableType", string(t.BillableType))
	}
	if t.InvoicedType != "" {
		query.Set("invoicedType", string(t.InvoicedType))
	}
	if t.Page > 0 {
		query.Set("page", strconv.FormatInt(t.Page, 10))
	}
	if t.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(t.PageSize, 10))
	}
	t.Fields.apply(query)
	t.CountMode.Apply(query)
	req.URL.RawQuery = query.Encode()
}

// TimelogListRequest represents the request body for loading multiple timelogs.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/get-projects-api-v3-time-json
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/get-projects-api-v3-tasks-task-id-time-json
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/get-projects-api-v3-projects-project-id-time-json
type TimelogListRequest struct {
	// Path contains the path parameters for the request.
	Path TimelogListRequestPath

	// Filters contains the filters for loading multiple timelogs.
	Filters TimelogListRequestFilters
}

// NewTimelogListRequest creates a new TimelogListRequest with default values.
func NewTimelogListRequest() TimelogListRequest {
	return TimelogListRequest{
		Filters: TimelogListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the TimelogListRequest.
func (t TimelogListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	var uri string
	switch {
	case t.Path.TaskID > 0:
		uri = fmt.Sprintf("%s/projects/api/v3/tasks/%d/time.json", server, t.Path.TaskID)
	case t.Path.ProjectID > 0:
		uri = fmt.Sprintf("%s/projects/api/v3/projects/%d/time.json", server, t.Path.ProjectID)
	default:
		uri = server + "/projects/api/v3/time.json"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	t.Filters.apply(req)

	return req, nil
}

// TimelogListResponse contains information by multiple timelogs matching the
// request filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/get-projects-api-v3-time-json
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/get-projects-api-v3-tasks-task-id-time-json
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/get-projects-api-v3-projects-project-id-time-json
//
// sparsefields:list
type TimelogListResponse struct {
	request TimelogListRequest

	Meta     twapi.ListMeta `json:"meta"`
	Timelogs []Timelog      `json:"timelogs"`
}

// HandleHTTPResponse handles the HTTP response for the TimelogListResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (t *TimelogListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list timelogs")
	}

	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode list timelogs response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (t *TimelogListResponse) SetRequest(req TimelogListRequest) {
	t.request = req
	t.Meta.ResolveCount(req.Filters.CountMode)
}

// Iterate returns the request set to the next page, if available. If there
// are no more pages, a nil request is returned.
func (t *TimelogListResponse) Iterate() *TimelogListRequest {
	if !t.Meta.Page.HasMore {
		return nil
	}
	req := t.request
	req.Filters.Page++
	return &req
}

// TimelogList retrieves multiple timelogs using the provided request and
// returns the response.
func TimelogList(
	ctx context.Context,
	engine *twapi.Engine,
	req TimelogListRequest,
) (*TimelogListResponse, error) {
	return twapi.Execute[TimelogListRequest, *TimelogListResponse](ctx, engine, req)
}
