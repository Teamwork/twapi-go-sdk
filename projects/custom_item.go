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
	_ twapi.HTTPRequester = (*CustomItemCreateRequest)(nil)
	_ twapi.HTTPResponser = (*CustomItemCreateResponse)(nil)
	_ twapi.HTTPRequester = (*CustomItemUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*CustomItemUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*CustomItemDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*CustomItemDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*CustomItemGetRequest)(nil)
	_ twapi.HTTPResponser = (*CustomItemGetResponse)(nil)
	_ twapi.HTTPRequester = (*CustomItemListRequest)(nil)
	_ twapi.HTTPResponser = (*CustomItemListResponse)(nil)
)

// CustomItemState represents the state of a custom item type.
type CustomItemState string

// Supported custom item states.
const (
	CustomItemStateActive   CustomItemState = "active"
	CustomItemStateArchived CustomItemState = "archived"
	CustomItemStateDeleted  CustomItemState = "deleted"
)

// CustomItemSideload identifies the related entities that can be requested
// alongside a custom item type via the API's include mechanism.
type CustomItemSideload string

// Supported custom item sideloads.
const (
	CustomItemSideloadCreatedBy          CustomItemSideload = "createdBy"
	CustomItemSideloadUpdatedBy          CustomItemSideload = "updatedBy"
	CustomItemSideloadDeletedBy          CustomItemSideload = "deletedBy"
	CustomItemSideloadCustomItemViews    CustomItemSideload = "customItemViews"
	CustomItemSideloadCustomItemFields   CustomItemSideload = "customItemFields"
	CustomItemSideloadCustomItemSections CustomItemSideload = "customItemSections"
)

// CustomItemOrderBy identifies the attributes a custom item list can be
// ordered by.
type CustomItemOrderBy string

// Supported custom item order-by values.
const (
	CustomItemOrderByName CustomItemOrderBy = "name"
)

// CustomItem is a user-defined entity type that can be added to a project to
// capture data that is not part of the built-in schema. Examples include
// "Contracts", "Leads", "Deals" or any other domain-specific record type. Each
// custom item type owns its own set of fields (see CustomItemField) and
// records (see CustomItemRecord), and is scoped to one or more projects.
//
// More information can be found at:
// https://support.teamwork.com/projects/projects-area/custom-items
type CustomItem struct {
	// ID is the unique identifier of the custom item type.
	ID int64 `json:"id"`

	// DisplayName is the display name of the custom item type, used in lists
	// and navigation.
	DisplayName string `json:"displayName"`

	// LabelSingular is the singular noun used to refer to a single record of
	// this type (for example "Contract").
	LabelSingular string `json:"labelSingular"`

	// LabelPlural is the plural noun used to refer to records of this type
	// (for example "Contracts").
	LabelPlural string `json:"labelPlural"`

	// State indicates whether the custom item type is active or deleted.
	State CustomItemState `json:"state"`

	// Relations lists the entities the custom item type is related to, such
	// as the projects it is available on.
	Relations []twapi.Relationship `json:"relations"`

	// Fields lists the field definitions that belong to this custom item type.
	// Each entry is a sideload reference to the underlying CustomItemField.
	Fields []twapi.Relationship `json:"fields"`

	// Views lists the views defined on this custom item type.
	Views []twapi.Relationship `json:"views"`

	// Sections lists the sections defined on this custom item type. Records
	// can optionally belong to a section.
	Sections []twapi.Relationship `json:"sections"`

	// CreatedBy identifies the user that created the custom item type.
	CreatedBy *twapi.Relationship `json:"createdBy"`

	// CreatedAt is the date and time when the custom item type was created.
	CreatedAt *time.Time `json:"createdAt"`

	// UpdatedBy identifies the user that last updated the custom item type.
	UpdatedBy *twapi.Relationship `json:"updatedBy"`

	// UpdatedAt is the date and time when the custom item type was last
	// updated.
	UpdatedAt *time.Time `json:"updatedAt"`

	// DeletedBy identifies the user that deleted the custom item type, if
	// applicable.
	DeletedBy *twapi.Relationship `json:"deletedBy"`

	// DeletedAt is the date and time when the custom item type was deleted,
	// if applicable.
	DeletedAt *time.Time `json:"deletedAt"`
}

// CustomItemOptions controls server-side side effects for a custom item write
// (notifications, webhooks). All fields are optional and default to the API's
// behaviour when nil.
type CustomItemOptions struct {
	// UseNotifyViaTWIM controls whether notifications are sent via Teamwork
	// Instant Messenger as part of this change.
	UseNotifyViaTWIM *bool `json:"useNotifyViaTWIM,omitempty"`

	// FireWebhook controls whether webhooks are fired as part of this change.
	FireWebhook *bool `json:"fireWebhook,omitempty"`

	// ApplyDefaultConfiguration applies the workspace default set of fields,
	// sections and views to the newly created custom item type. Only honoured
	// on create.
	ApplyDefaultConfiguration *bool `json:"applyDefaultConfiguration,omitempty"`
}

// CustomItemCreateRequestPath contains the path parameters for creating a
// custom item type.
type CustomItemCreateRequestPath struct {
	// ProjectID is the unique identifier of the project the custom item type
	// will be created on.
	ProjectID int64
}

// CustomItemCreateRequest represents the request body for creating a new
// custom item type on a project.
type CustomItemCreateRequest struct {
	// Path contains the path parameters for the request.
	Path CustomItemCreateRequestPath `json:"-"`

	// DisplayName is the display name of the custom item type. This field is
	// required.
	DisplayName string `json:"displayName"`

	// Description is an optional human-readable description for the custom
	// item type.
	Description *string `json:"description,omitempty"`

	// LabelSingular is the singular label (for example "Contract"). When
	// omitted, the API derives it from DisplayName.
	LabelSingular *string `json:"labelSingular,omitempty"`

	// LabelPlural is the plural label (for example "Contracts"). When
	// omitted, the API derives it from DisplayName.
	LabelPlural *string `json:"labelPlural,omitempty"`

	// Options controls notification and webhook side effects.
	Options CustomItemOptions `json:"-"`
}

// NewCustomItemCreateRequest creates a new CustomItemCreateRequest with the
// required project ID and display name populated.
func NewCustomItemCreateRequest(projectID int64, displayName string) CustomItemCreateRequest {
	return CustomItemCreateRequest{
		Path:        CustomItemCreateRequestPath{ProjectID: projectID},
		DisplayName: displayName,
	}
}

// HTTPRequest creates an HTTP request for the CustomItemCreateRequest.
func (c CustomItemCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if c.Path.ProjectID == 0 {
		return nil, fmt.Errorf("a project ID is required to create a custom item type")
	}
	uri := fmt.Sprintf("%s/projects/api/v3/projects/%d/customitems.json", server, c.Path.ProjectID)

	payload := struct {
		CustomItem CustomItemCreateRequest `json:"customItem"`
		Options    CustomItemOptions       `json:"customItemOptions"`
	}{CustomItem: c, Options: c.Options}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create custom item request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// CustomItemCreateResponse represents the response body for creating a new
// custom item type.
type CustomItemCreateResponse struct {
	// CustomItem is the created custom item type.
	CustomItem CustomItem `json:"customItem"`
}

// HandleHTTPResponse handles the HTTP response for the
// CustomItemCreateResponse. If some unexpected HTTP status code is returned by
// the API, a twapi.HTTPError is returned.
func (c *CustomItemCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create custom item")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode create custom item response: %w", err)
	}
	if c.CustomItem.ID == 0 {
		return fmt.Errorf("create custom item response does not contain a valid identifier")
	}
	return nil
}

// CustomItemCreate creates a new custom item type using the provided request
// and returns the response.
func CustomItemCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomItemCreateRequest,
) (*CustomItemCreateResponse, error) {
	return twapi.Execute[CustomItemCreateRequest, *CustomItemCreateResponse](ctx, engine, req)
}

// CustomItemUpdateRequestPath contains the path parameters for updating a
// custom item type.
type CustomItemUpdateRequestPath struct {
	// ID is the unique identifier of the custom item type to be updated.
	ID int64
}

// CustomItemUpdateRequest represents the request body for updating a custom
// item type. All non-path fields are optional. Fields left as nil are not
// modified.
type CustomItemUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path CustomItemUpdateRequestPath `json:"-"`

	// DisplayName is the new display name of the custom item type.
	DisplayName *string `json:"displayName,omitempty"`

	// Description is an optional new description for the custom item type.
	Description *string `json:"description,omitempty"`

	// LabelSingular is the new singular label.
	LabelSingular *string `json:"labelSingular,omitempty"`

	// LabelPlural is the new plural label.
	LabelPlural *string `json:"labelPlural,omitempty"`

	// Options controls notification and webhook side effects. The
	// ApplyDefaultConfiguration flag is ignored on update.
	Options CustomItemOptions `json:"-"`
}

// NewCustomItemUpdateRequest creates a new CustomItemUpdateRequest with the
// provided custom item ID populated.
func NewCustomItemUpdateRequest(customItemID int64) CustomItemUpdateRequest {
	return CustomItemUpdateRequest{
		Path: CustomItemUpdateRequestPath{ID: customItemID},
	}
}

// HTTPRequest creates an HTTP request for the CustomItemUpdateRequest.
func (c CustomItemUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if c.Path.ID == 0 {
		return nil, fmt.Errorf("a custom item ID is required to update a custom item type")
	}
	uri := fmt.Sprintf("%s/projects/api/v3/customitems/%d.json", server, c.Path.ID)

	payload := struct {
		CustomItem CustomItemUpdateRequest `json:"customItem"`
		Options    CustomItemOptions       `json:"customItemOptions"`
	}{CustomItem: c, Options: c.Options}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update custom item request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// CustomItemUpdateResponse represents the response body for updating a custom
// item type.
type CustomItemUpdateResponse struct {
	// CustomItem is the updated custom item type.
	CustomItem CustomItem `json:"customItem"`
}

// HandleHTTPResponse handles the HTTP response for the
// CustomItemUpdateResponse. If some unexpected HTTP status code is returned
// by the API, a twapi.HTTPError is returned.
func (c *CustomItemUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update custom item")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode update custom item response: %w", err)
	}
	return nil
}

// CustomItemUpdate updates a custom item type using the provided request and
// returns the response.
func CustomItemUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomItemUpdateRequest,
) (*CustomItemUpdateResponse, error) {
	return twapi.Execute[CustomItemUpdateRequest, *CustomItemUpdateResponse](ctx, engine, req)
}

// CustomItemDeleteRequestPath contains the path parameters for deleting a
// custom item type.
type CustomItemDeleteRequestPath struct {
	// ID is the unique identifier of the custom item type to be deleted.
	ID int64
}

// CustomItemDeleteRequest represents the request body for deleting a custom
// item type. Deleting a type also deletes its records and fields.
type CustomItemDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path CustomItemDeleteRequestPath

	// Options controls notification and webhook side effects.
	Options CustomItemOptions `json:"-"`
}

// NewCustomItemDeleteRequest creates a new CustomItemDeleteRequest with the
// provided custom item ID.
func NewCustomItemDeleteRequest(customItemID int64) CustomItemDeleteRequest {
	return CustomItemDeleteRequest{
		Path: CustomItemDeleteRequestPath{ID: customItemID},
	}
}

// HTTPRequest creates an HTTP request for the CustomItemDeleteRequest.
func (c CustomItemDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if c.Path.ID == 0 {
		return nil, fmt.Errorf("a custom item ID is required to delete a custom item type")
	}
	uri := fmt.Sprintf("%s/projects/api/v3/customitems/%d.json", server, c.Path.ID)

	payload := struct {
		Options CustomItemOptions `json:"customItemOptions"`
	}{Options: c.Options}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode delete custom item request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// CustomItemDeleteResponse represents the response body for deleting a custom
// item type.
type CustomItemDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the
// CustomItemDeleteResponse.
func (c *CustomItemDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusNoContent {
		return twapi.NewHTTPError(resp, "failed to delete custom item")
	}
	return nil
}

// CustomItemDelete deletes a custom item type using the provided request and
// returns the response.
func CustomItemDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomItemDeleteRequest,
) (*CustomItemDeleteResponse, error) {
	return twapi.Execute[CustomItemDeleteRequest, *CustomItemDeleteResponse](ctx, engine, req)
}

// CustomItemGetRequestPath contains the path parameters for loading a single
// custom item type.
type CustomItemGetRequestPath struct {
	// ID is the unique identifier of the custom item type to be retrieved.
	ID int64
}

// CustomItemGetRequest represents the request body for loading a single
// custom item type.
type CustomItemGetRequest struct {
	// Path contains the path parameters for the request.
	Path CustomItemGetRequestPath

	// Include is an optional list of related entities to sideload alongside
	// the custom item. Use the CustomItemSideload constants.
	Include []CustomItemSideload
}

// NewCustomItemGetRequest creates a new CustomItemGetRequest with the
// provided custom item ID.
func NewCustomItemGetRequest(customItemID int64) CustomItemGetRequest {
	return CustomItemGetRequest{
		Path: CustomItemGetRequestPath{ID: customItemID},
	}
}

// HTTPRequest creates an HTTP request for the CustomItemGetRequest.
func (c CustomItemGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if c.Path.ID == 0 {
		return nil, fmt.Errorf("a custom item ID is required to retrieve a custom item type")
	}
	uri := fmt.Sprintf("%s/projects/api/v3/customitems/%d.json", server, c.Path.ID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	if len(c.Include) > 0 {
		includes := make([]string, len(c.Include))
		for i, include := range c.Include {
			includes[i] = string(include)
		}
		query := req.URL.Query()
		query.Set("include", strings.Join(includes, ","))
		req.URL.RawQuery = query.Encode()
	}

	return req, nil
}

// CustomItemGetResponse contains all the information related to a custom item
// type.
type CustomItemGetResponse struct {
	// CustomItem is the retrieved custom item type.
	CustomItem CustomItem `json:"customItem"`

	// Included carries any sideloaded entities requested via Include.
	Included CustomItemIncluded `json:"included"`
}

// HandleHTTPResponse handles the HTTP response for the CustomItemGetResponse.
func (c *CustomItemGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve custom item")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode retrieve custom item response: %w", err)
	}
	return nil
}

// CustomItemGet retrieves a single custom item type using the provided
// request.
func CustomItemGet(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomItemGetRequest,
) (*CustomItemGetResponse, error) {
	return twapi.Execute[CustomItemGetRequest, *CustomItemGetResponse](ctx, engine, req)
}

// CustomItemListRequestPath contains the path parameters for listing the
// custom item types on a project.
type CustomItemListRequestPath struct {
	// ProjectID is the unique identifier of the project to list custom item
	// types for.
	ProjectID int64
}

// CustomItemListRequestFilters contains the filters for listing custom item
// types on a project.
type CustomItemListRequestFilters struct {
	// SearchTerm filters custom items by display name or labels.
	SearchTerm string

	// IDs restricts the result to the given custom item IDs.
	IDs []int64

	// OrderBy sorts the results. Use the CustomItemOrderBy constants.
	OrderBy CustomItemOrderBy

	// OrderMode is the sort direction.
	OrderMode twapi.OrderMode

	// IncludeArchivedProjects includes custom items from archived projects in
	// the result.
	IncludeArchivedProjects *bool

	// ShowDeleted includes deleted custom items in the result.
	ShowDeleted *bool

	// Include is an optional list of related entities to sideload. Use the
	// CustomItemSideload constants.
	Include []CustomItemSideload

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of custom items to retrieve per page. Defaults
	// to 50.
	PageSize int64
}

func (c CustomItemListRequestFilters) apply(req *http.Request) {
	query := req.URL.Query()
	if c.SearchTerm != "" {
		query.Set("searchTerm", c.SearchTerm)
	}
	if len(c.IDs) > 0 {
		ids := make([]string, len(c.IDs))
		for i, id := range c.IDs {
			ids[i] = strconv.FormatInt(id, 10)
		}
		query.Set("ids", strings.Join(ids, ","))
	}
	if c.OrderBy != "" {
		query.Set("orderBy", string(c.OrderBy))
	}
	if c.OrderMode != "" {
		query.Set("orderMode", string(c.OrderMode))
	}
	if c.IncludeArchivedProjects != nil {
		query.Set("includeArchivedProjects", strconv.FormatBool(*c.IncludeArchivedProjects))
	}
	if c.ShowDeleted != nil {
		query.Set("showDeleted", strconv.FormatBool(*c.ShowDeleted))
	}
	if len(c.Include) > 0 {
		includes := make([]string, len(c.Include))
		for i, include := range c.Include {
			includes[i] = string(include)
		}
		query.Set("include", strings.Join(includes, ","))
	}
	if c.Page > 0 {
		query.Set("page", strconv.FormatInt(c.Page, 10))
	}
	if c.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(c.PageSize, 10))
	}
	query.Set("skipCounts", "true")
	req.URL.RawQuery = query.Encode()
}

// CustomItemListRequest represents the request body for listing the custom
// item types on a project.
type CustomItemListRequest struct {
	// Path contains the path parameters for the request.
	Path CustomItemListRequestPath

	// Filters contains the filters for listing custom item types.
	Filters CustomItemListRequestFilters
}

// NewCustomItemListRequest creates a new CustomItemListRequest scoped to the
// provided project with default pagination values.
func NewCustomItemListRequest(projectID int64) CustomItemListRequest {
	return CustomItemListRequest{
		Path: CustomItemListRequestPath{ProjectID: projectID},
		Filters: CustomItemListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the CustomItemListRequest.
func (c CustomItemListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if c.Path.ProjectID == 0 {
		return nil, fmt.Errorf("a project ID is required to list custom item types")
	}
	uri := fmt.Sprintf("%s/projects/api/v3/projects/%d/customitems.json", server, c.Path.ProjectID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	c.Filters.apply(req)

	return req, nil
}

// CustomItemListResponse contains information about multiple custom item
// types on a project.
type CustomItemListResponse struct {
	request CustomItemListRequest

	// Meta contains pagination information for the response.
	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`

	// CustomItems is the list of custom item types matching the request
	// filters.
	CustomItems []CustomItem `json:"customItems"`

	// Included carries any sideloaded entities requested via Filters.Include.
	Included CustomItemIncluded `json:"included"`
}

// CustomItemIncluded carries the sideloaded entities returned alongside a
// custom item type. Fields are populated only when the matching include
// option is requested.
type CustomItemIncluded struct {
	// Fields is the set of CustomItemField definitions for this custom item
	// type, keyed by stringified field ID.
	Fields map[string]CustomItemField `json:"customItemFields,omitempty"`

	// Sections is the set of CustomItemSection definitions for this custom
	// item type, keyed by stringified section ID.
	Sections map[string]CustomItemSection `json:"customItemSections,omitempty"`
}

// CustomItemSection is a section that records of a given custom item type can
// optionally belong to.
type CustomItemSection struct {
	// ID is the unique identifier of the section.
	ID int64 `json:"id"`

	// Name is the display name of the section.
	Name string `json:"name"`

	// DisplayOrder is the relative position of the section within the type.
	DisplayOrder float64 `json:"displayOrder"`
}

// HandleHTTPResponse handles the HTTP response for the CustomItemListResponse.
func (c *CustomItemListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list custom items")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode list custom items response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response, enabling pagination
// via Iterate.
func (c *CustomItemListResponse) SetRequest(req CustomItemListRequest) {
	c.request = req
}

// Iterate returns the request set to the next page, if available. If there
// are no more pages, a nil request is returned.
func (c *CustomItemListResponse) Iterate() *CustomItemListRequest {
	if !c.Meta.Page.HasMore {
		return nil
	}
	req := c.request
	req.Filters.Page++
	return &req
}

// CustomItemList retrieves the custom item types on a project using the
// provided request.
func CustomItemList(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomItemListRequest,
) (*CustomItemListResponse, error) {
	return twapi.Execute[CustomItemListRequest, *CustomItemListResponse](ctx, engine, req)
}
