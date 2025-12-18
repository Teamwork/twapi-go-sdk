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
	_ twapi.HTTPRequester = (*SkillCreateRequest)(nil)
	_ twapi.HTTPResponser = (*SkillCreateResponse)(nil)
	_ twapi.HTTPRequester = (*SkillUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*SkillUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*SkillDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*SkillDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*SkillGetRequest)(nil)
	_ twapi.HTTPResponser = (*SkillGetResponse)(nil)
	_ twapi.HTTPRequester = (*SkillListRequest)(nil)
	_ twapi.HTTPResponser = (*SkillListResponse)(nil)
)

// Skill represents a specific capability, area of expertise, or proficiency
// that can be assigned to users to describe what they are good at or qualified
// to work on. Skills help teams understand the strengths available across the
// organization and make it easier to match the right skills to the right work
// when planning projects, assigning tasks, or managing resources. By
// associating skills with users and leveraging them in planning and reporting,
// Teamwork enables more effective workload distribution, better project
// outcomes, and clearer visibility into whether the team has the capabilities
// needed to deliver upcoming work.
//
// More information can be found at:
// https://support.teamwork.com/projects/planning/skills
type Skill struct {
	// ID is the unique identifier of the skill.
	ID int64 `json:"id"`

	// Name is the name of the skill.
	Name string `json:"name"`

	// CreatedByUserID is the user who created this skill.
	CreatedByUserID int64 `json:"createdByUser"`

	// CreatedAt is the date and time when the skill was created.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedByUserID is the user who last updated this skill.
	UpdatedByUserID *int64 `json:"updatedByUser"`

	// UpdatedAt is the date and time when the skill was last updated.
	UpdatedAt *time.Time `json:"updatedAt"`

	// DeletedByUserID is the user who deleted this skill.
	DeletedByUserID *int64 `json:"deletedByUser"`

	// DeletedAt is the date and time when the skill was deleted.
	DeletedAt *time.Time `json:"deletedAt"`

	// Users contains the list of users associated with this skill.
	Users []twapi.Relationship `json:"users"`
}

// SkillCreateRequest represents the request body for creating a new
// skill.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/skills/post-projects-api-v3-skills-json
type SkillCreateRequest struct {
	// Name is the name of the skill.
	Name string `json:"name"`

	// UserIDs contains the list of user IDs to be associated with this skill.
	UserIDs []int64 `json:"userIds"`
}

// NewSkillCreateRequest reates a new SkillCreateRequest with the provided name.
func NewSkillCreateRequest(name string) SkillCreateRequest {
	return SkillCreateRequest{
		Name: name,
	}
}

// HTTPRequest creates an HTTP request for the SkillCreateRequest.
func (s SkillCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/skills.json"

	payload := struct {
		Skill SkillCreateRequest `json:"skill"`
	}{Skill: s}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create skill request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// SkillCreateResponse represents the response body for creating a new skill.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/skills/post-projects-api-v3-skills-json
type SkillCreateResponse struct {
	// Skill contains the created skill information.
	Skill Skill `json:"skill"`
}

// HandleHTTPResponse handles the HTTP response for the SkillCreateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (s *SkillCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create skill")
	}
	if err := json.NewDecoder(resp.Body).Decode(s); err != nil {
		return fmt.Errorf("failed to decode create skill response: %w", err)
	}
	if s.Skill.ID == 0 {
		return fmt.Errorf("create skill response does not contain a valid identifier")
	}
	return nil
}

// SkillCreate creates a new skill using the provided request and returns the
// response.
func SkillCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req SkillCreateRequest,
) (*SkillCreateResponse, error) {
	return twapi.Execute[SkillCreateRequest, *SkillCreateResponse](ctx, engine, req)
}

// SkillUpdateRequestPath contains the path parameters for updating a skill.
type SkillUpdateRequestPath struct {
	// ID is the unique identifier of the skill to be updated.
	ID int64
}

// SkillUpdateRequest represents the request body for updating a skill. Besides
// the identifier, all other fields are optional. When a field is not provided,
// it will not be modified.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/skills/patch-projects-api-v3-skills-id-json
type SkillUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path SkillUpdateRequestPath `json:"-"`

	// Name is the name of the skill.
	Name *string `json:"name,omitempty"`

	// UserIDs contains the list of user IDs to be associated with this skill.
	UserIDs []int64 `json:"userIds,omitempty"`
}

// NewSkillUpdateRequest creates a new SkillUpdateRequest with the provided
// skill ID. The ID is required to update a skill.
func NewSkillUpdateRequest(skillID int64) SkillUpdateRequest {
	return SkillUpdateRequest{
		Path: SkillUpdateRequestPath{
			ID: skillID,
		},
	}
}

// HTTPRequest creates an HTTP request for the SkillUpdateRequest.
func (s SkillUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/skills/" + strconv.FormatInt(s.Path.ID, 10) + ".json"

	payload := struct {
		Skill SkillUpdateRequest `json:"skill"`
	}{Skill: s}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update skill request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// SkillUpdateResponse represents the response body for updating a skill.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/skills/patch-projects-api-v3-skills-id-json
type SkillUpdateResponse struct {
	// Skill contains the updated skill information.
	Skill Skill `json:"skill"`
}

// HandleHTTPResponse handles the HTTP response for the SkillUpdateResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (s *SkillUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update skill")
	}
	if err := json.NewDecoder(resp.Body).Decode(s); err != nil {
		return fmt.Errorf("failed to decode update skill response: %w", err)
	}
	return nil
}

// SkillUpdate updates a skill using the provided request and returns the
// response.
func SkillUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req SkillUpdateRequest,
) (*SkillUpdateResponse, error) {
	return twapi.Execute[SkillUpdateRequest, *SkillUpdateResponse](ctx, engine, req)
}

// SkillDeleteRequestPath contains the path parameters for deleting a skill.
type SkillDeleteRequestPath struct {
	// ID is the unique identifier of the skill to be deleted.
	ID int64
}

// SkillDeleteRequest represents the request body for deleting a skill.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/skills/delete-projects-api-v3-skills-id-json
type SkillDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path SkillDeleteRequestPath
}

// NewSkillDeleteRequest creates a new SkillDeleteRequest with the provided
// skill ID.
func NewSkillDeleteRequest(skillID int64) SkillDeleteRequest {
	return SkillDeleteRequest{
		Path: SkillDeleteRequestPath{
			ID: skillID,
		},
	}
}

// HTTPRequest creates an HTTP request for the SkillDeleteRequest.
func (u SkillDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/skills/" + strconv.FormatInt(u.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// SkillDeleteResponse represents the response body for deleting a skill.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/skills/delete-projects-api-v3-skills-id-json
type SkillDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the SkillDeleteResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (u *SkillDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusNoContent {
		return twapi.NewHTTPError(resp, "failed to delete skill")
	}
	return nil
}

// SkillDelete deletes a skill using the provided request and returns the
// response.
func SkillDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req SkillDeleteRequest,
) (*SkillDeleteResponse, error) {
	return twapi.Execute[SkillDeleteRequest, *SkillDeleteResponse](ctx, engine, req)
}

// SkillGetRequestPath contains the path parameters for loading a single skill.
type SkillGetRequestPath struct {
	// ID is the unique identifier of the skill to be retrieved.
	ID int64 `json:"id"`
}

// SkillGetRequest represents the request body for loading a single skill.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/skills/get-projects-api-v3-skills-skill-id-json
type SkillGetRequest struct {
	// Path contains the path parameters for the request.
	Path SkillGetRequestPath
}

// NewSkillGetRequest creates a new SkillGetRequest with the provided skill ID.
// The ID is required to load a skill.
func NewSkillGetRequest(skillID int64) SkillGetRequest {
	return SkillGetRequest{
		Path: SkillGetRequestPath{
			ID: skillID,
		},
	}
}

// HTTPRequest creates an HTTP request for the SkillGetRequest.
func (s SkillGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/skills/" + strconv.FormatInt(s.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// SkillGetResponse contains all the information related to a skill.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/skills/get-projects-api-v3-skills-skill-id-json
type SkillGetResponse struct {
	Skill Skill `json:"skill"`
}

// HandleHTTPResponse handles the HTTP response for the SkillGetResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (s *SkillGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve skill")
	}

	if err := json.NewDecoder(resp.Body).Decode(s); err != nil {
		return fmt.Errorf("failed to decode retrieve skill response: %w", err)
	}
	return nil
}

// SkillGet retrieves a single skill using the provided request and returns the
// response.
func SkillGet(
	ctx context.Context,
	engine *twapi.Engine,
	req SkillGetRequest,
) (*SkillGetResponse, error) {
	return twapi.Execute[SkillGetRequest, *SkillGetResponse](ctx, engine, req)
}

// SkillListRequestPath contains the path parameters for loading multiple
// skills.
type SkillListRequestPath struct{}

// SkillListRequestSideload contains the possible sideload options when loading
// multiple skills.
type SkillListRequestSideload string

// List of possible sideload options for SkillListRequestSideload.
const (
	SkillListRequestSideloadUsers SkillListRequestSideload = "users"
)

// SkillListRequestFilters contains the filters for loading multiple
// skills.
type SkillListRequestFilters struct {
	// SearchTerm is an optional search term to filter skills by name or assigned
	// users' names.
	SearchTerm string

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of skills to retrieve per page. Defaults to 50.
	PageSize int64

	// Include contains additional related information to include in the response
	// as a sideload.
	Include []SkillListRequestSideload
}

// SkillListRequest represents the request body for loading multiple skills.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/skills/get-projects-api-v3-skills-json
type SkillListRequest struct {
	// Path contains the path parameters for the request.
	Path SkillListRequestPath

	// Filters contains the filters for loading multiple skills.
	Filters SkillListRequestFilters
}

// NewSkillListRequest creates a new SkillListRequest with default values.
func NewSkillListRequest() SkillListRequest {
	return SkillListRequest{
		Filters: SkillListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the SkillListRequest.
func (s SkillListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/skills.json"

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

// SkillListResponse contains information by multiple skills matching the
// request filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/skills/get-projects-api-v3-skills-json
type SkillListResponse struct {
	request SkillListRequest

	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`
	Skills []Skill `json:"skills"`
}

// HandleHTTPResponse handles the HTTP response for the SkillListResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (s *SkillListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list skills")
	}

	if err := json.NewDecoder(resp.Body).Decode(s); err != nil {
		return fmt.Errorf("failed to decode list skills response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (s *SkillListResponse) SetRequest(req SkillListRequest) {
	s.request = req
}

// Iterate returns the request set to the next page, if available. If there are
// no more pages, a nil request is returned.
func (s *SkillListResponse) Iterate() *SkillListRequest {
	if !s.Meta.Page.HasMore {
		return nil
	}
	req := s.request
	req.Filters.Page++
	return &req
}

// SkillList retrieves multiple skills using the provided request and returns
// the response.
func SkillList(
	ctx context.Context,
	engine *twapi.Engine,
	req SkillListRequest,
) (*SkillListResponse, error) {
	return twapi.Execute[SkillListRequest, *SkillListResponse](ctx, engine, req)
}
