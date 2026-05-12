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
	_ twapi.HTTPRequester = (*CustomFieldCreateRequest)(nil)
	_ twapi.HTTPResponser = (*CustomFieldCreateResponse)(nil)
	_ twapi.HTTPRequester = (*CustomFieldUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*CustomFieldUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*CustomFieldDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*CustomFieldDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*CustomFieldGetRequest)(nil)
	_ twapi.HTTPResponser = (*CustomFieldGetResponse)(nil)
	_ twapi.HTTPRequester = (*CustomFieldListRequest)(nil)
	_ twapi.HTTPResponser = (*CustomFieldListResponse)(nil)
)

// CustomFieldType represents the data type of a custom field.
type CustomFieldType string

// Supported custom field types.
const (
	CustomFieldTypeTextShort     CustomFieldType = "text-short"
	CustomFieldTypeTextLong      CustomFieldType = "text-long"
	CustomFieldTypeNumberDecimal CustomFieldType = "number-decimal"
	CustomFieldTypeNumberInteger CustomFieldType = "number-integer"
	CustomFieldTypeFinancial     CustomFieldType = "financial"
	CustomFieldTypeDropdown      CustomFieldType = "dropdown"
	CustomFieldTypeMultiselect   CustomFieldType = "multiselect"
	CustomFieldTypeCheckbox      CustomFieldType = "checkbox"
	CustomFieldTypePercentage    CustomFieldType = "percentage"
	CustomFieldTypeURL           CustomFieldType = "url"
	CustomFieldTypeTeamworkURL   CustomFieldType = "tw-url"
	CustomFieldTypeUser          CustomFieldType = "user"
	CustomFieldTypeRating        CustomFieldType = "rating"
	CustomFieldTypeDate          CustomFieldType = "date"
	CustomFieldTypePhone         CustomFieldType = "phone"
	CustomFieldTypeEmail         CustomFieldType = "email"
	CustomFieldTypeStatus        CustomFieldType = "status"
)

// CustomFieldEntity represents the type of entity a custom field is attached
// to.
type CustomFieldEntity string

// Supported custom field entities.
const (
	CustomFieldEntityGlobal  CustomFieldEntity = "all"
	CustomFieldEntityProject CustomFieldEntity = "project"
	CustomFieldEntityTask    CustomFieldEntity = "task"
	CustomFieldEntityCompany CustomFieldEntity = "company"
	CustomFieldEntityUser    CustomFieldEntity = "user"
)

// CustomFieldUnit defines the unit of numeric value.
type CustomFieldUnit string

// Supported custom field value units.
const (
	CustomFieldUnitCurrency            CustomFieldUnit = "currency"
	CustomFieldUnitDuration            CustomFieldUnit = "duration"
	CustomFieldUnitDate                CustomFieldUnit = "date"
	CustomFieldUnitPercent             CustomFieldUnit = "percent"
	CustomFieldUnitCurrencyPerDuration CustomFieldUnit = "currency/duration"
	CustomFieldUnitDurationPerCurrency CustomFieldUnit = "duration/currency"
)

// UnmarshalJSON implements the json.Unmarshaler interface for
// CustomFieldUnit converting from ID to string.
func (c *CustomFieldUnit) UnmarshalJSON(data []byte) error {
	var id int64
	if err := json.Unmarshal(data, &id); err != nil {
		return err
	}
	switch id {
	case 1:
		*c = CustomFieldUnitCurrency
	case 2:
		*c = CustomFieldUnitDuration
	case 3:
		*c = CustomFieldUnitDate
	case 4:
		*c = CustomFieldUnitPercent
	case 1001:
		*c = CustomFieldUnitCurrencyPerDuration
	case 1002:
		*c = CustomFieldUnitDurationPerCurrency
	default:
		return fmt.Errorf("unknown custom field value unit ID: %d", id)
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface for CustomFieldUnit
// converting from string to ID.
func (c CustomFieldUnit) MarshalJSON() ([]byte, error) {
	var id int64
	switch c {
	case CustomFieldUnitCurrency:
		id = 1
	case CustomFieldUnitDuration:
		id = 2
	case CustomFieldUnitDate:
		id = 3
	case CustomFieldUnitPercent:
		id = 4
	case CustomFieldUnitCurrencyPerDuration:
		id = 1001
	case CustomFieldUnitDurationPerCurrency:
		id = 1002
	default:
		return nil, fmt.Errorf("unknown custom field value unit: %s", c)
	}
	return json.Marshal(id)
}

// CustomFieldOptions represents type-specific options for a custom field. The
// structure of the options varies based on the custom field type.
type CustomFieldOptions interface {
	options()
}

// CustomFieldOptionsDropdown store the options for a dropdown type.
type CustomFieldOptionsDropdown struct {
	// Choices is the list of choices available for the dropdown.
	Choices []CustomFieldOptionsDropdownChoice `json:"choices"`
}

func (CustomFieldOptionsDropdown) options() {}

// CustomFieldOptionsDropdownChoice store the properties of a single choice
// in the dropdown.
type CustomFieldOptionsDropdownChoice struct {
	// Value is the display value of the choice.
	Value string `json:"value"`

	// Color is the hexadecimal color code associated with the choice, without the
	// leading "#".
	Color twapi.HexColor `json:"color"`
}

// CustomFieldOptionsRating store the options for rating type.
type CustomFieldOptionsRating struct {
	// Icon is the name of the icon used for the rating, such as "star" or
	// "heart".
	Icon string `json:"icon"`

	// Color is the hexadecimal color code associated with the rating, without the
	// leading "#".
	Color twapi.HexColor `json:"color"`
}

func (CustomFieldOptionsRating) options() {}

// CustomFieldOptionsNumberDecimal store the options for decimal type.
type CustomFieldOptionsNumberDecimal struct {
	// DecimalPoints is the number of decimal points allowed for the decimal type.
	DecimalPoints *int `json:"decimals,omitempty"`
}

func (CustomFieldOptionsNumberDecimal) options() {}

type customField struct {
	// ID is the unique identifier of the custom field.
	ID int64 `json:"id"`

	// Name is the display name of the custom field.
	Name string `json:"name"`

	// Description is an optional description for the custom field.
	Description string `json:"description"`

	// Type is the data type of the custom field. Common values include "text",
	// "number", "date", "dropdown", "checkbox", "url", "status", "currency" and
	// "formula".
	Type CustomFieldType `json:"type"`

	// Entity is the type of entity this custom field can be applied to. Common
	// values include "project", "task", "company" and "user".
	Entity CustomFieldEntity `json:"entity"`

	// Required indicates whether the custom field must have a value when set on
	// an entity.
	Required bool `json:"required"`

	// Formula is the formula expression for "formula" type custom fields.
	Formula *string `json:"formula"`

	// CurrencyCode is the ISO currency code for "currency" type custom fields.
	CurrencyCode *string `json:"currencyCode"`

	// Unit is the unit identifier associated with the custom field, when
	// applicable.
	Unit *CustomFieldUnit `json:"unitId"`

	// Project is the project the custom field is scoped to. It is nil for
	// installation-level custom fields.
	Project *twapi.Relationship `json:"project"`

	// CreatedBy is the unique identifier of the user who created the custom
	// field.
	CreatedBy int64 `json:"createdBy"`

	// CreatedAt is the date and time when the custom field was created.
	CreatedAt *time.Time `json:"createdAt"`

	// UpdatedBy is the unique identifier of the user who last updated the custom
	// field.
	UpdatedBy *int64 `json:"updatedBy"`

	// UpdatedAt is the date and time when the custom field was last updated.
	UpdatedAt *time.Time `json:"updatedAt"`

	// Deleted indicates whether the custom field has been deleted.
	Deleted bool `json:"deleted"`

	// DeletedBy is the unique identifier of the user who deleted the custom
	// field, if applicable.
	DeletedBy *int64 `json:"deletedBy"`

	// DeletedAt is the date and time when the custom field was deleted, if
	// applicable.
	DeletedAt *time.Time `json:"deletedAt"`
}

// CustomField is a user-defined attribute that can be attached to projects,
// tasks, companies and other entities to capture data that is not part of the
// built-in schema. Custom fields can be scoped at the installation level so
// they are available across the workspace, or at the project level so they are
// only available for entities in that project.
//
// More information can be found at:
// https://support.teamwork.com/projects/custom-fields/create-and-manage-custom-fields
// https://support.teamwork.com/projects/custom-fields/use-custom-fields
type CustomField struct {
	customField

	// Options holds the available options for dropdown-style custom fields.
	Options CustomFieldOptions `json:"options"`
}

// UnmarshalJSON implements the json.Unmarshaler interface for CustomField to
// handle different options.
func (c *CustomField) UnmarshalJSON(data []byte) error {
	var customField customField
	if err := json.Unmarshal(data, &customField); err != nil {
		return err
	}
	*c = CustomField{customField: customField}

	switch c.Type {
	case CustomFieldTypeDropdown:
		var options CustomFieldOptionsDropdown
		if err := json.Unmarshal(data, &options); err != nil {
			return fmt.Errorf("failed to decode dropdown options: %w", err)
		}
		c.Options = options
	case CustomFieldTypeRating:
		var options CustomFieldOptionsRating
		if err := json.Unmarshal(data, &options); err != nil {
			return fmt.Errorf("failed to decode rating options: %w", err)
		}
		c.Options = options
	case CustomFieldTypeNumberDecimal:
		var options CustomFieldOptionsNumberDecimal
		if err := json.Unmarshal(data, &options); err != nil {
			return fmt.Errorf("failed to decode number decimal options: %w", err)
		}
		c.Options = options
	}
	return nil
}

// CustomFieldCreateRequest represents the request body for creating a new
// custom field.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/custom-fields/post-projects-api-v3-customfields-json
type CustomFieldCreateRequest struct {
	// Name is the display name of the custom field. This field is required.
	Name string `json:"name"`

	// Type is the data type of the custom field. This field is required. See
	// CustomFieldType for the supported values.
	Type CustomFieldType `json:"type"`

	// Entity is the type of entity this custom field can be applied to. This
	// field is required. See CustomFieldEntity for the supported values.
	Entity CustomFieldEntity `json:"entity"`

	// Description is an optional description for the custom field.
	Description *string `json:"description,omitempty"`

	// Required indicates whether the custom field must have a value when set on
	// an entity.
	Required *bool `json:"required,omitempty"`

	// ProjectID is the unique identifier of the project the custom field is
	// scoped to. When omitted, the custom field is created at the installation
	// level.
	ProjectID *int64 `json:"projectId,omitempty"`

	// Options holds the available options for dropdown-style custom fields.
	Options CustomFieldOptions `json:"options,omitempty"`

	// Formula is the formula expression for "formula" type custom fields.
	Formula *string `json:"formula,omitempty"`

	// CurrencyCode is the ISO currency code for "currency" type custom fields.
	CurrencyCode *string `json:"currencyCode,omitempty"`

	// Unit is the unit identifier associated with the custom field, when
	// applicable.
	Unit *CustomFieldUnit `json:"unitId,omitempty"`
}

// NewCustomFieldCreateRequest creates a new CustomFieldCreateRequest with the
// required fields populated.
func NewCustomFieldCreateRequest(
	name string,
	fieldType CustomFieldType,
	entity CustomFieldEntity,
) CustomFieldCreateRequest {
	return CustomFieldCreateRequest{
		Name:   name,
		Type:   fieldType,
		Entity: entity,
	}
}

// HTTPRequest creates an HTTP request for the CustomFieldCreateRequest.
func (c CustomFieldCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/customfields.json"

	payload := struct {
		CustomField CustomFieldCreateRequest `json:"customfield"`
	}{CustomField: c}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create custom field request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// CustomFieldCreateResponse represents the response body for creating a new
// custom field.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/custom-fields/post-projects-api-v3-customfields-json
type CustomFieldCreateResponse struct {
	// CustomField is the created custom field.
	CustomField CustomField `json:"customfield"`
}

// HandleHTTPResponse handles the HTTP response for the
// CustomFieldCreateResponse. If some unexpected HTTP status code is returned by
// the API, a twapi.HTTPError is returned.
func (c *CustomFieldCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create custom field")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode create custom field response: %w", err)
	}
	if c.CustomField.ID == 0 {
		return fmt.Errorf("create custom field response does not contain a valid identifier")
	}
	return nil
}

// CustomFieldCreate creates a new custom field using the provided request and
// returns the response.
func CustomFieldCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomFieldCreateRequest,
) (*CustomFieldCreateResponse, error) {
	return twapi.Execute[CustomFieldCreateRequest, *CustomFieldCreateResponse](ctx, engine, req)
}

// CustomFieldUpdateRequestPath contains the path parameters for updating a
// custom field.
type CustomFieldUpdateRequestPath struct {
	// ID is the unique identifier of the custom field to be updated.
	ID int64
}

// CustomFieldUpdateRequest represents the request body for updating a custom
// field. Besides the identifier, all other fields are optional. When a field is
// not provided, it will not be modified.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/custom-fields/patch-projects-api-v3-customfields-custom-field-id-json
type CustomFieldUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path CustomFieldUpdateRequestPath `json:"-"`

	// Name is the display name of the custom field.
	Name *string `json:"name,omitempty"`

	// Description is an optional description for the custom field.
	Description *string `json:"description,omitempty"`

	// Required indicates whether the custom field must have a value when set on
	// an entity.
	Required *bool `json:"required,omitempty"`

	// Options holds the available options for dropdown-style custom fields.
	Options CustomFieldOptions `json:"options,omitempty"`

	// Formula is the formula expression for "formula" type custom fields.
	Formula *string `json:"formula,omitempty"`

	// CurrencyCode is the ISO currency code for "currency" type custom fields.
	CurrencyCode *string `json:"currencyCode,omitempty"`

	// Unit is the unit identifier associated with the custom field, when
	// applicable.
	Unit *CustomFieldUnit `json:"unitId,omitempty"`
}

// NewCustomFieldUpdateRequest creates a new CustomFieldUpdateRequest with the
// provided custom field ID. The ID is required to update a custom field.
func NewCustomFieldUpdateRequest(customFieldID int64) CustomFieldUpdateRequest {
	return CustomFieldUpdateRequest{
		Path: CustomFieldUpdateRequestPath{
			ID: customFieldID,
		},
	}
}

// HTTPRequest creates an HTTP request for the CustomFieldUpdateRequest.
func (c CustomFieldUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/customfields/" + strconv.FormatInt(c.Path.ID, 10) + ".json"

	payload := struct {
		CustomField CustomFieldUpdateRequest `json:"customfield"`
	}{CustomField: c}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update custom field request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// CustomFieldUpdateResponse represents the response body for updating a custom
// field.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/custom-fields/patch-projects-api-v3-customfields-custom-field-id-json
type CustomFieldUpdateResponse struct {
	// CustomField is the updated custom field.
	CustomField CustomField `json:"customfield"`
}

// HandleHTTPResponse handles the HTTP response for the
// CustomFieldUpdateResponse. If some unexpected HTTP status code is returned by
// the API, a twapi.HTTPError is returned.
func (c *CustomFieldUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update custom field")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode update custom field response: %w", err)
	}
	return nil
}

// CustomFieldUpdate updates a custom field using the provided request and
// returns the response.
func CustomFieldUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomFieldUpdateRequest,
) (*CustomFieldUpdateResponse, error) {
	return twapi.Execute[CustomFieldUpdateRequest, *CustomFieldUpdateResponse](ctx, engine, req)
}

// CustomFieldDeleteRequestPath contains the path parameters for deleting a
// custom field.
type CustomFieldDeleteRequestPath struct {
	// ID is the unique identifier of the custom field to be deleted.
	ID int64
}

// CustomFieldDeleteRequest represents the request body for deleting a custom
// field.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/custom-fields/delete-projects-api-v3-customfields-custom-field-id-json
type CustomFieldDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path CustomFieldDeleteRequestPath
}

// NewCustomFieldDeleteRequest creates a new CustomFieldDeleteRequest with the
// provided custom field ID.
func NewCustomFieldDeleteRequest(customFieldID int64) CustomFieldDeleteRequest {
	return CustomFieldDeleteRequest{
		Path: CustomFieldDeleteRequestPath{
			ID: customFieldID,
		},
	}
}

// HTTPRequest creates an HTTP request for the CustomFieldDeleteRequest.
func (c CustomFieldDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/customfields/" + strconv.FormatInt(c.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// CustomFieldDeleteResponse represents the response body for deleting a custom
// field.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/custom-fields/delete-projects-api-v3-customfields-custom-field-id-json
type CustomFieldDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the
// CustomFieldDeleteResponse. If some unexpected HTTP status code is returned by
// the API, a twapi.HTTPError is returned.
func (c *CustomFieldDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusNoContent {
		return twapi.NewHTTPError(resp, "failed to delete custom field")
	}
	return nil
}

// CustomFieldDelete deletes a custom field using the provided request and
// returns the response.
func CustomFieldDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomFieldDeleteRequest,
) (*CustomFieldDeleteResponse, error) {
	return twapi.Execute[CustomFieldDeleteRequest, *CustomFieldDeleteResponse](ctx, engine, req)
}

// CustomFieldGetRequestPath contains the path parameters for loading a single
// custom field.
type CustomFieldGetRequestPath struct {
	// ID is the unique identifier of the custom field to be retrieved.
	ID int64
}

// CustomFieldGetRequest represents the request body for loading a single
// custom field.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/custom-fields/get-projects-api-v3-customfields-custom-field-id-json
type CustomFieldGetRequest struct {
	// Path contains the path parameters for the request.
	Path CustomFieldGetRequestPath
}

// NewCustomFieldGetRequest creates a new CustomFieldGetRequest with the
// provided custom field ID.
func NewCustomFieldGetRequest(customFieldID int64) CustomFieldGetRequest {
	return CustomFieldGetRequest{
		Path: CustomFieldGetRequestPath{
			ID: customFieldID,
		},
	}
}

// HTTPRequest creates an HTTP request for the CustomFieldGetRequest.
func (c CustomFieldGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/customfields/" + strconv.FormatInt(c.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// CustomFieldGetResponse contains all the information related to a custom
// field.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/custom-fields/get-projects-api-v3-customfields-custom-field-id-json
type CustomFieldGetResponse struct {
	CustomField CustomField `json:"customfield"`
}

// HandleHTTPResponse handles the HTTP response for the CustomFieldGetResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (c *CustomFieldGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve custom field")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode retrieve custom field response: %w", err)
	}
	return nil
}

// CustomFieldGet retrieves a single custom field using the provided request and
// returns the response.
func CustomFieldGet(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomFieldGetRequest,
) (*CustomFieldGetResponse, error) {
	return twapi.Execute[CustomFieldGetRequest, *CustomFieldGetResponse](ctx, engine, req)
}

// CustomFieldListRequestFilters contains the filters for loading multiple
// custom fields.
type CustomFieldListRequestFilters struct {
	// SearchTerm is an optional search term to filter custom fields by name.
	SearchTerm string

	// IDs is an optional list of custom field IDs to retrieve.
	IDs []int64

	// Entities is an optional list of entity types to filter custom fields by.
	// See CustomFieldEntity for the supported values.
	Entities []CustomFieldEntity

	// ProjectIDs is an optional list of project IDs to filter custom fields by.
	ProjectIDs []int64

	// OnlySiteLevel indicates whether to return only installation-level custom
	// fields.
	OnlySiteLevel *bool

	// OnlyProjectLevel indicates whether to return only project-level custom
	// fields.
	OnlyProjectLevel *bool

	// IncludeSiteLevel indicates whether to also include installation-level
	// custom fields when filtering by project.
	IncludeSiteLevel *bool

	// ShowDeleted indicates whether to include deleted custom fields in the
	// results. Defaults to false.
	ShowDeleted *bool

	// OrderBy is the field to sort the results by. Valid values are "name",
	// "project", "dateCreated" and "dateUpdated".
	OrderBy string

	// OrderMode is the direction to sort the results in. See twapi.OrderMode for
	// the supported values.
	OrderMode twapi.OrderMode

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of custom fields to retrieve per page. Defaults to
	// 50.
	PageSize int64
}

func (c CustomFieldListRequestFilters) apply(req *http.Request) {
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
	if len(c.Entities) > 0 {
		entities := make([]string, len(c.Entities))
		for i, entity := range c.Entities {
			entities[i] = string(entity)
		}
		query.Set("entities", strings.Join(entities, ","))
	}
	if len(c.ProjectIDs) > 0 {
		ids := make([]string, len(c.ProjectIDs))
		for i, id := range c.ProjectIDs {
			ids[i] = strconv.FormatInt(id, 10)
		}
		query.Set("projectIds", strings.Join(ids, ","))
	}
	if c.OnlySiteLevel != nil {
		query.Set("onlySiteLevel", strconv.FormatBool(*c.OnlySiteLevel))
	}
	if c.OnlyProjectLevel != nil {
		query.Set("onlyProjectLevel", strconv.FormatBool(*c.OnlyProjectLevel))
	}
	if c.IncludeSiteLevel != nil {
		query.Set("includeSiteLevel", strconv.FormatBool(*c.IncludeSiteLevel))
	}
	if c.ShowDeleted != nil {
		query.Set("showDeleted", strconv.FormatBool(*c.ShowDeleted))
	}
	if c.OrderBy != "" {
		query.Set("orderBy", c.OrderBy)
	}
	if c.OrderMode != "" {
		query.Set("orderMode", string(c.OrderMode))
	}
	if c.Page > 0 {
		query.Set("page", strconv.FormatInt(c.Page, 10))
	}
	if c.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(c.PageSize, 10))
	}
	req.URL.RawQuery = query.Encode()
}

// CustomFieldListRequest represents the request body for loading multiple
// custom fields.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/custom-fields/get-projects-api-v3-customfields-json
type CustomFieldListRequest struct {
	// Filters contains the filters for loading multiple custom fields.
	Filters CustomFieldListRequestFilters
}

// NewCustomFieldListRequest creates a new CustomFieldListRequest with default
// values.
func NewCustomFieldListRequest() CustomFieldListRequest {
	return CustomFieldListRequest{
		Filters: CustomFieldListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the CustomFieldListRequest.
func (c CustomFieldListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/customfields.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	c.Filters.apply(req)

	return req, nil
}

// CustomFieldListResponse contains information about multiple custom fields
// matching the request filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/custom-fields/get-projects-api-v3-customfields-json
type CustomFieldListResponse struct {
	request CustomFieldListRequest

	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`
	CustomFields []CustomField `json:"customfields"`
}

// HandleHTTPResponse handles the HTTP response for the CustomFieldListResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (c *CustomFieldListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list custom fields")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode list custom fields response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (c *CustomFieldListResponse) SetRequest(req CustomFieldListRequest) {
	c.request = req
}

// Iterate returns the request set to the next page, if available. If there are
// no more pages, a nil request is returned.
func (c *CustomFieldListResponse) Iterate() *CustomFieldListRequest {
	if !c.Meta.Page.HasMore {
		return nil
	}
	req := c.request
	req.Filters.Page++
	return &req
}

// CustomFieldList retrieves multiple custom fields using the provided request
// and returns the response.
func CustomFieldList(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomFieldListRequest,
) (*CustomFieldListResponse, error) {
	return twapi.Execute[CustomFieldListRequest, *CustomFieldListResponse](ctx, engine, req)
}
