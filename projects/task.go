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
	_ twapi.HTTPRequester = (*TaskCreateRequest)(nil)
	_ twapi.HTTPResponser = (*TaskCreateResponse)(nil)
	_ twapi.HTTPRequester = (*TaskUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*TaskUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*TaskDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*TaskDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*TaskGetRequest)(nil)
	_ twapi.HTTPResponser = (*TaskGetResponse)(nil)
	_ twapi.HTTPRequester = (*TaskListRequest)(nil)
	_ twapi.HTTPResponser = (*TaskListResponse)(nil)
)

// Task represents an individual unit of work assigned to one or more team
// members within a project. Each task can include details such as a title,
// description, priority, estimated time, assignees, and due date, along with
// the ability to attach files, leave comments, track time, and set dependencies
// on other tasks. Tasks are organized within task lists, helping structure and
// sequence work logically. They serve as the building blocks of project
// management in Teamwork, allowing teams to collaborate, monitor progress, and
// ensure accountability throughout the project's lifecycle.
//
// More information can be found at:
// https://support.teamwork.com/projects/getting-started/tasks-overview
type Task struct {
	// ID is the unique identifier of the task.
	ID int64 `json:"id"`

	// Name is the name of the task.
	Name string `json:"name"`

	// Description is the description of the task.
	Description *string `json:"description"`

	// DescriptionContentType is the content type of the description. It can be
	// "TEXT" or "HTML".
	DescriptionContentType *string `json:"descriptionContentType"`

	// Priority is the priority of the task. It can be "none", "low", "medium" or
	// "high".
	Priority *string `json:"priority"`

	// Progress is the progress of the task, in percentage (0-100).
	Progress int64 `json:"progress"`

	// StartAt is the date and time when the task is scheduled to start.
	StartAt *time.Time `json:"startDate"`

	// DueAt is the date and time when the task is scheduled to be completed.
	DueAt *time.Time `json:"dueDate"`

	// EstimatedMinutes is the estimated time to complete the task, in minutes.
	EstimatedMinutes int64 `json:"estimateMinutes"`

	// Tasklist is the relationship to the tasklist containing this task.
	Tasklist twapi.Relationship `json:"tasklist"`

	// Assignees is the list of users, teams or clients/companies assigned to this
	// task.
	Assignees []twapi.Relationship `json:"assignees"`

	// Tags is the list of tags associated with this task.
	Tags []twapi.Relationship `json:"tags"`

	// CreatedBy is the ID of the user who created the task.
	CreatedBy *int64 `json:"createdBy"`

	// CreatedAt is the date and time when the task was created.
	CreatedAt *time.Time `json:"createdAt"`

	// UpdatedBy is the ID of the user who last updated the task.
	UpdatedBy *int64 `json:"updatedBy"`

	// UpdatedAt is the date and time when the task was last updated.
	UpdatedAt time.Time `json:"updatedAt"`

	// DeletedBy is the ID of the user who deleted the task, if it was deleted.
	DeletedBy *int64 `json:"deletedBy"`

	// DeletedAt is the date and time when the task was deleted, if it was
	// deleted.
	DeletedAt *time.Time `json:"deletedAt"`

	// CompletedBy is the ID of the user who completed the task, if it was
	// completed.
	CompletedBy *int64 `json:"completedBy,omitempty"`

	// CompletedAt is the date and time when the task was completed, if it was
	// completed.
	CompletedAt *time.Time `json:"completedAt,omitempty"`

	// Status is the status of the task. It can be "new", "reopened", "completed"
	// or "deleted".
	Status string `json:"status"`
}

// TaskUpdateRequestPath contains the path parameters for creating a
// task.
type TaskCreateRequestPath struct {
	// TasklistID is the unique identifier of the tasklist that will contain the
	// task.
	TasklistID int64
}

// TaskCreateRequest represents the request body for creating a new
// task.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/tasks/post-projects-api-v3-tasklists-tasklist-id-tasks-json
type TaskCreateRequest struct {
	// Path contains the path parameters for the request.
	Path TaskCreateRequestPath `json:"-"`

	// Name is the name of the task
	Name string `json:"name"`

	// Description is an optional description of the task.
	Description *string `json:"description,omitempty"`

	// Priority is the priority of the task. It can be "none", "low", "medium" or
	// "high".
	Priority *string `json:"priority,omitempty"`

	// Progress is the progress of the task, in percentage (0-100).
	Progress *int64 `json:"progress,omitempty"`

	// StartAt is the date and time when the task is scheduled to start.
	StartAt *twapi.Date `json:"startAt,omitempty"`

	// DueAt is the date and time when the task is scheduled to be completed.
	DueAt *twapi.Date `json:"dueAt,omitempty"`

	// EstimatedMinutes is the estimated time to complete the task, in minutes.
	EstimatedMinutes *int64 `json:"estimatedMinutes,omitempty"`

	// Assignees is the list of users, teams or clients/companies assigned to this
	// task.
	Assignees *UserGroups `json:"assignees,omitempty"`

	// TagIDs is the list of tag IDs associated with this task.
	TagIDs []int64 `json:"tagIds,omitempty"`
}

// NewTaskCreateRequest creates a new TaskCreateRequest with the provided name
// in a specific tasklist.
func NewTaskCreateRequest(tasklistID int64, name string) TaskCreateRequest {
	return TaskCreateRequest{
		Path: TaskCreateRequestPath{
			TasklistID: tasklistID,
		},
		Name: name,
	}
}

// HTTPRequest creates an HTTP request for the TaskCreateRequest.
func (t TaskCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/tasklists/%d/tasks.json", server, t.Path.TasklistID)

	payload := struct {
		Task TaskCreateRequest `json:"task"`
	}{Task: t}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create task request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// TaskCreateResponse represents the response body for creating a new task.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/tasks/post-projects-api-v3-tasklists-tasklist-id-tasks-json
type TaskCreateResponse struct {
	// Task is the created task.
	Task Task `json:"task"`
}

// HandleHTTPResponse handles the HTTP response for the TaskCreateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (t *TaskCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create task")
	}
	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode create task response: %w", err)
	}
	if t.Task.ID == 0 {
		return fmt.Errorf("create task response does not contain a valid identifier")
	}
	return nil
}

// TaskCreate creates a new task using the provided request and returns the
// response.
func TaskCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req TaskCreateRequest,
) (*TaskCreateResponse, error) {
	return twapi.Execute[TaskCreateRequest, *TaskCreateResponse](ctx, engine, req)
}

// TaskUpdateRequestPath contains the path parameters for updating a task.
type TaskUpdateRequestPath struct {
	// ID is the unique identifier of the task to be updated.
	ID int64
}

// TaskUpdateRequest represents the request body for updating a task.
// Besides the identifier, all other fields are optional. When a field is not
// provided, it will not be modified.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/tasks/patch-projects-api-v3-tasks-task-id-json
type TaskUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path TaskUpdateRequestPath `json:"-"`

	// Name is the name of the task
	Name *string `json:"name,omitempty"`

	// Description is an optional description of the task.
	Description *string `json:"description,omitempty"`

	// Priority is the priority of the task. It can be "none", "low", "medium" or
	// "high".
	Priority *string `json:"priority,omitempty"`

	// Progress is the progress of the task, in percentage (0-100).
	Progress *int64 `json:"progress,omitempty"`

	// StartAt is the date and time when the task is scheduled to start.
	StartAt *twapi.Date `json:"startAt,omitempty"`

	// DueAt is the date and time when the task is scheduled to be completed.
	DueAt *twapi.Date `json:"dueAt,omitempty"`

	// EstimatedMinutes is the estimated time to complete the task, in minutes.
	EstimatedMinutes *int64 `json:"estimatedMinutes,omitempty"`

	// TasklistID is the identifier of the tasklist that will contain the task. If
	// provided, the task will be moved to this tasklist.
	TasklistID *int64 `json:"tasklistId,omitempty"`

	// Assignees is the list of users, teams or clients/companies assigned to this
	// task.
	Assignees *UserGroups `json:"assignees,omitempty"`

	// TagIDs is the list of tag IDs associated with this task.
	TagIDs []int64 `json:"tagIds,omitempty"`
}

// NewTaskUpdateRequest creates a new TaskUpdateRequest with the
// provided task ID. The ID is required to update a task.
func NewTaskUpdateRequest(taskID int64) TaskUpdateRequest {
	return TaskUpdateRequest{
		Path: TaskUpdateRequestPath{
			ID: taskID,
		},
	}
}

// HTTPRequest creates an HTTP request for the TaskUpdateRequest.
func (t TaskUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/tasks/" + strconv.FormatInt(t.Path.ID, 10) + ".json"

	payload := struct {
		Task TaskUpdateRequest `json:"task"`
	}{Task: t}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update task request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// TaskUpdateResponse represents the response body for updating a task.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/tasks/patch-projects-api-v3-tasks-task-id-json
type TaskUpdateResponse struct {
	// Task is the updated task.
	Task Task `json:"task"`
}

// HandleHTTPResponse handles the HTTP response for the TaskUpdateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (t *TaskUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update task")
	}
	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode update task response: %w", err)
	}
	return nil
}

// TaskUpdate updates a task using the provided request and returns the
// response.
func TaskUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req TaskUpdateRequest,
) (*TaskUpdateResponse, error) {
	return twapi.Execute[TaskUpdateRequest, *TaskUpdateResponse](ctx, engine, req)
}

// TaskDeleteRequestPath contains the path parameters for deleting a task.
type TaskDeleteRequestPath struct {
	// ID is the unique identifier of the task to be deleted.
	ID int64
}

// TaskDeleteRequest represents the request body for deleting a task.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/tasks/delete-projects-api-v3-tasks-task-id-json
type TaskDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path TaskDeleteRequestPath
}

// NewTaskDeleteRequest creates a new TaskDeleteRequest with the
// provided task ID.
func NewTaskDeleteRequest(taskID int64) TaskDeleteRequest {
	return TaskDeleteRequest{
		Path: TaskDeleteRequestPath{
			ID: taskID,
		},
	}
}

// HTTPRequest creates an HTTP request for the TaskDeleteRequest.
func (t TaskDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/tasks/" + strconv.FormatInt(t.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// TaskDeleteResponse represents the response body for deleting a task.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/tasks/delete-projects-api-v3-tasks-task-id-json
type TaskDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the TaskDeleteResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (t *TaskDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to delete task")
	}
	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode delete task response: %w", err)
	}
	return nil
}

// TaskDelete deletes a task using the provided request and returns the
// response.
func TaskDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req TaskDeleteRequest,
) (*TaskDeleteResponse, error) {
	return twapi.Execute[TaskDeleteRequest, *TaskDeleteResponse](ctx, engine, req)
}

// TaskGetRequestPath contains the path parameters for loading a single task.
type TaskGetRequestPath struct {
	// ID is the unique identifier of the task to be retrieved.
	ID int64 `json:"id"`
}

// TaskGetRequest represents the request body for loading a single task.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/tasks/get-projects-api-v3-tasks-task-id-json
type TaskGetRequest struct {
	// Path contains the path parameters for the request.
	Path TaskGetRequestPath
}

// NewTaskGetRequest creates a new TaskGetRequest with the provided
// task ID. The ID is required to load a task.
func NewTaskGetRequest(taskID int64) TaskGetRequest {
	return TaskGetRequest{
		Path: TaskGetRequestPath{
			ID: taskID,
		},
	}
}

// HTTPRequest creates an HTTP request for the TaskGetRequest.
func (t TaskGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/tasks/" + strconv.FormatInt(t.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// TaskGetResponse contains all the information related to a task.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/tasks/get-projects-api-v3-tasks-task-id-json
type TaskGetResponse struct {
	Task Task `json:"task"`
}

// HandleHTTPResponse handles the HTTP response for the TaskGetResponse. If some
// unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (t *TaskGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve task")
	}

	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode retrieve task response: %w", err)
	}
	return nil
}

// TaskGet retrieves a single task using the provided request and returns the
// response.
func TaskGet(
	ctx context.Context,
	engine *twapi.Engine,
	req TaskGetRequest,
) (*TaskGetResponse, error) {
	return twapi.Execute[TaskGetRequest, *TaskGetResponse](ctx, engine, req)
}

// TaskListRequestPath contains the path parameters for loading multiple tasks.
type TaskListRequestPath struct {
	// ProjectID is the unique identifier of the project whose tasks are to be
	// retrieved.
	ProjectID int64
	// TasklistID is the unique identifier of the tasklist whose tasks are to be
	// retrieved. If provided, the ProjectID is ignored.
	TasklistID int64
}

// TaskListRequestFilters contains the filters for loading multiple tasks.
type TaskListRequestFilters struct {
	// SearchTerm is an optional search term to filter tasks by name, description
	// or tasklist's name.
	SearchTerm string

	// TagIDs is an optional list of tag IDs to filter tasks by tags.
	TagIDs []int64

	// MatchAllTags is an optional flag to indicate if all tags must match. If set
	// to true, only tasks matching all specified tags will be returned.
	MatchAllTags *bool

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of tasks to retrieve per page. Defaults to 50.
	PageSize int64
}

// TaskListRequest represents the request body for loading multiple tasks.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/tasks/get-projects-api-v3-tasks-json
// https://apidocs.teamwork.com/docs/teamwork/v3/tasks/get-projects-api-v3-projects-project-id-tasks-json
// https://apidocs.teamwork.com/docs/teamwork/v3/tasks/get-projects-api-v3-tasklists-tasklist-id-tasks-json
type TaskListRequest struct {
	// Path contains the path parameters for the request.
	Path TaskListRequestPath

	// Filters contains the filters for loading multiple tasks.
	Filters TaskListRequestFilters
}

// NewTaskListRequest creates a new TaskListRequest with default values.
func NewTaskListRequest() TaskListRequest {
	return TaskListRequest{
		Filters: TaskListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the TaskListRequest.
func (t TaskListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	var uri string
	switch {
	case t.Path.TasklistID > 0:
		uri = fmt.Sprintf("%s/projects/api/v3/tasklists/%d/tasks.json", server, t.Path.TasklistID)
	case t.Path.ProjectID > 0:
		uri = fmt.Sprintf("%s/projects/api/v3/projects/%d/tasks.json", server, t.Path.ProjectID)
	default:
		uri = server + "/projects/api/v3/tasks.json"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	if t.Filters.SearchTerm != "" {
		query.Set("searchTerm", t.Filters.SearchTerm)
	}
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

// TaskListResponse contains information by multiple tasks matching the request
// filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/tasks/get-projects-api-v3-tasks-json
// https://apidocs.teamwork.com/docs/teamwork/v3/tasks/get-projects-api-v3-projects-project-id-tasks-json
// https://apidocs.teamwork.com/docs/teamwork/v3/tasks/get-projects-api-v3-tasklists-tasklist-id-tasks-json
type TaskListResponse struct {
	request TaskListRequest

	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`
	Tasks []Task `json:"tasks"`
}

// HandleHTTPResponse handles the HTTP response for the TaskListResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (t *TaskListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list tasks")
	}

	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode list tasks response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (t *TaskListResponse) SetRequest(req TaskListRequest) {
	t.request = req
}

// Iterate returns the request set to the next page, if available. If there
// are no more pages, a nil request is returned.
func (t *TaskListResponse) Iterate() *TaskListRequest {
	if !t.Meta.Page.HasMore {
		return nil
	}
	req := t.request
	req.Filters.Page++
	return &req
}

// TaskList retrieves multiple tasks using the provided request and
// returns the response.
func TaskList(
	ctx context.Context,
	engine *twapi.Engine,
	req TaskListRequest,
) (*TaskListResponse, error) {
	return twapi.Execute[TaskListRequest, *TaskListResponse](ctx, engine, req)
}
