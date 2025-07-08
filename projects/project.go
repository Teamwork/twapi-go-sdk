package projects

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var (
	_ twapi.HTTPRequester = (*CreateProjectRequest)(nil)
	_ twapi.HTTPResponser = (*CreateProjectResponse)(nil)
)

// CreateProjectRequest represents the request body for creating a new project.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/projects/post-projects-json
type CreateProjectRequest struct {
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

// HTTPRequest creates an HTTP request for the CreateProjectRequest.
func (c CreateProjectRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects.json"

	payload := struct {
		Project CreateProjectRequest `json:"project"`
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

// CreateProjectResponse represents the response body for creating a new
// project.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/projects/post-projects-json
type CreateProjectResponse struct {
	// ID is the unique identifier of the created project.
	ID LegacyNumber `json:"id"`
}

// HandleHTTPResponse handles the HTTP response for the CreateProjectResponse.
func (c *CreateProjectResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		if len(body) == 0 {
			body = []byte("no response body")
		}
		return fmt.Errorf("failed to create project (%q): %s", resp.Status, string(body))
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode create project response: %w", err)
	}
	if c.ID == 0 {
		return fmt.Errorf("create project response does not contain a valid identifier")
	}
	return nil
}

// CreateProject creates a new project using the provided request and returns
// the response.
func CreateProject(
	ctx context.Context,
	engine *twapi.Engine,
	req CreateProjectRequest,
) (*CreateProjectResponse, error) {
	return twapi.Execute[*CreateProjectResponse](ctx, engine, req)
}
