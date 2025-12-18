package projects

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var (
	_ twapi.HTTPRequester = (*UserAssignJobRoleRequest)(nil)
	_ twapi.HTTPResponser = (*UserAssignJobRoleResponse)(nil)
)

// UserAssignJobRoleRequestPath contains the path parameters for assigning users
// to a job role.
type UserAssignJobRoleRequestPath struct {
	// ID is the unique identifier of the job role.
	ID int64
}

// UserAssignJobRoleRequest represents the request body for assigning a job role
// to a user.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/job-roles/post-projects-api-v3-jobroles-id-people-json
type UserAssignJobRoleRequest struct {
	// Path contains the path parameters for the request.
	Path UserAssignJobRoleRequestPath `json:"-"`

	// IDs are the unique identifiers of the users to assign the job role to.
	IDs []int64 `json:"users"`

	// IsPrimary indicates whether this job role is the primary role for the
	// provided users.
	IsPrimary bool `json:"isPrimary"`
}

// NewUserAssignJobRoleRequest creates a new UserAssignJobRoleRequest with the
// provided job role ID.
func NewUserAssignJobRoleRequest(jobRoleID int64) UserAssignJobRoleRequest {
	return UserAssignJobRoleRequest{
		Path: UserAssignJobRoleRequestPath{
			ID: jobRoleID,
		},
	}
}

// HTTPRequest creates an HTTP request for the UserAssignJobRoleRequest.
func (s UserAssignJobRoleRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/jobroles/%d/people.json", server, s.Path.ID)

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(s); err != nil {
		return nil, fmt.Errorf("failed to encode create job role request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// UserAssignJobRoleResponse represents the response body for assigning a job
// role to users.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/job-roles/post-projects-api-v3-jobroles-id-people-json
type UserAssignJobRoleResponse struct {
	// JobRole contains the created job role information.
	JobRole JobRole `json:"jobRole"`
}

// HandleHTTPResponse handles the HTTP response for the JobRoleCreateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (s *UserAssignJobRoleResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to assign job role")
	}
	if err := json.NewDecoder(resp.Body).Decode(s); err != nil {
		return fmt.Errorf("failed to decode assign job role response: %w", err)
	}
	if s.JobRole.ID == 0 {
		return fmt.Errorf("assign job role response does not contain a valid identifier")
	}
	return nil
}

// UserAssignJobRole assigns a job role to users using the provided request and
// returns the response.
func UserAssignJobRole(
	ctx context.Context,
	engine *twapi.Engine,
	req UserAssignJobRoleRequest,
) (*UserAssignJobRoleResponse, error) {
	return twapi.Execute[UserAssignJobRoleRequest, *UserAssignJobRoleResponse](ctx, engine, req)
}

// UserUnassignJobRoleRequestPath contains the path parameters for unassigning
// users to a job role.
type UserUnassignJobRoleRequestPath struct {
	// ID is the unique identifier of the job role.
	ID int64
}

// UserUnassignJobRoleRequest represents the request body for unassigning a job
// role.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/job-roles/delete-projects-api-v3-jobroles-id-people-json
type UserUnassignJobRoleRequest struct {
	// Path contains the path parameters for the request.
	Path UserUnassignJobRoleRequestPath `json:"-"`

	// IDs are the unique identifiers of the users to unassign the job role to.
	IDs []int64 `json:"users"`

	// IsPrimary indicates whether this job role is the primary role for the
	// provided users.
	IsPrimary bool `json:"isPrimary"`
}

// NewUserUnassignJobRoleRequest creates a new UserUnassignJobRoleRequest with the
// provided job role ID.
func NewUserUnassignJobRoleRequest(jobRoleID int64) UserUnassignJobRoleRequest {
	return UserUnassignJobRoleRequest{
		Path: UserUnassignJobRoleRequestPath{
			ID: jobRoleID,
		},
	}
}

// HTTPRequest creates an HTTP request for the UserUnassignJobRoleRequest.
func (s UserUnassignJobRoleRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/jobroles/%d/people.json", server, s.Path.ID)

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(s); err != nil {
		return nil, fmt.Errorf("failed to encode create job role request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// UserUnassignJobRoleResponse represents the response body for unassigning a
// job role.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/job-roles/delete-projects-api-v3-jobroles-id-people-json
type UserUnassignJobRoleResponse struct{}

// HandleHTTPResponse handles the HTTP response for the JobRoleCreateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (s *UserUnassignJobRoleResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusNoContent {
		return twapi.NewHTTPError(resp, "failed to unassign job role")
	}
	return nil
}

// UserUnassignJobRole unassigns a job role from users using the provided
// request and returns the response.
func UserUnassignJobRole(
	ctx context.Context,
	engine *twapi.Engine,
	req UserUnassignJobRoleRequest,
) (*UserUnassignJobRoleResponse, error) {
	return twapi.Execute[UserUnassignJobRoleRequest, *UserUnassignJobRoleResponse](ctx, engine, req)
}
