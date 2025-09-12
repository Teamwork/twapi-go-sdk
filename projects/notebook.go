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
	_ twapi.HTTPRequester = (*NotebookCreateRequest)(nil)
	_ twapi.HTTPResponser = (*NotebookCreateResponse)(nil)
	_ twapi.HTTPRequester = (*NotebookUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*NotebookUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*NotebookDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*NotebookDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*NotebookGetRequest)(nil)
	_ twapi.HTTPResponser = (*NotebookGetResponse)(nil)
	_ twapi.HTTPRequester = (*NotebookListRequest)(nil)
	_ twapi.HTTPResponser = (*NotebookListResponse)(nil)
)

// NotebookType defines the notebook type.
type NotebookType string

// List of possible notebook types.
const (
	// NotebookTypeMarkdown indicates a notebook with markdown content.
	NotebookTypeMarkdown NotebookType = "MARKDOWN"

	// NotebookTypeHTML indicates a notebook with HTML content.
	NotebookTypeHTML NotebookType = "HTML"
)

// Notebook is a space where teams can create, share, and organize written
// content in a structured way. It’s commonly used for documenting processes,
// storing meeting notes, capturing research, or drafting ideas that need to be
// revisited and refined over time. Unlike quick messages or task comments,
// notebooks provide a more permanent and organized format that can be easily
// searched and referenced, helping teams maintain a centralized source of
// knowledge and ensuring important information remains accessible to everyone
// who needs it.
//
// More information can be found at:
// https://support.teamwork.com/projects/notebooks
type Notebook struct {
	// ID is the unique identifier of the notebook.
	ID int64 `json:"id"`

	// Name is the name of the notebook.
	Name string `json:"name"`

	// Description is the description of the notebook.
	Description string `json:"description"`

	// Contents is the contents of the notebook.
	Contents *string `json:"contents,omitempty"` // can be optionally hidden on lists

	// Type is the type of the notebook. It can be "MARKDOWN" or "HTML".
	Type NotebookType `json:"type"`

	// Project is the project associated with the notebook.
	Project twapi.Relationship `json:"project"`

	// Tags is the list of tags associated with the notebook.
	Tags []twapi.Relationship `json:"tags"`

	// CreatedAt is the date and time when the notebook was created.
	CreatedAt *time.Time `json:"createdAt"`

	// CreatedBy is the user who created the notebook.
	CreatedBy *int64 `json:"createdBy"`

	// UpdatedAt is the date and time when the notebook was last updated.
	UpdatedAt *time.Time `json:"updatedAt"`

	// UpdatedBy is the user who last updated the notebook.
	UpdatedBy *int64 `json:"updatedBy"`

	// DeletedAt is the date and time when the notebook was deleted, if it was
	// deleted.
	DeletedAt *time.Time `json:"deletedAt"`

	// DeletedBy is the ID of the user who deleted the notebook, if it was
	// deleted.
	DeletedBy *int64 `json:"deletedBy"`

	// Deleted indicates whether the notebook has been deleted.
	Deleted bool `json:"deleted"`
}

// NotebookUpdateRequestPath contains the path parameters for creating a
// notebook.
type NotebookCreateRequestPath struct {
	// ProjectID is the unique identifier of the project that will contain the
	// notebook.
	ProjectID int64
}

// NotebookCreateRequest represents the request body for creating a new
// notebook.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/notebooks/post-projects-api-v3-projects-project-id-notebooks-json
type NotebookCreateRequest struct {
	// Path contains the path parameters for the request.
	Path NotebookCreateRequestPath `json:"-"`

	// Name is the name of the notebook. This field is required and must not be
	// empty.
	Name string `json:"name"`

	// Description is the description of the notebook.
	Description *string `json:"description,omitempty"`

	// Contents is the contents of the notebook. This field is required and must
	// not be empty.
	Contents string `json:"contents"`

	// Type is the type of the notebook. It can be "MARKDOWN" or "HTML". This
	// field is required.
	Type NotebookType `json:"type"`

	// TagIDs is the list of tags associated with the notebook.
	TagIDs []int64 `json:"tagIds,omitempty"`
}

// NewNotebookCreateRequest creates a new NotebookCreateRequest with the
// provided required fields.
func NewNotebookCreateRequest(projectID int64, name, contents string, typ NotebookType) NotebookCreateRequest {
	return NotebookCreateRequest{
		Path: NotebookCreateRequestPath{
			ProjectID: projectID,
		},
		Name:     name,
		Contents: contents,
		Type:     typ,
	}
}

// HTTPRequest creates an HTTP request for the NotebookCreateRequest.
func (m NotebookCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/projects/%d/notebooks.json", server, m.Path.ProjectID)

	payload := struct {
		Notebook NotebookCreateRequest `json:"notebook"`
	}{Notebook: m}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create notebook request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// NotebookCreateResponse represents the response body for creating a new
// notebook.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/notebooks/post-projects-api-v3-projects-project-id-notebooks-json
type NotebookCreateResponse struct {
	// Notebook is the created notebook.
	Notebook Notebook `json:"notebook"`
}

// HandleHTTPResponse handles the HTTP response for the NotebookCreateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (n *NotebookCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create notebook")
	}
	if err := json.NewDecoder(resp.Body).Decode(n); err != nil {
		return fmt.Errorf("failed to decode create notebook response: %w", err)
	}
	if n.Notebook.ID == 0 {
		return fmt.Errorf("create notebook response does not contain a valid identifier")
	}
	return nil
}

// NotebookCreate creates a new notebook using the provided request and returns
// the response.
func NotebookCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req NotebookCreateRequest,
) (*NotebookCreateResponse, error) {
	return twapi.Execute[NotebookCreateRequest, *NotebookCreateResponse](ctx, engine, req)
}

// NotebookUpdateRequestPath contains the path parameters for updating a
// notebook.
type NotebookUpdateRequestPath struct {
	// ID is the unique identifier of the notebook to be updated.
	ID int64
}

// NotebookUpdateRequest represents the request body for updating a notebook.
// Besides the identifier, all other fields are optional. When a field is not
// provided, it will not be modified.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/notebooks/patch-projects-api-v3-notebooks-notebook-id-json
type NotebookUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path NotebookUpdateRequestPath `json:"-"`

	// Name is the name of the notebook.
	Name *string `json:"name,omitempty"`

	// Description is the description of the notebook.
	Description *string `json:"description,omitempty"`

	// Contents is the contents of the notebook.
	Contents *string `json:"contents,omitempty"`

	// Type is the type of the notebook. It can be "MARKDOWN" or "HTML".
	Type *NotebookType `json:"type,omitempty"`

	// TagIDs is the list of tags associated with the notebook.
	TagIDs []int64 `json:"tagIds,omitempty"`
}

// NewNotebookUpdateRequest creates a new NotebookUpdateRequest with the
// provided notebook ID. The ID is required to update a notebook.
func NewNotebookUpdateRequest(notebookID int64) NotebookUpdateRequest {
	return NotebookUpdateRequest{
		Path: NotebookUpdateRequestPath{
			ID: notebookID,
		},
	}
}

// HTTPRequest creates an HTTP request for the NotebookUpdateRequest.
func (m NotebookUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/notebooks/" + strconv.FormatInt(m.Path.ID, 10) + ".json"

	payload := struct {
		Notebook NotebookUpdateRequest `json:"notebook"`
	}{Notebook: m}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update notebook request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// NotebookUpdateResponse represents the response body for updating a notebook.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/notebooks/patch-projects-api-v3-notebooks-notebook-id-json
type NotebookUpdateResponse struct {
	// Notebook is the updated notebook.
	Notebook Notebook `json:"notebook"`
}

// HandleHTTPResponse handles the HTTP response for the NotebookUpdateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (m *NotebookUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update notebook")
	}
	if err := json.NewDecoder(resp.Body).Decode(m); err != nil {
		return fmt.Errorf("failed to decode update notebook response: %w", err)
	}
	return nil
}

// NotebookUpdate updates a notebook using the provided request and returns
// the response.
func NotebookUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req NotebookUpdateRequest,
) (*NotebookUpdateResponse, error) {
	return twapi.Execute[NotebookUpdateRequest, *NotebookUpdateResponse](ctx, engine, req)
}

// NotebookDeleteRequestPath contains the path parameters for deleting a
// notebook.
type NotebookDeleteRequestPath struct {
	// ID is the unique identifier of the notebook to be deleted.
	ID int64
}

// NotebookDeleteRequest represents the request body for deleting a notebook.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/notebooks/delete-projects-api-v3-notebooks-notebook-id-json
type NotebookDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path NotebookDeleteRequestPath
}

// NewNotebookDeleteRequest creates a new NotebookDeleteRequest with the
// provided notebook ID.
func NewNotebookDeleteRequest(notebookID int64) NotebookDeleteRequest {
	return NotebookDeleteRequest{
		Path: NotebookDeleteRequestPath{
			ID: notebookID,
		},
	}
}

// HTTPRequest creates an HTTP request for the NotebookDeleteRequest.
func (m NotebookDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/notebooks/" + strconv.FormatInt(m.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NotebookDeleteResponse represents the response body for deleting a notebook.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/notebooks/delete-projects-api-v3-notebooks-notebook-id-json
type NotebookDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the NotebookDeleteResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (m *NotebookDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusNoContent {
		return twapi.NewHTTPError(resp, "failed to delete notebook")
	}
	return nil
}

// NotebookDelete deletes a notebook using the provided request and returns
// the response.
func NotebookDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req NotebookDeleteRequest,
) (*NotebookDeleteResponse, error) {
	return twapi.Execute[NotebookDeleteRequest, *NotebookDeleteResponse](ctx, engine, req)
}

// NotebookGetRequestPath contains the path parameters for loading a single
// notebook.
type NotebookGetRequestPath struct {
	// ID is the unique identifier of the notebook to be retrieved.
	ID int64 `json:"id"`
}

// NotebookGetRequest represents the request body for loading a single notebook.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/notebooks/get-projects-api-v3-notebooks-notebook-id-json
type NotebookGetRequest struct {
	// Path contains the path parameters for the request.
	Path NotebookGetRequestPath
}

// NewNotebookGetRequest creates a new NotebookGetRequest with the provided
// notebook ID. The ID is required to load a notebook.
func NewNotebookGetRequest(notebookID int64) NotebookGetRequest {
	return NotebookGetRequest{
		Path: NotebookGetRequestPath{
			ID: notebookID,
		},
	}
}

// HTTPRequest creates an HTTP request for the NotebookGetRequest.
func (m NotebookGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/notebooks/" + strconv.FormatInt(m.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NotebookGetResponse contains all the information related to a notebook.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/notebooks/get-projects-api-v3-notebooks-notebook-id-json
type NotebookGetResponse struct {
	Notebook Notebook `json:"notebook"`
}

// HandleHTTPResponse handles the HTTP response for the NotebookGetResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (m *NotebookGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve notebook")
	}

	if err := json.NewDecoder(resp.Body).Decode(m); err != nil {
		return fmt.Errorf("failed to decode retrieve notebook response: %w", err)
	}
	return nil
}

// NotebookGet retrieves a single notebook using the provided request and
// returns the response.
func NotebookGet(
	ctx context.Context,
	engine *twapi.Engine,
	req NotebookGetRequest,
) (*NotebookGetResponse, error) {
	return twapi.Execute[NotebookGetRequest, *NotebookGetResponse](ctx, engine, req)
}

// NotebookListRequestFilters contains the filters for loading multiple
// notebooks.
type NotebookListRequestFilters struct {
	// ProjectIDs is an optional list of project IDs to filter notebooks by
	// projects. If provided, only notebooks belonging to the specified projects
	// will be returned.
	ProjectIDs []int64

	// SearchTerm is an optional search term to filter notebooks by name or
	// description.
	SearchTerm string

	// TagIDs is an optional list of tag IDs to filter notebooks by tags.
	TagIDs []int64

	// MatchAllTags is an optional flag to indicate if all tags must match. If set
	// to true, only notebooks matching all specified tags will be returned.
	MatchAllTags *bool

	// IncludeContents is an optional flag to indicate if the contents of the
	// notebooks should be included in the response. If set to true, the contents
	// will be included; otherwise, they will be omitted. Defaults to true.
	IncludeContents *bool

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of notebooks to retrieve per page. Defaults to 50.
	PageSize int64
}

// NotebookListRequest represents the request body for loading multiple notebooks.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/notebooks/get-projects-api-v3-notebooks-json
type NotebookListRequest struct {
	// Filters contains the filters for loading multiple notebooks.
	Filters NotebookListRequestFilters
}

// NewNotebookListRequest creates a new NotebookListRequest with default values.
func NewNotebookListRequest() NotebookListRequest {
	return NotebookListRequest{
		Filters: NotebookListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the NotebookListRequest.
func (m NotebookListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/notebooks.json"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	query := req.URL.Query()
	if len(m.Filters.ProjectIDs) > 0 {
		projectIDs := make([]string, len(m.Filters.ProjectIDs))
		for i, id := range m.Filters.ProjectIDs {
			projectIDs[i] = strconv.FormatInt(id, 10)
		}
		query.Set("projectIds", strings.Join(projectIDs, ","))
	}
	if m.Filters.SearchTerm != "" {
		query.Set("searchTerm", m.Filters.SearchTerm)
	}
	if len(m.Filters.TagIDs) > 0 {
		tagIDs := make([]string, len(m.Filters.TagIDs))
		for i, id := range m.Filters.TagIDs {
			tagIDs[i] = strconv.FormatInt(id, 10)
		}
		query.Set("tagIds", strings.Join(tagIDs, ","))
	}
	if m.Filters.MatchAllTags != nil {
		query.Set("matchAllTags", strconv.FormatBool(*m.Filters.MatchAllTags))
	}
	if m.Filters.IncludeContents != nil {
		query.Set("includeContents", strconv.FormatBool(*m.Filters.IncludeContents))
	}
	if m.Filters.Page > 0 {
		query.Set("page", strconv.FormatInt(m.Filters.Page, 10))
	}
	if m.Filters.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(m.Filters.PageSize, 10))
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// NotebookListResponse contains information by multiple notebooks matching the
// request filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/notebooks/get-projects-api-v3-notebooks-json
type NotebookListResponse struct {
	request NotebookListRequest

	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`
	Notebooks []Notebook `json:"notebooks"`
}

// HandleHTTPResponse handles the HTTP response for the NotebookListResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (m *NotebookListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list notebooks")
	}

	if err := json.NewDecoder(resp.Body).Decode(m); err != nil {
		return fmt.Errorf("failed to decode list notebooks response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (m *NotebookListResponse) SetRequest(req NotebookListRequest) {
	m.request = req
}

// Iterate returns the request set to the next page, if available. If there
// are no more pages, a nil request is returned.
func (m *NotebookListResponse) Iterate() *NotebookListRequest {
	if !m.Meta.Page.HasMore {
		return nil
	}
	req := m.request
	req.Filters.Page++
	return &req
}

// NotebookList retrieves multiple notebooks using the provided request and
// returns the response.
func NotebookList(
	ctx context.Context,
	engine *twapi.Engine,
	req NotebookListRequest,
) (*NotebookListResponse, error) {
	return twapi.Execute[NotebookListRequest, *NotebookListResponse](ctx, engine, req)
}
