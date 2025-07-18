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
	_ twapi.HTTPRequester = (*ProjectMemberAddRequest)(nil)
	_ twapi.HTTPResponser = (*ProjectMemberAddResponse)(nil)
)

// ProjectMemberAddRequestPath contains the path parameters for adding users as
// project members.
type ProjectMemberAddRequestPath struct {
	ProjectID int64 `json:"projectId"`
}

// ProjectMemberAddRequest represents the request body for adding users as
// project members.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/people/put-projects-api-v3-projects-project-id-people-json
type ProjectMemberAddRequest struct {
	// Path contains the path parameters for the request.
	Path ProjectMemberAddRequestPath `json:"-"`

	// UserIDs is a list of user IDs to add as project members.
	UserIDs []int64 `json:"userIds"`
}

// NewProjectMemberAddRequest creates a new ProjectMemberAddRequest with the
// provided project and user IDs.
func NewProjectMemberAddRequest(projectID int64, userIDs ...int64) ProjectMemberAddRequest {
	return ProjectMemberAddRequest{
		Path: ProjectMemberAddRequestPath{
			ProjectID: projectID,
		},
		UserIDs: userIDs,
	}
}

// HTTPRequest creates an HTTP request for the ProjectMemberAddRequest.
func (u ProjectMemberAddRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/projects/" + strconv.FormatInt(u.Path.ProjectID, 10) + "/people.json"

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(u); err != nil {
		return nil, fmt.Errorf("failed to encode project members request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// ProjectMemberAddResponse represents the response body for adding users as
// project members.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/people/put-projects-api-v3-projects-project-id-people-json
type ProjectMemberAddResponse struct{}

// HandleHTTPResponse handles the HTTP response for the
// ProjectMemberAddResponse. If some unexpected HTTP status code is returned by
// the API, a twapi.HTTPError is returned.
func (u *ProjectMemberAddResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to add project members")
	}
	if err := json.NewDecoder(resp.Body).Decode(u); err != nil {
		return fmt.Errorf("failed to decode project members response: %w", err)
	}
	return nil
}

// ProjectMemberAdd adds users to a project.
func ProjectMemberAdd(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectMemberAddRequest,
) (*ProjectMemberAddResponse, error) {
	return twapi.Execute[ProjectMemberAddRequest, *ProjectMemberAddResponse](ctx, engine, req)
}
