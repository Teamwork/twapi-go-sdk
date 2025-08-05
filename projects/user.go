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
	_ twapi.HTTPRequester = (*UserCreateRequest)(nil)
	_ twapi.HTTPResponser = (*UserCreateResponse)(nil)
	_ twapi.HTTPRequester = (*UserUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*UserUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*UserDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*UserDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*UserGetRequest)(nil)
	_ twapi.HTTPResponser = (*UserGetResponse)(nil)
	_ twapi.HTTPRequester = (*UserGetMeRequest)(nil)
	_ twapi.HTTPResponser = (*UserGetMeResponse)(nil)
	_ twapi.HTTPRequester = (*UserListRequest)(nil)
	_ twapi.HTTPResponser = (*UserListResponse)(nil)
)

// User is an individual who has access to one or more projects within a
// Teamwork site, typically as a team member, collaborator, or administrator.
// Users can be assigned tasks, participate in discussions, log time, share
// files, and interact with other members depending on their permission levels.
// Each user has a unique profile that defines their role, visibility, and
// access to features and project data. Users can belong to clients/companies or
// teams within the system, and their permissions can be customized to control
// what actions they can perform or what information they can see.
//
// More information can be found at:
// https://support.teamwork.com/projects/getting-started/people-overview
type User struct {
	// ID is the unique identifier of the user.
	ID int64 `json:"id"`

	// FirstName is the first name of the user.
	FirstName string `json:"firstName"`

	// LastName is the last name of the user.
	LastName string `json:"lastName"`

	// Title is the title of the user (e.g. "Senior Developer").
	Title *string `json:"title"`

	// Email is the email address of the user.
	Email string `json:"email"`

	// Admin indicates whether the user is an administrator.
	Admin bool `json:"isAdmin"`

	// Type is the type of user. Possible values are "account", "collaborator" or "contact".
	Type string `json:"type"`

	// Cost is the hourly cost, to your company, to employ this user.
	Cost *twapi.Money `json:"userCost"`

	// Rate is the individual's hourly rate. This is what you charge for someone's
	// time on a project.
	Rate *twapi.Money `json:"userRate"`

	// Company is the client/company the user belongs to.
	Company twapi.Relationship `json:"company"`

	// JobRoles are the job roles assigned to the user.
	JobRoles []twapi.Relationship `json:"jobRoles,omitempty"`

	// Skills are the skills assigned to the user.
	Skills []twapi.Relationship `json:"skills,omitempty"`

	// Deleted indicates whether the user has been deleted.
	Deleted bool `json:"deleted"`

	// CreatedBy is the user who created this user.
	CreatedBy *twapi.Relationship `json:"createdBy"`

	// CreatedAt is the date and time when the user was created.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedBy is the user who last updated this user.
	UpdatedBy *twapi.Relationship `json:"updatedBy"`

	// UpdatedAt is the date and time when the user was last updated.
	UpdatedAt *time.Time `json:"updatedAt"`
}

// UserCreateRequest represents the request body for creating a new
// user.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/people/post-people-json
type UserCreateRequest struct {
	// FirstName is the first name of the user.
	FirstName string `json:"first-name"`

	// LastName is the last name of the user.
	LastName string `json:"last-name"`

	// Title is the title of the user (e.g. "Senior Developer").
	Title *string `json:"title,omitempty"`

	// Email is the email address of the user.
	Email string `json:"email-address"`

	// Admin indicates whether the user is an administrator. By default it is
	// false.
	Admin *bool `json:"administrator,omitempty"`

	// Type is the type of user. Possible values are "account", "collaborator" or
	// "contact". By default it is "account".
	Type *string `json:"user-type,omitempty"`

	// Company is the client/company the user belongs to. By default is the same
	// from the logged user creating the new user.
	CompanyID *int64 `json:"company-id,omitempty"`
}

// NewUserCreateRequest creates a new UserCreateRequest with the
// provided name in a specific project.
func NewUserCreateRequest(firstName, lastName, email string) UserCreateRequest {
	return UserCreateRequest{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}
}

// HTTPRequest creates an HTTP request for the UserCreateRequest.
func (u UserCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/people.json"

	payload := struct {
		User UserCreateRequest `json:"person"`
	}{User: u}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create user request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// UserCreateResponse represents the response body for creating a new
// user.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/people/post-people-json
type UserCreateResponse struct {
	// ID is the unique identifier of the created user.
	ID LegacyNumber `json:"id"`
}

// HandleHTTPResponse handles the HTTP response for the UserCreateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (u *UserCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create user")
	}
	if err := json.NewDecoder(resp.Body).Decode(u); err != nil {
		return fmt.Errorf("failed to decode create user response: %w", err)
	}
	if u.ID == 0 {
		return fmt.Errorf("create user response does not contain a valid identifier")
	}
	return nil
}

// UserCreate creates a new user using the provided request and returns
// the response.
func UserCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req UserCreateRequest,
) (*UserCreateResponse, error) {
	return twapi.Execute[UserCreateRequest, *UserCreateResponse](ctx, engine, req)
}

// UserUpdateRequestPath contains the path parameters for updating a user.
type UserUpdateRequestPath struct {
	// ID is the unique identifier of the user to be updated.
	ID int64
}

// UserUpdateRequest represents the request body for updating a user. Besides
// the identifier, all other fields are optional. When a field is not provided,
// it will not be modified.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/people/put-people-id-json
type UserUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path UserUpdateRequestPath `json:"-"`

	// FirstName is the first name of the user.
	FirstName *string `json:"first-name,omitempty"`

	// LastName is the last name of the user.
	LastName *string `json:"last-name,omitempty"`

	// Title is the title of the user (e.g. "Senior Developer").
	Title *string `json:"title,omitempty"`

	// Email is the email address of the user.
	Email *string `json:"email-address,omitempty"`

	// Admin indicates whether the user is an administrator. By default it is
	// false.
	Admin *bool `json:"administrator,omitempty"`

	// Type is the type of user. Possible values are "account", "collaborator" or
	// "contact". By default it is "account".
	Type *string `json:"user-type,omitempty"`

	// Company is the client/company the user belongs to. By default is the same
	// from the logged user creating the new user.
	CompanyID *int64 `json:"company-id,omitempty"`
}

// NewUserUpdateRequest creates a new UserUpdateRequest with the
// provided user ID. The ID is required to update a user.
func NewUserUpdateRequest(userID int64) UserUpdateRequest {
	return UserUpdateRequest{
		Path: UserUpdateRequestPath{
			ID: userID,
		},
	}
}

// HTTPRequest creates an HTTP request for the UserUpdateRequest.
func (u UserUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/people/" + strconv.FormatInt(u.Path.ID, 10) + ".json"

	payload := struct {
		User UserUpdateRequest `json:"person"`
	}{User: u}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update user request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// UserUpdateResponse represents the response body for updating a user.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/people/put-people-id-json
type UserUpdateResponse struct{}

// HandleHTTPResponse handles the HTTP response for the UserUpdateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (u *UserUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update user")
	}
	if err := json.NewDecoder(resp.Body).Decode(u); err != nil {
		return fmt.Errorf("failed to decode update user response: %w", err)
	}
	return nil
}

// UserUpdate updates a user using the provided request and returns the
// response.
func UserUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req UserUpdateRequest,
) (*UserUpdateResponse, error) {
	return twapi.Execute[UserUpdateRequest, *UserUpdateResponse](ctx, engine, req)
}

// UserDeleteRequestPath contains the path parameters for deleting a user.
type UserDeleteRequestPath struct {
	// ID is the unique identifier of the user to be deleted.
	ID int64
}

// UserDeleteRequest represents the request body for deleting a user.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/people/delete-people-id-json
type UserDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path UserDeleteRequestPath
}

// NewUserDeleteRequest creates a new UserDeleteRequest with the
// provided user ID.
func NewUserDeleteRequest(userID int64) UserDeleteRequest {
	return UserDeleteRequest{
		Path: UserDeleteRequestPath{
			ID: userID,
		},
	}
}

// HTTPRequest creates an HTTP request for the UserDeleteRequest.
func (u UserDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/people/" + strconv.FormatInt(u.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// UserDeleteResponse represents the response body for deleting a user.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/people/delete-people-id-json
type UserDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the UserDeleteResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (u *UserDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to delete user")
	}
	if err := json.NewDecoder(resp.Body).Decode(u); err != nil {
		return fmt.Errorf("failed to decode delete user response: %w", err)
	}
	return nil
}

// UserDelete deletes a user using the provided request and returns the
// response.
func UserDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req UserDeleteRequest,
) (*UserDeleteResponse, error) {
	return twapi.Execute[UserDeleteRequest, *UserDeleteResponse](ctx, engine, req)
}

// UserGetRequestPath contains the path parameters for loading a single user.
type UserGetRequestPath struct {
	// ID is the unique identifier of the user to be retrieved.
	ID int64 `json:"id"`
}

// UserGetRequest represents the request body for loading a single user.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/person/get-projects-api-v3-people-person-id-json
type UserGetRequest struct {
	// Path contains the path parameters for the request.
	Path UserGetRequestPath
}

// NewUserGetRequest creates a new UserGetRequest with the provided
// user ID. The ID is required to load a user.
func NewUserGetRequest(userID int64) UserGetRequest {
	return UserGetRequest{
		Path: UserGetRequestPath{
			ID: userID,
		},
	}
}

// HTTPRequest creates an HTTP request for the UserGetRequest.
func (u UserGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/people/" + strconv.FormatInt(u.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// UserGetResponse contains all the information related to a user.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/person/get-projects-api-v3-people-person-id-json
type UserGetResponse struct {
	User User `json:"person"`
}

// HandleHTTPResponse handles the HTTP response for the UserGetResponse. If some
// unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (u *UserGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve user")
	}

	if err := json.NewDecoder(resp.Body).Decode(u); err != nil {
		return fmt.Errorf("failed to decode retrieve user response: %w", err)
	}
	return nil
}

// UserGet retrieves a single user using the provided request and returns the
// response.
func UserGet(
	ctx context.Context,
	engine *twapi.Engine,
	req UserGetRequest,
) (*UserGetResponse, error) {
	return twapi.Execute[UserGetRequest, *UserGetResponse](ctx, engine, req)
}

// UserGetMeRequest represents the request body for loading the logged user.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/people/get-projects-api-v3-me-json
type UserGetMeRequest struct {
}

// NewUserGetMeRequest creates a new UserGetMeRequest.
func NewUserGetMeRequest() UserGetMeRequest {
	return UserGetMeRequest{}
}

// HTTPRequest creates an HTTP request for the UserGetMeRequest.
func (u UserGetMeRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/me.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// UserGetMeResponse contains all the information related to the logged user.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/people/get-projects-api-v3-me-json
type UserGetMeResponse struct {
	User User `json:"person"`
}

// HandleHTTPResponse handles the HTTP response for the UserGetMeResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (u *UserGetMeResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve user")
	}

	if err := json.NewDecoder(resp.Body).Decode(u); err != nil {
		return fmt.Errorf("failed to decode retrieve user response: %w", err)
	}
	return nil
}

// UserGetMe retrieves the logged user using the provided request and returns
// the response.
func UserGetMe(
	ctx context.Context,
	engine *twapi.Engine,
	req UserGetMeRequest,
) (*UserGetMeResponse, error) {
	return twapi.Execute[UserGetMeRequest, *UserGetMeResponse](ctx, engine, req)
}

// UserListRequestPath contains the path parameters for loading multiple users.
type UserListRequestPath struct {
	// ProjectID is the unique identifier of the project whose members are to be
	// retrieved.
	ProjectID int64
}

// UserListRequestFilters contains the filters for loading multiple
// users.
type UserListRequestFilters struct {
	// SearchTerm is an optional search term to filter users by name or e-mail.
	SearchTerm string

	// Type is an optional filter to load only users of a specific type. Possible
	// values are "account", "collaborator" or "contact".
	Type string

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of users to retrieve per page. Defaults to 50.
	PageSize int64
}

// UserListRequest represents the request body for loading multiple users.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/people/get-projects-api-v3-people-json
// https://apidocs.teamwork.com/docs/teamwork/v3/people/get-projects-api-v3-projects-project-id-people-json
type UserListRequest struct {
	// Path contains the path parameters for the request.
	Path UserListRequestPath

	// Filters contains the filters for loading multiple users.
	Filters UserListRequestFilters
}

// NewUserListRequest creates a new UserListRequest with default values.
func NewUserListRequest() UserListRequest {
	return UserListRequest{
		Filters: UserListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the UserListRequest.
func (u UserListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	var uri string
	switch {
	case u.Path.ProjectID > 0:
		uri = fmt.Sprintf("%s/projects/api/v3/projects/%d/people.json", server, u.Path.ProjectID)
	default:
		uri = server + "/projects/api/v3/people.json"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	if u.Filters.SearchTerm != "" {
		query.Set("searchTerm", u.Filters.SearchTerm)
	}
	if u.Filters.Type != "" {
		query.Set("userType", u.Filters.Type)
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

// UserListResponse contains information by multiple users matching the request
// filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/people/get-projects-api-v3-people-json
// https://apidocs.teamwork.com/docs/teamwork/v3/people/get-projects-api-v3-projects-project-id-people-json
type UserListResponse struct {
	request UserListRequest

	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`
	Users []User `json:"people"`
}

// HandleHTTPResponse handles the HTTP response for the UserListResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (u *UserListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list users")
	}

	if err := json.NewDecoder(resp.Body).Decode(u); err != nil {
		return fmt.Errorf("failed to decode list users response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (u *UserListResponse) SetRequest(req UserListRequest) {
	u.request = req
}

// Iterate returns the request set to the next page, if available. If there are
// no more pages, a nil request is returned.
func (u *UserListResponse) Iterate() *UserListRequest {
	if !u.Meta.Page.HasMore {
		return nil
	}
	req := u.request
	req.Filters.Page++
	return &req
}

// UserList retrieves multiple users using the provided request and returns the
// response.
func UserList(
	ctx context.Context,
	engine *twapi.Engine,
	req UserListRequest,
) (*UserListResponse, error) {
	return twapi.Execute[UserListRequest, *UserListResponse](ctx, engine, req)
}
