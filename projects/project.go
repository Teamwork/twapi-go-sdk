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
	_ twapi.HTTPRequester = (*ProjectCreateRequest)(nil)
	_ twapi.HTTPResponser = (*ProjectCreateResponse)(nil)
	_ twapi.HTTPRequester = (*ProjectUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*ProjectUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*ProjectDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*ProjectDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*ProjectGetRequest)(nil)
	_ twapi.HTTPResponser = (*ProjectGetResponse)(nil)
	_ twapi.HTTPRequester = (*ProjectListRequest)(nil)
	_ twapi.HTTPResponser = (*ProjectListResponse)(nil)
)

// Project serves as the central workspace for organizing and managing a
// specific piece of work or initiative. Each project provides a dedicated area
// where teams can plan tasks, assign responsibilities, set deadlines, and track
// progress toward shared goals. Projects include tools for communication, file
// sharing, milestones, and time tracking, allowing teams to stay aligned and
// informed throughout the entire lifecycle of the work. Whether it's a product
// launch, client engagement, or internal initiative, projects in Teamwork.com
// help teams structure their efforts, collaborate more effectively, and deliver
// results with greater visibility and accountability.
//
// More information can be found at:
// https://support.teamwork.com/projects/getting-started/projects-overview
type Project struct {
	// ID is the unique identifier of the project.
	ID int64 `json:"id"`

	// Description is an optional description of the project.
	Description *string `json:"description"`

	// Name is the name of the project.
	Name string `json:"name"`

	// StartAt is the start date of the project.
	StartAt *time.Time `json:"startAt"`

	// EndAt is the end date of the project.
	EndAt *time.Time `json:"endAt"`

	// Company is the company associated with the project.
	Company twapi.Relationship `json:"company"`

	// Owner is the user who owns the project.
	Owner *twapi.Relationship `json:"projectOwner"`

	// Tags is a list of tags associated with the project.
	Tags []twapi.Relationship `json:"tags"`

	// CreatedAt is the date and time when the project was created.
	CreatedAt *time.Time `json:"createdAt"`

	// CreatedBy is the ID of the user who created the project.
	CreatedBy *int64 `json:"createdBy"`

	// UpdatedAt is the date and time when the project was last updated.
	UpdatedAt *time.Time `json:"updatedAt"`

	// UpdatedBy is the ID of the user who last updated the project.
	UpdatedBy *int64 `json:"updatedBy"`

	// CompletedAt is the date and time when the project was completed.
	CompletedAt *time.Time `json:"completedAt"`

	// CompletedBy is the ID of the user who completed the project.
	CompletedBy *int64 `json:"completedBy"`

	// Status is the status of the project. It can be "active", "inactive"
	// (archived) or "deleted".
	Status string `json:"status"`

	// Type is the type of the project. It can be "normal", "tasklists-template",
	// "projects-template", "personal", "holder-project", "tentative" or
	// "global-messages".
	Type string `json:"type"`
}

// ProjectCreateRequest represents the request body for creating a new project.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/projects/post-projects-json
type ProjectCreateRequest struct {
	// Name is the name of the project.
	Name string `json:"name"`

	// Description is an optional description of the project.
	Description *string `json:"description,omitempty"`

	// StartAt is an optional start date for the project. By default it doesn't
	// have a start date.
	StartAt *LegacyDate `json:"start-date,omitempty"`

	// EndAt is an optional end date for the project. By default it doesn't have
	// an end date.
	EndAt *LegacyDate `json:"end-date,omitempty"`

	// CompanyID is an optional ID of the company/client associated with the
	// project. By default it is the ID of the company of the logged user
	// creating the project.
	CompanyID int64 `json:"companyId"`

	// OwnerID is an optional ID of the user who owns the project. By default it
	// is the ID of the logged user creating the project.
	OwnerID *int64 `json:"projectOwnerId,omitempty"`

	// TagIDs is an optional list of tag IDs associated with the project.
	TagIDs []int64 `json:"tagIds,omitempty"`
}

// NewProjectCreateRequest creates a new ProjectCreateRequest with the
// provided name. The name is required to create a new project.
func NewProjectCreateRequest(name string) ProjectCreateRequest {
	return ProjectCreateRequest{Name: name}
}

// HTTPRequest creates an HTTP request for the ProjectCreateRequest.
func (p ProjectCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects.json"

	payload := struct {
		Project ProjectCreateRequest `json:"project"`
	}{Project: p}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create project request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// ProjectCreateResponse represents the response body for creating a new
// project.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/projects/post-projects-json
type ProjectCreateResponse struct {
	// ID is the unique identifier of the created project.
	ID LegacyNumber `json:"id"`
}

// HandleHTTPResponse handles the HTTP response for the ProjectCreateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (p *ProjectCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create project")
	}
	if err := json.NewDecoder(resp.Body).Decode(p); err != nil {
		return fmt.Errorf("failed to decode create project response: %w", err)
	}
	if p.ID == 0 {
		return fmt.Errorf("create project response does not contain a valid identifier")
	}
	return nil
}

// ProjectCreate creates a new project using the provided request and returns
// the response.
func ProjectCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectCreateRequest,
) (*ProjectCreateResponse, error) {
	return twapi.Execute[ProjectCreateRequest, *ProjectCreateResponse](ctx, engine, req)
}

// ProjectUpdateRequestPath contains the path parameters for updating a project.
type ProjectUpdateRequestPath struct {
	// ID is the unique identifier of the project to be updated.
	ID int64
}

// ProjectUpdateRequest represents the request body for updating a project.
// Besides the identifier, all other fields are optional. When a field is not
// provided, it will not be modified.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/projects/put-projects-id-json
type ProjectUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path ProjectUpdateRequestPath

	// Name is the name of the project.
	Name *string `json:"name,omitempty"`

	// Description is the project description.
	Description *string `json:"description,omitempty"`

	// StartAt is the start date for the project.
	StartAt *LegacyDate `json:"start-date,omitempty"`

	// EndAt is the end date for the project.
	EndAt *LegacyDate `json:"end-date,omitempty"`

	// CompanyID is the company/client associated with the project.
	CompanyID *int64 `json:"companyId,omitempty"`

	// OwnerID is the ID of the user who owns the project.
	OwnerID *int64 `json:"projectOwnerId,omitempty"`

	// TagIDs is the list of tag IDs associated with the project.
	TagIDs []int64 `json:"tagIds,omitempty"`
}

// NewProjectUpdateRequest creates a new ProjectUpdateRequest with the
// provided project ID. The ID is required to update a project.
func NewProjectUpdateRequest(projectID int64) ProjectUpdateRequest {
	return ProjectUpdateRequest{
		Path: ProjectUpdateRequestPath{
			ID: projectID,
		},
	}
}

// HTTPRequest creates an HTTP request for the ProjectUpdateRequest.
func (p ProjectUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/" + strconv.FormatInt(p.Path.ID, 10) + ".json"

	payload := struct {
		Project ProjectUpdateRequest `json:"project"`
	}{Project: p}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update project request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// ProjectUpdateResponse represents the response body for updating a project.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/projects/put-projects-id-json
type ProjectUpdateResponse struct{}

// HandleHTTPResponse handles the HTTP response for the ProjectUpdateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (p *ProjectUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update project")
	}
	if err := json.NewDecoder(resp.Body).Decode(p); err != nil {
		return fmt.Errorf("failed to decode update project response: %w", err)
	}
	return nil
}

// ProjectUpdate updates a project using the provided request and returns the
// response.
func ProjectUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectUpdateRequest,
) (*ProjectUpdateResponse, error) {
	return twapi.Execute[ProjectUpdateRequest, *ProjectUpdateResponse](ctx, engine, req)
}

// ProjectDeleteRequestPath contains the path parameters for deleting a project.
type ProjectDeleteRequestPath struct {
	// ID is the unique identifier of the project to be deleted.
	ID int64
}

// ProjectDeleteRequest represents the request body for deleting a project.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/projects/delete-projects-id-json
type ProjectDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path ProjectDeleteRequestPath
}

// NewProjectDeleteRequest creates a new ProjectDeleteRequest with the
// provided project ID.
func NewProjectDeleteRequest(projectID int64) ProjectDeleteRequest {
	return ProjectDeleteRequest{
		Path: ProjectDeleteRequestPath{
			ID: projectID,
		},
	}
}

// HTTPRequest creates an HTTP request for the ProjectDeleteRequest.
func (p ProjectDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/" + strconv.FormatInt(p.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// ProjectDeleteResponse represents the response body for deleting a project.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/projects/delete-projects-id-json
type ProjectDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the ProjectDeleteResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (p *ProjectDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to delete project")
	}
	if err := json.NewDecoder(resp.Body).Decode(p); err != nil {
		return fmt.Errorf("failed to decode delete project response: %w", err)
	}
	return nil
}

// ProjectDelete deletes a project using the provided request and returns the
// response.
func ProjectDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectDeleteRequest,
) (*ProjectDeleteResponse, error) {
	return twapi.Execute[ProjectDeleteRequest, *ProjectDeleteResponse](ctx, engine, req)
}

// ProjectGetRequestPath contains the path parameters for loading a single
// project.
type ProjectGetRequestPath struct {
	// ID is the unique identifier of the project to be retrieved.
	ID int64 `json:"id"`
}

// ProjectGetRequest represents the request body for loading a single project.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/projects/get-projects-api-v3-projects-project-id-json
type ProjectGetRequest struct {
	// Path contains the path parameters for the request.
	Path ProjectGetRequestPath
}

// NewProjectGetRequest creates a new ProjectGetRequest with the provided
// project ID. The ID is required to load a project.
func NewProjectGetRequest(projectID int64) ProjectGetRequest {
	return ProjectGetRequest{
		Path: ProjectGetRequestPath{
			ID: projectID,
		},
	}
}

// HTTPRequest creates an HTTP request for the ProjectGetRequest.
func (p ProjectGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/projects/" + strconv.FormatInt(p.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// ProjectGetResponse contains all the information related to a project.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/projects/get-projects-api-v3-projects-project-id-json
type ProjectGetResponse struct {
	Project Project `json:"project"`
}

// HandleHTTPResponse handles the HTTP response for the ProjectGetResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (p *ProjectGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve project")
	}

	if err := json.NewDecoder(resp.Body).Decode(p); err != nil {
		return fmt.Errorf("failed to decode retrieve project response: %w", err)
	}
	return nil
}

// ProjectGet retrieves a single project using the provided request and returns
// the response.
func ProjectGet(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectGetRequest,
) (*ProjectGetResponse, error) {
	return twapi.Execute[ProjectGetRequest, *ProjectGetResponse](ctx, engine, req)
}

// ProjectListRequestFilters contains the filters for loading multiple projects.
type ProjectListRequestFilters struct {
	// SearchTerm is an optional search term to filter projects by name or
	// description.
	SearchTerm string

	// TagIDs is an optional list of tag IDs to filter projects by tags.
	TagIDs []int64

	// MatchAllTags is an optional flag to indicate if all tags must match. If
	// set to true, only projects matching all specified tags will be returned.
	MatchAllTags *bool

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of projects to retrieve per page. Defaults to 50.
	PageSize int64
}

// ProjectListRequest represents the request body for loading multiple projects.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/projects/get-projects-api-v3-projects-json
type ProjectListRequest struct {
	// Filters contains the filters for loading multiple projects.
	Filters ProjectListRequestFilters
}

// NewProjectListRequest creates a new ProjectListRequest with default values.
func NewProjectListRequest() ProjectListRequest {
	return ProjectListRequest{
		Filters: ProjectListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the ProjectListRequest.
func (p ProjectListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/projects.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	if p.Filters.SearchTerm != "" {
		query.Set("searchTerm", p.Filters.SearchTerm)
	}
	if len(p.Filters.TagIDs) > 0 {
		tagIDs := make([]string, len(p.Filters.TagIDs))
		for i, id := range p.Filters.TagIDs {
			tagIDs[i] = strconv.FormatInt(id, 10)
		}
		query.Set("projectTagIds", strings.Join(tagIDs, ","))
	}
	if p.Filters.MatchAllTags != nil {
		query.Set("matchAllProjectTags", strconv.FormatBool(*p.Filters.MatchAllTags))
	}
	if p.Filters.Page > 0 {
		query.Set("page", strconv.FormatInt(p.Filters.Page, 10))
	}
	if p.Filters.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(p.Filters.PageSize, 10))
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// ProjectListResponse contains information by multiple projects matching the
// request filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/projects/get-projects-api-v3-projects-json
type ProjectListResponse struct {
	request ProjectListRequest

	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`
	Projects []Project `json:"projects"`
}

// HandleHTTPResponse handles the HTTP response for the ProjectListResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (p *ProjectListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list projects")
	}

	if err := json.NewDecoder(resp.Body).Decode(p); err != nil {
		return fmt.Errorf("failed to decode list projects response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (p *ProjectListResponse) SetRequest(req ProjectListRequest) {
	p.request = req
}

// Iterate returns the request set to the next page, if available. If there
// are no more pages, a nil request is returned.
func (p *ProjectListResponse) Iterate() *ProjectListRequest {
	if !p.Meta.Page.HasMore {
		return nil
	}
	req := p.request
	req.Filters.Page++
	return &req
}

// ProjectList retrieves multiple projects using the provided request
// and returns the response.
func ProjectList(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectListRequest,
) (*ProjectListResponse, error) {
	return twapi.Execute[ProjectListRequest, *ProjectListResponse](ctx, engine, req)
}
