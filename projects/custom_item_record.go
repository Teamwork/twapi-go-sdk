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
	_ twapi.HTTPRequester = (*CustomItemRecordCreateRequest)(nil)
	_ twapi.HTTPResponser = (*CustomItemRecordCreateResponse)(nil)
	_ twapi.HTTPRequester = (*CustomItemRecordUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*CustomItemRecordUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*CustomItemRecordDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*CustomItemRecordDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*CustomItemRecordBulkDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*CustomItemRecordBulkDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*CustomItemRecordGetRequest)(nil)
	_ twapi.HTTPResponser = (*CustomItemRecordGetResponse)(nil)
	_ twapi.HTTPRequester = (*CustomItemRecordListRequest)(nil)
	_ twapi.HTTPResponser = (*CustomItemRecordListResponse)(nil)
)

// CustomItemRecordState reports whether a record is active or has been
// deleted.
type CustomItemRecordState string

// Supported custom item record states.
const (
	CustomItemRecordStateActive   CustomItemRecordState = "active"
	CustomItemRecordStateArchived CustomItemRecordState = "archived"
	CustomItemRecordStateDeleted  CustomItemRecordState = "deleted"
)

// CustomItemRecordSideload identifies the related entities that can be
// requested alongside a custom item record via the API's include mechanism.
type CustomItemRecordSideload string

// Supported custom item record sideloads.
const (
	CustomItemRecordSideloadCreatedBy          CustomItemRecordSideload = "createdBy"
	CustomItemRecordSideloadUpdatedBy          CustomItemRecordSideload = "updatedBy"
	CustomItemRecordSideloadDeletedBy          CustomItemRecordSideload = "deletedBy"
	CustomItemRecordSideloadCustomItems        CustomItemRecordSideload = "customItems"
	CustomItemRecordSideloadCustomItemFields   CustomItemRecordSideload = "customItemFields"
	CustomItemRecordSideloadCustomItemSections CustomItemRecordSideload = "customItemSections"
	CustomItemRecordSideloadUsers              CustomItemRecordSideload = "users"
)

// CustomItemRecordOrderBy identifies the attributes a custom item record
// list can be ordered by.
type CustomItemRecordOrderBy string

// Supported custom item record order-by values.
const (
	CustomItemRecordOrderByDisplayOrder    CustomItemRecordOrderBy = "displayorder"
	CustomItemRecordOrderByName            CustomItemRecordOrderBy = "name"
	CustomItemRecordOrderByCustomItemField CustomItemRecordOrderBy = "customitemfield"
)

// CustomItemRecordFieldValues is the field values payload of a record, keyed
// by each field's TwID. The concrete type of the value depends on the
// field's type — strings for text, numbers for numeric, option TwIDs for
// dropdowns, slices of TwIDs for multiselect, ISO-8601 strings for dates,
// and so on.
type CustomItemRecordFieldValues map[string]any

// CustomItemRecord is a single row of a custom item type. Records carry a
// per-field value map keyed by field TwID (see CustomItemField.TwID).
type CustomItemRecord struct {
	// ID is the unique identifier of the record.
	ID int64 `json:"id"`

	// CustomItem is the relationship to the custom item type this record
	// belongs to.
	CustomItem *twapi.Relationship `json:"customItem"`

	// Section is the relationship to the section this record belongs to, if
	// any.
	Section *twapi.Relationship `json:"section"`

	// ParentID is the ID of the parent record, when records are nested.
	ParentID *int64 `json:"parentId"`

	// Name is the display name of the record.
	Name string `json:"name"`

	// State indicates whether the record is active or deleted.
	State CustomItemRecordState `json:"state"`

	// DisplayOrder is the relative position of the record within its
	// section (or the type when sectionless).
	DisplayOrder float64 `json:"displayOrder"`

	// FieldValues carries the values for each custom field on the record,
	// keyed by the field's TwID. Resolve a TwID to a human-readable field
	// name via CustomItemField.TwID.
	FieldValues CustomItemRecordFieldValues `json:"fieldValues"`

	// CreatedBy identifies the user that created the record.
	CreatedBy *twapi.Relationship `json:"createdBy"`

	// CreatedAt is the date and time when the record was created.
	CreatedAt *time.Time `json:"createdAt"`

	// UpdatedBy identifies the user that last updated the record.
	UpdatedBy *twapi.Relationship `json:"updatedBy"`

	// UpdatedAt is the date and time when the record was last updated.
	UpdatedAt *time.Time `json:"updatedAt"`

	// DeletedBy identifies the user that deleted the record, if applicable.
	DeletedBy *twapi.Relationship `json:"deletedBy"`

	// DeletedAt is the date and time when the record was deleted, if
	// applicable.
	DeletedAt *time.Time `json:"deletedAt"`
}

// CustomItemRecordOptions controls server-side side effects for a record
// write.
type CustomItemRecordOptions struct {
	// UseNotifyViaTWIM controls whether notifications are sent via Teamwork
	// Instant Messenger as part of this change.
	UseNotifyViaTWIM *bool `json:"useNotifyViaTWIM,omitempty"`

	// FireWebhook controls whether webhooks are fired as part of this change.
	FireWebhook *bool `json:"fireWebhook,omitempty"`
}

// CustomItemRecordCreateRequestPath contains the path parameters for
// creating a custom item record.
type CustomItemRecordCreateRequestPath struct {
	// CustomItemID is the unique identifier of the custom item type that
	// will own the new record.
	CustomItemID int64
}

// CustomItemRecordCreateRequest represents the request body for creating a
// new record on a custom item type.
type CustomItemRecordCreateRequest struct {
	// Path contains the path parameters for the request.
	Path CustomItemRecordCreateRequestPath `json:"-"`

	// Name is the display name of the record. This field is required.
	Name string `json:"name"`

	// SectionID places the record in the given section. Use twapi.NullInt64()
	// to place it outside any section.
	SectionID twapi.NullableInt64 `json:"sectionId,omitzero"`

	// PositionAfterID places the new record after the record with the given
	// ID. Nil appends to the end of the section.
	PositionAfterID *int64 `json:"positionAfterId,omitempty"`

	// FieldValues are the custom-field values for the record, keyed by
	// field TwID.
	FieldValues CustomItemRecordFieldValues `json:"fieldValues,omitempty"`

	// RecordOptions controls notification and webhook side effects.
	RecordOptions CustomItemRecordOptions `json:"-"`
}

// NewCustomItemRecordCreateRequest creates a new
// CustomItemRecordCreateRequest with the required custom item ID and record
// name populated.
func NewCustomItemRecordCreateRequest(customItemID int64, name string) CustomItemRecordCreateRequest {
	return CustomItemRecordCreateRequest{
		Path: CustomItemRecordCreateRequestPath{CustomItemID: customItemID},
		Name: name,
	}
}

// HTTPRequest creates an HTTP request for the CustomItemRecordCreateRequest.
func (c CustomItemRecordCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if c.Path.CustomItemID == 0 {
		return nil, fmt.Errorf("a custom item ID is required to create a custom item record")
	}
	uri := fmt.Sprintf("%s/projects/api/v3/customitems/%d/records.json", server, c.Path.CustomItemID)

	payload := struct {
		Record  CustomItemRecordCreateRequest `json:"customItemRecord"`
		Options CustomItemRecordOptions       `json:"customItemRecordOptions"`
	}{Record: c, Options: c.RecordOptions}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create custom item record request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// CustomItemRecordCreateResponse represents the response body for creating
// a new record.
type CustomItemRecordCreateResponse struct {
	// CustomItemRecord is the created record.
	CustomItemRecord CustomItemRecord `json:"customItemRecord"`
}

// HandleHTTPResponse handles the HTTP response for the
// CustomItemRecordCreateResponse.
func (c *CustomItemRecordCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create custom item record")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode create custom item record response: %w", err)
	}
	if c.CustomItemRecord.ID == 0 {
		return fmt.Errorf("create custom item record response does not contain a valid identifier")
	}
	return nil
}

// CustomItemRecordCreate creates a new custom item record using the
// provided request.
func CustomItemRecordCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomItemRecordCreateRequest,
) (*CustomItemRecordCreateResponse, error) {
	return twapi.Execute[CustomItemRecordCreateRequest, *CustomItemRecordCreateResponse](ctx, engine, req)
}

// CustomItemRecordUpdateRequestPath contains the path parameters for
// updating a custom item record.
type CustomItemRecordUpdateRequestPath struct {
	// CustomItemID is the unique identifier of the custom item type the
	// record belongs to.
	CustomItemID int64

	// ID is the unique identifier of the record to be updated.
	ID int64
}

// CustomItemRecordUpdateRequest represents the request body for updating a
// custom item record. All non-path fields are optional.
type CustomItemRecordUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path CustomItemRecordUpdateRequestPath `json:"-"`

	// Name is the new display name of the record.
	Name *string `json:"name,omitempty"`

	// SectionID moves the record to the given section. Use twapi.NullInt64()
	// to remove the record from any section.
	SectionID twapi.NullableInt64 `json:"sectionId,omitzero"`

	// PositionAfterID moves the record after the record with the given ID.
	PositionAfterID *int64 `json:"positionAfterId,omitempty"`

	// FieldValues partially updates the custom-field values on the record,
	// keyed by field TwID. Fields not included are left unchanged.
	FieldValues CustomItemRecordFieldValues `json:"fieldValues,omitempty"`

	// RecordOptions controls notification and webhook side effects.
	RecordOptions CustomItemRecordOptions `json:"-"`
}

// NewCustomItemRecordUpdateRequest creates a new
// CustomItemRecordUpdateRequest with the required path parameters populated.
func NewCustomItemRecordUpdateRequest(customItemID, recordID int64) CustomItemRecordUpdateRequest {
	return CustomItemRecordUpdateRequest{
		Path: CustomItemRecordUpdateRequestPath{
			CustomItemID: customItemID,
			ID:           recordID,
		},
	}
}

// HTTPRequest creates an HTTP request for the CustomItemRecordUpdateRequest.
func (c CustomItemRecordUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if c.Path.CustomItemID == 0 {
		return nil, fmt.Errorf("a custom item ID is required to update a custom item record")
	}
	if c.Path.ID == 0 {
		return nil, fmt.Errorf("a record ID is required to update a custom item record")
	}
	uri := fmt.Sprintf("%s/projects/api/v3/customitems/%d/records/%d.json",
		server, c.Path.CustomItemID, c.Path.ID)

	payload := struct {
		Record  CustomItemRecordUpdateRequest `json:"customItemRecord"`
		Options CustomItemRecordOptions       `json:"customItemRecordOptions"`
	}{Record: c, Options: c.RecordOptions}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update custom item record request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// CustomItemRecordUpdateResponse represents the response body for updating
// a custom item record.
type CustomItemRecordUpdateResponse struct {
	// CustomItemRecord is the updated record.
	CustomItemRecord CustomItemRecord `json:"customItemRecord"`
}

// HandleHTTPResponse handles the HTTP response for the
// CustomItemRecordUpdateResponse. The API returns 201 Created on a
// successful update.
func (c *CustomItemRecordUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update custom item record")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode update custom item record response: %w", err)
	}
	return nil
}

// CustomItemRecordUpdate updates a custom item record using the provided
// request.
func CustomItemRecordUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomItemRecordUpdateRequest,
) (*CustomItemRecordUpdateResponse, error) {
	return twapi.Execute[CustomItemRecordUpdateRequest, *CustomItemRecordUpdateResponse](ctx, engine, req)
}

// CustomItemRecordDeleteRequestPath contains the path parameters for
// deleting a custom item record.
type CustomItemRecordDeleteRequestPath struct {
	// CustomItemID is the unique identifier of the custom item type the
	// record belongs to.
	CustomItemID int64

	// ID is the unique identifier of the record to be deleted.
	ID int64
}

// CustomItemRecordDeleteRequest represents the request body for deleting a
// custom item record.
type CustomItemRecordDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path CustomItemRecordDeleteRequestPath

	// RecordOptions controls notification and webhook side effects.
	RecordOptions CustomItemRecordOptions `json:"-"`
}

// NewCustomItemRecordDeleteRequest creates a new
// CustomItemRecordDeleteRequest with the required path parameters populated.
func NewCustomItemRecordDeleteRequest(customItemID, recordID int64) CustomItemRecordDeleteRequest {
	return CustomItemRecordDeleteRequest{
		Path: CustomItemRecordDeleteRequestPath{
			CustomItemID: customItemID,
			ID:           recordID,
		},
	}
}

// HTTPRequest creates an HTTP request for the CustomItemRecordDeleteRequest.
func (c CustomItemRecordDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if c.Path.CustomItemID == 0 {
		return nil, fmt.Errorf("a custom item ID is required to delete a custom item record")
	}
	if c.Path.ID == 0 {
		return nil, fmt.Errorf("a record ID is required to delete a custom item record")
	}
	uri := fmt.Sprintf("%s/projects/api/v3/customitems/%d/records/%d.json",
		server, c.Path.CustomItemID, c.Path.ID)

	payload := struct {
		Options CustomItemRecordOptions `json:"customItemRecordOptions"`
	}{Options: c.RecordOptions}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode delete custom item record request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// CustomItemRecordDeleteResponse represents the response body for deleting
// a custom item record.
type CustomItemRecordDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the
// CustomItemRecordDeleteResponse.
func (c *CustomItemRecordDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusNoContent {
		return twapi.NewHTTPError(resp, "failed to delete custom item record")
	}
	return nil
}

// CustomItemRecordDelete deletes a custom item record using the provided
// request.
func CustomItemRecordDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomItemRecordDeleteRequest,
) (*CustomItemRecordDeleteResponse, error) {
	return twapi.Execute[CustomItemRecordDeleteRequest, *CustomItemRecordDeleteResponse](ctx, engine, req)
}

// CustomItemRecordBulkDeleteRequestPath contains the path parameters for
// bulk-deleting custom item records.
type CustomItemRecordBulkDeleteRequestPath struct {
	// CustomItemID is the unique identifier of the custom item type the
	// records belong to.
	CustomItemID int64
}

// CustomItemRecordBulkDeleteRequest represents the request body for
// deleting many records of a custom item type in a single call.
type CustomItemRecordBulkDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path CustomItemRecordBulkDeleteRequestPath `json:"-"`

	// IDs is the list of record IDs to delete.
	IDs []int64 `json:"customItemRecordIds"`

	// RecordOptions controls notification and webhook side effects.
	RecordOptions CustomItemRecordOptions `json:"-"`
}

// NewCustomItemRecordBulkDeleteRequest creates a new
// CustomItemRecordBulkDeleteRequest with the given custom item ID and
// record IDs populated.
func NewCustomItemRecordBulkDeleteRequest(customItemID int64, recordIDs []int64) CustomItemRecordBulkDeleteRequest {
	return CustomItemRecordBulkDeleteRequest{
		Path: CustomItemRecordBulkDeleteRequestPath{CustomItemID: customItemID},
		IDs:  recordIDs,
	}
}

// HTTPRequest creates an HTTP request for the
// CustomItemRecordBulkDeleteRequest.
func (c CustomItemRecordBulkDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if c.Path.CustomItemID == 0 {
		return nil, fmt.Errorf("a custom item ID is required to bulk delete custom item records")
	}
	if len(c.IDs) == 0 {
		return nil, fmt.Errorf("at least one record ID is required to bulk delete custom item records")
	}
	uri := fmt.Sprintf("%s/projects/api/v3/customitems/%d/records/bulk/delete.json", server, c.Path.CustomItemID)

	payload := struct {
		IDs     []int64                 `json:"customItemRecordIds"`
		Options CustomItemRecordOptions `json:"customItemRecordOptions"`
	}{IDs: c.IDs, Options: c.RecordOptions}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode bulk delete custom item records request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// CustomItemRecordBulkDeleteResponse represents the response body for
// bulk-deleting custom item records.
type CustomItemRecordBulkDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the
// CustomItemRecordBulkDeleteResponse.
func (c *CustomItemRecordBulkDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to bulk delete custom item records")
	}
	return nil
}

// CustomItemRecordBulkDelete deletes many custom item records in a single
// API call.
func CustomItemRecordBulkDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomItemRecordBulkDeleteRequest,
) (*CustomItemRecordBulkDeleteResponse, error) {
	return twapi.Execute[CustomItemRecordBulkDeleteRequest, *CustomItemRecordBulkDeleteResponse](ctx, engine, req)
}

// CustomItemRecordGetRequestPath contains the path parameters for loading a
// single custom item record.
type CustomItemRecordGetRequestPath struct {
	// CustomItemID is the unique identifier of the custom item type the
	// record belongs to.
	CustomItemID int64

	// ID is the unique identifier of the record to be retrieved.
	ID int64
}

// CustomItemRecordGetRequest represents the request body for loading a
// single custom item record.
type CustomItemRecordGetRequest struct {
	// Path contains the path parameters for the request.
	Path CustomItemRecordGetRequestPath

	// Include is an optional list of related entities to sideload. Use the
	// CustomItemRecordSideload constants.
	Include []CustomItemRecordSideload
}

// NewCustomItemRecordGetRequest creates a new CustomItemRecordGetRequest
// with the required path parameters populated.
func NewCustomItemRecordGetRequest(customItemID, recordID int64) CustomItemRecordGetRequest {
	return CustomItemRecordGetRequest{
		Path: CustomItemRecordGetRequestPath{
			CustomItemID: customItemID,
			ID:           recordID,
		},
	}
}

// HTTPRequest creates an HTTP request for the CustomItemRecordGetRequest.
func (c CustomItemRecordGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if c.Path.CustomItemID == 0 {
		return nil, fmt.Errorf("a custom item ID is required to retrieve a custom item record")
	}
	if c.Path.ID == 0 {
		return nil, fmt.Errorf("a record ID is required to retrieve a custom item record")
	}
	uri := fmt.Sprintf("%s/projects/api/v3/customitems/%d/records/%d.json",
		server, c.Path.CustomItemID, c.Path.ID)

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

// CustomItemRecordGetResponse contains the information related to a single
// custom item record.
type CustomItemRecordGetResponse struct {
	// CustomItemRecord is the retrieved record.
	CustomItemRecord CustomItemRecord `json:"customItemRecord"`
}

// HandleHTTPResponse handles the HTTP response for the
// CustomItemRecordGetResponse.
func (c *CustomItemRecordGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve custom item record")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode retrieve custom item record response: %w", err)
	}
	return nil
}

// CustomItemRecordGet retrieves a single custom item record using the
// provided request.
func CustomItemRecordGet(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomItemRecordGetRequest,
) (*CustomItemRecordGetResponse, error) {
	return twapi.Execute[CustomItemRecordGetRequest, *CustomItemRecordGetResponse](ctx, engine, req)
}

// CustomItemRecordListRequestPath contains the path parameters for listing
// the records of a custom item type.
type CustomItemRecordListRequestPath struct {
	// CustomItemID is the unique identifier of the custom item type to list
	// records for.
	CustomItemID int64
}

// CustomItemRecordListRequestFilters contains the filters for listing the
// records of a custom item type.
type CustomItemRecordListRequestFilters struct {
	// SearchTerm filters records by name.
	SearchTerm string

	// IDs restricts the result to the given record IDs.
	IDs []int64

	// SectionIDs restricts the result to records in the given sections.
	SectionIDs []int64

	// ShowDeleted includes deleted records in the result.
	ShowDeleted *bool

	// OrderBy sorts the result. Use the CustomItemRecordOrderBy constants.
	OrderBy CustomItemRecordOrderBy

	// OrderMode is the sort direction.
	OrderMode twapi.OrderMode

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of records to retrieve per page. Defaults to
	// 50.
	PageSize int64
}

func (c CustomItemRecordListRequestFilters) apply(req *http.Request) {
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
	if len(c.SectionIDs) > 0 {
		ids := make([]string, len(c.SectionIDs))
		for i, id := range c.SectionIDs {
			ids[i] = strconv.FormatInt(id, 10)
		}
		query.Set("sectionIds", strings.Join(ids, ","))
	}
	if c.ShowDeleted != nil {
		query.Set("showDeleted", strconv.FormatBool(*c.ShowDeleted))
	}
	if c.OrderBy != "" {
		query.Set("orderBy", string(c.OrderBy))
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

// CustomItemRecordListRequest represents the request body for listing
// records of a custom item type.
type CustomItemRecordListRequest struct {
	// Path contains the path parameters for the request.
	Path CustomItemRecordListRequestPath

	// Filters contains the filters for listing custom item records.
	Filters CustomItemRecordListRequestFilters
}

// NewCustomItemRecordListRequest creates a new CustomItemRecordListRequest
// scoped to the given custom item type with default pagination values.
func NewCustomItemRecordListRequest(customItemID int64) CustomItemRecordListRequest {
	return CustomItemRecordListRequest{
		Path: CustomItemRecordListRequestPath{CustomItemID: customItemID},
		Filters: CustomItemRecordListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the CustomItemRecordListRequest.
func (c CustomItemRecordListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if c.Path.CustomItemID == 0 {
		return nil, fmt.Errorf("a custom item ID is required to list custom item records")
	}
	uri := fmt.Sprintf("%s/projects/api/v3/customitems/%d/records.json", server, c.Path.CustomItemID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	c.Filters.apply(req)

	return req, nil
}

// CustomItemRecordListResponse contains information about multiple records
// of a custom item type.
type CustomItemRecordListResponse struct {
	request CustomItemRecordListRequest

	// Meta contains pagination information for the response.
	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`

	// CustomItemRecords is the list of records matching the request
	// filters.
	CustomItemRecords []CustomItemRecord `json:"customItemRecords"`
}

// HandleHTTPResponse handles the HTTP response for the
// CustomItemRecordListResponse.
func (c *CustomItemRecordListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list custom item records")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode list custom item records response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response, enabling
// pagination via Iterate.
func (c *CustomItemRecordListResponse) SetRequest(req CustomItemRecordListRequest) {
	c.request = req
}

// Iterate returns the request set to the next page, if available. If there
// are no more pages, a nil request is returned.
func (c *CustomItemRecordListResponse) Iterate() *CustomItemRecordListRequest {
	if !c.Meta.Page.HasMore {
		return nil
	}
	req := c.request
	req.Filters.Page++
	return &req
}

// CustomItemRecordList retrieves the records of a custom item type using
// the provided request.
func CustomItemRecordList(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomItemRecordListRequest,
) (*CustomItemRecordListResponse, error) {
	return twapi.Execute[CustomItemRecordListRequest, *CustomItemRecordListResponse](ctx, engine, req)
}
