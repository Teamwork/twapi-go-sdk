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
	_ twapi.HTTPRequester = (*ProjectCreateRequest)(nil)
	_ twapi.HTTPResponser = (*ProjectCreateResponse)(nil)
	_ twapi.HTTPRequester = (*ProjectUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*ProjectUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*ProjectDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*ProjectDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*ProjectGetRequest)(nil)
	_ twapi.HTTPResponser = (*ProjectGetResponse)(nil)
	_ twapi.HTTPRequester = (*ProjectListRequest)(nil)
	_ twapi.HTTPResponser = (*ProjectListResponse)(nil)
)

// ProjectStatus represents the status of a project. It can be "active",
// "archived" (also seen as "inactive") or "deleted". For update operations,
// only "active" and "archived" are valid values, as "deleted" is used to
// indicate a deleted project in the response.
type ProjectStatus string

const (
	// ProjectStatusActive indicates that the project is active and ongoing.
	ProjectStatusActive ProjectStatus = "active"

	// ProjectStatusArchived indicates that the project is archived and no longer
	// active, but still exists in the system.
	ProjectStatusArchived ProjectStatus = "archived"

	// ProjectStatusInactive is an alias for ProjectStatusArchived, as the API may
	// return either "archived" or "inactive" to indicate an archived project.
	ProjectStatusInactive ProjectStatus = "inactive"

	// ProjectStatusDeleted is only used in responses to indicate a deleted
	// project. It cannot be set in update requests, as deletion is performed
	// using the delete endpoint.
	ProjectStatusDeleted ProjectStatus = "deleted"
)

// Project serves as the central workspace for organizing and managing a
// specific piece of work or initiative. Each project provides a dedicated area
// where teams can plan tasks, assign responsibilities, set deadlines, and track
// progress toward shared goals. Projects include tools for communication, file
// sharing, milestones, and time tracking, allowing teams to stay aligned and
// informed throughout the entire lifecycle of the work. Whether it's a product
// launch, client engagement, or internal initiative, projects in Teamwork.com
// help teams structure their efforts, collaborate more effectively, and deliver
// results with greater visibility and accountability.
//
// More information can be found at:
// https://support.teamwork.com/projects/getting-started/projects-overview
//
// sparsefields:gen
type Project struct {
	// ID is the unique identifier of the project.
	ID int64 `json:"id"`

	// Description is an optional description of the project.
	Description *string `json:"description"`

	// Name is the name of the project.
	Name string `json:"name"`

	// StartAt is the start date of the project.
	StartAt *time.Time `json:"startAt"`

	// EndAt is the end date of the project.
	EndAt *time.Time `json:"endAt"`

	// Category is the category associated with the project.
	Category *twapi.Relationship `json:"category"`

	// Company is the company associated with the project.
	Company twapi.Relationship `json:"company"`

	// Owner is the user who owns the project.
	Owner *twapi.Relationship `json:"projectOwner"`

	// Tags is a list of tags associated with the project.
	Tags []twapi.Relationship `json:"tags"`

	// LastWorkedAt is the date and time when the project was last worked on.
	//
	// Attention: This field isn't populated when the project is retrieved as a
	// sideload.
	LastWorkedAt *time.Time `json:"lastWorkedOn"`

	// CreatedAt is the date and time when the project was created.
	CreatedAt *time.Time `json:"createdAt"`

	// CreatedBy is the ID of the user who created the project.
	CreatedBy *int64 `json:"createdBy"`

	// UpdatedAt is the date and time when the project was last updated.
	UpdatedAt *time.Time `json:"updatedAt"`

	// UpdatedBy is the ID of the user who last updated the project.
	UpdatedBy *int64 `json:"updatedBy"`

	// CompletedAt is the date and time when the project was completed.
	CompletedAt *time.Time `json:"completedAt"`

	// CompletedBy is the ID of the user who completed the project.
	CompletedBy *int64 `json:"completedBy"`

	// Status is the status of the project. It can be "active", "archived" or
	// "deleted".
	Status ProjectStatus `json:"status"`

	// IsBillable indicates whether time logged to the project is billable by
	// default.
	IsBillable *bool `json:"isBillable"`

	// Type is the type of the project. It can be "normal", "tasklists-template",
	// "projects-template", "personal", "holder-project", "tentative" or
	// "global-messages".
	Type string `json:"type"`
}

// ProjectCreateRequest represents the request body for creating a new project.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/projects/post-projects-json
type ProjectCreateRequest struct {
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

	// CategoryID is an optional ID of the category associated with the project.
	CategoryID *int64 `json:"category-id,omitempty"`

	// CompanyID is an optional ID of the company/client associated with the
	// project. By default it is the ID of the company of the logged user
	// creating the project.
	CompanyID int64 `json:"companyId"`

	// OwnerID is an optional ID of the user who owns the project. By default it
	// is the ID of the logged user creating the project.
	OwnerID *int64 `json:"projectOwnerId,omitempty"`

	// TagIDs is an optional list of tag IDs associated with the project.
	TagIDs []int64 `json:"tagIds,omitempty"`
}

// NewProjectCreateRequest creates a new ProjectCreateRequest with the
// provided name. The name is required to create a new project.
func NewProjectCreateRequest(name string) ProjectCreateRequest {
	return ProjectCreateRequest{Name: name}
}

// HTTPRequest creates an HTTP request for the ProjectCreateRequest.
func (p ProjectCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects.json"

	payload := struct {
		Project ProjectCreateRequest `json:"project"`
	}{Project: p}

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

// ProjectCreateResponse represents the response body for creating a new
// project.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/projects/post-projects-json
type ProjectCreateResponse struct {
	// ID is the unique identifier of the created project.
	ID LegacyNumber `json:"id"`
}

// HandleHTTPResponse handles the HTTP response for the ProjectCreateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (p *ProjectCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create project")
	}
	if err := json.NewDecoder(resp.Body).Decode(p); err != nil {
		return fmt.Errorf("failed to decode create project response: %w", err)
	}
	if p.ID == 0 {
		return fmt.Errorf("create project response does not contain a valid identifier")
	}
	return nil
}

// ProjectCreate creates a new project using the provided request and returns
// the response.
func ProjectCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectCreateRequest,
) (*ProjectCreateResponse, error) {
	return twapi.Execute[ProjectCreateRequest, *ProjectCreateResponse](ctx, engine, req)
}

// ProjectUpdateRequestPath contains the path parameters for updating a project.
type ProjectUpdateRequestPath struct {
	// ID is the unique identifier of the project to be updated.
	ID int64
}

// ProjectUpdateRequest represents the request body for updating a project.
// Besides the identifier, all other fields are optional. When a field is not
// provided, it will not be modified.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/projects/put-projects-id-json
type ProjectUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path ProjectUpdateRequestPath

	// Name is the name of the project.
	Name *string `json:"name,omitempty"`

	// Description is the project description.
	Description *string `json:"description,omitempty"`

	// StartAt is the start date for the project.
	StartAt *LegacyDate `json:"start-date,omitempty"`

	// EndAt is the end date for the project.
	EndAt *LegacyDate `json:"end-date,omitempty"`

	// CategoryID is the ID of the category associated with the project.
	CategoryID *int64 `json:"category-id,omitempty"`

	// CompanyID is the company/client associated with the project.
	CompanyID *int64 `json:"companyId,omitempty"`

	// OwnerID is the ID of the user who owns the project.
	OwnerID *int64 `json:"projectOwnerId,omitempty"`

	// TagIDs is the list of tag IDs associated with the project.
	TagIDs []int64 `json:"tagIds,omitempty"`

	// Status is the status of the project. It can be "active" or "archived". The
	// "deleted" status is not allowed in update requests, as deletion is
	// performed using the delete endpoint.
	Status *ProjectStatus `json:"status,omitempty"`
}

// NewProjectUpdateRequest creates a new ProjectUpdateRequest with the
// provided project ID. The ID is required to update a project.
func NewProjectUpdateRequest(projectID int64) ProjectUpdateRequest {
	return ProjectUpdateRequest{
		Path: ProjectUpdateRequestPath{
			ID: projectID,
		},
	}
}

// HTTPRequest creates an HTTP request for the ProjectUpdateRequest.
func (p ProjectUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/" + strconv.FormatInt(p.Path.ID, 10) + ".json"

	payload := struct {
		Project ProjectUpdateRequest `json:"project"`
	}{Project: p}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update project request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// ProjectUpdateResponse represents the response body for updating a project.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/projects/put-projects-id-json
type ProjectUpdateResponse struct{}

// HandleHTTPResponse handles the HTTP response for the ProjectUpdateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (p *ProjectUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update project")
	}
	if err := json.NewDecoder(resp.Body).Decode(p); err != nil {
		return fmt.Errorf("failed to decode update project response: %w", err)
	}
	return nil
}

// ProjectUpdate updates a project using the provided request and returns the
// response.
func ProjectUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectUpdateRequest,
) (*ProjectUpdateResponse, error) {
	return twapi.Execute[ProjectUpdateRequest, *ProjectUpdateResponse](ctx, engine, req)
}

// ProjectDeleteRequestPath contains the path parameters for deleting a project.
type ProjectDeleteRequestPath struct {
	// ID is the unique identifier of the project to be deleted.
	ID int64
}

// ProjectDeleteRequest represents the request body for deleting a project.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/projects/delete-projects-id-json
type ProjectDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path ProjectDeleteRequestPath
}

// NewProjectDeleteRequest creates a new ProjectDeleteRequest with the
// provided project ID.
func NewProjectDeleteRequest(projectID int64) ProjectDeleteRequest {
	return ProjectDeleteRequest{
		Path: ProjectDeleteRequestPath{
			ID: projectID,
		},
	}
}

// HTTPRequest creates an HTTP request for the ProjectDeleteRequest.
func (p ProjectDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/" + strconv.FormatInt(p.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// ProjectDeleteResponse represents the response body for deleting a project.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/projects/delete-projects-id-json
type ProjectDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the ProjectDeleteResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (p *ProjectDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to delete project")
	}
	if err := json.NewDecoder(resp.Body).Decode(p); err != nil {
		return fmt.Errorf("failed to decode delete project response: %w", err)
	}
	return nil
}

// ProjectDelete deletes a project using the provided request and returns the
// response.
func ProjectDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectDeleteRequest,
) (*ProjectDeleteResponse, error) {
	return twapi.Execute[ProjectDeleteRequest, *ProjectDeleteResponse](ctx, engine, req)
}

// ProjectCloneRequestPath contains the path parameters for cloning a project.
type ProjectCloneRequestPath struct {
	// ID is the unique identifier of the project to be cloned.
	ID int64
}

// ProjectCloneAction represents the action to perform when cloning a project.
type ProjectCloneAction string

const (
	// ProjectCloneActionCopy indicates that the project should be copied,
	// creating a new project with the same details.
	ProjectCloneActionCopy ProjectCloneAction = "COPY"

	// ProjectCloneActionMove indicates that the project should be moved,
	// effectively renaming it and keeping the same project ID.
	ProjectCloneActionMove ProjectCloneAction = "MOVE"
)

// ProjectCloneTemplateDateTarget represents the target date type for cloning a
// project from a template.
type ProjectCloneTemplateDateTarget string

const (
	// ProjectCloneTemplateDateTargetStart indicates that the target date
	// represents the start date of the new project. This is the default behavior
	// when cloning from a template.
	ProjectCloneTemplateDateTargetStart ProjectCloneTemplateDateTarget = "start"

	// ProjectCloneTemplateDateTargetEnd indicates that the target date represents
	// the end date of the new project. When this option is selected, the start
	// date of the new project is calculated by subtracting the duration of the
	// template project from the target end date.
	ProjectCloneTemplateDateTargetEnd ProjectCloneTemplateDateTarget = "end"
)

// ProjectCloneRequest represents the request body for copying/moving a project.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/projects/post-projects-id-clone-json
type ProjectCloneRequest struct {
	// Path contains the path parameters for the request.
	Path ProjectCloneRequestPath

	// Name for the new cloned project. If not provided the new project will have
	// the same name as the original project with an incremental number suffix.
	Name *string `json:"cloneProjectName,omitempty"`

	// Action to perform when cloning a project. By default, the project will be
	// copied.
	Action *ProjectCloneAction `json:"cloneProjectAction,omitempty"`

	// Description is an optional description for the new cloned project. If not
	// provided, the new project will have the same description as the original
	// project.
	Description *string `json:"description,omitempty"`

	// CompanyID is an optional ID of the company/client associated with the new
	// cloned project. By default it will be the same as the original project.
	CompanyID *int64 `json:"companyId,omitempty"`

	// CopyFiles indicates whether to copy the files from the original project to
	// the new project. By default it is set to false.
	CopyFiles *bool `json:"copyFiles,omitempty"`

	// CopyMessages indicates whether to copy the messages from the original
	// project to the new project. By default it is set to false.
	CopyMessages *bool `json:"copyMessages,omitempty"`

	// CopyMilestones indicates whether to copy the milestones from the original
	// project to the new project. By default it is set to false.
	CopyMilestones *bool `json:"copyMilestones,omitempty"`

	// CopyTasks indicates whether to copy the tasks from the original project to
	// the new project. By default it is set to false.
	CopyTasks *bool `json:"copyTasks,omitempty"`

	// CopyTasklists indicates whether to copy the tasklists from the original
	// project to the new project. By default it is set to false.
	CopyTasklists *bool `json:"copyTasklists,omitempty"`

	// CopyNotebooks indicates whether to copy the notebooks from the original
	// project to the new project. By default it is set to false.
	CopyNotebooks *bool `json:"copyNotebooks,omitempty"`

	// CopyLinks indicates whether to copy the links from the original project to
	// the new project. By default it is set to false.
	CopyLinks *bool `json:"copyLinks,omitempty"`

	// CopyComments indicates whether to copy the comments from the original
	// project to the new project. By default it is set to false.
	CopyComments *bool `json:"copyComments,omitempty"`

	// CopyFollowers indicates whether to copy the followers from the original
	// project to the new project. By default it is set to false.
	CopyFollowers *bool `json:"copyFollowers,omitempty"`

	// CopyInvoices indicates whether to copy the invoices from the original
	// project to the new project. By default it is set to false.
	CopyInvoices *bool `json:"copyInvoices,omitempty"`

	// CopyTimelogs indicates whether to copy the timelogs from the original
	// project to the new project. By default it is set to false.
	CopyTimelogs *bool `json:"copyTimelogs,omitempty"`

	// CopyExpenses indicates whether to copy the expenses from the original
	// project to the new project. By default it is set to false.
	CopyExpenses *bool `json:"copyExpenses,omitempty"`

	// CopyWebhooks indicates whether to copy the webhooks from the original
	// project to the new project. By default it is set to false.
	CopyWebhooks *bool `json:"copyWebhooks,omitempty"`

	// CopyProjectRoles indicates whether to copy the project roles from the
	// original project to the new project. By default it is set to false.
	CopyProjectRoles *bool `json:"copyProjectRoles,omitempty"`

	// CopyCustomFields indicates whether to copy the custom fields from the
	// original project to the new project. By default it is set to true.
	CopyCustomFields *bool `json:"copyCustomFields,omitempty"`

	// CopyCustomItems indicates whether to copy the custom items from the
	// original project to the new project. By default it is set to false.
	CopyCustomItems *bool `json:"copyCustomItems,omitempty"`

	// CopyProjectUpdates indicates whether to copy the project updates from the
	// original project to the new project. By default it is set to false.
	CopyProjectUpdates *bool `json:"copyProjectUpdates,omitempty"`

	// CopyRisks indicates whether to copy the risks from the original project to
	// the new project. By default it is set to false.
	CopyRisks *bool `json:"copyRisks,omitempty"`

	// CopyForms indicates whether to copy the forms from the original project to
	// the new project. By default it is set to false.
	CopyForms *bool `json:"copyForms,omitempty"`

	// CopyAutomations indicates whether to copy the automations from the original
	// project to the new project. By default it is set to false.
	CopyAutomations *bool `json:"copyAutomations,omitempty"`

	// CopyPeople indicates whether to copy the people from the original project
	// to the new project. By default it is set to false.
	CopyPeople *bool `json:"copyPeople,omitempty"`

	// CopyProjectPrivacy indicates whether to copy the project privacy settings
	// from the original project to the new project. By default it is set to false.
	CopyProjectPrivacy *bool `json:"copyProjectPrivacy,omitempty"`

	// CopyBudgets indicates whether to copy the budgets from the original project
	// to the new project. By default it is set to false.
	CopyBudgets *bool `json:"copyBudgets,omitempty"`

	// CopyAllocations indicates whether to copy the allocations from the original
	// project to the new project. By default it is set to false.
	CopyAllocations *bool `json:"copyAllocations,omitempty"`

	// CopyLogo indicates whether to copy the logo from the original project to
	// the new project. By default it is set to false.
	CopyLogo *bool `json:"copyLogo,omitempty"`

	// CopyProjectPreferences indicates whether to copy the project preferences
	// from the original project to the new project. By default it is set to
	// false.
	CopyProjectPreferences *bool `json:"copyProjectPrefs,omitempty"`

	// NewFromTemplate indicates whether the new project should be a regular one
	// created from a template. By default it is set to false.
	NewFromTemplate *bool `json:"newFromTemplate,omitempty"`

	// ToTemplate indicates whether the new project should be set as a template.
	// By default it is set to false.
	ToTemplate *bool `json:"toTemplate,omitempty"`

	// TemplateDateTarget specifies whether targetDate represents the project's
	// start or end date. Accepted values: 'start' (default) or 'end'. When 'end',
	// the start date is calculated by subtracting the template project's duration
	// from targetDate. Only applicable when newFromTemplate=true.
	TemplateDateTarget *ProjectCloneTemplateDateTarget `json:"templateDateTarget,omitempty"`

	// TargetDate is the desired start or end date for the cloned project
	// (determined by templateDateTarget). Used only when creating a project from
	// a template (newFromTemplate=true). Accepted format: YYYYMMDD string.
	// Defaults to the current user date if omitted.
	TargetDate *LegacyDate `json:"targetDate,omitempty"`

	// DaysOffset is the number of days to shift all scheduled dates in the cloned
	// project relative to the base date. When cloning from a template, it defines
	// the project duration span. When copying an existing project, it shifts the
	// original start and end dates by this many days. If omitted, calculated
	// automatically from the source project's date range.
	DaysOffset *int64 `json:"daysOffset,omitempty"`
}

// NewProjectCloneRequest creates a new ProjectCloneRequest with the provided project
// ID. The ID is required to clone a project.
func NewProjectCloneRequest(projectID int64) ProjectCloneRequest {
	return ProjectCloneRequest{
		Path: ProjectCloneRequestPath{
			ID: projectID,
		},
	}
}

// HTTPRequest creates an HTTP request for the ProjectCloneRequest.
func (p ProjectCloneRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/" + strconv.FormatInt(p.Path.ID, 10) + "/clone.json"

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(p); err != nil {
		return nil, fmt.Errorf("failed to encode clone project request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// ProjectCloneResponse represents the response body for cloning a project.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/projects/post-projects-id-clone-json
type ProjectCloneResponse struct {
	// ID is the unique identifier of the cloned project.
	ID LegacyNumber `json:"projectId"`
}

// HandleHTTPResponse handles the HTTP response for the ProjectCloneResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (p *ProjectCloneResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to clone project")
	}
	if err := json.NewDecoder(resp.Body).Decode(p); err != nil {
		return fmt.Errorf("failed to decode clone project response: %w", err)
	}
	if p.ID == 0 {
		return fmt.Errorf("clone project response does not contain a valid identifier")
	}
	return nil
}

// ProjectClone clones a project using the provided request and returns the
// response.
func ProjectClone(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectCloneRequest,
) (*ProjectCloneResponse, error) {
	return twapi.Execute[ProjectCloneRequest, *ProjectCloneResponse](ctx, engine, req)
}

// ProjectRequestSideload contains the possible sideload options when loading
// projects.
//
// Only ProjectRequestSideloadProjectCategories,
// ProjectRequestSideloadCustomFields and ProjectRequestSideloadCustomFieldValues
// are decoded into the typed Included struct of the get and list responses. The
// remaining values are accepted by the endpoint and reach the wire, but their
// payloads are discarded until the responses grow a slot for them.
type ProjectRequestSideload string

// List of possible sideload options for ProjectRequestSideload.
const (
	ProjectRequestSideloadProjectCategories ProjectRequestSideload = "projectCategories"
	ProjectRequestSideloadCustomFields      ProjectRequestSideload = "customfields"
	ProjectRequestSideloadCustomFieldValues ProjectRequestSideload = "customfieldprojects"
	ProjectRequestSideloadProjectOwners     ProjectRequestSideload = "projectOwners"
	ProjectRequestSideloadProjectUpdates    ProjectRequestSideload = "projectUpdates"
	ProjectRequestSideloadCompanies         ProjectRequestSideload = "companies"
	ProjectRequestSideloadCompanyCountries  ProjectRequestSideload = "companies.countries"
	ProjectRequestSideloadCompanyIndustries ProjectRequestSideload = "companies.industries"
	ProjectRequestSideloadCompanyCurrencies ProjectRequestSideload = "companies.currencies"
	ProjectRequestSideloadPortfolioCards    ProjectRequestSideload = "portfolioCards"
	ProjectRequestSideloadPortfolioBoards   ProjectRequestSideload = "portfolioBoards"
	ProjectRequestSideloadPortfolioColumns  ProjectRequestSideload = "portfolioColumns"
	ProjectRequestSideloadTags              ProjectRequestSideload = "tags"
	ProjectRequestSideloadProjectBudgets    ProjectRequestSideload = "projectBudgets"
	ProjectRequestSideloadLatestActivities  ProjectRequestSideload = "activities.latest"
	ProjectRequestSideloadUsers             ProjectRequestSideload = "users"
	ProjectRequestSideloadCreatedBy         ProjectRequestSideload = "createdBy"
	ProjectRequestSideloadUpdatedBy         ProjectRequestSideload = "updatedBy"
	ProjectRequestSideloadCompletedBy       ProjectRequestSideload = "completedBy"
	ProjectRequestSideloadDeletedBy         ProjectRequestSideload = "deletedBy"
	ProjectRequestSideloadArchivedBy        ProjectRequestSideload = "archivedBy"
	ProjectRequestSideloadEmailDropboxes    ProjectRequestSideload = "projectEmailDropboxes"
	ProjectRequestSideloadWorkflows         ProjectRequestSideload = "workflows"
	ProjectRequestSideloadWorkflowStages    ProjectRequestSideload = "workflows.stages"
)

// ProjectRequestFilters contains the filters for loading projects.
type ProjectRequestFilters struct {
	// Include specifies related resources to include.
	Include []ProjectRequestSideload
}

func (p ProjectRequestFilters) apply(req *http.Request) {
	query := req.URL.Query()
	querySetStrings(query, "include", p.Include)
	req.URL.RawQuery = query.Encode()
}

// ProjectGetRequestPath contains the path parameters for loading a single
// project.
type ProjectGetRequestPath struct {
	// ID is the unique identifier of the project to be retrieved.
	ID int64 `json:"id"`
}

// ProjectGetRequest represents the request body for loading a single project.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/projects/get-projects-api-v3-projects-project-id-json
type ProjectGetRequest struct {
	// Path contains the path parameters for the request.
	Path ProjectGetRequestPath

	// Filters contains the filters for loading a single project.
	Filters ProjectRequestFilters

	// Fields restricts the attributes returned for the project and each of its
	// sideloads. Each slot of ProjectGetFields is a separate
	// `fields[entity]=…` selection; populated slots restrict the response,
	// empty slots return the API default. Use the generated ProjectField
	// constants to ensure values match real attributes.
	Fields ProjectGetFields
}

// NewProjectGetRequest creates a new ProjectGetRequest with the provided
// project ID. The ID is required to load a project.
func NewProjectGetRequest(projectID int64) ProjectGetRequest {
	return ProjectGetRequest{
		Path: ProjectGetRequestPath{
			ID: projectID,
		},
	}
}

// HTTPRequest creates an HTTP request for the ProjectGetRequest.
func (p ProjectGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/projects/" + strconv.FormatInt(p.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	p.Filters.apply(req)

	query := req.URL.Query()
	p.Fields.apply(query)
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// ProjectGetResponse contains all the information related to a project.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/projects/get-projects-api-v3-projects-project-id-json
//
// sparsefields:get
type ProjectGetResponse struct {
	Project Project `json:"project"`

	// Included contains related objects included in the response.
	Included struct {
		// ProjectCategories contains the categories associated with the project.
		//
		// The key is the string representation of the project category ID.
		ProjectCategories map[string]ProjectCategory `json:"projectCategories,omitempty"`
		// CustomFields contains the custom fields associated with the project.
		//
		// The key is the string representation of the custom field ID.
		CustomFields map[string]CustomField `json:"customfields,omitempty"`
		// CustomFieldValues contains the values of the custom fields associated
		// with the project.
		//
		// The key is the string representation of the custom field value ID.
		CustomFieldValues map[string]CustomFieldValue `json:"customfieldProjects,omitempty"`
	} `json:"included"`
}

// HandleHTTPResponse handles the HTTP response for the ProjectGetResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (p *ProjectGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve project")
	}

	if err := json.NewDecoder(resp.Body).Decode(p); err != nil {
		return fmt.Errorf("failed to decode retrieve project response: %w", err)
	}
	return nil
}

// ProjectGet retrieves a single project using the provided request and returns
// the response.
func ProjectGet(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectGetRequest,
) (*ProjectGetResponse, error) {
	return twapi.Execute[ProjectGetRequest, *ProjectGetResponse](ctx, engine, req)
}

// ProjectOrderBy identifies the attributes a project list can be ordered by.
type ProjectOrderBy string

// Supported project order-by values.
const (
	ProjectOrderByBudgetUsed          ProjectOrderBy = "budgetused"
	ProjectOrderByCategoryName        ProjectOrderBy = "categoryname"
	ProjectOrderByCompanyName         ProjectOrderBy = "companyname"
	ProjectOrderByCreatorName         ProjectOrderBy = "creatorname"
	ProjectOrderByCustomField         ProjectOrderBy = "customfield"
	ProjectOrderByDateCreated         ProjectOrderBy = "datecreated"
	ProjectOrderByDueDate             ProjectOrderBy = "duedate"
	ProjectOrderByHealth              ProjectOrderBy = "health"
	ProjectOrderByLastActivity        ProjectOrderBy = "lastactivity"
	ProjectOrderByLastWorkedOn        ProjectOrderBy = "lastworkedon"
	ProjectOrderByMobileSpecial       ProjectOrderBy = "mobilespecial"
	ProjectOrderByName                ProjectOrderBy = "name"
	ProjectOrderByNameCaseInsensitive ProjectOrderBy = "namecaseinsensitive"
	ProjectOrderByOwnerCompany        ProjectOrderBy = "ownercompany"
	ProjectOrderByOwnerName           ProjectOrderBy = "ownername"
	ProjectOrderByStarred             ProjectOrderBy = "starred"
	ProjectOrderByStarredCompanyName  ProjectOrderBy = "starredcompanyname"
	ProjectOrderByStarredFirst        ProjectOrderBy = "starredfirst"
	ProjectOrderByStartDate           ProjectOrderBy = "startdate"
	ProjectOrderByStatus              ProjectOrderBy = "status"
	ProjectOrderByTaskCompletion      ProjectOrderBy = "taskcompletion"
	ProjectOrderByID                  ProjectOrderBy = "id"
)

// ProjectType identifies the kinds of project a list can be restricted to.
//
// It covers the values the projectType filter accepts, which is not the same set
// a project's own type can report: a response may also carry "holder-project" or
// "global-messages".
type ProjectType string

// Supported project type filters.
const (
	// ProjectTypeAll returns every type of project, templates included.
	ProjectTypeAll ProjectType = "all"

	// ProjectTypeNormal returns ordinary projects.
	ProjectTypeNormal ProjectType = "normal"

	// ProjectTypeTasklistsTemplate returns tasklist templates.
	ProjectTypeTasklistsTemplate ProjectType = "tasklists-template"

	// ProjectTypeProjectsTemplate returns project templates.
	ProjectTypeProjectsTemplate ProjectType = "projects-template"

	// ProjectTypePersonal returns the caller's personal project.
	ProjectTypePersonal ProjectType = "personal"

	// ProjectTypeTentative returns tentative projects.
	ProjectTypeTentative ProjectType = "tentative"
)

// ProjectListStatus identifies the states a project list can be filtered by.
//
// These are not the values of ProjectStatus, which is what a project reports
// about itself. This set describes the progress buckets the list endpoint
// filters on, so "current", "late" and "upcoming" appear here and "archived"
// does not — archived projects are selected with
// ProjectListRequestFilters.IncludeArchivedProjects or OnlyArchivedProjects
// instead.
type ProjectListStatus string

// Supported project list status filters.
const (
	ProjectListStatusActive    ProjectListStatus = "active"
	ProjectListStatusCurrent   ProjectListStatus = "current"
	ProjectListStatusLate      ProjectListStatus = "late"
	ProjectListStatusUpcoming  ProjectListStatus = "upcoming"
	ProjectListStatusCompleted ProjectListStatus = "completed"
	ProjectListStatusDeleted   ProjectListStatus = "deleted"
)

// ProjectHealth identifies the health ratings a project list can be filtered by.
// The API reports and filters health as a number rather than a name.
type ProjectHealth int64

// Supported project health filters.
const (
	// ProjectHealthNotSet matches projects whose health has never been reported.
	ProjectHealthNotSet ProjectHealth = 0

	// ProjectHealthBad matches projects reported as being in bad health.
	ProjectHealthBad ProjectHealth = 1

	// ProjectHealthOK matches projects reported as being in acceptable health.
	ProjectHealthOK ProjectHealth = 2

	// ProjectHealthGood matches projects reported as being in good health.
	ProjectHealthGood ProjectHealth = 3
)

// ProjectFeature identifies the features a project list can be filtered by. A
// project matches when the feature is enabled for it.
type ProjectFeature string

// Supported project feature filters.
const (
	ProjectFeatureList       ProjectFeature = "list"
	ProjectFeatureBoard      ProjectFeature = "board"
	ProjectFeatureGantt      ProjectFeature = "gantt"
	ProjectFeatureTable      ProjectFeature = "table"
	ProjectFeatureDashboard  ProjectFeature = "dashboard"
	ProjectFeatureMilestones ProjectFeature = "milestones"
	ProjectFeatureMessages   ProjectFeature = "messages"
	ProjectFeatureFiles      ProjectFeature = "files"
	ProjectFeatureTime       ProjectFeature = "time"
	ProjectFeatureNotebooks  ProjectFeature = "notebooks"
	ProjectFeatureRisks      ProjectFeature = "risks"
	ProjectFeatureLinks      ProjectFeature = "links"
	ProjectFeatureBilling    ProjectFeature = "billing"
	ProjectFeatureComments   ProjectFeature = "comments"
	ProjectFeaturePeople     ProjectFeature = "people"
	ProjectFeatureSettings   ProjectFeature = "settings"
)

// ProjectTimeMode identifies which time figures the profitability calculation is
// based on. It is only read when
// ProjectListRequestFilters.IncludeProjectProfitability is set.
type ProjectTimeMode string

// Supported project profitability time modes.
const (
	// ProjectTimeModeTimelogs bases profitability on the time actually logged.
	ProjectTimeModeTimelogs ProjectTimeMode = "timelogs"

	// ProjectTimeModeEstimated bases profitability on estimated time.
	ProjectTimeModeEstimated ProjectTimeMode = "estimated"
)

// ProjectListRequestFilters contains the filters for loading multiple projects.
//
// The endpoint also documents parameters that turn the response into a
// downloadable report (reportType, reportFormat, reportTimezone,
// isReportDownload and the selectedColumns it exports). They are deliberately
// absent: ProjectListResponse decodes JSON, so a request that asked for a CSV or
// PDF body could not be handled. The undocumented `filter` object is absent for
// the same reason of not being modellable.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/projects/get-projects-api-v3-projects-json
type ProjectListRequestFilters struct {
	ProjectRequestFilters

	// SearchTerm is an optional search term to filter projects by name or
	// description.
	SearchTerm string

	// SearchCompanies extends SearchTerm to the name of the company owning the
	// project.
	SearchCompanies *bool

	// ProjectIDs is an optional list of project IDs to restrict the results to.
	// Providing it normally bypasses the other filters
	ProjectIDs []int64

	// ExcludeProjectIDs is an optional list of project IDs to leave out of the
	// results.
	ExcludeProjectIDs []int64

	// ProjectType restricts the results to one kind of project. Use the
	// ProjectType constants.
	ProjectType ProjectType

	// ProjectStatuses is an optional list of progress states to filter projects
	// by. Use the ProjectListStatus constants.
	ProjectStatuses []ProjectListStatus

	// ProjectHealths is an optional list of health ratings to filter projects by.
	// Use the ProjectHealth constants.
	ProjectHealths []ProjectHealth

	// ProjectCategoryIDs is an optional list of project category IDs to filter
	// projects by categories.
	ProjectCategoryIDs []int64

	// IncludeSubCategories extends ProjectCategoryIDs to the categories nested
	// under the ones given.
	IncludeSubCategories *bool

	// ProjectCompanyIDs is an optional list of company IDs to filter projects by
	// the company that owns them.
	ProjectCompanyIDs []int64

	// ProjectOwnerIDs is an optional list of user IDs to filter projects by their
	// owner.
	ProjectOwnerIDs []int64

	// UserIDs is an optional list of user IDs to filter projects by.
	UserIDs []int64

	// TeamIDs is an optional list of team IDs to filter projects by, matching the
	// projects containing users of those teams.
	TeamIDs []int64

	// TagIDs is an optional list of tag IDs to filter projects by tags.
	TagIDs []int64

	// MatchAllTags is an optional flag to indicate if all tags must match. If
	// set to true, only projects matching all specified tags will be returned.
	MatchAllTags *bool

	// ExcludeTagIDs is an optional list of tag IDs whose projects are left out of
	// the results.
	ExcludeTagIDs []int64

	// MatchAllExcludedTags requires a project to carry every tag of
	// ExcludeTagIDs before it is left out, rather than any one of them.
	MatchAllExcludedTags *bool

	// UpdatedAfter is an optional filter to retrieve projects updated after a
	// specific date and time. It is sent in RFC3339 format.
	UpdatedAfter *time.Time

	// NotCompletedBefore is an optional filter to retrieve projects that were
	// still open on the given date.
	NotCompletedBefore *twapi.Date

	// OnlyStarredProjects restricts the results to the projects the calling user
	// has starred.
	OnlyStarredProjects *bool

	// OnlyProjectsWithAdminAccess restricts the results to the projects the
	// calling user administers. Defaults to false.
	OnlyProjectsWithAdminAccess *bool

	// HideObservedProjects leaves out the projects where the calling user is only
	// an observer. Defaults to false.
	HideObservedProjects *bool

	// IncludeArchivedProjects returns archived projects alongside the active
	// ones. Defaults to false.
	IncludeArchivedProjects *bool

	// OnlyArchivedProjects restricts the results to archived projects. Defaults
	// to false.
	OnlyArchivedProjects *bool

	// IncludeTentativeProjects returns tentative projects alongside the normal
	// ones.
	IncludeTentativeProjects *bool

	// IncludeCustomFields adds the projects' custom field values to the response.
	// Defaults to false.
	IncludeCustomFields *bool

	// IncludeCustomFieldIDs narrows IncludeCustomFields to the given custom
	// fields.
	IncludeCustomFieldIDs []int64

	// UseFormulaFields evaluates formula custom fields instead of returning them
	// unresolved.
	UseFormulaFields *bool

	// IncludeOverallStats adds the projects' related counts to the response.
	IncludeOverallStats *bool

	// IncludeProjectStats adds per-project counts for tasks, columns, billing events and
	// milestones to the response.
	IncludeProjectStats *bool

	// IncludeProjectDates adds the earliest start and latest end date of each
	// project to the response.
	IncludeProjectDates *bool

	// IncludeProjectUserInfo adds the data that depends on the calling user, such
	// as whether the project is starred. Defaults to false.
	IncludeProjectUserInfo *bool

	// IncludeProjectProfitability adds each project's profitability to the
	// response. TimeMode selects which time figures it is based on.
	IncludeProjectProfitability *bool

	// TimeMode selects the time figures the profitability calculation uses. Use
	// the ProjectTimeMode constants. It is only read when
	// IncludeProjectProfitability is set.
	TimeMode ProjectTimeMode

	// OrderBy is the field to sort the results by. Use the ProjectOrderBy
	// constants. The endpoint defaults to name.
	OrderBy ProjectOrderBy

	// OrderMode is the direction to sort the results in. See twapi.OrderMode for
	// the supported values. The endpoint defaults to ascending.
	OrderMode twapi.OrderMode

	// OrderByCustomFieldID selects the custom field to sort by. It is only used
	// when OrderBy is ProjectOrderByCustomField.
	OrderByCustomFieldID int64

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of projects to retrieve per page. Defaults to 50.
	PageSize int64

	// CountMode selects whether the API computes the exact number of projects
	// matching the filters, reported in Meta.Page.Count. Defaults to
	// twapi.ListCountModeDefault, which leaves the decision to the API. An
	// unscoped exact count over every project in the installation is the
	// expensive one to ask for.
	CountMode twapi.ListCountMode

	// Fields restricts the attributes returned for the project and each of its
	// sideloads. Each slot of ProjectListFields is a separate `fields[entity]=…`
	// selection; populated slots restrict the response, empty slots return the
	// API default. Use the generated ProjectField / ProjectCategoryField /
	// CustomFieldField / CustomFieldValueField constants to ensure values match
	// real attributes.
	Fields ProjectListFields
}

func (p ProjectListRequestFilters) apply(req *http.Request) {
	p.ProjectRequestFilters.apply(req)

	query := req.URL.Query()

	querySetString(query, "searchTerm", p.SearchTerm)
	querySetBool(query, "searchCompanies", p.SearchCompanies)

	querySetInt64s(query, "projectIds", p.ProjectIDs)
	querySetInt64s(query, "excludeProjectIds", p.ExcludeProjectIDs)

	querySetString(query, "projectType", p.ProjectType)
	querySetStrings(query, "projectStatuses", p.ProjectStatuses)
	querySetInt64s(query, "projectHealths", p.ProjectHealths)

	querySetInt64s(query, "projectCategoryIds", p.ProjectCategoryIDs)
	querySetBool(query, "includeSubCategories", p.IncludeSubCategories)
	querySetInt64s(query, "projectCompanyIds", p.ProjectCompanyIDs)
	querySetInt64s(query, "projectOwnerIds", p.ProjectOwnerIDs)

	querySetInt64s(query, "usersWithExplicitMembershipIds", p.UserIDs)
	querySetInt64s(query, "teamIds", p.TeamIDs)

	querySetInt64s(query, "projectTagIds", p.TagIDs)
	querySetBool(query, "matchAllProjectTags", p.MatchAllTags)
	querySetInt64s(query, "excludeTagIds", p.ExcludeTagIDs)
	querySetBool(query, "matchAllExcludedTags", p.MatchAllExcludedTags)

	querySetTimestamp(query, "updatedAfter", p.UpdatedAfter)
	querySetDate(query, "notCompletedBefore", p.NotCompletedBefore)

	querySetBool(query, "onlyStarredProjects", p.OnlyStarredProjects)
	querySetBool(query, "onlyProjectsWithAdminAccess", p.OnlyProjectsWithAdminAccess)
	querySetBool(query, "hideObservedProjects", p.HideObservedProjects)

	querySetBool(query, "includeArchivedProjects", p.IncludeArchivedProjects)
	querySetBool(query, "onlyArchivedProjects", p.OnlyArchivedProjects)
	querySetBool(query, "includeTentativeProjects", p.IncludeTentativeProjects)

	querySetBool(query, "includeCustomFields", p.IncludeCustomFields)
	querySetInt64s(query, "includeCustomFieldIds", p.IncludeCustomFieldIDs)
	querySetBool(query, "useFormulaFields", p.UseFormulaFields)

	querySetBool(query, "includeCounts", p.IncludeOverallStats)
	querySetBool(query, "includeStats", p.IncludeProjectStats)
	querySetBool(query, "includeProjectDates", p.IncludeProjectDates)
	querySetBool(query, "includeProjectUserInfo", p.IncludeProjectUserInfo)
	querySetBool(query, "includeProjectProfitability", p.IncludeProjectProfitability)
	querySetString(query, "timeMode", p.TimeMode)

	querySetString(query, "orderBy", p.OrderBy)
	querySetString(query, "orderMode", p.OrderMode)
	querySetInt64(query, "orderByCustomFieldId", p.OrderByCustomFieldID)

	querySetInt64(query, "page", p.Page)
	querySetInt64(query, "pageSize", p.PageSize)

	p.Fields.apply(query)
	p.CountMode.Apply(query)
	req.URL.RawQuery = query.Encode()
}

// ProjectListRequest represents the request body for loading multiple projects.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/projects/get-projects-api-v3-projects-json
type ProjectListRequest struct {
	// Filters contains the filters for loading multiple projects.
	Filters ProjectListRequestFilters
}

// NewProjectListRequest creates a new ProjectListRequest with default values.
func NewProjectListRequest() ProjectListRequest {
	return ProjectListRequest{
		Filters: ProjectListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the ProjectListRequest.
func (p ProjectListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/projects.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	p.Filters.apply(req)

	return req, nil
}

// ProjectListResponse contains information by multiple projects matching the
// request filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/projects/get-projects-api-v3-projects-json
//
// sparsefields:list
type ProjectListResponse struct {
	request ProjectListRequest

	// Meta contains metadata about the response, including pagination details.
	Meta twapi.ListMeta `json:"meta"`

	// Projects is the list of projects matching the request filters.
	Projects []Project `json:"projects"`

	// Included contains related objects included in the response.
	Included struct {
		// ProjectCategories contains the categories associated with the project.
		//
		// The key is the string representation of the project category ID.
		ProjectCategories map[string]ProjectCategory `json:"projectCategories,omitempty"`
		// CustomFields contains the custom fields associated with the project.
		//
		// The key is the string representation of the custom field ID.
		CustomFields map[string]CustomField `json:"customfields,omitempty"`
		// CustomFieldValues contains the values of the custom fields associated
		// with the project.
		//
		// The key is the string representation of the custom field value ID.
		CustomFieldValues map[string]CustomFieldValue `json:"customfieldProjects,omitempty"`
	} `json:"included"`
}

// HandleHTTPResponse handles the HTTP response for the ProjectListResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (p *ProjectListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list projects")
	}

	if err := json.NewDecoder(resp.Body).Decode(p); err != nil {
		return fmt.Errorf("failed to decode list projects response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (p *ProjectListResponse) SetRequest(req ProjectListRequest) {
	p.request = req
	p.Meta.ResolveCount(req.Filters.CountMode)
}

// Iterate returns the request set to the next page, if available. If there
// are no more pages, a nil request is returned.
func (p *ProjectListResponse) Iterate() *ProjectListRequest {
	if !p.Meta.Page.HasMore {
		return nil
	}
	req := p.request
	req.Filters.Page++
	return &req
}

// ProjectList retrieves multiple projects using the provided request
// and returns the response.
func ProjectList(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectListRequest,
) (*ProjectListResponse, error) {
	return twapi.Execute[ProjectListRequest, *ProjectListResponse](ctx, engine, req)
}
