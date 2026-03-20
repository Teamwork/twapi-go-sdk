package projects

import (
	"context"
	"net/http"
	"net/url"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var (
	_ twapi.HTTPRequester = (*ProjectTemplateCreateRequest)(nil)
	_ twapi.HTTPRequester = (*ProjectTemplateListRequest)(nil)
)

// ProjectTemplateCreateRequest represents the request body for creating a new
// project template.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/project-templates/post-projects-template-json
type ProjectTemplateCreateRequest struct {
	ProjectCreateRequest
}

// NewProjectTemplateCreateRequest creates a new ProjectTemplateCreateRequest with the
// provided name. The name is required to create a new project template.
func NewProjectTemplateCreateRequest(name string) ProjectTemplateCreateRequest {
	return ProjectTemplateCreateRequest{
		ProjectCreateRequest: NewProjectCreateRequest(name),
	}
}

// HTTPRequest creates an HTTP request for the ProjectTemplateCreateRequest.
func (p ProjectTemplateCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	req, err := p.ProjectCreateRequest.HTTPRequest(ctx, server)
	if err != nil {
		return nil, err
	}

	req.URL, err = url.Parse(server + "/projects/template.json")
	if err != nil {
		return nil, err
	}

	return req, nil
}

// ProjectTemplateCreate creates a new project template using the provided
// request and returns the response.
func ProjectTemplateCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectTemplateCreateRequest,
) (*ProjectCreateResponse, error) {
	return twapi.Execute[ProjectTemplateCreateRequest, *ProjectCreateResponse](ctx, engine, req)
}

// ProjectTemplateListRequest represents the request body for loading multiple
// project templates.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/projects/get-projects-api-v3-projects-templates-json
type ProjectTemplateListRequest struct {
	ProjectListRequest
}

// NewProjectTemplateListRequest creates a new ProjectTemplateListRequest with
// default values.
func NewProjectTemplateListRequest() ProjectTemplateListRequest {
	return ProjectTemplateListRequest{
		ProjectListRequest: NewProjectListRequest(),
	}
}

// HTTPRequest creates an HTTP request for the ProjectTemplateListRequest.
func (p ProjectTemplateListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	req, err := p.ProjectListRequest.HTTPRequest(ctx, server)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	req.URL, err = url.Parse(server + "/projects/api/v3/projects/templates.json")
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// ProjectTemplateList retrieves multiple project templates using the provided
// request and returns the response.
func ProjectTemplateList(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectTemplateListRequest,
) (*ProjectListResponse, error) {
	return twapi.Execute[ProjectTemplateListRequest, *ProjectListResponse](ctx, engine, req)
}
