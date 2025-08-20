package projects

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var (
	_ twapi.HTTPRequester = (*TimerCreateRequest)(nil)
	_ twapi.HTTPResponser = (*TimerCreateResponse)(nil)
	_ twapi.HTTPRequester = (*TimerUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*TimerUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*TimerPauseRequest)(nil)
	_ twapi.HTTPResponser = (*TimerPauseResponse)(nil)
	_ twapi.HTTPRequester = (*TimerResumeRequest)(nil)
	_ twapi.HTTPResponser = (*TimerResumeResponse)(nil)
	_ twapi.HTTPRequester = (*TimerCompleteRequest)(nil)
	_ twapi.HTTPResponser = (*TimerCompleteResponse)(nil)
	_ twapi.HTTPRequester = (*TimerDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*TimerDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*TimerGetRequest)(nil)
	_ twapi.HTTPResponser = (*TimerGetResponse)(nil)
	_ twapi.HTTPRequester = (*TimerListRequest)(nil)
	_ twapi.HTTPResponser = (*TimerListResponse)(nil)
)

// Timer is a built-in tool that allows users to accurately track the time they
// spend working on specific tasks, projects, or client work. Instead of
// manually recording hours, users can start, pause, and stop timers directly
// within the platform or through the desktop and mobile apps, ensuring precise
// time logs without interrupting their workflow. Once recorded, these entries
// are automatically linked to the relevant task or project, making it easier to
// monitor productivity, manage billable hours, and generate detailed reports
// for both internal tracking and client invoicing.
//
// More information can be found at:
// https://support.teamwork.com/projects/time-tracking/multiple-timers
type Timer struct {
	// ID is the unique identifier of the timer.
	ID int64 `json:"id"`

	// Description is a brief summary of the timer's purpose.
	Description string `json:"description"`

	// Running indicates whether the timer is currently running.
	Running bool `json:"running"`

	// Billable indicates whether the timer is billable.
	Billable bool `json:"billable"`

	// User is the user associated with the timer.
	User twapi.Relationship `json:"user"`

	// Task is the task associated with the timer.
	Task *twapi.Relationship `json:"task"`

	// Project is the project associated with the timer.
	Project twapi.Relationship `json:"project"`

	// Timelog is the timelog associated with the timer.
	Timelog *twapi.Relationship `json:"timelog,omitempty"`

	// CreatedAt is the date and time when the timer was created.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt is the date and time when the timer was last updated.
	UpdatedAt time.Time `json:"updatedAt"`

	// DeletedAt is the date and time when the timer was deleted.
	DeletedAt *time.Time `json:"deletedAt"`

	// Deleted indicates whether the timer has been deleted.
	Deleted bool `json:"deleted"`

	// Duration is the total duration of the timer in seconds.
	Duration int64 `json:"duration"`

	// LastStartedAt is the date and time when the timer was last started.
	LastStartedAt time.Time `json:"lastStartedAt"`

	// LastIntervalAt is the date and time when the last interval ended.
	LastIntervalAt *time.Time `json:"timerLastIntervalEnd,omitempty"`

	// Intervals is a list of time intervals for the timer.
	Intervals []struct {
		// ID is the unique identifier of the interval.
		ID int64 `json:"id"`

		// From is the start time of the interval.
		From time.Time `json:"from"`

		// To is the end time of the interval.
		To time.Time `json:"to"`

		// Duration is the total duration of the interval in seconds.
		Duration int64 `json:"duration"`
	} `json:"intervals"`
}

// TimerCreateRequest represents the request body for creating a new
// timer.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/post-projects-api-v3-me-timers-json
type TimerCreateRequest struct {
	// Description is a brief summary of the timer's purpose.
	Description *string `json:"description"`

	// Billable indicates whether the timer is billable.
	Billable *bool `json:"isBillable"`

	// Running indicates whether the timer is currently running.
	Running *bool `json:"isRunning"`

	// Seconds is the total duration of the timer in seconds.
	Seconds *int64 `json:"seconds"`

	// StopRunningTimers indicates whether to stop all running timers.
	StopRunningTimers *bool `json:"stopRunningTimers"`

	// ProjectID is the unique identifier of the project associated with the
	// timer. The ProjectID must be provided.
	ProjectID int64 `json:"projectId"`

	// TaskID is the unique identifier of the task associated with the timer.
	TaskID *int64 `json:"taskId"`
}

// NewTimerCreateRequest creates a new TimerCreateRequest with the provided
// project.
func NewTimerCreateRequest(projectID int64) TimerCreateRequest {
	return TimerCreateRequest{
		ProjectID: projectID,
	}
}

// HTTPRequest creates an HTTP request for the TimerCreateRequest.
func (t TimerCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/me/timers.json"

	payload := struct {
		Timer TimerCreateRequest `json:"timer"`
	}{Timer: t}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create timer request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// TimerCreateResponse represents the response body for creating a new timer.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/post-projects-api-v3-me-timers-json
type TimerCreateResponse struct {
	// Timer represents the created timer.
	Timer Timer `json:"timer"`
}

// HandleHTTPResponse handles the HTTP response for the TimerCreateResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (t *TimerCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create timer")
	}
	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode create timer response: %w", err)
	}
	if t.Timer.ID == 0 {
		return fmt.Errorf("create timer response does not contain a valid identifier")
	}
	return nil
}

// TimerCreate creates a new timer using the provided request and returns the
// response.
func TimerCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req TimerCreateRequest,
) (*TimerCreateResponse, error) {
	return twapi.Execute[TimerCreateRequest, *TimerCreateResponse](ctx, engine, req)
}

// TimerUpdateRequestPath contains the path parameters for updating a timer.
type TimerUpdateRequestPath struct {
	// ID is the unique identifier of the timer to be updated.
	ID int64
}

// TimerUpdateRequest represents the request body for updating a timer. Besides
// the identifier, all other fields are optional. When a field is not provided,
// it will not be modified.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/patch-projects-api-v3-time-timer-id-json
type TimerUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path TimerUpdateRequestPath `json:"-"`

	// Description is a brief summary of the timer's purpose.
	Description *string `json:"description"`

	// Billable indicates whether the timer is billable.
	Billable *bool `json:"isBillable"`

	// Running indicates whether the timer is currently running.
	Running *bool `json:"isRunning"`

	// StopRunningTimers indicates whether to stop all running timers.
	StopRunningTimers *bool `json:"stopRunningTimers"`

	// ProjectID is the unique identifier of the project associated with the
	// timer. The ProjectID must be provided.
	ProjectID int64 `json:"projectId"`

	// TaskID is the unique identifier of the task associated with the timer.
	TaskID *int64 `json:"taskId"`
}

// NewTimerUpdateRequest creates a new TimerUpdateRequest with the
// provided timer ID. The ID is required to update a timer.
func NewTimerUpdateRequest(timerID int64) TimerUpdateRequest {
	return TimerUpdateRequest{
		Path: TimerUpdateRequestPath{
			ID: timerID,
		},
	}
}

// HTTPRequest creates an HTTP request for the TimerUpdateRequest.
func (t TimerUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/me/timers/" + strconv.FormatInt(t.Path.ID, 10) + ".json"

	payload := struct {
		Timer TimerUpdateRequest `json:"timer"`
	}{Timer: t}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update timer request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// TimerUpdateResponse represents the response body for updating a timer.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/patch-projects-api-v3-time-timer-id-json
type TimerUpdateResponse struct {
	// Timer represents the updated timer.
	Timer Timer `json:"timer"`
}

// HandleHTTPResponse handles the HTTP response for the TimerUpdateResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (t *TimerUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update timer")
	}
	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode update timer response: %w", err)
	}
	return nil
}

// TimerUpdate updates a timer using the provided request and returns the
// response.
func TimerUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req TimerUpdateRequest,
) (*TimerUpdateResponse, error) {
	return twapi.Execute[TimerUpdateRequest, *TimerUpdateResponse](ctx, engine, req)
}

// TimerPauseRequestPath contains the path parameters for pausing a timer.
type TimerPauseRequestPath struct {
	// ID is the unique identifier of the timer to be paused.
	ID int64
}

// TimerPauseRequest represents the request body for pausing a timer.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/put-projects-api-v3-me-timers-timer-id-pause-json
type TimerPauseRequest struct {
	// Path contains the path parameters for the request.
	Path TimerPauseRequestPath `json:"-"`
}

// NewTimerPauseRequest creates a new TimerPauseRequest with the
// provided timer ID. The ID is required to pause a timer.
func NewTimerPauseRequest(timerID int64) TimerPauseRequest {
	return TimerPauseRequest{
		Path: TimerPauseRequestPath{
			ID: timerID,
		},
	}
}

// HTTPRequest creates an HTTP request for the TimerPauseRequest.
func (t TimerPauseRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/me/timers/" + strconv.FormatInt(t.Path.ID, 10) + "/pause.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// TimerPauseResponse represents the response body for pausing a timer.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/put-projects-api-v3-me-timers-timer-id-pause-json
type TimerPauseResponse struct {
	// Timer represents the paused timer.
	Timer Timer `json:"timer"`
}

// HandleHTTPResponse handles the HTTP response for the TimerPauseResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (t *TimerPauseResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to pause timer")
	}
	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode pause timer response: %w", err)
	}
	return nil
}

// TimerPause pauses a timer using the provided request and returns the
// response.
func TimerPause(
	ctx context.Context,
	engine *twapi.Engine,
	req TimerPauseRequest,
) (*TimerPauseResponse, error) {
	return twapi.Execute[TimerPauseRequest, *TimerPauseResponse](ctx, engine, req)
}

// TimerResumeRequestPath contains the path parameters for resuming a timer.
type TimerResumeRequestPath struct {
	// ID is the unique identifier of the timer to be resumed.
	ID int64
}

// TimerResumeRequest represents the request body for resuming a timer.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/put-projects-api-v3-me-timers-timer-id-resume-json
type TimerResumeRequest struct {
	// Path contains the path parameters for the request.
	Path TimerResumeRequestPath `json:"-"`
}

// NewTimerResumeRequest creates a new TimerResumeRequest with the provided
// timer ID. The ID is required to resume a timer.
func NewTimerResumeRequest(timerID int64) TimerResumeRequest {
	return TimerResumeRequest{
		Path: TimerResumeRequestPath{
			ID: timerID,
		},
	}
}

// HTTPRequest creates an HTTP request for the TimerResumeRequest.
func (t TimerResumeRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/me/timers/" + strconv.FormatInt(t.Path.ID, 10) + "/resume.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// TimerResumeResponse represents the response body for resuming a timer.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/put-projects-api-v3-me-timers-timer-id-resume-json
type TimerResumeResponse struct {
	// Timer represents the resumed timer.
	Timer Timer `json:"timer"`
}

// HandleHTTPResponse handles the HTTP response for the TimerResumeResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (t *TimerResumeResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to resume timer")
	}
	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode resume timer response: %w", err)
	}
	return nil
}

// TimerResume resumes a timer using the provided request and returns the
// response.
func TimerResume(
	ctx context.Context,
	engine *twapi.Engine,
	req TimerResumeRequest,
) (*TimerResumeResponse, error) {
	return twapi.Execute[TimerResumeRequest, *TimerResumeResponse](ctx, engine, req)
}

// TimerCompleteRequestPath contains the path parameters for completing a timer.
type TimerCompleteRequestPath struct {
	// ID is the unique identifier of the timer to be completed.
	ID int64
}

// TimerCompleteRequest represents the request body for completing a timer.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/put-projects-api-v3-me-timers-timer-id-complete-json
type TimerCompleteRequest struct {
	// Path contains the path parameters for the request.
	Path TimerCompleteRequestPath `json:"-"`
}

// NewTimerCompleteRequest creates a new TimerCompleteRequest with the provided
// timer ID. The ID is required to complete a timer.
func NewTimerCompleteRequest(timerID int64) TimerCompleteRequest {
	return TimerCompleteRequest{
		Path: TimerCompleteRequestPath{
			ID: timerID,
		},
	}
}

// HTTPRequest creates an HTTP request for the TimerCompleteRequest.
func (t TimerCompleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/me/timers/" + strconv.FormatInt(t.Path.ID, 10) + "/complete.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// TimerCompleteResponse represents the response body for completing a timer.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/put-projects-api-v3-me-timers-timer-id-complete-json
type TimerCompleteResponse struct {
	// Timer represents the completed timer.
	Timer Timer `json:"timer"`
}

// HandleHTTPResponse handles the HTTP response for the TimerCompleteResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (t *TimerCompleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to complete timer")
	}
	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode complete timer response: %w", err)
	}
	return nil
}

// TimerComplete completes a timer using the provided request and returns the
// response.
func TimerComplete(
	ctx context.Context,
	engine *twapi.Engine,
	req TimerCompleteRequest,
) (*TimerCompleteResponse, error) {
	return twapi.Execute[TimerCompleteRequest, *TimerCompleteResponse](ctx, engine, req)
}

// TimerDeleteRequestPath contains the path parameters for deleting a timer.
type TimerDeleteRequestPath struct {
	// ID is the unique identifier of the timer to be deleted.
	ID int64
}

// TimerDeleteRequest represents the request body for deleting a timer.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/delete-projects-api-v3-time-timer-id-json
type TimerDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path TimerDeleteRequestPath
}

// NewTimerDeleteRequest creates a new TimerDeleteRequest with the
// provided timer ID.
func NewTimerDeleteRequest(timerID int64) TimerDeleteRequest {
	return TimerDeleteRequest{
		Path: TimerDeleteRequestPath{
			ID: timerID,
		},
	}
}

// HTTPRequest creates an HTTP request for the TimerDeleteRequest.
func (t TimerDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/me/timers/" + strconv.FormatInt(t.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// TimerDeleteResponse represents the response body for deleting a timer.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/delete-projects-api-v3-time-timer-id-json
type TimerDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the TimerDeleteResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (t *TimerDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusNoContent {
		return twapi.NewHTTPError(resp, "failed to delete timer")
	}
	return nil
}

// TimerDelete deletes a timer using the provided request and returns the
// response.
func TimerDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req TimerDeleteRequest,
) (*TimerDeleteResponse, error) {
	return twapi.Execute[TimerDeleteRequest, *TimerDeleteResponse](ctx, engine, req)
}

// TimerGetRequestPath contains the path parameters for loading a single timer.
type TimerGetRequestPath struct {
	// ID is the unique identifier of the timer to be retrieved.
	ID int64 `json:"id"`
}

// TimerGetRequest represents the request body for loading a single timer.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/get-projects-api-v3-timers-timer-id-json
type TimerGetRequest struct {
	// Path contains the path parameters for the request.
	Path TimerGetRequestPath
}

// NewTimerGetRequest creates a new TimerGetRequest with the provided
// timer ID. The ID is required to load a timer.
func NewTimerGetRequest(timerID int64) TimerGetRequest {
	return TimerGetRequest{
		Path: TimerGetRequestPath{
			ID: timerID,
		},
	}
}

// HTTPRequest creates an HTTP request for the TimerGetRequest.
func (t TimerGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/timers/" + strconv.FormatInt(t.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// TimerGetResponse contains all the information related to a timer.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/get-projects-api-v3-timers-timer-id-json
type TimerGetResponse struct {
	Timer Timer `json:"timer"`
}

// HandleHTTPResponse handles the HTTP response for the TimerGetResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (t *TimerGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve timer")
	}

	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode retrieve timer response: %w", err)
	}
	return nil
}

// TimerGet retrieves a single timer using the provided request and returns the
// response.
func TimerGet(
	ctx context.Context,
	engine *twapi.Engine,
	req TimerGetRequest,
) (*TimerGetResponse, error) {
	return twapi.Execute[TimerGetRequest, *TimerGetResponse](ctx, engine, req)
}

// TimerListRequestFilters contains the filters for loading multiple
// timers.
type TimerListRequestFilters struct {
	// UserID is the unique identifier of the user whose timers are to be
	// retrieved.
	UserID int64

	// ProjectID is the unique identifier of the project whose timers are to be
	// retrieved.
	ProjectID int64

	// TaskID is the unique identifier of the task whose timers are to be
	// retrieved.
	TaskID int64

	// RunningTimersOnly is a flag to indicate whether to retrieve only running
	// timers.
	RunningTimersOnly bool

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of timers to retrieve per page. Defaults to 50.
	PageSize int64
}

// TimerListRequest represents the request body for loading multiple timers.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/get-projects-api-v3-timers-json
type TimerListRequest struct {
	// Filters contains the filters for loading multiple timers.
	Filters TimerListRequestFilters
}

// NewTimerListRequest creates a new TimerListRequest with default values.
func NewTimerListRequest() TimerListRequest {
	return TimerListRequest{
		Filters: TimerListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the TimerListRequest.
func (t TimerListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/timers.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	if t.Filters.UserID > 0 {
		query.Set("userId", strconv.FormatInt(t.Filters.UserID, 10))
	}
	if t.Filters.ProjectID > 0 {
		query.Set("projectId", strconv.FormatInt(t.Filters.ProjectID, 10))
	}
	if t.Filters.TaskID > 0 {
		query.Set("taskId", strconv.FormatInt(t.Filters.TaskID, 10))
	}
	if t.Filters.RunningTimersOnly {
		query.Set("runningTimersOnly", strconv.FormatBool(t.Filters.RunningTimersOnly))
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

// TimerListResponse contains information by multiple timers matching the
// request filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/time-tracking/get-projects-api-v3-timers-json
type TimerListResponse struct {
	request TimerListRequest

	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`
	Timers []Timer `json:"timers"`
}

// HandleHTTPResponse handles the HTTP response for the TimerListResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (u *TimerListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list timers")
	}

	if err := json.NewDecoder(resp.Body).Decode(u); err != nil {
		return fmt.Errorf("failed to decode list timers response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (u *TimerListResponse) SetRequest(req TimerListRequest) {
	u.request = req
}

// Iterate returns the request set to the next page, if available. If there are
// no more pages, a nil request is returned.
func (u *TimerListResponse) Iterate() *TimerListRequest {
	if !u.Meta.Page.HasMore {
		return nil
	}
	req := u.request
	req.Filters.Page++
	return &req
}

// TimerList retrieves multiple timers using the provided request and returns
// the response.
func TimerList(
	ctx context.Context,
	engine *twapi.Engine,
	req TimerListRequest,
) (*TimerListResponse, error) {
	return twapi.Execute[TimerListRequest, *TimerListResponse](ctx, engine, req)
}
