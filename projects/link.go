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
	_ twapi.HTTPRequester = (*LinkCreateRequest)(nil)
	_ twapi.HTTPResponser = (*LinkCreateResponse)(nil)
	_ twapi.HTTPRequester = (*LinkUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*LinkUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*LinkDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*LinkDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*LinkGetRequest)(nil)
	_ twapi.HTTPResponser = (*LinkGetResponse)(nil)
	_ twapi.HTTPRequester = (*LinkListRequest)(nil)
	_ twapi.HTTPResponser = (*LinkListResponse)(nil)
)

// Link is a saved URL attached to a project, task, or other item, allowing
// users to quickly reference and access external resources (such as documents,
// tools, or websites) directly within their workflow.
//
// More information can be found at:
// https://support.teamwork.com/projects/links/links-explained
type Link struct {
	// ID is the unique identifier of the link.
	ID LegacyNumber `json:"id"`

	// Title is the name of the link, which provides a brief summary of the
	// content and purpose of the link, allowing team members to quickly
	// understand the topic at a glance.
	Title string `json:"name"`

	// Description is longer text that provides more detailed information about
	// the link, allowing team members to understand the context, background, and
	// specifics of the link's content.
	Description string `json:"description"`

	// Code is the URL of the link, which is the actual web address that team
	// members can click on to access the linked resource, enabling seamless
	// integration of external content within the project environment.
	Code string `json:"code"`

	// ProjectID is the unique identifier of the project associated with the link.
	ProjectID LegacyNumber `json:"project-id"`

	// Tags is the list of tags associated with the link.
	Tags []struct {
		// ID is the unique identifier of the tag.
		ID LegacyNumber `json:"id"`

		// Name is the name of the tag, which provides a brief label or keyword that
		// categorizes the link, allowing team members to quickly identify and group
		// related links based on shared characteristics or themes.
		Name string `json:"name"`

		// ProjectID is the unique identifier of the project associated with the
		// tag.
		ProjectID *LegacyNumber `json:"projectId"`
	} `json:"tags"`

	// CreatedByUserID is the user who created this link.
	CreatedByUserID LegacyNumber `json:"created-by-userId"`

	// CreatedAt is the date and time when the link was created.
	CreatedAt time.Time `json:"created-date"`

	// UpdatedByUserID is the user who last updated this link.
	UpdatedByUserID *LegacyNumber `json:"updated-by-userId"`

	// UpdatedAt is the date and time when the link was last updated.
	UpdatedAt *time.Time `json:"updated-date"`
}

// linkNotifier is an interface that represents the different options for
// notifying users when a link is created or updated. It is implemented by
// LinkNotifyAll and LinkNotifyGroup.
type linkNotifier interface {
	linkNotifier()
}

// LinkNotifyAll is a link notifier that notifies all users from the project.
type LinkNotifyAll struct{}

// NewLinkNotifyAll creates a new LinkNotifyAll notifier.
func NewLinkNotifyAll() LinkNotifyAll {
	return LinkNotifyAll{}
}

// linkNotifier is a marker method to indicate that LinkNotifyAll implements the
// linkNotifier interface.
func (LinkNotifyAll) linkNotifier() {}

// MarshalJSON encodes the LinkNotifyAll as a string "ALL", which is the value
// expected by the API to notify all users.
func (LinkNotifyAll) MarshalJSON() ([]byte, error) {
	return []byte(`"ALL"`), nil
}

// LinkNotifyGroup is a link notifier that notifies a specific group of users,
// teams and/or companies.
type LinkNotifyGroup LegacyUserGroups

// NewLinkNotifyGroup creates a new LinkNotifyGroup notifier with the provided
// group of users, teams and/or companies.
func NewLinkNotifyGroup(groups LegacyUserGroups) LinkNotifyGroup {
	return LinkNotifyGroup(groups)
}

// linkNotifier is a marker method to indicate that LinkNotifyGroup implements
// the linkNotifier interface.
func (LinkNotifyGroup) linkNotifier() {}

// MarshalJSON encodes the LinkNotifyGroup as a string containing the encoded
// group of users, teams and/or companies, which is the value expected by the
// API to notify a specific group.
func (c LinkNotifyGroup) MarshalJSON() ([]byte, error) {
	return LegacyUserGroups(c).MarshalJSON()
}

// LinkUpdateRequestPath contains the path parameters for creating a
// link.
type LinkCreateRequestPath struct {
	// ProjectID is the unique identifier of the project that will contain the
	// link.
	ProjectID int64
}

// LinkCreateRequest represents the request body for creating a new link.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/links/post-projects-id-links-json
type LinkCreateRequest struct {
	// Path contains the path parameters for the request.
	Path LinkCreateRequestPath `json:"-"`

	// Title is the name of the link, which provides a brief summary of the
	// content and purpose of the link, allowing team members to quickly
	// understand the topic at a glance.
	Title *string `json:"name,omitempty"`

	// Description is longer text that provides more detailed information about
	// the link, allowing team members to understand the context, background, and
	// specifics of the link's content.
	Description *string `json:"description,omitempty"`

	// Code is the URL of the link, which is the actual web address that team
	// members can click on to access the linked resource, enabling seamless
	// integration of external content within the project environment.
	Code string `json:"code"`

	// TagIDs is an optional list of tag IDs to associate with the link.
	TagIDs LegacyNumericList `json:"tagIds,omitempty"`

	// NotifyCurrentUser indicates whether the user creating the link should be
	// notified about the new link. If not provided, it defaults to false.
	NotifyCurrentUser *bool `json:"notify-current-user,omitempty"`

	// Notify is the link notifier that specifies who should be notified about the
	// new link. It can be a LinkNotifyAll, or LinkNotifyGroup. If not provided,
	// no notifications will be sent.
	Notify linkNotifier `json:"notify,omitempty"`
}

// NewLinkCreateRequest creates a new LinkCreateRequest with the provided
// required fields.
func NewLinkCreateRequest(projectID int64, code string) LinkCreateRequest {
	return LinkCreateRequest{
		Path: LinkCreateRequestPath{
			ProjectID: projectID,
		},
		Code: code,
	}
}

// HTTPRequest creates an HTTP request for the LinkCreateRequest.
func (l LinkCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/%d/links.json", server, l.Path.ProjectID)

	payload := struct {
		Link LinkCreateRequest `json:"link"`
	}{Link: l}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create link request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// LinkCreateResponse represents the response body for creating a new
// link.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/links/post-projects-id-links-json
type LinkCreateResponse struct {
	// ID is the unique identifier of the created link.
	ID LegacyNumber `json:"id"`
}

// HandleHTTPResponse handles the HTTP response for the LinkCreateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (l *LinkCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create link")
	}
	if err := json.NewDecoder(resp.Body).Decode(l); err != nil {
		return fmt.Errorf("failed to decode create link response: %w", err)
	}
	if l.ID == 0 {
		return fmt.Errorf("create link response does not contain a valid identifier")
	}
	return nil
}

// LinkCreate creates a new link using the provided request and returns the
// response.
func LinkCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req LinkCreateRequest,
) (*LinkCreateResponse, error) {
	return twapi.Execute[LinkCreateRequest, *LinkCreateResponse](ctx, engine, req)
}

// LinkUpdateRequestPath contains the path parameters for updating a link.
type LinkUpdateRequestPath struct {
	// ID is the unique identifier of the link to be updated.
	ID int64
}

// LinkUpdateRequest represents the request body for updating a link.
// Besides the identifier, all other fields are optional. When a field is not
// provided, it will not be modified.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/links/put-links-id-json
type LinkUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path LinkUpdateRequestPath `json:"-"`

	// Title is the name of the link, which provides a brief summary of the
	// content and purpose of the link, allowing team members to quickly
	// understand the topic at a glance.
	Title *string `json:"name,omitempty"`

	// Description is longer text that provides more detailed information about
	// the link, allowing team members to understand the context, background, and
	// specifics of the link's content.
	Description *string `json:"description,omitempty"`

	// Code is the URL of the link, which is the actual web address that team
	// members can click on to access the linked resource, enabling seamless
	// integration of external content within the project environment.
	Code *string `json:"code,omitempty"`

	// TagIDs is an optional list of tag IDs to associate with the link.
	TagIDs LegacyNumericList `json:"tagIds,omitempty"`

	// NotifyCurrentUser indicates whether the user creating the link should be
	// notified about the new link. If not provided, it defaults to false.
	NotifyCurrentUser *bool `json:"notify-current-user,omitempty"`

	// Notify is the link notifier that specifies who should be notified about the
	// new link. It can be a LinkNotifyAll, or LinkNotifyGroup. If not provided,
	// no notifications will be sent.
	Notify linkNotifier `json:"notify,omitempty"`
}

// NewLinkUpdateRequest creates a new LinkUpdateRequest with the provided
// link ID. The ID is required to update a link.
func NewLinkUpdateRequest(linkID int64) LinkUpdateRequest {
	return LinkUpdateRequest{
		Path: LinkUpdateRequestPath{
			ID: linkID,
		},
	}
}

// HTTPRequest creates an HTTP request for the LinkUpdateRequest.
func (l LinkUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/links/" + strconv.FormatInt(l.Path.ID, 10) + ".json"

	payload := struct {
		Link LinkUpdateRequest `json:"link"`
	}{Link: l}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update link request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// LinkUpdateResponse represents the response body for updating a link.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/links/put-links-id-json
type LinkUpdateResponse struct{}

// HandleHTTPResponse handles the HTTP response for the LinkUpdateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (l *LinkUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update link")
	}
	if err := json.NewDecoder(resp.Body).Decode(l); err != nil {
		return fmt.Errorf("failed to decode update link response: %w", err)
	}
	return nil
}

// LinkUpdate updates a link using the provided request and returns the
// response.
func LinkUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req LinkUpdateRequest,
) (*LinkUpdateResponse, error) {
	return twapi.Execute[LinkUpdateRequest, *LinkUpdateResponse](ctx, engine, req)
}

// LinkDeleteRequestPath contains the path parameters for deleting a link.
type LinkDeleteRequestPath struct {
	// ID is the unique identifier of the link to be deleted.
	ID int64
}

// LinkDeleteRequest represents the request body for deleting a link.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/links/delete-links-id-json
type LinkDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path LinkDeleteRequestPath
}

// NewLinkDeleteRequest creates a new LinkDeleteRequest with the provided link
// ID.
func NewLinkDeleteRequest(linkID int64) LinkDeleteRequest {
	return LinkDeleteRequest{
		Path: LinkDeleteRequestPath{
			ID: linkID,
		},
	}
}

// HTTPRequest creates an HTTP request for the LinkDeleteRequest.
func (l LinkDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/links/" + strconv.FormatInt(l.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// LinkDeleteResponse represents the response body for deleting a link.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/links/delete-links-id-json
type LinkDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the LinkDeleteResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (l *LinkDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to delete link")
	}
	if err := json.NewDecoder(resp.Body).Decode(l); err != nil {
		return fmt.Errorf("failed to decode delete link response: %w", err)
	}
	return nil
}

// LinkDelete deletes a link using the provided request and returns the
// response.
func LinkDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req LinkDeleteRequest,
) (*LinkDeleteResponse, error) {
	return twapi.Execute[LinkDeleteRequest, *LinkDeleteResponse](ctx, engine, req)
}

// LinkGetRequestPath contains the path parameters for loading a single
// link.
type LinkGetRequestPath struct {
	// ID is the unique identifier of the link to be retrieved.
	ID int64 `json:"id"`
}

// LinkGetRequest represents the request body for loading a single link.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/links/get-links-id-json
type LinkGetRequest struct {
	// Path contains the path parameters for the request.
	Path LinkGetRequestPath
}

// NewLinkGetRequest creates a new LinkGetRequest with the provided link ID. The
// ID is required to load a link.
func NewLinkGetRequest(linkID int64) LinkGetRequest {
	return LinkGetRequest{
		Path: LinkGetRequestPath{
			ID: linkID,
		},
	}
}

// HTTPRequest creates an HTTP request for the LinkGetRequest.
func (l LinkGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/links/" + strconv.FormatInt(l.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// LinkGetResponse contains all the information related to a link.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/links/get-links-id-json
type LinkGetResponse struct {
	Link Link `json:"link"`
}

// HandleHTTPResponse handles the HTTP response for the LinkGetResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (l *LinkGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve link")
	}

	if err := json.NewDecoder(resp.Body).Decode(l); err != nil {
		return fmt.Errorf("failed to decode retrieve link response: %w", err)
	}
	return nil
}

// LinkGet retrieves a single link using the provided request and returns
// the response.
func LinkGet(
	ctx context.Context,
	engine *twapi.Engine,
	req LinkGetRequest,
) (*LinkGetResponse, error) {
	return twapi.Execute[LinkGetRequest, *LinkGetResponse](ctx, engine, req)
}

// LinkListRequestFilters contains the filters for loading multiple
// links.
type LinkListRequestFilters struct {
	// SearchTerm is an optional search term to filter links by title or
	// description.
	SearchTerm string

	// ProjectID is an optional project ID to filter links by project.
	ProjectID int64

	// TagIDs is an optional list of tag IDs to filter links by tags.
	TagIDs []int64

	// MatchAllTags is an optional flag to indicate if all tags must match. If set
	// to true, only links matching all specified tags will be returned.
	MatchAllTags *bool

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of links to retrieve per page. Defaults to 50.
	PageSize int64
}

func (l LinkListRequestFilters) apply(req *http.Request) {
	query := req.URL.Query()
	if l.SearchTerm != "" {
		query.Set("filterText", l.SearchTerm)
	}
	if l.ProjectID > 0 {
		query.Set("projectId", strconv.FormatInt(l.ProjectID, 10))
	}
	if len(l.TagIDs) > 0 {
		tagIDs := make([]string, len(l.TagIDs))
		for i, id := range l.TagIDs {
			tagIDs[i] = strconv.FormatInt(id, 10)
		}
		query.Set("filterTagIds", strings.Join(tagIDs, ","))
	}
	if l.MatchAllTags != nil {
		query.Set("matchAllTags", strconv.FormatBool(*l.MatchAllTags))
	}
	if l.Page > 0 {
		query.Set("page", strconv.FormatInt(l.Page, 10))
	}
	if l.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(l.PageSize, 10))
	}
	req.URL.RawQuery = query.Encode()
}

// LinkListRequest represents the request body for loading multiple links.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/links/get-links-json
type LinkListRequest struct {
	// Filters contains the filters for loading multiple links.
	Filters LinkListRequestFilters
}

// NewLinkListRequest creates a new LinkListRequest with default values.
func NewLinkListRequest() LinkListRequest {
	return LinkListRequest{
		Filters: LinkListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the LinkListRequest.
func (l LinkListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/links.json"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	l.Filters.apply(req)

	return req, nil
}

// LinkListResponse contains information by multiple links matching the request
// filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/links/get-links-json
type LinkListResponse struct {
	request LinkListRequest
	hasMore bool

	// Links is the list of links matching the request filters. This field is not
	// directly decoded from the API response, but is populated in the
	// HandleHTTPResponse method, which decodes the response and extracts the
	// links from it.
	Links []Link `json:"links"`
}

// HandleHTTPResponse handles the HTTP response for the LinkListResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (l *LinkListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list links")
	}

	page, _ := strconv.ParseInt(resp.Header.Get("X-Page"), 10, 64)
	pages, _ := strconv.ParseInt(resp.Header.Get("X-Pages"), 10, 64)
	l.hasMore = pages > page

	var links struct {
		Project struct {
			Links []Link `json:"links"`
		} `json:"project"`
		Projects []struct {
			Links []Link `json:"links"`
		} `json:"projects"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&links); err != nil {
		return fmt.Errorf("failed to decode list links response: %w", err)
	}

	l.Links = links.Project.Links
	for _, project := range links.Projects {
		l.Links = append(l.Links, project.Links...)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (l *LinkListResponse) SetRequest(req LinkListRequest) {
	l.request = req
}

// Iterate returns the request set to the next page, if available. If there are
// no more pages, a nil request is returned.
func (l *LinkListResponse) Iterate() *LinkListRequest {
	if !l.hasMore {
		return nil
	}
	req := l.request
	req.Filters.Page++
	return &req
}

// LinkList retrieves multiple links using the provided request and returns the
// response.
func LinkList(
	ctx context.Context,
	engine *twapi.Engine,
	req LinkListRequest,
) (*LinkListResponse, error) {
	return twapi.Execute[LinkListRequest, *LinkListResponse](ctx, engine, req)
}
