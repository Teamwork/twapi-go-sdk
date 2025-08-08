package projects

import (
	"bytes"
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
	_ twapi.HTTPRequester = (*TimelogCreateRequest)(nil)
	_ twapi.HTTPResponser = (*TimelogCreateResponse)(nil)
	_ twapi.HTTPRequester = (*TimelogUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*TimelogUpdateResponse)(nil)
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

	// Billable indicates whether the timelog is billable or not.
	Billable bool `json:"isBillable"`

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

	return req, nil
}

// TimelogGetResponse contains all the information related to a timelog.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/get-projects-api-v3-time-timelog-id-json
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

// TimelogListRequestFilters contains the filters for loading multiple
// timelogs.
type TimelogListRequestFilters struct {
	// TagIDs is an optional list of tag IDs to filter the timelogs by. If provided,
	// only timelogs associated with these tags will be returned.
	TagIDs []int64

	// MatchAllTags indicates whether to match all tags or any tag. If true, only
	// timelogs that have all the specified tags will be returned. If false,
	// timelogs that have at least one of the specified tags will be returned.
	MatchAllTags *bool

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of timelogs to retrieve per page. Defaults to 50.
	PageSize int64
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

	query := req.URL.Query()
	if len(t.Filters.TagIDs) > 0 {
		tagIDs := make([]string, len(t.Filters.TagIDs))
		for i, id := range t.Filters.TagIDs {
			tagIDs[i] = strconv.FormatInt(id, 10)
		}
		query.Set("tagIds", strings.Join(tagIDs, ","))
	}
	if t.Filters.MatchAllTags != nil {
		query.Set("matchAllTags", strconv.FormatBool(*t.Filters.MatchAllTags))
	}
	if t.Filters.Page > 0 {
		query.Set("page", strconv.FormatInt(t.Filters.Page, 10))
	}
	if t.Filters.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(t.Filters.PageSize, 10))
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// TimelogListResponse contains information by multiple timelogs matching the
// request filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/get-projects-api-v3-time-json
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/get-projects-api-v3-tasks-task-id-time-json
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/get-projects-api-v3-projects-project-id-time-json
type TimelogListResponse struct {
	request TimelogListRequest

	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`
	Timelogs []Timelog `json:"timelogs"`
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
