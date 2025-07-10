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
	_ twapi.HTTPRequester = (*ProjectRetrieveRequest)(nil)
	_ twapi.HTTPResponser = (*ProjectRetrieveResponse)(nil)
	_ twapi.HTTPRequester = (*ProjectRetrieveManyRequest)(nil)
	_ twapi.HTTPResponser = (*ProjectRetrieveManyResponse)(nil)
)

// Project represents a project in Teamwork.
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

	// Tags is an optional list of tag IDs associated with the project.
	Tags []int64 `json:"tagIds,omitempty"`
}

// HTTPRequest creates an HTTP request for the ProjectCreateRequest.
func (c ProjectCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects.json"

	payload := struct {
		Project ProjectCreateRequest `json:"project"`
	}{Project: c}

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
func (c *ProjectCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create project")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode create project response: %w", err)
	}
	if c.ID == 0 {
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
	return twapi.Execute[*ProjectCreateResponse](ctx, engine, req)
}

// ProjectUpdateRequest represents the request body for updating a project.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/projects/put-projects-id-json
type ProjectUpdateRequest struct {
	Path struct {
		// ID is the unique identifier of the project to be updated.
		ID int64
	}

	ProjectCreateRequest
}

// HTTPRequest creates an HTTP request for the ProjectUpdateRequest.
func (c ProjectUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/" + strconv.FormatInt(c.Path.ID, 10) + ".json"

	payload := struct {
		Project ProjectUpdateRequest `json:"project"`
	}{Project: c}

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
func (c *ProjectUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update project")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode update project response: %w", err)
	}
	return nil
}

// ProjectUpdate creates a new project using the provided request and returns
// the response.
func ProjectUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectUpdateRequest,
) (*ProjectUpdateResponse, error) {
	return twapi.Execute[*ProjectUpdateResponse](ctx, engine, req)
}

// ProjectDeleteRequest represents the request body for deleting a project.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/projects/delete-projects-id-json
type ProjectDeleteRequest struct {
	Path struct {
		// ID is the unique identifier of the project to be deleted.
		ID int64
	}
}

// HTTPRequest creates an HTTP request for the ProjectDeleteRequest.
func (c ProjectDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/" + strconv.FormatInt(c.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// ProjectDeleteResponse represents the response body for deleting a project.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/projects/delete-projects-id-json
type ProjectDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the ProjectDeleteResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (c *ProjectDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to delete project")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode delete project response: %w", err)
	}
	return nil
}

// ProjectDelete creates a new project using the provided request and returns
// the response.
func ProjectDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectDeleteRequest,
) (*ProjectDeleteResponse, error) {
	return twapi.Execute[*ProjectDeleteResponse](ctx, engine, req)
}

// ProjectRetrieveRequest represents the request body for loading a single
// project.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/projects/get-projects-api-v3-projects-project-id-json
type ProjectRetrieveRequest struct {
	Path struct {
		// ID is the unique identifier of the project to be retrieved.
		ID int64 `json:"id"`
	}
}

// HTTPRequest creates an HTTP request for the ProjectRetrieveRequest.
func (p ProjectRetrieveRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/projects/" + strconv.FormatInt(p.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// ProjectRetrieveResponse contains all the information related to a project.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/projects/get-projects-api-v3-projects-project-id-json
type ProjectRetrieveResponse struct {
	Project Project `json:"project"`
}

// HandleHTTPResponse handles the HTTP response for the ProjectRetrieveResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (p *ProjectRetrieveResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve project")
	}

	if err := json.NewDecoder(resp.Body).Decode(p); err != nil {
		return fmt.Errorf("failed to decode retrieve project response: %w", err)
	}
	return nil
}

// ProjectRetrieve retrieves a single project using the provided request and
// returns the response.
func ProjectRetrieve(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectRetrieveRequest,
) (*ProjectRetrieveResponse, error) {
	return twapi.Execute[*ProjectRetrieveResponse](ctx, engine, req)
}

// ProjectRetrieveManyRequest represents the request body for loading multiple
// projects.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/projects/get-projects-api-v3-projects-json
type ProjectRetrieveManyRequest struct {
	Filters struct {
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
}

// HTTPRequest creates an HTTP request for the ProjectRetrieveManyRequest.
func (p ProjectRetrieveManyRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
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

// ProjectRetrieveManyResponse contains information by multiple projects
// matching the request filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/projects/get-projects-api-v3-projects-json
type ProjectRetrieveManyResponse struct {
	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`
	Projects []Project `json:"projects"`
}

// HandleHTTPResponse handles the HTTP response for the
// ProjectRetrieveManyResponse. If some unexpected HTTP status code is returned
// by the API, a twapi.HTTPError is returned.
func (p *ProjectRetrieveManyResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve many projects")
	}

	if err := json.NewDecoder(resp.Body).Decode(p); err != nil {
		return fmt.Errorf("failed to decode retrieve many projects response: %w", err)
	}
	return nil
}

// ProjectRetrieveMany retrieves multiple projects using the provided request
// and returns the response.
func ProjectRetrieveMany(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectRetrieveManyRequest,
) (*ProjectRetrieveManyResponse, error) {
	return twapi.Execute[*ProjectRetrieveManyResponse](ctx, engine, req)
}
