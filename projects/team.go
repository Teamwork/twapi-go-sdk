package projects

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var (
	_ twapi.HTTPRequester = (*TeamCreateRequest)(nil)
	_ twapi.HTTPResponser = (*TeamCreateResponse)(nil)
	_ twapi.HTTPRequester = (*TeamUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*TeamUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*TeamDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*TeamDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*TeamGetRequest)(nil)
	_ twapi.HTTPResponser = (*TeamGetResponse)(nil)
	_ twapi.HTTPRequester = (*TeamListRequest)(nil)
	_ twapi.HTTPResponser = (*TeamListResponse)(nil)
)

// Team is a group of users who are organized together to collaborate more
// efficiently on projects and tasks. Teams help structure work by grouping
// individuals with similar roles, responsibilities, or departmental functions,
// making it easier to assign work, track progress, and manage communication. By
// using teams, organizations can streamline project planning and ensure the
// right people are involved in the right parts of a project, enhancing clarity
// and accountability across the platform.
type Team struct {
	// ID is the unique identifier of the team.
	ID LegacyNumber `json:"id"`

	// Name is the name of the team.
	Name string `json:"name"`

	// Description is an optional description of the team.
	Description *string `json:"description"`

	// Handle is the unique handle of the team, used in mentions.
	Handle string `json:"handle"`

	// LogoURL is the URL of the team's logo image.
	LogoURL *string `json:"logoUrl"`

	// LogoIcon is the icon of the team's logo, if available.
	LogoIcon *string `json:"logoIcon"`

	// LogoColor is the color of the team's logo, if available.
	LogoColor *string `json:"logoColor"`

	// ProjectID is the unique identifier of the project this team belongs to.
	// This is only set when the team is a project team.
	ProjectID LegacyNumber `json:"projectId"`

	// Company is the client/company the team belongs to.
	Company *struct {
		ID   LegacyNumber `json:"id"`
		Name string       `json:"name"`
	} `json:"company"`

	// ParentTeam is the parent team of this team, if available.
	ParentTeam *struct {
		ID     LegacyNumber `json:"id"`
		Name   string       `json:"name"`
		Handle string       `json:"handle"`
	} `json:"parentTeam"`

	// RootTeam is the root team of this team, if available.
	RootTeam *struct {
		ID     LegacyNumber `json:"id"`
		Name   string       `json:"name"`
		Handle string       `json:"handle"`
	} `json:"rootTeam"`

	// Members is the list of members in this team.
	Members []LegacyRelationship `json:"members"`

	// CreatedBy is the team who created this team.
	CreatedBy LegacyNumber `json:"createdByUserId"`

	// CreatedAt is the date and time when the team was created.
	CreatedAt time.Time `json:"dateCreated"`

	// UpdatedBy is the team who last updated this team.
	UpdatedBy LegacyNumber `json:"updatedByUserId"`

	// UpdatedAt is the date and time when the team was last updated.
	UpdatedAt time.Time `json:"dateUpdated"`

	// Deleted indicates whether the team has been deleted.
	Deleted bool `json:"deleted"`

	// DeletedAt is the date and time when the team was deleted, if applicable.
	DeletedAt *twapi.OptionalDateTime `json:"deletedDate"`
}

// TeamCreateRequest represents the request body for creating a new
// team.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/teams/post-teams-json
type TeamCreateRequest struct {
	// Name is the name of the team.
	Name string `json:"name"`

	// Handle is the unique handle of the team, used in mentions. It must not have
	// spaces or special characters.
	Handle *string `json:"handle,omitempty"`

	// Description is an optional description of the team.
	Description *string `json:"description,omitempty"`

	// ParentTeamID is the unique identifier of the parent team. If not provided,
	// the team will be created as a root team.
	ParentTeamID *int64 `json:"parentTeamId,omitempty"`

	// Company is the client/company the team belongs to. By default is the same
	// from the logged team creating the new team.
	CompanyID *int64 `json:"companyId,omitempty"`

	// ProjectID is the unique identifier of the project this team belongs to.
	ProjectID *int64 `json:"projectId,omitempty"`

	// UserIDs is the list of user IDs to be added as members of the team.
	UserIDs LegacyNumericList `json:"userIds,omitempty"`
}

// NewTeamCreateRequest creates a new TeamCreateRequest with the provided name.
func NewTeamCreateRequest(name string) TeamCreateRequest {
	return TeamCreateRequest{
		Name: name,
	}
}

// HTTPRequest creates an HTTP request for the TeamCreateRequest.
func (u TeamCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/teams.json"

	payload := struct {
		Team TeamCreateRequest `json:"team"`
	}{Team: u}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create team request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// TeamCreateResponse represents the response body for creating a new team.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/teams/post-teams-json
type TeamCreateResponse struct {
	// ID is the unique identifier of the created team.
	ID LegacyNumber `json:"id"`
}

// HandleHTTPResponse handles the HTTP response for the TeamCreateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (u *TeamCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to create team")
	}
	if err := json.NewDecoder(resp.Body).Decode(u); err != nil {
		return fmt.Errorf("failed to decode create team response: %w", err)
	}
	if u.ID == 0 {
		return fmt.Errorf("create team response does not contain a valid identifier")
	}
	return nil
}

// TeamCreate creates a new team using the provided request and returns the
// response.
func TeamCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req TeamCreateRequest,
) (*TeamCreateResponse, error) {
	return twapi.Execute[TeamCreateRequest, *TeamCreateResponse](ctx, engine, req)
}

// TeamUpdateRequestPath contains the path parameters for updating a team.
type TeamUpdateRequestPath struct {
	// ID is the unique identifier of the team to be updated.
	ID int64
}

// TeamUpdateRequest represents the request body for updating a team. Besides
// the identifier, all other fields are optional. When a field is not provided,
// it will not be modified.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/teams/put-teams-id-json
type TeamUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path TeamUpdateRequestPath `json:"-"`

	// Name is the name of the team.
	Name *string `json:"name,omitempty"`

	// Handle is the unique handle of the team, used in mentions. It must not have
	// spaces or special characters.
	Handle *string `json:"handle,omitempty"`

	// Description is an optional description of the team.
	Description *string `json:"description,omitempty"`

	// CompanyID is the unique identifier of the company the team belongs to.
	CompanyID *int64 `json:"companyId,omitempty"`

	// ProjectID is the unique identifier of the project this team belongs to.
	ProjectID *int64 `json:"projectId,omitempty"`

	// UserIDs is the list of user IDs to be added as members of the team.
	UserIDs LegacyNumericList `json:"userIds,omitempty"`
}

// NewTeamUpdateRequest creates a new TeamUpdateRequest with the
// provided team ID. The ID is required to update a team.
func NewTeamUpdateRequest(teamID int64) TeamUpdateRequest {
	return TeamUpdateRequest{
		Path: TeamUpdateRequestPath{
			ID: teamID,
		},
	}
}

// HTTPRequest creates an HTTP request for the TeamUpdateRequest.
func (u TeamUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/teams/" + strconv.FormatInt(u.Path.ID, 10) + ".json"

	payload := struct {
		Team TeamUpdateRequest `json:"team"`
	}{Team: u}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update team request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// TeamUpdateResponse represents the response body for updating a team.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/teams/put-teams-id-json
type TeamUpdateResponse struct{}

// HandleHTTPResponse handles the HTTP response for the TeamUpdateResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (u *TeamUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update team")
	}
	if err := json.NewDecoder(resp.Body).Decode(u); err != nil {
		return fmt.Errorf("failed to decode update team response: %w", err)
	}
	return nil
}

// TeamUpdate updates a team using the provided request and returns the
// response.
func TeamUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req TeamUpdateRequest,
) (*TeamUpdateResponse, error) {
	return twapi.Execute[TeamUpdateRequest, *TeamUpdateResponse](ctx, engine, req)
}

// TeamDeleteRequestPath contains the path parameters for deleting a team.
type TeamDeleteRequestPath struct {
	// ID is the unique identifier of the team to be deleted.
	ID int64
}

// TeamDeleteRequest represents the request body for deleting a team.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/teams/delete-teams-id-json
type TeamDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path TeamDeleteRequestPath
}

// NewTeamDeleteRequest creates a new TeamDeleteRequest with the
// provided team ID.
func NewTeamDeleteRequest(teamID int64) TeamDeleteRequest {
	return TeamDeleteRequest{
		Path: TeamDeleteRequestPath{
			ID: teamID,
		},
	}
}

// HTTPRequest creates an HTTP request for the TeamDeleteRequest.
func (u TeamDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/teams/" + strconv.FormatInt(u.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// TeamDeleteResponse represents the response body for deleting a team.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/teams/delete-teams-id-json
type TeamDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the TeamDeleteResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (u *TeamDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to delete team")
	}
	if err := json.NewDecoder(resp.Body).Decode(u); err != nil {
		return fmt.Errorf("failed to decode delete team response: %w", err)
	}
	return nil
}

// TeamDelete deletes a team using the provided request and returns the
// response.
func TeamDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req TeamDeleteRequest,
) (*TeamDeleteResponse, error) {
	return twapi.Execute[TeamDeleteRequest, *TeamDeleteResponse](ctx, engine, req)
}

// TeamGetRequestPath contains the path parameters for loading a single team.
type TeamGetRequestPath struct {
	// ID is the unique identifier of the team to be retrieved.
	ID int64 `json:"id"`
}

// TeamGetRequest represents the request body for loading a single team.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/teams/get-teams-id-json
type TeamGetRequest struct {
	// Path contains the path parameters for the request.
	Path TeamGetRequestPath
}

// NewTeamGetRequest creates a new TeamGetRequest with the provided
// team ID. The ID is required to load a team.
func NewTeamGetRequest(teamID int64) TeamGetRequest {
	return TeamGetRequest{
		Path: TeamGetRequestPath{
			ID: teamID,
		},
	}
}

// HTTPRequest creates an HTTP request for the TeamGetRequest.
func (u TeamGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/teams/" + strconv.FormatInt(u.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// TeamGetResponse contains all the information related to a team.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/teams/get-teams-id-json
type TeamGetResponse struct {
	Team Team `json:"team"`
}

// HandleHTTPResponse handles the HTTP response for the TeamGetResponse. If some
// unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (u *TeamGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve team")
	}

	if err := json.NewDecoder(resp.Body).Decode(u); err != nil {
		return fmt.Errorf("failed to decode retrieve team response: %w", err)
	}
	return nil
}

// TeamGet retrieves a single team using the provided request and returns the
// response.
func TeamGet(
	ctx context.Context,
	engine *twapi.Engine,
	req TeamGetRequest,
) (*TeamGetResponse, error) {
	return twapi.Execute[TeamGetRequest, *TeamGetResponse](ctx, engine, req)
}

// TeamListRequestPath contains the path parameters for loading multiple teams.
type TeamListRequestPath struct {
	// ProjectID is the unique identifier of the project to load teams for.
	ProjectID int64

	// CompanyID is the unique identifier of the company to load teams for.
	CompanyID int64
}

// TeamListRequestFilters contains the filters for loading multiple
// teams.
type TeamListRequestFilters struct {
	// SearchTerm is an optional search term to filter teams by name or e-mail.
	SearchTerm string

	// IncludeCompanyTeams indicates whether to include client/company teams in
	// the response. By default client/company teams are not included.
	IncludeCompanyTeams bool

	// IncludeProjectTeams indicates whether to include project teams in the
	// response. By default project teams are not included.
	IncludeProjectTeams bool

	// IncludeSubteams indicates whether to include subteams in the response. By
	// default sub-teams are not included.
	IncludeSubteams bool

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of teams to retrieve per page. Defaults to 50.
	PageSize int64
}

// TeamListRequest represents the request body for loading multiple teams.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/teams/get-teams-json
// https://apidocs.teamwork.com/docs/teamwork/v1/teams/get-projects-id-teams-json
// https://apidocs.teamwork.com/docs/teamwork/v1/teams/get-companies-id-teams-json
type TeamListRequest struct {
	// Path contains the path parameters for the request.
	Path TeamListRequestPath

	// Filters contains the filters for loading multiple teams.
	Filters TeamListRequestFilters
}

// NewTeamListRequest creates a new TeamListRequest with default values.
func NewTeamListRequest() TeamListRequest {
	return TeamListRequest{
		Filters: TeamListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the TeamListRequest.
func (u TeamListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	var uri string
	switch {
	case u.Path.ProjectID > 0:
		uri = fmt.Sprintf("%s/projects/%d/teams.json", server, u.Path.ProjectID)
	case u.Path.CompanyID > 0:
		uri = fmt.Sprintf("%s/companies/%d/teams.json", server, u.Path.CompanyID)
	default:
		uri = server + "/teams.json"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	if u.Filters.SearchTerm != "" {
		query.Set("searchTerm", u.Filters.SearchTerm)
	}
	if u.Filters.IncludeCompanyTeams {
		query.Set("includeCompanyTeams", "true")
	}
	if u.Filters.IncludeProjectTeams {
		query.Set("includeProjectTeams", "true")
	}
	if u.Filters.IncludeSubteams {
		query.Set("includeSubteams", "true")
	}
	if u.Filters.Page > 0 {
		query.Set("page", strconv.FormatInt(u.Filters.Page, 10))
	}
	if u.Filters.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(u.Filters.PageSize, 10))
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// TeamListResponse contains information by multiple teams matching the request
// filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/teams/get-teams-json
// https://apidocs.teamwork.com/docs/teamwork/v1/teams/get-projects-id-teams-json
// https://apidocs.teamwork.com/docs/teamwork/v1/teams/get-companies-id-teams-json
type TeamListResponse struct {
	request TeamListRequest
	hasMore bool

	Teams []Team `json:"teams"`
}

// HandleHTTPResponse handles the HTTP response for the TeamListResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (u *TeamListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list teams")
	}

	page, _ := strconv.ParseInt(resp.Header.Get("X-Page"), 10, 64)
	pages, _ := strconv.ParseInt(resp.Header.Get("X-Pages"), 10, 64)
	u.hasMore = pages > page

	if err := json.NewDecoder(resp.Body).Decode(u); err != nil {
		return fmt.Errorf("failed to decode list teams response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (u *TeamListResponse) SetRequest(req TeamListRequest) {
	u.request = req
}

// Iterate returns the request set to the next page, if available. If there are
// no more pages, a nil request is returned.
func (u *TeamListResponse) Iterate() *TeamListRequest {
	if !u.hasMore {
		return nil
	}
	req := u.request
	req.Filters.Page++
	return &req
}

// TeamList retrieves multiple teams using the provided request and returns the
// response.
func TeamList(
	ctx context.Context,
	engine *twapi.Engine,
	req TeamListRequest,
) (*TeamListResponse, error) {
	return twapi.Execute[TeamListRequest, *TeamListResponse](ctx, engine, req)
}
