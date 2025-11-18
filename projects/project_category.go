package projects

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var (
	_ twapi.HTTPRequester = (*ProjectCategoryCreateRequest)(nil)
	_ twapi.HTTPResponser = (*ProjectCategoryCreateResponse)(nil)
	_ twapi.HTTPRequester = (*ProjectCategoryUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*ProjectCategoryUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*ProjectCategoryDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*ProjectCategoryDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*ProjectCategoryGetRequest)(nil)
	_ twapi.HTTPResponser = (*ProjectCategoryGetResponse)(nil)
	_ twapi.HTTPRequester = (*ProjectCategoryListRequest)(nil)
	_ twapi.HTTPResponser = (*ProjectCategoryListResponse)(nil)
)

// ProjectCategory is a way to group and label related projects so teams can
// organize their work more clearly across the platform. By assigning a
// category, you create a higher-level structure that makes it easier to filter,
// report on, and navigate multiple projects, ensuring that departments,
// workflows, or strategic areas remain neatly aligned and easier to manage.
//
// More information can be found at:
// https://support.teamwork.com/projects/glossary/categories
type ProjectCategory struct {
	// ID is the unique identifier of the projectCategory.
	ID int64 `json:"id"`

	// Name is the name of the projectCategory.
	Name string `json:"name"`

	// Color is the color associated with the projectCategory.
	Color string `json:"color"`

	// Parent is the relationship to the parent projectCategory, if any.
	Parent *twapi.Relationship `json:"parent"`

	// Count is the number of projects associated with the projectCategory.
	Count int64 `json:"count"`
}

// ProjectCategoryCreateRequest represents the request body for creating a new
// project category.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/project-categories/post-project-categories-json
type ProjectCategoryCreateRequest struct {
	// Name is the name of the project category. This field is required.
	Name string `json:"name"`

	// Color is the optional color associated with the project category. When
	// provided, it should be a valid hex color code (e.g., "#FF5733").
	Color *string `json:"color,omitempty"`

	// ParentID is the optional ID of the parent project category.
	ParentID *int64 `json:"parent-id,omitempty"`
}

// NewProjectCategoryCreateRequest creates a new ProjectCategoryCreateRequest
// with the provided name. The name is required to create a new project category.
func NewProjectCategoryCreateRequest(name string) ProjectCategoryCreateRequest {
	return ProjectCategoryCreateRequest{Name: name}
}

// HTTPRequest creates an HTTP request for the ProjectCategoryCreateRequest.
func (p ProjectCategoryCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projectcategories.json"

	payload := struct {
		ProjectCategory ProjectCategoryCreateRequest `json:"category"`
	}{ProjectCategory: p}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create projectCategory request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// ProjectCategoryCreateResponse represents the response body for creating a new
// project category.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/project-categories/post-project-categories-json
type ProjectCategoryCreateResponse struct {
	// ID is the unique identifier of the created project category.
	ID LegacyNumber `json:"categoryId"`
}

// HandleHTTPResponse handles the HTTP response for the
// ProjectCategoryCreateResponse. If some unexpected HTTP status code is
// returned by the API, a twapi.HTTPError is returned.
func (p *ProjectCategoryCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create project category")
	}
	if err := json.NewDecoder(resp.Body).Decode(p); err != nil {
		return fmt.Errorf("failed to decode create project category response: %w", err)
	}
	if p.ID == 0 {
		return fmt.Errorf("create project category response does not contain a valid identifier")
	}
	return nil
}

// ProjectCategoryCreate creates a new project category using the provided
// request and returns the response.
func ProjectCategoryCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectCategoryCreateRequest,
) (*ProjectCategoryCreateResponse, error) {
	return twapi.Execute[ProjectCategoryCreateRequest, *ProjectCategoryCreateResponse](ctx, engine, req)
}

// ProjectCategoryUpdateRequestPath contains the path parameters for updating a
// project category.
type ProjectCategoryUpdateRequestPath struct {
	// ID is the unique identifier of the project category to be updated.
	ID int64
}

// ProjectCategoryUpdateRequest represents the request body for updating a
// project category. Besides the identifier, all other fields are optional. When
// a field is not provided, it will not be modified.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/project-categories/put-project-categories-id-json
type ProjectCategoryUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path ProjectCategoryUpdateRequestPath

	// Name is the name of the project category.
	Name *string `json:"name,omitempty"`

	// ParentID is the optional ID of the parent project category.
	ParentID *int64 `json:"parent-id,omitempty"`

	// Color is the optional color associated with the project category. When
	// provided, it should be a valid hex color code (e.g., "#FF5733").
	Color *string `json:"color,omitempty"`
}

// NewProjectCategoryUpdateRequest creates a new ProjectCategoryUpdateRequest
// with the provided project category ID. The ID is required to update a project
// category.
func NewProjectCategoryUpdateRequest(projectCategoryID int64) ProjectCategoryUpdateRequest {
	return ProjectCategoryUpdateRequest{
		Path: ProjectCategoryUpdateRequestPath{
			ID: projectCategoryID,
		},
	}
}

// HTTPRequest creates an HTTP request for the ProjectCategoryUpdateRequest.
func (p ProjectCategoryUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projectcategories/" + strconv.FormatInt(p.Path.ID, 10) + ".json"

	payload := struct {
		ProjectCategory ProjectCategoryUpdateRequest `json:"category"`
	}{ProjectCategory: p}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update project category request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// ProjectCategoryUpdateResponse represents the response body for updating a
// project category.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/project-categories/put-project-categories-id-json
type ProjectCategoryUpdateResponse struct{}

// HandleHTTPResponse handles the HTTP response for the
// ProjectCategoryUpdateResponse. If some unexpected HTTP status code is
// returned by the API, a twapi.HTTPError is returned.
func (p *ProjectCategoryUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update projectCategory")
	}
	if err := json.NewDecoder(resp.Body).Decode(p); err != nil {
		return fmt.Errorf("failed to decode update projectCategory response: %w", err)
	}
	return nil
}

// ProjectCategoryUpdate updates a project category using the provided request
// and returns the response.
func ProjectCategoryUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectCategoryUpdateRequest,
) (*ProjectCategoryUpdateResponse, error) {
	return twapi.Execute[ProjectCategoryUpdateRequest, *ProjectCategoryUpdateResponse](ctx, engine, req)
}

// ProjectCategoryDeleteRequestPath contains the path parameters for deleting a
// project category.
type ProjectCategoryDeleteRequestPath struct {
	// ID is the unique identifier of the project category to be deleted.
	ID int64
}

// ProjectCategoryDeleteRequest represents the request body for deleting a
// project category.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/project-categories/delete-project-categories-id-json
type ProjectCategoryDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path ProjectCategoryDeleteRequestPath
}

// NewProjectCategoryDeleteRequest creates a new ProjectCategoryDeleteRequest
// with the provided project category ID.
func NewProjectCategoryDeleteRequest(projectCategoryID int64) ProjectCategoryDeleteRequest {
	return ProjectCategoryDeleteRequest{
		Path: ProjectCategoryDeleteRequestPath{
			ID: projectCategoryID,
		},
	}
}

// HTTPRequest creates an HTTP request for the ProjectCategoryDeleteRequest.
func (p ProjectCategoryDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projectcategories/" + strconv.FormatInt(p.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// ProjectCategoryDeleteResponse represents the response body for deleting a
// project category.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/project-categories/delete-project-categories-id-json
type ProjectCategoryDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the
// ProjectCategoryDeleteResponse. If some unexpected HTTP status code is
// returned by the API, a twapi.HTTPError is returned.
func (p *ProjectCategoryDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to delete project category")
	}
	if err := json.NewDecoder(resp.Body).Decode(p); err != nil {
		return fmt.Errorf("failed to decode delete project category response: %w", err)
	}
	return nil
}

// ProjectCategoryDelete deletes a project category using the provided request
// and returns the response.
func ProjectCategoryDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectCategoryDeleteRequest,
) (*ProjectCategoryDeleteResponse, error) {
	return twapi.Execute[ProjectCategoryDeleteRequest, *ProjectCategoryDeleteResponse](ctx, engine, req)
}

// ProjectCategoryGetRequestPath contains the path parameters for loading a
// single project category.
type ProjectCategoryGetRequestPath struct {
	// ID is the unique identifier of the project category to be retrieved.
	ID int64 `json:"id"`
}

// ProjectCategoryGetRequest represents the request body for loading a single
// project category.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/categories/get-projects-api-v3-projectcategories-category-id-json
type ProjectCategoryGetRequest struct {
	// Path contains the path parameters for the request.
	Path ProjectCategoryGetRequestPath
}

// NewProjectCategoryGetRequest creates a new ProjectCategoryGetRequest with the
// provided project category ID. The ID is required to load a project category.
func NewProjectCategoryGetRequest(projectCategoryID int64) ProjectCategoryGetRequest {
	return ProjectCategoryGetRequest{
		Path: ProjectCategoryGetRequestPath{
			ID: projectCategoryID,
		},
	}
}

// HTTPRequest creates an HTTP request for the ProjectCategoryGetRequest.
func (p ProjectCategoryGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/projectcategories/" + strconv.FormatInt(p.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// ProjectCategoryGetResponse contains all the information related to a
// project category.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/categories/get-projects-api-v3-projectcategories-category-id-json
type ProjectCategoryGetResponse struct {
	ProjectCategory ProjectCategory `json:"projectCategory"`
}

// HandleHTTPResponse handles the HTTP response for the ProjectCategoryGetResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (p *ProjectCategoryGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve project category")
	}

	if err := json.NewDecoder(resp.Body).Decode(p); err != nil {
		return fmt.Errorf("failed to decode retrieve project category response: %w", err)
	}
	return nil
}

// ProjectCategoryGet retrieves a single project category using the provided
// request and returns the response.
func ProjectCategoryGet(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectCategoryGetRequest,
) (*ProjectCategoryGetResponse, error) {
	return twapi.Execute[ProjectCategoryGetRequest, *ProjectCategoryGetResponse](ctx, engine, req)
}

// ProjectCategoryListRequestFilters contains the filters for loading multiple
// project categories.
type ProjectCategoryListRequestFilters struct {
	// SearchTerm is an optional search term to filter project categories by name.
	SearchTerm string

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of project categories to retrieve per page. Defaults
	// to 50.
	PageSize int64
}

// ProjectCategoryListRequest represents the request body for loading multiple
// project categories.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/categories/get-projects-api-v3-projectcategories-json
type ProjectCategoryListRequest struct {
	// Filters contains the filters for loading multiple project categories.
	Filters ProjectCategoryListRequestFilters
}

// NewProjectCategoryListRequest creates a new ProjectCategoryListRequest with
// default values.
func NewProjectCategoryListRequest() ProjectCategoryListRequest {
	return ProjectCategoryListRequest{
		Filters: ProjectCategoryListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the ProjectCategoryListRequest.
func (p ProjectCategoryListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/projectcategories.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	if p.Filters.SearchTerm != "" {
		query.Set("searchTerm", p.Filters.SearchTerm)
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

// ProjectCategoryListResponse contains information by multiple project
// categories matching the request filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/categories/get-projects-api-v3-projectcategories-json
type ProjectCategoryListResponse struct {
	request ProjectCategoryListRequest

	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`
	ProjectCategories []ProjectCategory `json:"projectCategories"`
}

// HandleHTTPResponse handles the HTTP response for the
// ProjectCategoryListResponse. If some unexpected HTTP status code is returned
// by the API, a twapi.HTTPError is returned.
func (p *ProjectCategoryListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list project categories")
	}

	if err := json.NewDecoder(resp.Body).Decode(p); err != nil {
		return fmt.Errorf("failed to decode list project categories response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (p *ProjectCategoryListResponse) SetRequest(req ProjectCategoryListRequest) {
	p.request = req
}

// Iterate returns the request set to the next page, if available. If there
// are no more pages, a nil request is returned.
func (p *ProjectCategoryListResponse) Iterate() *ProjectCategoryListRequest {
	if !p.Meta.Page.HasMore {
		return nil
	}
	req := p.request
	req.Filters.Page++
	return &req
}

// ProjectCategoryList retrieves multiple project categories using the provided
// request and returns the response.
func ProjectCategoryList(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectCategoryListRequest,
) (*ProjectCategoryListResponse, error) {
	return twapi.Execute[ProjectCategoryListRequest, *ProjectCategoryListResponse](ctx, engine, req)
}
