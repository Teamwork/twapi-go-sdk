package projects

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var (
	_ twapi.HTTPRequester = (*ProjectStatusUpdateListRequest)(nil)
	_ twapi.HTTPResponser = (*ProjectStatusUpdateListResponse)(nil)
)

// ProjectStatusUpdateReactions contains the reactions left on a project update.
// It is only populated when ProjectStatusUpdateListRequestFilters.Reactions is
// set.
type ProjectStatusUpdateReactions struct {
	// Counts is the number of reactions of each of the five fixed kinds.
	Counts struct {
		Like    int64 `json:"like"`
		Dislike int64 `json:"dislike"`
		Joy     int64 `json:"joy"`
		Frown   int64 `json:"frown"`
		Heart   int64 `json:"heart"`
	} `json:"counts"`

	// Mine lists the fixed kinds the calling user reacted with.
	Mine []string `json:"mine"`

	// ExtendedReactions counts the arbitrary emoji reactions.
	ExtendedReactions []struct {
		Emoji string `json:"emoji"`
		Count int64  `json:"count"`
		Mine  bool   `json:"mine"`
	} `json:"extendedReactions"`
}

// ProjectStatusUpdate is the status update posted on a project: the prose shown
// on the project dashboard, together with the health rating it reports.
//
// The API calls it a "project update" and keys it as `projectUpdates` in the
// response and in fields[…] selections. The Go type carries "status" because
// ProjectUpdate is already the function that edits a project.
//
// More information can be found at:
// https://support.teamwork.com/projects/project-updates/adding-a-project-update
//
// sparsefields:gen
type ProjectStatusUpdate struct {
	// ID is the unique identifier of the project update.
	ID int64 `json:"id"`

	// Text is the body of the update, in Markdown. Emoji codes in it are
	// converted to characters unless
	// ProjectStatusUpdateListRequestFilters.Emoji is set to false.
	Text string `json:"text"`

	// Health is the health rating the update reports. Use the ProjectHealth
	// constants: the API models health as an integer whose meaning is positional.
	Health ProjectHealth `json:"health"`

	// HealthLabel is the name the installation gives to Health. It is configured
	// per installation, so it is not one of a fixed set of names — build a label
	// from this field rather than from Health.
	HealthLabel string `json:"healthLabel"`

	// Color is the colour of Health, as a hex value with a leading number sign.
	// It is empty when the health is not set, which is why it is a
	// twapi.OptionalHexColor: twapi.HexColor rejects the empty value.
	Color twapi.OptionalHexColor `json:"color"`

	// ProjectID is the identifier of the project the update belongs to.
	ProjectID int64 `json:"projectId"`

	// Project is the relationship to the project the update belongs to.
	Project twapi.Relationship `json:"project"`

	// CreatedBy is the identifier of the user who posted the update. Sideload
	// ProjectStatusUpdateListRequestSideloadCreatedBy to resolve the name.
	CreatedBy int64 `json:"createdBy"`

	// CreatedAt is the date and time when the update was posted.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt is the date and time when the update was last edited.
	UpdatedAt time.Time `json:"updatedAt"`

	// IsActive reports whether this is the project's current update. Each project
	// has at most one, and the older updates are its history.
	IsActive *bool `json:"isActive"`

	// Deleted reports whether the update was deleted.
	Deleted bool `json:"deleted"`

	// DeletedAt is the date and time when the update was deleted.
	DeletedAt *time.Time `json:"deletedAt"`

	// DeletedBy is the identifier of the user who deleted the update.
	DeletedBy *int64 `json:"deletedBy"`

	// LikeFromUserIDs are the identifiers of the users who liked the update.
	LikeFromUserIDs []int64 `json:"likeFromUserIDs"`

	// LikeFromUsers are the relationships to the users who liked the update.
	LikeFromUsers []twapi.Relationship `json:"likeFromUsers"`

	// Reactions are the reactions left on the update. It is only populated when
	// ProjectStatusUpdateListRequestFilters.Reactions is set.
	Reactions *ProjectStatusUpdateReactions `json:"reactions"`
}

// ProjectStatusUpdateListRequestPath contains the path parameters for listing
// project updates.
type ProjectStatusUpdateListRequestPath struct {
	// ProjectID optionally scopes the request to a single project, reaching
	// /projects/{projectId}/updates.json instead of the site-wide
	// /projects/updates.json. The path segment holds one identifier only; use
	// ProjectStatusUpdateListRequestFilters.ProjectIDs to scope the request to
	// several projects.
	ProjectID int64
}

// ProjectStatusUpdateListRequestSideload represents related objects that can be
// included in the response.
type ProjectStatusUpdateListRequestSideload string

// List of valid sideload values for ProjectStatusUpdateListRequest.
const (
	ProjectStatusUpdateListRequestSideloadProjects  ProjectStatusUpdateListRequestSideload = "projects"
	ProjectStatusUpdateListRequestSideloadCreatedBy ProjectStatusUpdateListRequestSideload = "createdBy"
	ProjectStatusUpdateListRequestSideloadDeletedBy ProjectStatusUpdateListRequestSideload = "deletedBy"
	ProjectStatusUpdateListRequestSideloadLikeUsers ProjectStatusUpdateListRequestSideload = "likes.users"
)

// ProjectStatusUpdateOrderBy identifies the attributes a project update list can
// be ordered by.
type ProjectStatusUpdateOrderBy string

// Supported project update order-by values.
const (
	ProjectStatusUpdateOrderByDate    ProjectStatusUpdateOrderBy = "date"
	ProjectStatusUpdateOrderByColor   ProjectStatusUpdateOrderBy = "color"
	ProjectStatusUpdateOrderByHealth  ProjectStatusUpdateOrderBy = "health"
	ProjectStatusUpdateOrderByProject ProjectStatusUpdateOrderBy = "project"
	ProjectStatusUpdateOrderByUser    ProjectStatusUpdateOrderBy = "user"
	ProjectStatusUpdateOrderByID      ProjectStatusUpdateOrderBy = "id"
)

// ProjectStatusUpdateListRequestFilters contains the filters for loading
// multiple project updates.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/project-updates/get-projects-api-v3-projects-updates-json
type ProjectStatusUpdateListRequestFilters struct {
	// ProjectIDs restricts the results to the updates of the given projects.
	// Providing it makes the endpoint ignore every other project filter in this
	// struct.
	ProjectIDs []int64

	// ProjectOwnerIDs restricts the results to the updates of the projects owned
	// by the given users.
	ProjectOwnerIDs []int64

	// ProjectCategoryIDs restricts the results to the updates of the projects in
	// the given categories.
	ProjectCategoryIDs []int64

	// ProjectTagIDs restricts the results to the updates of the projects carrying
	// the given tags.
	ProjectTagIDs []int64

	// MatchAllProjectTags requires a project to carry every tag of ProjectTagIDs
	// rather than any one of them.
	MatchAllProjectTags *bool

	// ProjectStatuses restricts the results to the updates of the projects in the
	// given progress states. Use the ProjectListStatus constants.
	ProjectStatuses []ProjectListStatus

	// ProjectHealths restricts the results to the updates reporting the given
	// health ratings. Use the ProjectHealth constants.
	ProjectHealths []ProjectHealth

	// OnlyStarredProjects restricts the results to the updates of the projects
	// the calling user has starred.
	OnlyStarredProjects *bool

	// ProjectCompanyIDs restricts the results to the updates of the projects
	// owned by the given companies.
	ProjectCompanyIDs []int64

	// IncludeArchivedProjects returns the updates of archived projects alongside
	// those of the active ones. Defaults to false.
	IncludeArchivedProjects *bool

	// ShowDeleted returns deleted updates alongside the live ones. Defaults to
	// false.
	ShowDeleted *bool

	// ActiveOnly restricts the results to each project's current update. The
	// endpoint defaults it to true, so an unfiltered call returns one update per
	// project rather than the whole history. Set it to false to read the history.
	ActiveOnly *bool

	// CreatedAfter restricts the results to the updates posted at or after this
	// moment. The boundary is inclusive: an update posted exactly on it matches.
	CreatedAfter *time.Time

	// UpdatedAfter restricts the results to the updates last edited after this
	// moment. The boundary is exclusive: an update edited exactly on it does not
	// match.
	UpdatedAfter *time.Time

	// Emoji converts the emoji codes in Text to characters. The endpoint defaults
	// it to true.
	Emoji *bool

	// Reactions populates ProjectStatusUpdate.Reactions. The endpoint defaults it
	// to false, and setting it drops any "likes" sideload from Include.
	Reactions *bool

	// OrderBy is the field to sort the results by. Use the
	// ProjectStatusUpdateOrderBy constants.
	OrderBy ProjectStatusUpdateOrderBy

	// OrderMode is the direction to sort the results in. See twapi.OrderMode for
	// the supported values.
	OrderMode twapi.OrderMode

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of updates to retrieve per page. Defaults to 50.
	PageSize int64

	// Include specifies sideloaded entities to include in the response.
	Include []ProjectStatusUpdateListRequestSideload

	// CountMode selects whether the API computes the exact number of project
	// updates matching the filters, reported in Meta.Page.Count. Defaults to
	// twapi.ListCountModeDefault, which leaves the decision to the API.
	CountMode twapi.ListCountMode

	// Fields restricts the attributes returned for the project update and each of
	// its sideloads. Each slot of ProjectStatusUpdateListFields is a separate
	// `fields[entity]=…` selection; populated slots restrict the response, empty
	// slots return the API default. Use the generated ProjectStatusUpdateField /
	// UserField / ProjectField constants to ensure values match real attributes.
	Fields ProjectStatusUpdateListFields
}

func (p ProjectStatusUpdateListRequestFilters) apply(req *http.Request) {
	query := req.URL.Query()

	querySetInt64s(query, "projectIds", p.ProjectIDs)
	querySetInt64s(query, "projectOwnerIds", p.ProjectOwnerIDs)
	querySetInt64s(query, "projectCategoryIds", p.ProjectCategoryIDs)
	querySetInt64s(query, "projectTagIds", p.ProjectTagIDs)
	querySetBool(query, "matchAllProjectTags", p.MatchAllProjectTags)
	querySetStrings(query, "projectStatuses", p.ProjectStatuses)
	querySetInt64s(query, "projectHealths", p.ProjectHealths)
	querySetBool(query, "onlyStarredProjects", p.OnlyStarredProjects)
	querySetInt64s(query, "projectCompanyIds", p.ProjectCompanyIDs)
	querySetBool(query, "includeArchivedProjects", p.IncludeArchivedProjects)

	querySetBool(query, "showDeleted", p.ShowDeleted)
	querySetBool(query, "activeOnly", p.ActiveOnly)
	querySetTimestamp(query, "createdAfter", p.CreatedAfter)
	querySetTimestamp(query, "updatedAfter", p.UpdatedAfter)
	querySetBool(query, "emoji", p.Emoji)
	querySetBool(query, "reactions", p.Reactions)

	querySetString(query, "orderBy", p.OrderBy)
	querySetString(query, "orderMode", p.OrderMode)
	querySetInt64(query, "page", p.Page)
	querySetInt64(query, "pageSize", p.PageSize)
	querySetStrings(query, "include", p.Include)

	p.Fields.apply(query)
	p.CountMode.Apply(query)
	req.URL.RawQuery = query.Encode()
}

// ProjectStatusUpdateListRequest represents the request for loading multiple
// project updates.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/project-updates/get-projects-api-v3-projects-updates-json
type ProjectStatusUpdateListRequest struct {
	// Path contains the path parameters for the request.
	Path ProjectStatusUpdateListRequestPath

	// Filters contains the filters for loading multiple project updates.
	Filters ProjectStatusUpdateListRequestFilters
}

// NewProjectStatusUpdateListRequest creates a new
// ProjectStatusUpdateListRequest with default values.
//
// Unlike most list requests it pins the ordering, to newest first. The endpoint
// answers an omitted orderMode that way today, but its reference documents the
// default as ascending, and behaviour that contradicts the reference is not a
// guarantee. Nothing depends on the previous order because the request type is
// new. Callers wanting the reverse set Filters.OrderMode themselves.
func NewProjectStatusUpdateListRequest() ProjectStatusUpdateListRequest {
	return ProjectStatusUpdateListRequest{
		Filters: ProjectStatusUpdateListRequestFilters{
			OrderBy:   ProjectStatusUpdateOrderByDate,
			OrderMode: twapi.OrderModeDescending,
			Page:      1,
			PageSize:  50,
		},
	}
}

// HTTPRequest creates an HTTP request for the ProjectStatusUpdateListRequest.
func (p ProjectStatusUpdateListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/projects/updates.json"
	if p.Path.ProjectID > 0 {
		uri = fmt.Sprintf("%s/projects/api/v3/projects/%d/updates.json", server, p.Path.ProjectID)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	p.Filters.apply(req)

	return req, nil
}

// ProjectStatusUpdateListResponse contains information by multiple project
// updates matching the request filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/project-updates/get-projects-api-v3-projects-updates-json
//
// sparsefields:list
type ProjectStatusUpdateListResponse struct {
	request ProjectStatusUpdateListRequest

	// Meta contains metadata about the response, including pagination details.
	Meta twapi.ListMeta `json:"meta"`

	// ProjectUpdates is the list of project updates matching the request filters.
	ProjectUpdates []ProjectStatusUpdate `json:"projectUpdates"`

	// Included contains related objects included in the response.
	Included struct {
		// Users contains the users referenced by the updates, such as their
		// authors and the users who liked them.
		//
		// The key is the string representation of the user ID.
		Users map[string]User `json:"users,omitempty"`

		// Projects contains the projects the updates belong to.
		//
		// The key is the string representation of the project ID.
		Projects map[string]Project `json:"projects,omitempty"`
	} `json:"included"`
}

// HandleHTTPResponse handles the HTTP response for the
// ProjectStatusUpdateListResponse. If some unexpected HTTP status code is
// returned by the API, a twapi.HTTPError is returned.
func (p *ProjectStatusUpdateListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list project updates")
	}

	if err := json.NewDecoder(resp.Body).Decode(p); err != nil {
		return fmt.Errorf("failed to decode list project updates response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (p *ProjectStatusUpdateListResponse) SetRequest(req ProjectStatusUpdateListRequest) {
	p.request = req
	p.Meta.ResolveCount(req.Filters.CountMode)
}

// Iterate returns the request set to the next page, if available. If there
// are no more pages, a nil request is returned.
func (p *ProjectStatusUpdateListResponse) Iterate() *ProjectStatusUpdateListRequest {
	if !p.Meta.Page.HasMore {
		return nil
	}
	req := p.request
	req.Filters.Page++
	return &req
}

// ProjectStatusUpdateList retrieves multiple project updates using the provided
// request and returns the response.
func ProjectStatusUpdateList(
	ctx context.Context,
	engine *twapi.Engine,
	req ProjectStatusUpdateListRequest,
) (*ProjectStatusUpdateListResponse, error) {
	return twapi.Execute[ProjectStatusUpdateListRequest, *ProjectStatusUpdateListResponse](ctx, engine, req)
}
