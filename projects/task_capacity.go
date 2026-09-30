package projects

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var (
	_ twapi.HTTPRequester = (*TaskCapacityListRequest)(nil)
	_ twapi.HTTPResponser = (*TaskCapacityListResponse)(nil)
	_ twapi.HTTPRequester = (*TaskCapacityCreateRequest)(nil)
	_ twapi.HTTPResponser = (*TaskCapacityCreateResponse)(nil)
	_ twapi.HTTPRequester = (*TaskCapacityReplaceRequest)(nil)
	_ twapi.HTTPResponser = (*TaskCapacityReplaceResponse)(nil)
	_ twapi.HTTPRequester = (*TaskCapacityDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*TaskCapacityDeleteResponse)(nil)
)

// TaskCapacity is one day of a user's custom split of a task's estimate, the
// "capacity split" the Workload planner edits.
//
// Without a split, Workload spreads a task's estimate evenly over the
// assignee's working days from start to due date, divided between the user
// assignees. A split replaces that for one user: each row is the time that user
// spends on the task on that day, a day without a row counts as zero, and the
// minutes are not divided between assignees. The rows need not add up to the
// estimate. Only a custom split is stored — a task on the even spread has no
// rows.
//
// The server maintains a split when the task changes: changing its start or
// due date moves the split with it (a longer range adds zero-minute days, a
// shorter one spreads the lost minutes over the days that remain), while
// changing its estimate or assignees, or clearing its dates, deletes every
// user's split on it.
//
// More information can be found at:
// https://support.teamwork.com/projects/workload/using-the-workload-planner
type TaskCapacity struct {
	// ID is the unique identifier of the row.
	ID int64 `json:"id"`

	// TaskID is the task the split belongs to.
	TaskID int64 `json:"taskId"`

	// UserID is the assignee the split applies to.
	UserID int64 `json:"userId"`

	// Date is the day this row covers.
	Date twapi.Date `json:"date"`

	// Minutes is the time the user spends on the task on Date.
	Minutes int64 `json:"minutes"`

	// Seconds is the same time in seconds. When positive it is what Workload
	// counts, as Seconds/60 minutes; zero means Minutes is used.
	Seconds int64 `json:"seconds"`
}

// TaskCapacityDate is one day of a split sent to the API.
type TaskCapacityDate struct {
	// Date is the day. It must fall within the task's start and due dates, or
	// be the due date itself when the task has no start date.
	Date twapi.Date `json:"date"`

	// Minutes is the time the user spends on the task on Date.
	Minutes int64 `json:"minutes"`

	// Seconds optionally gives the same time with second precision. When
	// positive it takes precedence over Minutes. It must not exceed 1,000,000.
	Seconds *int64 `json:"seconds,omitempty"`
}

// TaskCapacityListRequestFilters contains the filters for loading the rows of
// task splits.
type TaskCapacityListRequestFilters struct {
	// TaskID restricts the rows to one task.
	TaskID int64

	// UserIDs restricts the rows to these users. The endpoint defaults to the
	// authenticated user, so a task's rows for other assignees are only
	// returned when they are named here.
	UserIDs []int64

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of rows to retrieve per page. Defaults to 50.
	PageSize int64

	// CountMode selects whether the API computes the exact number of rows
	// matching the filters, reported in Meta.Page.Count. Defaults to
	// twapi.ListCountModeDefault, which leaves the decision to the API.
	CountMode twapi.ListCountMode
}

func (t TaskCapacityListRequestFilters) apply(req *http.Request) {
	query := req.URL.Query()
	if t.TaskID > 0 {
		query.Set("taskId", strconv.FormatInt(t.TaskID, 10))
	}
	if len(t.UserIDs) > 0 {
		userIDs := make([]string, len(t.UserIDs))
		for i, id := range t.UserIDs {
			userIDs[i] = strconv.FormatInt(id, 10)
		}
		query.Set("userIds", strings.Join(userIDs, ","))
	}
	if t.Page > 0 {
		query.Set("page", strconv.FormatInt(t.Page, 10))
	}
	if t.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(t.PageSize, 10))
	}
	t.CountMode.Apply(query)
	req.URL.RawQuery = query.Encode()
}

// TaskCapacityListRequest represents the request for loading the rows of task
// splits. The endpoint does not honour a fields[...] selection.
type TaskCapacityListRequest struct {
	// Filters contains the filters for loading the rows.
	Filters TaskCapacityListRequestFilters
}

// NewTaskCapacityListRequest creates a new TaskCapacityListRequest for the
// split of a task among the given users.
func NewTaskCapacityListRequest(taskID int64, userIDs ...int64) TaskCapacityListRequest {
	return TaskCapacityListRequest{
		Filters: TaskCapacityListRequestFilters{
			TaskID:   taskID,
			UserIDs:  userIDs,
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the TaskCapacityListRequest.
func (t TaskCapacityListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/tasks/capacity.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	t.Filters.apply(req)

	return req, nil
}

// TaskCapacityListResponse contains the rows of task splits matching the
// request filters.
type TaskCapacityListResponse struct {
	request TaskCapacityListRequest

	Meta       twapi.ListMeta `json:"meta"`
	Capacities []TaskCapacity `json:"capacities"`
}

// HandleHTTPResponse handles the HTTP response for the
// TaskCapacityListResponse. If some unexpected HTTP status code is returned by
// the API, a twapi.HTTPError is returned.
func (t *TaskCapacityListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list task capacities")
	}

	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode list task capacities response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (t *TaskCapacityListResponse) SetRequest(req TaskCapacityListRequest) {
	t.request = req
	t.Meta.ResolveCount(req.Filters.CountMode)
}

// Iterate returns the request set to the next page, if available. If there are
// no more pages, a nil request is returned.
func (t *TaskCapacityListResponse) Iterate() *TaskCapacityListRequest {
	if !t.Meta.Page.HasMore {
		return nil
	}
	req := t.request
	req.Filters.Page++
	return &req
}

// TaskCapacityList retrieves the rows of task splits using the provided request
// and returns the response.
func TaskCapacityList(
	ctx context.Context,
	engine *twapi.Engine,
	req TaskCapacityListRequest,
) (*TaskCapacityListResponse, error) {
	return twapi.Execute[TaskCapacityListRequest, *TaskCapacityListResponse](ctx, engine, req)
}

// TaskCapacityRequestPath contains the path parameters for writing a task's
// split.
type TaskCapacityRequestPath struct {
	// TaskID is the task whose split is written.
	TaskID int64
}

// TaskCapacityWrite is the split one user is given on a task. It is the body
// shared by TaskCapacityCreateRequest and TaskCapacityReplaceRequest.
type TaskCapacityWrite struct {
	// UserID is the assignee the split applies to. The user must be assigned
	// to the task, directly or through a team.
	UserID int64 `json:"userId"`

	// Dates are the days of the split, one per date. A day of the task's
	// range that is left out counts as zero.
	Dates []TaskCapacityDate `json:"dates"`
}

func encodeTaskCapacityWrite(
	ctx context.Context,
	method, server string,
	path TaskCapacityRequestPath,
	write TaskCapacityWrite,
) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/tasks/%d/capacity.json", server, path.TaskID)

	payload := struct {
		TaskCapacity TaskCapacityWrite `json:"taskCapacity"`
	}{TaskCapacity: write}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode task capacity request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// TaskCapacityCreateRequest represents the request for giving a user a split on
// a task that has none for them yet. A date the user already has a row for is
// answered with 400, so use TaskCapacityReplaceRequest to change an existing
// split.
//
// The task needs an estimate and a due date; a task with only a start date
// takes its milestone's due date, which the endpoint then saves on the task.
type TaskCapacityCreateRequest struct {
	// Path contains the path parameters for the request.
	Path TaskCapacityRequestPath `json:"-"`

	TaskCapacityWrite
}

// NewTaskCapacityCreateRequest creates a new TaskCapacityCreateRequest with the
// provided split.
func NewTaskCapacityCreateRequest(taskID, userID int64, dates []TaskCapacityDate) TaskCapacityCreateRequest {
	return TaskCapacityCreateRequest{
		Path: TaskCapacityRequestPath{TaskID: taskID},
		TaskCapacityWrite: TaskCapacityWrite{
			UserID: userID,
			Dates:  dates,
		},
	}
}

// HTTPRequest creates an HTTP request for the TaskCapacityCreateRequest.
func (t TaskCapacityCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	return encodeTaskCapacityWrite(ctx, http.MethodPost, server, t.Path, t.TaskCapacityWrite)
}

// TaskCapacityCreateResponse contains the rows stored by a create.
type TaskCapacityCreateResponse struct {
	// Capacities are the stored rows. It is empty when the split sent was the
	// even spread of the whole estimate over two or more days, which the
	// endpoint answers with 204 and does not store.
	Capacities []TaskCapacity `json:"capacities"`
}

// HandleHTTPResponse handles the HTTP response for the
// TaskCapacityCreateResponse. If some unexpected HTTP status code is returned
// by the API, a twapi.HTTPError is returned.
func (t *TaskCapacityCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusNoContent:
		return nil
	case http.StatusCreated:
	default:
		return twapi.NewHTTPError(resp, "failed to create task capacity")
	}

	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode create task capacity response: %w", err)
	}
	return nil
}

// TaskCapacityCreate gives a user a split on a task using the provided request
// and returns the response.
func TaskCapacityCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req TaskCapacityCreateRequest,
) (*TaskCapacityCreateResponse, error) {
	return twapi.Execute[TaskCapacityCreateRequest, *TaskCapacityCreateResponse](ctx, engine, req)
}

// TaskCapacityReplaceRequest represents the request for replacing a user's
// whole split on a task. The user must already have one, or the endpoint
// answers 404; use TaskCapacityCreateRequest for the first split. Sending the
// even spread of the whole estimate over two or more days removes the split,
// returning the task to the default spread.
type TaskCapacityReplaceRequest struct {
	// Path contains the path parameters for the request.
	Path TaskCapacityRequestPath `json:"-"`

	TaskCapacityWrite
}

// NewTaskCapacityReplaceRequest creates a new TaskCapacityReplaceRequest with
// the provided split.
func NewTaskCapacityReplaceRequest(taskID, userID int64, dates []TaskCapacityDate) TaskCapacityReplaceRequest {
	return TaskCapacityReplaceRequest{
		Path: TaskCapacityRequestPath{TaskID: taskID},
		TaskCapacityWrite: TaskCapacityWrite{
			UserID: userID,
			Dates:  dates,
		},
	}
}

// HTTPRequest creates an HTTP request for the TaskCapacityReplaceRequest.
func (t TaskCapacityReplaceRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	return encodeTaskCapacityWrite(ctx, http.MethodPut, server, t.Path, t.TaskCapacityWrite)
}

// TaskCapacityReplaceResponse contains the user's split after a replace.
type TaskCapacityReplaceResponse struct {
	// Capacities are the user's rows after the replace. It is empty when the
	// split was removed.
	Capacities []TaskCapacity `json:"capacities"`
}

// HandleHTTPResponse handles the HTTP response for the
// TaskCapacityReplaceResponse. If some unexpected HTTP status code is returned
// by the API, a twapi.HTTPError is returned.
func (t *TaskCapacityReplaceResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to replace task capacity")
	}

	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode replace task capacity response: %w", err)
	}
	return nil
}

// TaskCapacityReplace replaces a user's split on a task using the provided
// request and returns the response.
func TaskCapacityReplace(
	ctx context.Context,
	engine *twapi.Engine,
	req TaskCapacityReplaceRequest,
) (*TaskCapacityReplaceResponse, error) {
	return twapi.Execute[TaskCapacityReplaceRequest, *TaskCapacityReplaceResponse](ctx, engine, req)
}

// TaskCapacityDeleteRequest represents the request for removing splits from a
// task, returning it to the even spread for those users. Removing a split that
// does not exist succeeds.
type TaskCapacityDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path TaskCapacityRequestPath `json:"-"`

	// UserIDs are the users whose split is removed. An empty list removes
	// every user's split on the task.
	UserIDs []int64 `json:"userIds"`
}

// NewTaskCapacityDeleteRequest creates a new TaskCapacityDeleteRequest for the
// given users' splits on a task. Name no users to remove them all.
func NewTaskCapacityDeleteRequest(taskID int64, userIDs ...int64) TaskCapacityDeleteRequest {
	return TaskCapacityDeleteRequest{
		Path:    TaskCapacityRequestPath{TaskID: taskID},
		UserIDs: userIDs,
	}
}

// HTTPRequest creates an HTTP request for the TaskCapacityDeleteRequest.
func (t TaskCapacityDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/tasks/%d/capacity.json", server, t.Path.TaskID)

	// the endpoint requires a body, and an empty list rather than null reads
	// unambiguously as "every user"
	userIDs := t.UserIDs
	if userIDs == nil {
		userIDs = []int64{}
	}
	payload := struct {
		UserIDs []int64 `json:"userIds"`
	}{UserIDs: userIDs}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode delete task capacity request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// TaskCapacityDeleteResponse represents the response for removing splits.
type TaskCapacityDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the
// TaskCapacityDeleteResponse. If some unexpected HTTP status code is returned
// by the API, a twapi.HTTPError is returned.
func (t *TaskCapacityDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusNoContent {
		return twapi.NewHTTPError(resp, "failed to delete task capacity")
	}
	return nil
}

// TaskCapacityDelete removes splits from a task using the provided request and
// returns the response.
func TaskCapacityDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req TaskCapacityDeleteRequest,
) (*TaskCapacityDeleteResponse, error) {
	return twapi.Execute[TaskCapacityDeleteRequest, *TaskCapacityDeleteResponse](ctx, engine, req)
}
