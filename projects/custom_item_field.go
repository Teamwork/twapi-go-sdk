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
	_ twapi.HTTPRequester = (*CustomItemFieldCreateRequest)(nil)
	_ twapi.HTTPResponser = (*CustomItemFieldCreateResponse)(nil)
	_ twapi.HTTPRequester = (*CustomItemFieldUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*CustomItemFieldUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*CustomItemFieldDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*CustomItemFieldDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*CustomItemFieldGetRequest)(nil)
	_ twapi.HTTPResponser = (*CustomItemFieldGetResponse)(nil)
	_ twapi.HTTPRequester = (*CustomItemFieldListRequest)(nil)
	_ twapi.HTTPResponser = (*CustomItemFieldListResponse)(nil)
)

// CustomItemFieldType identifies the data type of a custom item field.
type CustomItemFieldType string

// Supported custom item field types.
const (
	CustomItemFieldTypeTextShort     CustomItemFieldType = "text-short"
	CustomItemFieldTypeTextLong      CustomItemFieldType = "text-long"
	CustomItemFieldTypeNumberDecimal CustomItemFieldType = "number-decimal"
	CustomItemFieldTypeNumberInteger CustomItemFieldType = "number-integer"
	CustomItemFieldTypeDropdown      CustomItemFieldType = "dropdown"
	CustomItemFieldTypeMultiselect   CustomItemFieldType = "multiselect"
	CustomItemFieldTypeCheckbox      CustomItemFieldType = "checkbox"
	CustomItemFieldTypeURL           CustomItemFieldType = "url"
	CustomItemFieldTypeUser          CustomItemFieldType = "user"
	CustomItemFieldTypeDate          CustomItemFieldType = "date"
	CustomItemFieldTypeTime          CustomItemFieldType = "time"
	CustomItemFieldTypeDateTime      CustomItemFieldType = "datetime"
)

// CustomItemFieldTwType is an additional sub-classification used by the API
// for built-in dropdown semantics (such as a "status" dropdown).
type CustomItemFieldTwType string

// Supported custom item field tw types.
const (
	CustomItemFieldTwTypeStatus CustomItemFieldTwType = "status"
)

// CustomItemFieldState reports whether a field is active or has been deleted.
type CustomItemFieldState string

// Supported custom item field states.
const (
	CustomItemFieldStateActive  CustomItemFieldState = "active"
	CustomItemFieldStateDeleted CustomItemFieldState = "deleted"
)

// CustomItemField is a single field (column) defined on a custom item type.
// Records of the parent custom item type carry a value for each field, keyed
// by the field's TwID.
type CustomItemField struct {
	// ID is the unique identifier of the field.
	ID int64 `json:"id"`

	// CustomItem is the relationship to the custom item type this field
	// belongs to.
	CustomItem *twapi.Relationship `json:"customItem"`

	// DisplayName is the human-readable name of the field.
	DisplayName string `json:"displayName"`

	// Type is the data type of the field.
	Type CustomItemFieldType `json:"type"`

	// Definition carries type-specific configuration as opaque JSON. For
	// number-decimal fields this includes precision and unit, for user
	// fields the user limit and source, and so on. Callers should consult
	// the API documentation for the shape of each type's definition.
	Definition map[string]any `json:"definition,omitempty"`

	// TwType optionally classifies dropdown fields as having special
	// semantics, such as representing a status.
	TwType *CustomItemFieldTwType `json:"twType,omitempty"`

	// TwID is the opaque string identifier used as the key for this field's
	// value in a CustomItemRecord.FieldValues map.
	TwID string `json:"twId"`

	// State indicates whether the field is active or deleted.
	State CustomItemFieldState `json:"state"`

	// DisplayOrder is the relative position of the field within its custom
	// item type.
	DisplayOrder float64 `json:"displayOrder"`

	// Options lists the available choices for dropdown and multiselect
	// fields. Empty for fields of other types.
	Options []CustomItemFieldOption `json:"options,omitempty"`

	// CreatedBy identifies the user that created the field.
	CreatedBy *twapi.Relationship `json:"createdBy"`

	// CreatedAt is the date and time when the field was created.
	CreatedAt *time.Time `json:"createdAt"`

	// UpdatedBy identifies the user that last updated the field.
	UpdatedBy *twapi.Relationship `json:"updatedBy"`

	// UpdatedAt is the date and time when the field was last updated.
	UpdatedAt *time.Time `json:"updatedAt"`

	// DeletedBy identifies the user that deleted the field, if applicable.
	DeletedBy *twapi.Relationship `json:"deletedBy"`

	// DeletedAt is the date and time when the field was deleted, if
	// applicable.
	DeletedAt *time.Time `json:"deletedAt"`
}

// CustomItemFieldOption is a single choice available on a dropdown or
// multiselect field.
type CustomItemFieldOption struct {
	// ID is the unique identifier of the option.
	ID int64 `json:"id"`

	// Label is the display value of the option.
	Label string `json:"label"`

	// TwID is the opaque string identifier the API uses to refer to this
	// option in record field values.
	TwID string `json:"twId"`

	// Color is the hex colour associated with the option (no leading "#").
	Color string `json:"color"`

	// DisplayOrder is the relative position of the option within its field.
	DisplayOrder float64 `json:"displayOrder"`

	// IsDefault indicates whether the option is the default for the field.
	IsDefault bool `json:"isDefault"`
}

// CustomItemFieldOptionInput describes a single dropdown or multiselect
// option supplied on field create or update. The PositionAfterID points to
// the option that the new option should be placed after; nil means append.
type CustomItemFieldOptionInput struct {
	// Label is the display value for the option. Required.
	Label *string `json:"label,omitempty"`

	// Color is the hex colour for the option (no leading "#").
	Color *string `json:"color,omitempty"`

	// PositionAfterID places this option after the given option ID. Nil
	// appends to the end of the option list.
	PositionAfterID *int64 `json:"positionAfterId,omitempty"`
}

// CustomItemFieldOptions controls server-side side effects for a custom item
// field write.
type CustomItemFieldOptions struct {
	// UseNotifyViaTWIM controls whether notifications are sent via Teamwork
	// Instant Messenger as part of this change.
	UseNotifyViaTWIM *bool `json:"useNotifyViaTWIM,omitempty"`

	// FireWebhook controls whether webhooks are fired as part of this change.
	FireWebhook *bool `json:"fireWebhook,omitempty"`
}

// CustomItemFieldCreateRequestPath contains the path parameters for creating
// a custom item field.
type CustomItemFieldCreateRequestPath struct {
	// CustomItemID is the unique identifier of the custom item type that
	// will own the new field.
	CustomItemID int64
}

// CustomItemFieldCreateRequest represents the request body for creating a
// new field on a custom item type.
type CustomItemFieldCreateRequest struct {
	// Path contains the path parameters for the request.
	Path CustomItemFieldCreateRequestPath `json:"-"`

	// DisplayName is the human-readable name of the field. This field is
	// required.
	DisplayName string `json:"displayName"`

	// Type is the data type of the field. This field is required.
	Type CustomItemFieldType `json:"type"`

	// Definition is type-specific configuration as opaque JSON. See
	// CustomItemField.Definition for examples.
	Definition map[string]any `json:"definition,omitempty"`

	// TwType optionally tags dropdown fields with built-in semantics, such
	// as representing a status.
	TwType *CustomItemFieldTwType `json:"twType,omitempty"`

	// Options lists the available choices for dropdown and multiselect
	// fields. Ignored for fields of other types.
	Options []CustomItemFieldOptionInput `json:"options,omitempty"`

	// PositionAfterID places the new field after the field with the given
	// ID. Nil appends to the end of the field list.
	PositionAfterID *int64 `json:"positionAfterId,omitempty"`

	// FieldOptions controls notification and webhook side effects.
	FieldOptions CustomItemFieldOptions `json:"-"`
}

// NewCustomItemFieldCreateRequest creates a new CustomItemFieldCreateRequest
// with the required custom item ID, display name and type populated.
func NewCustomItemFieldCreateRequest(
	customItemID int64,
	displayName string,
	fieldType CustomItemFieldType,
) CustomItemFieldCreateRequest {
	return CustomItemFieldCreateRequest{
		Path:        CustomItemFieldCreateRequestPath{CustomItemID: customItemID},
		DisplayName: displayName,
		Type:        fieldType,
	}
}

// HTTPRequest creates an HTTP request for the CustomItemFieldCreateRequest.
func (c CustomItemFieldCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if c.Path.CustomItemID == 0 {
		return nil, fmt.Errorf("a custom item ID is required to create a custom item field")
	}
	uri := fmt.Sprintf("%s/projects/api/v3/customitems/%d/fields.json", server, c.Path.CustomItemID)

	payload := struct {
		Field   CustomItemFieldCreateRequest `json:"customItemField"`
		Options CustomItemFieldOptions       `json:"customItemFieldOptions"`
	}{Field: c, Options: c.FieldOptions}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create custom item field request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// CustomItemFieldCreateResponse represents the response body for creating a
// new custom item field.
type CustomItemFieldCreateResponse struct {
	// CustomItemField is the created field.
	CustomItemField CustomItemField `json:"customItemField"`
}

// HandleHTTPResponse handles the HTTP response for the
// CustomItemFieldCreateResponse.
func (c *CustomItemFieldCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create custom item field")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode create custom item field response: %w", err)
	}
	if c.CustomItemField.ID == 0 {
		return fmt.Errorf("create custom item field response does not contain a valid identifier")
	}
	return nil
}

// CustomItemFieldCreate creates a new custom item field using the provided
// request.
func CustomItemFieldCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomItemFieldCreateRequest,
) (*CustomItemFieldCreateResponse, error) {
	return twapi.Execute[CustomItemFieldCreateRequest, *CustomItemFieldCreateResponse](ctx, engine, req)
}

// CustomItemFieldUpdateRequestPath contains the path parameters for updating
// a custom item field.
type CustomItemFieldUpdateRequestPath struct {
	// CustomItemID is the unique identifier of the custom item type the
	// field belongs to.
	CustomItemID int64

	// ID is the unique identifier of the field to be updated.
	ID int64
}

// CustomItemFieldUpdateRequest represents the request body for updating a
// custom item field. All non-path fields are optional.
type CustomItemFieldUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path CustomItemFieldUpdateRequestPath `json:"-"`

	// DisplayName is the new display name of the field.
	DisplayName *string `json:"displayName,omitempty"`

	// Definition replaces the type-specific configuration of the field.
	Definition map[string]any `json:"definition,omitempty"`

	// PositionAfterID moves the field after the field with the given ID.
	PositionAfterID *int64 `json:"positionAfterId,omitempty"`

	// FieldOptions controls notification and webhook side effects.
	FieldOptions CustomItemFieldOptions `json:"-"`
}

// NewCustomItemFieldUpdateRequest creates a new CustomItemFieldUpdateRequest
// with the required path parameters populated.
func NewCustomItemFieldUpdateRequest(customItemID, fieldID int64) CustomItemFieldUpdateRequest {
	return CustomItemFieldUpdateRequest{
		Path: CustomItemFieldUpdateRequestPath{
			CustomItemID: customItemID,
			ID:           fieldID,
		},
	}
}

// HTTPRequest creates an HTTP request for the CustomItemFieldUpdateRequest.
func (c CustomItemFieldUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if c.Path.CustomItemID == 0 {
		return nil, fmt.Errorf("a custom item ID is required to update a custom item field")
	}
	if c.Path.ID == 0 {
		return nil, fmt.Errorf("a field ID is required to update a custom item field")
	}
	uri := fmt.Sprintf("%s/projects/api/v3/customitems/%d/fields/%d.json",
		server, c.Path.CustomItemID, c.Path.ID)

	payload := struct {
		Field   CustomItemFieldUpdateRequest `json:"customItemField"`
		Options CustomItemFieldOptions       `json:"customItemFieldOptions"`
	}{Field: c, Options: c.FieldOptions}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update custom item field request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// CustomItemFieldUpdateResponse represents the response body for updating a
// custom item field.
type CustomItemFieldUpdateResponse struct {
	// CustomItemField is the updated field.
	CustomItemField CustomItemField `json:"customItemField"`
}

// HandleHTTPResponse handles the HTTP response for the
// CustomItemFieldUpdateResponse.
func (c *CustomItemFieldUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update custom item field")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode update custom item field response: %w", err)
	}
	return nil
}

// CustomItemFieldUpdate updates a custom item field using the provided
// request.
func CustomItemFieldUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomItemFieldUpdateRequest,
) (*CustomItemFieldUpdateResponse, error) {
	return twapi.Execute[CustomItemFieldUpdateRequest, *CustomItemFieldUpdateResponse](ctx, engine, req)
}

// CustomItemFieldDeleteRequestPath contains the path parameters for deleting
// a custom item field.
type CustomItemFieldDeleteRequestPath struct {
	// CustomItemID is the unique identifier of the custom item type the
	// field belongs to.
	CustomItemID int64

	// ID is the unique identifier of the field to be deleted.
	ID int64
}

// CustomItemFieldDeleteRequest represents the request body for deleting a
// custom item field.
type CustomItemFieldDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path CustomItemFieldDeleteRequestPath

	// FieldOptions controls notification and webhook side effects.
	FieldOptions CustomItemFieldOptions `json:"-"`
}

// NewCustomItemFieldDeleteRequest creates a new CustomItemFieldDeleteRequest
// with the required path parameters populated.
func NewCustomItemFieldDeleteRequest(customItemID, fieldID int64) CustomItemFieldDeleteRequest {
	return CustomItemFieldDeleteRequest{
		Path: CustomItemFieldDeleteRequestPath{
			CustomItemID: customItemID,
			ID:           fieldID,
		},
	}
}

// HTTPRequest creates an HTTP request for the CustomItemFieldDeleteRequest.
func (c CustomItemFieldDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if c.Path.CustomItemID == 0 {
		return nil, fmt.Errorf("a custom item ID is required to delete a custom item field")
	}
	if c.Path.ID == 0 {
		return nil, fmt.Errorf("a field ID is required to delete a custom item field")
	}
	uri := fmt.Sprintf("%s/projects/api/v3/customitems/%d/fields/%d.json",
		server, c.Path.CustomItemID, c.Path.ID)

	payload := struct {
		Options CustomItemFieldOptions `json:"customItemFieldOptions"`
	}{Options: c.FieldOptions}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode delete custom item field request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// CustomItemFieldDeleteResponse represents the response body for deleting a
// custom item field.
type CustomItemFieldDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the
// CustomItemFieldDeleteResponse.
func (c *CustomItemFieldDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusNoContent {
		return twapi.NewHTTPError(resp, "failed to delete custom item field")
	}
	return nil
}

// CustomItemFieldDelete deletes a custom item field using the provided
// request.
func CustomItemFieldDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomItemFieldDeleteRequest,
) (*CustomItemFieldDeleteResponse, error) {
	return twapi.Execute[CustomItemFieldDeleteRequest, *CustomItemFieldDeleteResponse](ctx, engine, req)
}

// CustomItemFieldGetRequestPath contains the path parameters for loading a
// single custom item field.
type CustomItemFieldGetRequestPath struct {
	// CustomItemID is the unique identifier of the custom item type the
	// field belongs to.
	CustomItemID int64

	// ID is the unique identifier of the field to be retrieved.
	ID int64
}

// CustomItemFieldGetRequest represents the request body for loading a single
// custom item field.
type CustomItemFieldGetRequest struct {
	// Path contains the path parameters for the request.
	Path CustomItemFieldGetRequestPath
}

// NewCustomItemFieldGetRequest creates a new CustomItemFieldGetRequest with
// the required path parameters populated.
func NewCustomItemFieldGetRequest(customItemID, fieldID int64) CustomItemFieldGetRequest {
	return CustomItemFieldGetRequest{
		Path: CustomItemFieldGetRequestPath{
			CustomItemID: customItemID,
			ID:           fieldID,
		},
	}
}

// HTTPRequest creates an HTTP request for the CustomItemFieldGetRequest.
func (c CustomItemFieldGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if c.Path.CustomItemID == 0 {
		return nil, fmt.Errorf("a custom item ID is required to retrieve a custom item field")
	}
	if c.Path.ID == 0 {
		return nil, fmt.Errorf("a field ID is required to retrieve a custom item field")
	}
	uri := fmt.Sprintf("%s/projects/api/v3/customitems/%d/fields/%d.json",
		server, c.Path.CustomItemID, c.Path.ID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// CustomItemFieldGetResponse contains the information related to a single
// custom item field.
type CustomItemFieldGetResponse struct {
	// CustomItemField is the retrieved field.
	CustomItemField CustomItemField `json:"customItemField"`
}

// HandleHTTPResponse handles the HTTP response for the
// CustomItemFieldGetResponse.
func (c *CustomItemFieldGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve custom item field")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode retrieve custom item field response: %w", err)
	}
	return nil
}

// CustomItemFieldGet retrieves a single custom item field using the provided
// request.
func CustomItemFieldGet(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomItemFieldGetRequest,
) (*CustomItemFieldGetResponse, error) {
	return twapi.Execute[CustomItemFieldGetRequest, *CustomItemFieldGetResponse](ctx, engine, req)
}

// CustomItemFieldListRequestPath contains the path parameters for listing
// the fields on a custom item type.
type CustomItemFieldListRequestPath struct {
	// CustomItemID is the unique identifier of the custom item type to list
	// fields for.
	CustomItemID int64
}

// CustomItemFieldListRequestFilters contains the filters for listing custom
// item fields.
type CustomItemFieldListRequestFilters struct {
	// SearchTerm filters fields by display name.
	SearchTerm string

	// IDs restricts the result to the given field IDs.
	IDs []int64

	// ShowDeleted includes deleted fields in the result.
	ShowDeleted *bool

	// OrderMode is the sort direction.
	OrderMode twapi.OrderMode

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of fields to retrieve per page. Defaults to 50.
	PageSize int64
}

func (c CustomItemFieldListRequestFilters) apply(req *http.Request) {
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
	if c.ShowDeleted != nil {
		query.Set("showDeleted", strconv.FormatBool(*c.ShowDeleted))
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
	query.Set("skipCounts", "true")
	req.URL.RawQuery = query.Encode()
}

// CustomItemFieldListRequest represents the request body for listing the
// fields on a custom item type.
type CustomItemFieldListRequest struct {
	// Path contains the path parameters for the request.
	Path CustomItemFieldListRequestPath

	// Filters contains the filters for listing custom item fields.
	Filters CustomItemFieldListRequestFilters
}

// NewCustomItemFieldListRequest creates a new CustomItemFieldListRequest
// scoped to the given custom item type with default pagination values.
func NewCustomItemFieldListRequest(customItemID int64) CustomItemFieldListRequest {
	return CustomItemFieldListRequest{
		Path: CustomItemFieldListRequestPath{CustomItemID: customItemID},
		Filters: CustomItemFieldListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the CustomItemFieldListRequest.
func (c CustomItemFieldListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if c.Path.CustomItemID == 0 {
		return nil, fmt.Errorf("a custom item ID is required to list custom item fields")
	}
	uri := fmt.Sprintf("%s/projects/api/v3/customitems/%d/fields.json", server, c.Path.CustomItemID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	c.Filters.apply(req)

	return req, nil
}

// CustomItemFieldListResponse contains information about multiple custom
// item fields on a custom item type.
type CustomItemFieldListResponse struct {
	request CustomItemFieldListRequest

	// Meta contains pagination information for the response.
	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`

	// CustomItemFields is the list of fields matching the request filters.
	CustomItemFields []CustomItemField `json:"customItemFields"`
}

// HandleHTTPResponse handles the HTTP response for the
// CustomItemFieldListResponse.
func (c *CustomItemFieldListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list custom item fields")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode list custom item fields response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response, enabling
// pagination via Iterate.
func (c *CustomItemFieldListResponse) SetRequest(req CustomItemFieldListRequest) {
	c.request = req
}

// Iterate returns the request set to the next page, if available. If there
// are no more pages, a nil request is returned.
func (c *CustomItemFieldListResponse) Iterate() *CustomItemFieldListRequest {
	if !c.Meta.Page.HasMore {
		return nil
	}
	req := c.request
	req.Filters.Page++
	return &req
}

// CustomItemFieldList retrieves the fields on a custom item type using the
// provided request.
func CustomItemFieldList(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomItemFieldListRequest,
) (*CustomItemFieldListResponse, error) {
	return twapi.Execute[CustomItemFieldListRequest, *CustomItemFieldListResponse](ctx, engine, req)
}
