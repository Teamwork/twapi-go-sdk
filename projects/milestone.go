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
	_ twapi.HTTPRequester = (*MilestoneCreateRequest)(nil)
	_ twapi.HTTPResponser = (*MilestoneCreateResponse)(nil)
	_ twapi.HTTPRequester = (*MilestoneUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*MilestoneUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*MilestoneDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*MilestoneDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*MilestoneGetRequest)(nil)
	_ twapi.HTTPResponser = (*MilestoneGetResponse)(nil)
	_ twapi.HTTPRequester = (*MilestoneListRequest)(nil)
	_ twapi.HTTPResponser = (*MilestoneListResponse)(nil)
)

// Milestone represents a significant point or goal within a project that marks
// the completion of a major phase or a key deliverable. It acts as a high-level
// indicator of progress, helping teams track whether work is advancing
// according to plan. Milestones are typically used to coordinate efforts across
// different tasks and task lists, providing a clear deadline or objective that
// multiple team members or departments can align around. They don't contain
// individual tasks themselves but serve as checkpoints to ensure the project is
// moving in the right direction.
//
// More information can be found at:
// https://support.teamwork.com/projects/getting-started/milestones-overview
type Milestone struct {
	// ID is the unique identifier of the milestone.
	ID int64 `json:"id"`

	// Name is the name of the milestone.
	Name string `json:"name"`

	// Description is the description of the milestone.
	Description string `json:"description"`

	// DueAt is the due date of the milestone.
	DueAt time.Time `json:"deadline"`

	// Project is the project associated with the milestone.
	Project twapi.Relationship `json:"project"`

	// Tasklists is the list of tasklists associated with the milestone.
	Tasklists []twapi.Relationship `json:"tasklists"`

	// Tags is the list of tags associated with the milestone.
	Tags []twapi.Relationship `json:"tags"`

	// ResponsibleParties is the list of assingees (users, teams and
	// clients/companies) responsible for the milestone.
	ResponsibleParties []twapi.Relationship `json:"responsibleParties"`

	// CreatedAt is the date and time when the milestone was created.
	CreatedAt *time.Time `json:"createdOn"`

	// UpdatedAt is the date and time when the milestone was last updated.
	UpdatedAt *time.Time `json:"lastChangedOn"`

	// DeletedAt is the date and time when the milestone was deleted, if it was
	// deleted.
	DeletedAt *time.Time `json:"deletedOn"`

	// CompletedAt is the date and time when the milestone was completed, if it
	// was completed.
	CompletedAt *time.Time `json:"completedOn"`

	// CompletedBy is the ID of the user who completed the milestone, if it was
	// completed.
	CompletedBy *int64 `json:"completedBy"`

	// Completed indicates whether the milestone is completed or not.
	Completed bool `json:"completed"`

	// Status is the status of the milestone. It can be "new", "reopened",
	// "completed" or "deleted".
	Status string `json:"status"`
}

// MilestoneUpdateRequestPath contains the path parameters for creating a
// milestone.
type MilestoneCreateRequestPath struct {
	// ProjectID is the unique identifier of the project that will contain the
	// milestone.
	ProjectID int64
}

// MilestoneCreateRequest represents the request body for creating a new
// milestone.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/task-lists/post-projects-id-milestones-json
type MilestoneCreateRequest struct {
	// Path contains the path parameters for the request.
	Path MilestoneCreateRequestPath `json:"-"`

	// Name is the name of the milestone.
	Name string `json:"title"`

	// Description is an optional description of the milestone.
	Description *string `json:"description,omitempty"`

	// DueAt is the due date of the milestone.
	DueAt LegacyDate `json:"deadline"`

	// TasklistIDs is an optional list of tasklist IDs to associate with the
	// milestone.
	TasklistIDs []int64 `json:"tasklistIds,omitempty"`

	// TagIDs is an optional list of tag IDs to associate with the milestone.
	TagIDs []int64 `json:"tagIds,omitempty"`

	// Assignees is a list of users, companies and teams responsible for the
	// milestone.
	Assignees LegacyUserGroups `json:"responsible-party-ids"`
}

// NewMilestoneCreateRequest creates a new MilestoneCreateRequest with the
// provided required fields.
func NewMilestoneCreateRequest(
	projectID int64,
	name string,
	dueAt LegacyDate,
	assignees LegacyUserGroups,
) MilestoneCreateRequest {
	return MilestoneCreateRequest{
		Path: MilestoneCreateRequestPath{
			ProjectID: projectID,
		},
		Name:      name,
		DueAt:     dueAt,
		Assignees: assignees,
	}
}

// HTTPRequest creates an HTTP request for the MilestoneCreateRequest.
func (m MilestoneCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/%d/milestones.json", server, m.Path.ProjectID)

	payload := struct {
		Milestone MilestoneCreateRequest `json:"milestone"`
	}{Milestone: m}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create milestone request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// MilestoneCreateResponse represents the response body for creating a new
// milestone.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/task-lists/post-projects-id-milestones-json
type MilestoneCreateResponse struct {
	// ID is the unique identifier of the created milestone.
	ID LegacyNumber `json:"milestoneId"`
}

// HandleHTTPResponse handles the HTTP response for the MilestoneCreateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (m *MilestoneCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create milestone")
	}
	if err := json.NewDecoder(resp.Body).Decode(m); err != nil {
		return fmt.Errorf("failed to decode create milestone response: %w", err)
	}
	if m.ID == 0 {
		return fmt.Errorf("create milestone response does not contain a valid identifier")
	}
	return nil
}

// MilestoneCreate creates a new milestone using the provided request and returns
// the response.
func MilestoneCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req MilestoneCreateRequest,
) (*MilestoneCreateResponse, error) {
	return twapi.Execute[MilestoneCreateRequest, *MilestoneCreateResponse](ctx, engine, req)
}

// MilestoneUpdateRequestPath contains the path parameters for updating a
// milestone.
type MilestoneUpdateRequestPath struct {
	// ID is the unique identifier of the milestone to be updated.
	ID int64
}

// MilestoneUpdateRequest represents the request body for updating a milestone.
// Besides the identifier, all other fields are optional. When a field is not
// provided, it will not be modified.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/task-lists/put-milestones-id-json
type MilestoneUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path MilestoneUpdateRequestPath `json:"-"`

	// Name is the name of the milestone.
	Name *string `json:"title,omitempty"`

	// Description is an optional description of the milestone.
	Description *string `json:"description,omitempty"`

	// DueAt is the due date of the milestone.
	DueAt *LegacyDate `json:"deadline,omitempty"`

	// TasklistIDs is an optional list of tasklist IDs to associate with the
	// milestone.
	TasklistIDs []int64 `json:"tasklistIds,omitempty"`

	// TagIDs is an optional list of tag IDs to associate with the milestone.
	TagIDs []int64 `json:"tagIds,omitempty"`

	// Assignees is a list of users, companies and teams responsible for the
	// milestone.
	Assignees *LegacyUserGroups `json:"responsible-party-ids,omitempty"`
}

// NewMilestoneUpdateRequest creates a new MilestoneUpdateRequest with the
// provided milestone ID. The ID is required to update a milestone.
func NewMilestoneUpdateRequest(milestoneID int64) MilestoneUpdateRequest {
	return MilestoneUpdateRequest{
		Path: MilestoneUpdateRequestPath{
			ID: milestoneID,
		},
	}
}

// HTTPRequest creates an HTTP request for the MilestoneUpdateRequest.
func (m MilestoneUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/milestones/" + strconv.FormatInt(m.Path.ID, 10) + ".json"

	payload := struct {
		Milestone MilestoneUpdateRequest `json:"milestone"`
	}{Milestone: m}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update milestone request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// MilestoneUpdateResponse represents the response body for updating a milestone.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/task-lists/put-milestones-id-json
type MilestoneUpdateResponse struct{}

// HandleHTTPResponse handles the HTTP response for the MilestoneUpdateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (m *MilestoneUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update milestone")
	}
	if err := json.NewDecoder(resp.Body).Decode(m); err != nil {
		return fmt.Errorf("failed to decode update milestone response: %w", err)
	}
	return nil
}

// MilestoneUpdate updates a milestone using the provided request and returns
// the response.
func MilestoneUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req MilestoneUpdateRequest,
) (*MilestoneUpdateResponse, error) {
	return twapi.Execute[MilestoneUpdateRequest, *MilestoneUpdateResponse](ctx, engine, req)
}

// MilestoneDeleteRequestPath contains the path parameters for deleting a
// milestone.
type MilestoneDeleteRequestPath struct {
	// ID is the unique identifier of the milestone to be deleted.
	ID int64
}

// MilestoneDeleteRequest represents the request body for deleting a milestone.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/task-lists/delete-milestones-id-json
type MilestoneDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path MilestoneDeleteRequestPath
}

// NewMilestoneDeleteRequest creates a new MilestoneDeleteRequest with the
// provided milestone ID.
func NewMilestoneDeleteRequest(milestoneID int64) MilestoneDeleteRequest {
	return MilestoneDeleteRequest{
		Path: MilestoneDeleteRequestPath{
			ID: milestoneID,
		},
	}
}

// HTTPRequest creates an HTTP request for the MilestoneDeleteRequest.
func (m MilestoneDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/milestones/" + strconv.FormatInt(m.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// MilestoneDeleteResponse represents the response body for deleting a milestone.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/task-lists/delete-milestones-id-json
type MilestoneDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the MilestoneDeleteResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (m *MilestoneDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to delete milestone")
	}
	if err := json.NewDecoder(resp.Body).Decode(m); err != nil {
		return fmt.Errorf("failed to decode delete milestone response: %w", err)
	}
	return nil
}

// MilestoneDelete deletes a milestone using the provided request and returns
// the response.
func MilestoneDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req MilestoneDeleteRequest,
) (*MilestoneDeleteResponse, error) {
	return twapi.Execute[MilestoneDeleteRequest, *MilestoneDeleteResponse](ctx, engine, req)
}

// MilestoneGetRequestPath contains the path parameters for loading a single
// milestone.
type MilestoneGetRequestPath struct {
	// ID is the unique identifier of the milestone to be retrieved.
	ID int64 `json:"id"`
}

// MilestoneGetRequest represents the request body for loading a single milestone.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/milestones/get-projects-api-v3-milestones-mileston-id-json
type MilestoneGetRequest struct {
	// Path contains the path parameters for the request.
	Path MilestoneGetRequestPath
}

// NewMilestoneGetRequest creates a new MilestoneGetRequest with the provided
// milestone ID. The ID is required to load a milestone.
func NewMilestoneGetRequest(milestoneID int64) MilestoneGetRequest {
	return MilestoneGetRequest{
		Path: MilestoneGetRequestPath{
			ID: milestoneID,
		},
	}
}

// HTTPRequest creates an HTTP request for the MilestoneGetRequest.
func (m MilestoneGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/milestones/" + strconv.FormatInt(m.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// MilestoneGetResponse contains all the information related to a milestone.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/milestones/get-projects-api-v3-milestones-mileston-id-json
type MilestoneGetResponse struct {
	Milestone Milestone `json:"milestone"`
}

// HandleHTTPResponse handles the HTTP response for the MilestoneGetResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (m *MilestoneGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve milestone")
	}

	if err := json.NewDecoder(resp.Body).Decode(m); err != nil {
		return fmt.Errorf("failed to decode retrieve milestone response: %w", err)
	}
	return nil
}

// MilestoneGet retrieves a single milestone using the provided request and
// returns the response.
func MilestoneGet(
	ctx context.Context,
	engine *twapi.Engine,
	req MilestoneGetRequest,
) (*MilestoneGetResponse, error) {
	return twapi.Execute[MilestoneGetRequest, *MilestoneGetResponse](ctx, engine, req)
}

// MilestoneListRequestPath contains the path parameters for loading multiple
// milestones.
type MilestoneListRequestPath struct {
	// ProjectID is the unique identifier of the project whose milestones are to
	// be retrieved.
	ProjectID int64
}

// MilestoneListRequestFilters contains the filters for loading multiple
// milestones.
type MilestoneListRequestFilters struct {
	// SearchTerm is an optional search term to filter milestones by name.
	SearchTerm string

	// TagIDs is an optional list of tag IDs to filter milestones by tags.
	TagIDs []int64

	// MatchAllTags is an optional flag to indicate if all tags must match. If set
	// to true, only milestones matching all specified tags will be returned.
	MatchAllTags *bool

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of milestones to retrieve per page. Defaults to 50.
	PageSize int64
}

// MilestoneListRequest represents the request body for loading multiple milestones.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/milestones/get-projects-api-v3-milestones-json
// https://apidocs.teamwork.com/docs/teamwork/v3/milestones/get-projects-api-v3-projects-project-id-milestones-json
type MilestoneListRequest struct {
	// Path contains the path parameters for the request.
	Path MilestoneListRequestPath

	// Filters contains the filters for loading multiple milestones.
	Filters MilestoneListRequestFilters
}

// NewMilestoneListRequest creates a new MilestoneListRequest with default values.
func NewMilestoneListRequest() MilestoneListRequest {
	return MilestoneListRequest{
		Filters: MilestoneListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the MilestoneListRequest.
func (m MilestoneListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	var uri string
	switch {
	case m.Path.ProjectID > 0:
		uri = fmt.Sprintf("%s/projects/api/v3/projects/%d/milestones.json", server, m.Path.ProjectID)
	default:
		uri = server + "/projects/api/v3/milestones.json"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	query := req.URL.Query()
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
	if m.Filters.Page > 0 {
		query.Set("page", strconv.FormatInt(m.Filters.Page, 10))
	}
	if m.Filters.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(m.Filters.PageSize, 10))
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// MilestoneListResponse contains information by multiple milestones matching the
// request filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/milestones/get-projects-api-v3-milestones-json
// https://apidocs.teamwork.com/docs/teamwork/v3/milestones/get-projects-api-v3-projects-project-id-milestones-json
type MilestoneListResponse struct {
	request MilestoneListRequest

	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`
	Milestones []Milestone `json:"milestones"`
}

// HandleHTTPResponse handles the HTTP response for the MilestoneListResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (m *MilestoneListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list milestones")
	}

	if err := json.NewDecoder(resp.Body).Decode(m); err != nil {
		return fmt.Errorf("failed to decode list milestones response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (m *MilestoneListResponse) SetRequest(req MilestoneListRequest) {
	m.request = req
}

// Iterate returns the request set to the next page, if available. If there
// are no more pages, a nil request is returned.
func (m *MilestoneListResponse) Iterate() *MilestoneListRequest {
	if !m.Meta.Page.HasMore {
		return nil
	}
	req := m.request
	req.Filters.Page++
	return &req
}

// MilestoneList retrieves multiple milestones using the provided request and
// returns the response.
func MilestoneList(
	ctx context.Context,
	engine *twapi.Engine,
	req MilestoneListRequest,
) (*MilestoneListResponse, error) {
	return twapi.Execute[MilestoneListRequest, *MilestoneListResponse](ctx, engine, req)
}
