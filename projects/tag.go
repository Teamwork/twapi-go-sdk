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
	_ twapi.HTTPRequester = (*TagCreateRequest)(nil)
	_ twapi.HTTPResponser = (*TagCreateResponse)(nil)
	_ twapi.HTTPRequester = (*TagUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*TagUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*TagDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*TagDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*TagGetRequest)(nil)
	_ twapi.HTTPResponser = (*TagGetResponse)(nil)
	_ twapi.HTTPRequester = (*TagListRequest)(nil)
	_ twapi.HTTPResponser = (*TagListResponse)(nil)
)

// Tag is a customizable label that can be applied to various items such as
// tasks, projects, milestones, messages, and more, to help categorize and
// organize work efficiently. Tags provide a flexible way to filter, search, and
// group related items across the platform, making it easier for teams to manage
// complex workflows, highlight priorities, or track themes and statuses. Since
// tags are user-defined, they adapt to each team’s specific needs and can be
// color-coded for better visual clarity.
//
// More information can be found at:
// https://support.teamwork.com/projects/glossary/tags-overview
type Tag struct {
	// ID is the unique identifier of the tag.
	ID int64 `json:"id"`

	// Name is the name of the tag.
	Name string `json:"name"`

	// Project is the project the tag belongs to.
	Project *twapi.Relationship `json:"project"`
}

// TagCreateRequest represents the request body for creating a new tag.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/tags/post-projects-api-v3-tags-json
type TagCreateRequest struct {
	// Name is the name of the tag. This field is required. It must be less than
	// 50 characters.
	Name string `json:"name"`

	// ProjectID is the unique identifier of the project the tag belongs to. This
	// is for project-scoped tags.
	ProjectID *int64 `json:"projectId,omitempty"`
}

// NewTagCreateRequest creates a new TagCreateRequest with the provided name.
func NewTagCreateRequest(name string) TagCreateRequest {
	return TagCreateRequest{
		Name: name,
	}
}

// HTTPRequest creates an HTTP request for the TagCreateRequest.
func (t TagCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/tags.json"

	payload := struct {
		Tag TagCreateRequest `json:"tag"`
	}{Tag: t}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create tag request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// TagCreateResponse represents the response body for creating a new tag.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/tags/post-projects-api-v3-tags-json
type TagCreateResponse struct {
	// Tag is the created tag.
	Tag Tag `json:"tag"`
}

// HandleHTTPResponse handles the HTTP response for the TagCreateResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (t *TagCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create tag")
	}
	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode create tag response: %w", err)
	}
	if t.Tag.ID == 0 {
		return fmt.Errorf("create tag response does not contain a valid identifier")
	}
	return nil
}

// TagCreate creates a new tag using the provided request and returns the
// response.
func TagCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req TagCreateRequest,
) (*TagCreateResponse, error) {
	return twapi.Execute[TagCreateRequest, *TagCreateResponse](ctx, engine, req)
}

// TagUpdateRequestPath contains the path parameters for updating a tag.
type TagUpdateRequestPath struct {
	// ID is the unique identifier of the tag to be updated.
	ID int64
}

// TagUpdateRequest represents the request body for updating a tag. Besides the
// identifier, all other fields are optional. When a field is not provided, it
// will not be modified.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/tags/patch-projects-api-v3-tags-tag-id-json
type TagUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path TagUpdateRequestPath `json:"-"`

	// Name is the name of the tag. It must be less than 50 characters when
	// provided.
	Name *string `json:"name,omitempty"`

	// ProjectID is the unique identifier of the project the tag belongs to. This
	// is for project-scoped tags.
	ProjectID *int64 `json:"projectId,omitempty"`
}

// NewTagUpdateRequest creates a new TagUpdateRequest with the provided tag ID.
// The ID is required to update a tag.
func NewTagUpdateRequest(tagID int64) TagUpdateRequest {
	return TagUpdateRequest{
		Path: TagUpdateRequestPath{
			ID: tagID,
		},
	}
}

// HTTPRequest creates an HTTP request for the TagUpdateRequest.
func (t TagUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/tags/" + strconv.FormatInt(t.Path.ID, 10) + ".json"

	payload := struct {
		Tag TagUpdateRequest `json:"tag"`
	}{Tag: t}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update tag request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// TagUpdateResponse represents the response body for updating a tag.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/tags/put-tags-id-json
type TagUpdateResponse struct {
	// Tag is the updated tag.
	Tag Tag `json:"tag"`
}

// HandleHTTPResponse handles the HTTP response for the TagUpdateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (t *TagUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update tag")
	}
	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode update tag response: %w", err)
	}
	return nil
}

// TagUpdate updates a tag using the provided request and returns the response.
func TagUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req TagUpdateRequest,
) (*TagUpdateResponse, error) {
	return twapi.Execute[TagUpdateRequest, *TagUpdateResponse](ctx, engine, req)
}

// TagDeleteRequestPath contains the path parameters for deleting a tag.
type TagDeleteRequestPath struct {
	// ID is the unique identifier of the tag to be deleted.
	ID int64
}

// TagDeleteRequest represents the request body for deleting a tag.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/tags/delete-projects-api-v3-tags-tag-id-json
type TagDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path TagDeleteRequestPath
}

// NewTagDeleteRequest creates a new TagDeleteRequest with the provided tag ID.
func NewTagDeleteRequest(tagID int64) TagDeleteRequest {
	return TagDeleteRequest{
		Path: TagDeleteRequestPath{
			ID: tagID,
		},
	}
}

// HTTPRequest creates an HTTP request for the TagDeleteRequest.
func (t TagDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/tags/" + strconv.FormatInt(t.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// TagDeleteResponse represents the response body for deleting a tag.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/tags/delete-tags-id-json
type TagDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the TagDeleteResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (t *TagDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusNoContent {
		return twapi.NewHTTPError(resp, "failed to delete tag")
	}
	return nil
}

// TagDelete deletes a tag using the provided request and returns the response.
func TagDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req TagDeleteRequest,
) (*TagDeleteResponse, error) {
	return twapi.Execute[TagDeleteRequest, *TagDeleteResponse](ctx, engine, req)
}

// TagGetRequestPath contains the path parameters for loading a single tag.
type TagGetRequestPath struct {
	// ID is the unique identifier of the tag to be retrieved.
	ID int64 `json:"id"`
}

// TagGetRequest represents the request body for loading a single tag.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/tags/get-projects-api-v3-tags-tag-id-json
type TagGetRequest struct {
	// Path contains the path parameters for the request.
	Path TagGetRequestPath
}

// NewTagGetRequest creates a new TagGetRequest with the provided tag ID. The ID
// is required to load a tag.
func NewTagGetRequest(tagID int64) TagGetRequest {
	return TagGetRequest{
		Path: TagGetRequestPath{
			ID: tagID,
		},
	}
}

// HTTPRequest creates an HTTP request for the TagGetRequest.
func (t TagGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/tags/" + strconv.FormatInt(t.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// TagGetResponse contains all the information related to a tag.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/tags/get-projects-api-v3-tags-tag-id-json
type TagGetResponse struct {
	Tag Tag `json:"tag"`
}

// HandleHTTPResponse handles the HTTP response for the TagGetResponse. If some
// unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (t *TagGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve tag")
	}

	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode retrieve tag response: %w", err)
	}
	return nil
}

// TagGet retrieves a single tag using the provided request and returns the
// response.
func TagGet(
	ctx context.Context,
	engine *twapi.Engine,
	req TagGetRequest,
) (*TagGetResponse, error) {
	return twapi.Execute[TagGetRequest, *TagGetResponse](ctx, engine, req)
}

// TagListRequestFilters contains the filters for loading multiple
// tags.
type TagListRequestFilters struct {
	// SearchTerm is an optional search term to filter tags by name.
	SearchTerm string

	// ItemType is the type of item the tag is associated with. Valid values are
	// 'project', 'task', 'tasklist', 'milestone', 'message', 'timelog',
	// 'notebook', 'file', 'company' and 'link'.
	ItemType string

	// ProjectIDs is an optional list of project IDs to filter tags by
	// belonging to specific projects.
	ProjectIDs []int64

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of tags to retrieve per page. Defaults to 50.
	PageSize int64
}

// TagListRequest represents the request body for loading multiple tags.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/tags/get-projects-api-v3-tags-json
type TagListRequest struct {
	// Filters contains the filters for loading multiple tags.
	Filters TagListRequestFilters
}

// NewTagListRequest creates a new TagListRequest with default values.
func NewTagListRequest() TagListRequest {
	return TagListRequest{
		Filters: TagListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the TagListRequest.
func (t TagListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/tags.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	if t.Filters.SearchTerm != "" {
		query.Set("searchTerm", t.Filters.SearchTerm)
	}
	if t.Filters.ItemType != "" {
		query.Set("itemType", t.Filters.ItemType)
	}
	if len(t.Filters.ProjectIDs) > 0 {
		projectIDs := make([]string, len(t.Filters.ProjectIDs))
		for i, id := range t.Filters.ProjectIDs {
			projectIDs[i] = strconv.FormatInt(id, 10)
		}
		query.Set("projectIds", strings.Join(projectIDs, ","))
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

// TagListResponse contains information by multiple tags matching the request
// filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/tags/get-projects-api-v3-tags-json
type TagListResponse struct {
	request TagListRequest

	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`
	Tags []Tag `json:"tags"`
}

// HandleHTTPResponse handles the HTTP response for the TagListResponse. If some
// unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (t *TagListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list tags")
	}

	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode list tags response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (t *TagListResponse) SetRequest(req TagListRequest) {
	t.request = req
}

// Iterate returns the request set to the next page, if available. If there are
// no more pages, a nil request is returned.
func (t *TagListResponse) Iterate() *TagListRequest {
	if !t.Meta.Page.HasMore {
		return nil
	}
	req := t.request
	req.Filters.Page++
	return &req
}

// TagList retrieves multiple tags using the provided request and returns the
// response.
func TagList(
	ctx context.Context,
	engine *twapi.Engine,
	req TagListRequest,
) (*TagListResponse, error) {
	return twapi.Execute[TagListRequest, *TagListResponse](ctx, engine, req)
}
