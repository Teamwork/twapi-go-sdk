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
	_ twapi.HTTPRequester = (*JobRoleCreateRequest)(nil)
	_ twapi.HTTPResponser = (*JobRoleCreateResponse)(nil)
	_ twapi.HTTPRequester = (*JobRoleUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*JobRoleUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*JobRoleDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*JobRoleDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*JobRoleGetRequest)(nil)
	_ twapi.HTTPResponser = (*JobRoleGetResponse)(nil)
	_ twapi.HTTPRequester = (*JobRoleListRequest)(nil)
	_ twapi.HTTPResponser = (*JobRoleListResponse)(nil)
)

// JobRole defines a user’s primary function or position within the
// organization, such as developer, designer, project manager, or account
// manager. It provides high-level context about what a person is generally
// responsible for, helping teams understand who does what across projects and
// departments. Job roles are commonly used in resource planning, capacity
// forecasting, and reporting, allowing managers to group work by role, plan
// future demand more accurately, and ensure the right mix of roles is available
// to deliver projects efficiently.
//
// More information can be found at:
// https://support.teamwork.com/projects/people/roles
type JobRole struct {
	// ID is the unique identifier of the job role.
	ID int64 `json:"id"`

	// Name is the name of the job role.
	Name string `json:"name"`

	// IsActive indicates whether the job role is active.
	IsActive bool `json:"isActive"`

	// CreatedByUserID is the user who created this job role.
	CreatedByUserID int64 `json:"createdByUser"`

	// CreatedAt is the date and time when the job role was created.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedByUserID is the user who last updated this job role.
	UpdatedByUserID *int64 `json:"updatedByUser"`

	// UpdatedAt is the date and time when the job role was last updated.
	UpdatedAt *time.Time `json:"updatedAt"`

	// DeletedByUserID is the user who deleted this job role.
	DeletedByUserID *int64 `json:"deletedByUser"`

	// DeletedAt is the date and time when the job role was deleted.
	DeletedAt *time.Time `json:"deletedAt"`

	// Users contains the list of users associated with this job role.
	Users []twapi.Relationship `json:"users"`

	// PrimaryUsers contains the list of users who have this job role as their
	// primary role.
	PrimaryUsers []twapi.Relationship `json:"primaryUsers"`
}

// JobRoleCreateRequest represents the request body for creating a new
// job role.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/job-roles/post-projects-api-v3-jobroles-json
type JobRoleCreateRequest struct {
	// Name is the name of the job role.
	Name string `json:"name"`
}

// NewJobRoleCreateRequest reates a new JobRoleCreateRequest with the provided name.
func NewJobRoleCreateRequest(name string) JobRoleCreateRequest {
	return JobRoleCreateRequest{
		Name: name,
	}
}

// HTTPRequest creates an HTTP request for the JobRoleCreateRequest.
func (s JobRoleCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/jobroles.json"

	payload := struct {
		JobRole JobRoleCreateRequest `json:"jobRole"`
	}{JobRole: s}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create job role request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// JobRoleCreateResponse represents the response body for creating a new job role.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/job-roles/post-projects-api-v3-jobroles-json
type JobRoleCreateResponse struct {
	// JobRole contains the created job role information.
	JobRole JobRole `json:"jobRole"`
}

// HandleHTTPResponse handles the HTTP response for the JobRoleCreateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (s *JobRoleCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create job role")
	}
	if err := json.NewDecoder(resp.Body).Decode(s); err != nil {
		return fmt.Errorf("failed to decode create job role response: %w", err)
	}
	if s.JobRole.ID == 0 {
		return fmt.Errorf("create job role response does not contain a valid identifier")
	}
	return nil
}

// JobRoleCreate creates a new job role using the provided request and returns
// the response.
func JobRoleCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req JobRoleCreateRequest,
) (*JobRoleCreateResponse, error) {
	return twapi.Execute[JobRoleCreateRequest, *JobRoleCreateResponse](ctx, engine, req)
}

// JobRoleUpdateRequestPath contains the path parameters for updating a job
// role.
type JobRoleUpdateRequestPath struct {
	// ID is the unique identifier of the job role to be updated.
	ID int64
}

// JobRoleUpdateRequest represents the request body for updating a job role.
// Besides the identifier, all other fields are optional. When a field is not
// provided, it will not be modified.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/job-roles/patch-projects-api-v3-jobroles-id-json
type JobRoleUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path JobRoleUpdateRequestPath `json:"-"`

	// Name is the name of the job role.
	Name *string `json:"name,omitempty"`
}

// NewJobRoleUpdateRequest creates a new JobRoleUpdateRequest with the provided
// job role ID. The ID is required to update a job role.
func NewJobRoleUpdateRequest(jobRoleID int64) JobRoleUpdateRequest {
	return JobRoleUpdateRequest{
		Path: JobRoleUpdateRequestPath{
			ID: jobRoleID,
		},
	}
}

// HTTPRequest creates an HTTP request for the JobRoleUpdateRequest.
func (s JobRoleUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/jobroles/" + strconv.FormatInt(s.Path.ID, 10) + ".json"

	payload := struct {
		JobRole JobRoleUpdateRequest `json:"jobRole"`
	}{JobRole: s}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update job role request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// JobRoleUpdateResponse represents the response body for updating a job role.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/job-roles/patch-projects-api-v3-jobroles-id-json
type JobRoleUpdateResponse struct {
	// JobRole contains the updated job role information.
	JobRole JobRole `json:"jobRole"`
}

// HandleHTTPResponse handles the HTTP response for the JobRoleUpdateResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (s *JobRoleUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update job role")
	}
	if err := json.NewDecoder(resp.Body).Decode(s); err != nil {
		return fmt.Errorf("failed to decode update job role response: %w", err)
	}
	return nil
}

// JobRoleUpdate updates a job role using the provided request and returns the
// response.
func JobRoleUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req JobRoleUpdateRequest,
) (*JobRoleUpdateResponse, error) {
	return twapi.Execute[JobRoleUpdateRequest, *JobRoleUpdateResponse](ctx, engine, req)
}

// JobRoleDeleteRequestPath contains the path parameters for deleting a job
// role.
type JobRoleDeleteRequestPath struct {
	// ID is the unique identifier of the job role to be deleted.
	ID int64
}

// JobRoleDeleteRequest represents the request body for deleting a job role.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/job-roles/delete-projects-api-v3-jobroles-id-json
type JobRoleDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path JobRoleDeleteRequestPath
}

// NewJobRoleDeleteRequest creates a new JobRoleDeleteRequest with the provided
// job role ID.
func NewJobRoleDeleteRequest(jobRoleID int64) JobRoleDeleteRequest {
	return JobRoleDeleteRequest{
		Path: JobRoleDeleteRequestPath{
			ID: jobRoleID,
		},
	}
}

// HTTPRequest creates an HTTP request for the JobRoleDeleteRequest.
func (u JobRoleDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/jobroles/" + strconv.FormatInt(u.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// JobRoleDeleteResponse represents the response body for deleting a job role.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/job-roles/delete-projects-api-v3-jobroles-id-json
type JobRoleDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the JobRoleDeleteResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (u *JobRoleDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusNoContent {
		return twapi.NewHTTPError(resp, "failed to delete job role")
	}
	return nil
}

// JobRoleDelete deletes a job role using the provided request and returns the
// response.
func JobRoleDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req JobRoleDeleteRequest,
) (*JobRoleDeleteResponse, error) {
	return twapi.Execute[JobRoleDeleteRequest, *JobRoleDeleteResponse](ctx, engine, req)
}

// JobRoleGetRequestPath contains the path parameters for loading a single job
// role.
type JobRoleGetRequestPath struct {
	// ID is the unique identifier of the job role to be retrieved.
	ID int64 `json:"id"`
}

// JobRoleGetRequest represents the request body for loading a single job role.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/job-roles/get-projects-api-v3-jobroles-job-role-id-json
type JobRoleGetRequest struct {
	// Path contains the path parameters for the request.
	Path JobRoleGetRequestPath
}

// NewJobRoleGetRequest creates a new JobRoleGetRequest with the provided job
// role ID. The ID is required to load a job role.
func NewJobRoleGetRequest(jobRoleID int64) JobRoleGetRequest {
	return JobRoleGetRequest{
		Path: JobRoleGetRequestPath{
			ID: jobRoleID,
		},
	}
}

// HTTPRequest creates an HTTP request for the JobRoleGetRequest.
func (s JobRoleGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/jobroles/" + strconv.FormatInt(s.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// JobRoleGetResponse contains all the information related to a job role.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/job-roles/get-projects-api-v3-jobroles-job-role-id-json
type JobRoleGetResponse struct {
	JobRole JobRole `json:"jobRole"`
}

// HandleHTTPResponse handles the HTTP response for the JobRoleGetResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (s *JobRoleGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve job role")
	}

	if err := json.NewDecoder(resp.Body).Decode(s); err != nil {
		return fmt.Errorf("failed to decode retrieve job role response: %w", err)
	}
	return nil
}

// JobRoleGet retrieves a single job role using the provided request and returns
// the response.
func JobRoleGet(
	ctx context.Context,
	engine *twapi.Engine,
	req JobRoleGetRequest,
) (*JobRoleGetResponse, error) {
	return twapi.Execute[JobRoleGetRequest, *JobRoleGetResponse](ctx, engine, req)
}

// JobRoleListRequestPath contains the path parameters for loading multiple
// job roles.
type JobRoleListRequestPath struct{}

// JobRoleListRequestSideload contains the possible sideload options when
// loading multiple job roles.
type JobRoleListRequestSideload string

// List of possible sideload options for JobRoleListRequestSideload.
const (
	JobRoleListRequestSideloadUsers      JobRoleListRequestSideload = "users"
	JobRoleListRequestSideloadCurrencies JobRoleListRequestSideload = "currencies"
)

// JobRoleListRequestFilters contains the filters for loading multiple job
// roles.
type JobRoleListRequestFilters struct {
	// SearchTerm is an optional search term to filter job roles by name or
	// assigned users' names.
	SearchTerm string

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of job roles to retrieve per page. Defaults to 50.
	PageSize int64

	// Include contains additional related information to include in the response
	// as a sideload.
	Include []JobRoleListRequestSideload
}

// JobRoleListRequest represents the request body for loading multiple job
// roles.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/job-roles/get-projects-api-v3-jobroles-json
type JobRoleListRequest struct {
	// Path contains the path parameters for the request.
	Path JobRoleListRequestPath

	// Filters contains the filters for loading multiple job roles.
	Filters JobRoleListRequestFilters
}

// NewJobRoleListRequest creates a new JobRoleListRequest with default values.
func NewJobRoleListRequest() JobRoleListRequest {
	return JobRoleListRequest{
		Filters: JobRoleListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the JobRoleListRequest.
func (s JobRoleListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/jobroles.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	if s.Filters.SearchTerm != "" {
		query.Set("searchTerm", s.Filters.SearchTerm)
	}
	if s.Filters.Page > 0 {
		query.Set("page", strconv.FormatInt(s.Filters.Page, 10))
	}
	if s.Filters.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(s.Filters.PageSize, 10))
	}
	if len(s.Filters.Include) > 0 {
		var include []string
		for _, sideload := range s.Filters.Include {
			include = append(include, string(sideload))
		}
		query.Set("include", strings.Join(include, ","))
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// JobRoleListResponse contains information by multiple job roles matching the
// request filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/job-roles/get-projects-api-v3-jobroles-json
type JobRoleListResponse struct {
	request JobRoleListRequest

	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`
	JobRoles []JobRole `json:"jobRoles"`
}

// HandleHTTPResponse handles the HTTP response for the JobRoleListResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (s *JobRoleListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list jobRoles")
	}

	if err := json.NewDecoder(resp.Body).Decode(s); err != nil {
		return fmt.Errorf("failed to decode list job roles response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (s *JobRoleListResponse) SetRequest(req JobRoleListRequest) {
	s.request = req
}

// Iterate returns the request set to the next page, if available. If there are
// no more pages, a nil request is returned.
func (s *JobRoleListResponse) Iterate() *JobRoleListRequest {
	if !s.Meta.Page.HasMore {
		return nil
	}
	req := s.request
	req.Filters.Page++
	return &req
}

// JobRoleList retrieves multiple job roles using the provided request and
// returns the response.
func JobRoleList(
	ctx context.Context,
	engine *twapi.Engine,
	req JobRoleListRequest,
) (*JobRoleListResponse, error) {
	return twapi.Execute[JobRoleListRequest, *JobRoleListResponse](ctx, engine, req)
}
