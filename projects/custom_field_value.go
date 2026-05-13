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
	_ twapi.HTTPRequester = (*CustomFieldValueCreateRequest)(nil)
	_ twapi.HTTPResponser = (*CustomFieldValueCreateResponse)(nil)
	_ twapi.HTTPRequester = (*CustomFieldValueUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*CustomFieldValueUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*CustomFieldValueDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*CustomFieldValueDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*CustomFieldValueGetRequest)(nil)
	_ twapi.HTTPResponser = (*CustomFieldValueGetResponse)(nil)
	_ twapi.HTTPRequester = (*CustomFieldValueListRequest)(nil)
	_ twapi.HTTPResponser = (*CustomFieldValueListResponse)(nil)
)

// CustomFieldValue is the value of a custom field set on a specific entity such
// as a project, task or company. Exactly one of Task, Project or Company will
// be populated, matching the entity the value belongs to.
//
// More information can be found at:
// https://support.teamwork.com/projects/custom-fields/create-and-manage-custom-fields
// https://support.teamwork.com/projects/custom-fields/use-custom-fields
type CustomFieldValue struct {
	// ID is the unique identifier of the custom field value entry.
	ID int64 `json:"id"`

	// Value is the value of the custom field. The concrete type depends on the
	// custom field definition: strings for text fields, numbers for number
	// fields, booleans for checkboxes, option IDs for dropdown fields, ISO-8601
	// dates for date fields, etc.
	Value any `json:"value"`

	// CurrencyCode is the currency code associated with the value when the
	// custom field is of currency type.
	CurrencyCode string `json:"currencySymbol"`

	// CountryCode is the country code associated with the value when the custom
	// field is of currency type.
	CountryCode string `json:"countryCode"`

	// CustomField is the relationship to the custom field this value is
	// associated with.
	CustomField *twapi.Relationship `json:"customfield"`

	// Task is set when the value belongs to a task.
	Task *twapi.Relationship `json:"task,omitempty"`

	// Project is set when the value belongs to a project.
	Project *twapi.Relationship `json:"project,omitempty"`

	// Company is set when the value belongs to a company.
	Company *twapi.Relationship `json:"company,omitempty"`

	// CreatedBy is the unique identifier of the user that created the value.
	CreatedBy *int64 `json:"createdBy"`

	// CreatedAt is the date and time when the value was created.
	CreatedAt *time.Time `json:"createdAt"`

	// UpdatedBy is the unique identifier of the user that last updated the
	// value.
	UpdatedBy *int64 `json:"updatedBy"`

	// UpdatedAt is the date and time when the value was last updated.
	UpdatedAt *time.Time `json:"updatedAt"`
}

// CustomFieldValueCreateRequestPath contains the path parameters for creating a
// custom field value. The Owner identifies which entity (task, project or
// company) the value will be attached to.
type CustomFieldValueCreateRequestPath struct {
	// Owner is the entity (task, project or company) to attach the value to.
	// It is populated by the NewTask…/NewProject…/NewCompany… request
	// constructors.
	Owner CustomFieldValueOwner
}

// CustomFieldValueCreateRequest represents the request body for setting a
// custom field value on a project, task or company.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/custom-fields/post-projects-api-v3-tasks-task-id-customfields-json
// https://apidocs.teamwork.com/docs/teamwork/v3/custom-fields/post-projects-api-v3-projects-project-id-customfields-json
// https://apidocs.teamwork.com/docs/teamwork/v3/custom-fields/post-projects-api-v3-companies-company-id-customfields-json
//
//nolint:lll
type CustomFieldValueCreateRequest struct {
	// Path contains the path parameters for the request.
	Path CustomFieldValueCreateRequestPath `json:"-"`

	// CustomFieldID is the unique identifier of the custom field this value
	// belongs to. This field is required.
	CustomFieldID int64 `json:"customfieldId"`

	// Value is the value to assign to the custom field. The concrete type
	// depends on the custom field definition: strings for text fields, numbers
	// for number fields, booleans for checkboxes, option IDs for dropdown
	// fields, ISO-8601 dates for date fields, etc.
	Value any `json:"value"`

	// CurrencyCode is the ISO currency code for currency-type custom field
	// values.
	CurrencyCode *string `json:"currencyCode,omitempty"`

	// CountryCode is the country code for currency-type custom field values.
	CountryCode *string `json:"countryCode,omitempty"`
}

// NewTaskCustomFieldValueCreateRequest creates a new
// CustomFieldValueCreateRequest for a task.
func NewTaskCustomFieldValueCreateRequest(taskID, customFieldID int64, value any) CustomFieldValueCreateRequest {
	return CustomFieldValueCreateRequest{
		Path:          CustomFieldValueCreateRequestPath{Owner: taskCustomFieldValueOwner{taskID: taskID}},
		CustomFieldID: customFieldID,
		Value:         value,
	}
}

// NewProjectCustomFieldValueCreateRequest creates a new
// CustomFieldValueCreateRequest for a project.
func NewProjectCustomFieldValueCreateRequest(projectID, customFieldID int64, value any) CustomFieldValueCreateRequest {
	return CustomFieldValueCreateRequest{
		Path:          CustomFieldValueCreateRequestPath{Owner: projectCustomFieldValueOwner{projectID: projectID}},
		CustomFieldID: customFieldID,
		Value:         value,
	}
}

// NewCompanyCustomFieldValueCreateRequest creates a new
// CustomFieldValueCreateRequest for a company.
func NewCompanyCustomFieldValueCreateRequest(companyID, customFieldID int64, value any) CustomFieldValueCreateRequest {
	return CustomFieldValueCreateRequest{
		Path:          CustomFieldValueCreateRequestPath{Owner: companyCustomFieldValueOwner{companyID: companyID}},
		CustomFieldID: customFieldID,
		Value:         value,
	}
}

// HTTPRequest creates an HTTP request for the CustomFieldValueCreateRequest.
func (c CustomFieldValueCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if c.Path.Owner == nil {
		return nil, fmt.Errorf("a task, project or company owner is required for a custom field value")
	}
	uri := fmt.Sprintf("%s/projects/api/v3/%s/%d/customfields.json",
		server, c.Path.Owner.label(), c.Path.Owner.id())

	payload := map[string]CustomFieldValueCreateRequest{
		c.Path.Owner.responseField(): c,
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create custom field value request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// CustomFieldValueCreateResponse represents the response body for setting a
// custom field value on a project, task or company.
type CustomFieldValueCreateResponse struct {
	// CustomFieldValue is the created custom field value.
	CustomFieldValue CustomFieldValue
}

// UnmarshalJSON decodes the response into the CustomFieldValue field
// regardless of the entity-specific wrapper key returned by the API.
func (c *CustomFieldValueCreateResponse) UnmarshalJSON(data []byte) error {
	return decodeCustomFieldValue(data, &c.CustomFieldValue)
}

// HandleHTTPResponse handles the HTTP response for the
// CustomFieldValueCreateResponse. If some unexpected HTTP status code is
// returned by the API, a twapi.HTTPError is returned.
func (c *CustomFieldValueCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create custom field value")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode create custom field value response: %w", err)
	}
	if c.CustomFieldValue.ID == 0 {
		return fmt.Errorf("create custom field value response does not contain a valid identifier")
	}
	return nil
}

// CustomFieldValueCreate sets a custom field value on a project, task or
// company.
func CustomFieldValueCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomFieldValueCreateRequest,
) (*CustomFieldValueCreateResponse, error) {
	return twapi.Execute[CustomFieldValueCreateRequest, *CustomFieldValueCreateResponse](ctx, engine, req)
}

// CustomFieldValueUpdateRequestPath contains the path parameters for updating a
// custom field value. The Owner identifies which entity (task, project or
// company) the value belongs to.
type CustomFieldValueUpdateRequestPath struct {
	// Owner is the entity (task, project or company) the value belongs to.
	// It is populated by the NewTask…/NewProject…/NewCompany… request
	// constructors.
	Owner CustomFieldValueOwner

	// ValueID is the unique identifier of the custom field whose value is
	// being updated.
	ValueID int64
}

// CustomFieldValueUpdateRequest represents the request body for updating a
// custom field value on a project, task or company.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/custom-fields/patch-projects-api-v3-tasks-task-id-customfields-custom-field-id-json
// https://apidocs.teamwork.com/docs/teamwork/v3/custom-fields/patch-projects-api-v3-projects-project-id-customfields-custom-field-id-json
// https://apidocs.teamwork.com/docs/teamwork/v3/custom-fields/patch-projects-api-v3-companies-company-id-customfields-custom-field-id-json
//
//nolint:lll
type CustomFieldValueUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path CustomFieldValueUpdateRequestPath `json:"-"`

	// CustomFieldID is the unique identifier of the custom field this value
	// belongs to. This field is required.
	CustomFieldID int64 `json:"customfieldId"`

	// Value is the new value to assign to the custom field. The concrete type
	// depends on the custom field definition.
	Value any `json:"value,omitempty"`

	// CurrencyCode is the ISO currency code for currency-type custom field
	// values.
	CurrencyCode *string `json:"currencyCode,omitempty"`

	// CountryCode is the country code for currency-type custom field values.
	CountryCode *string `json:"countryCode,omitempty"`
}

// NewTaskCustomFieldValueUpdateRequest creates a new
// CustomFieldValueUpdateRequest for a task.
func NewTaskCustomFieldValueUpdateRequest(
	taskID, customFieldID, valueID int64,
	value any,
) CustomFieldValueUpdateRequest {
	return CustomFieldValueUpdateRequest{
		Path: CustomFieldValueUpdateRequestPath{
			Owner:   taskCustomFieldValueOwner{taskID: taskID},
			ValueID: valueID,
		},
		CustomFieldID: customFieldID,
		Value:         value,
	}
}

// NewProjectCustomFieldValueUpdateRequest creates a new
// CustomFieldValueUpdateRequest for a project.
func NewProjectCustomFieldValueUpdateRequest(
	projectID, customFieldID, valueID int64,
	value any,
) CustomFieldValueUpdateRequest {
	return CustomFieldValueUpdateRequest{
		Path: CustomFieldValueUpdateRequestPath{
			Owner:   projectCustomFieldValueOwner{projectID: projectID},
			ValueID: valueID,
		},
		CustomFieldID: customFieldID,
		Value:         value,
	}
}

// NewCompanyCustomFieldValueUpdateRequest creates a new
// CustomFieldValueUpdateRequest for a company.
func NewCompanyCustomFieldValueUpdateRequest(
	companyID, customFieldID, valueID int64,
	value any,
) CustomFieldValueUpdateRequest {
	return CustomFieldValueUpdateRequest{
		Path: CustomFieldValueUpdateRequestPath{
			Owner:   companyCustomFieldValueOwner{companyID: companyID},
			ValueID: valueID,
		},
		CustomFieldID: customFieldID,
		Value:         value,
	}
}

// HTTPRequest creates an HTTP request for the CustomFieldValueUpdateRequest.
func (c CustomFieldValueUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if c.Path.Owner == nil {
		return nil, fmt.Errorf("a task, project or company owner is required for a custom field value")
	}
	if c.Path.ValueID == 0 {
		return nil, fmt.Errorf("custom field value ID is required to update a custom field value")
	}
	uri := fmt.Sprintf("%s/projects/api/v3/%s/%d/customfields/%d.json",
		server, c.Path.Owner.label(), c.Path.Owner.id(), c.Path.ValueID)

	payload := map[string]CustomFieldValueUpdateRequest{
		c.Path.Owner.responseField(): c,
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update custom field value request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// CustomFieldValueUpdateResponse represents the response body for updating a
// custom field value.
type CustomFieldValueUpdateResponse struct {
	// CustomFieldValue is the updated custom field value.
	CustomFieldValue CustomFieldValue
}

// UnmarshalJSON decodes the response into the CustomFieldValue field regardless
// of the entity-specific wrapper key returned by the API.
func (c *CustomFieldValueUpdateResponse) UnmarshalJSON(data []byte) error {
	return decodeCustomFieldValue(data, &c.CustomFieldValue)
}

// HandleHTTPResponse handles the HTTP response for the
// CustomFieldValueUpdateResponse. If some unexpected HTTP status code is
// returned by the API, a twapi.HTTPError is returned.
func (c *CustomFieldValueUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update custom field value")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode update custom field value response: %w", err)
	}
	return nil
}

// CustomFieldValueUpdate updates a custom field value on a project, task or
// company.
func CustomFieldValueUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomFieldValueUpdateRequest,
) (*CustomFieldValueUpdateResponse, error) {
	return twapi.Execute[CustomFieldValueUpdateRequest, *CustomFieldValueUpdateResponse](ctx, engine, req)
}

// CustomFieldValueDeleteRequestPath contains the path parameters for deleting a
// custom field value. The Owner identifies which entity (task, project or
// company) the value belongs to.
type CustomFieldValueDeleteRequestPath struct {
	// Owner is the entity (task, project or company) the value belongs to.
	// It is populated by the NewTask…/NewProject…/NewCompany… request
	// constructors.
	Owner CustomFieldValueOwner

	// ValueID is the unique identifier of the custom field value.
	ValueID int64
}

// CustomFieldValueDeleteRequest represents the request body for clearing a
// custom field value from a project, task or company.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/custom-fields/delete-projects-api-v3-tasks-task-id-customfields-custom-field-id-json
// https://apidocs.teamwork.com/docs/teamwork/v3/custom-fields/delete-projects-api-v3-projects-project-id-customfields-custom-field-id-json
// https://apidocs.teamwork.com/docs/teamwork/v3/custom-fields/delete-projects-api-v3-companies-company-id-customfields-custom-field-id-json
//
//nolint:lll
type CustomFieldValueDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path CustomFieldValueDeleteRequestPath
}

// NewTaskCustomFieldValueDeleteRequest creates a new
// CustomFieldValueDeleteRequest for a task.
func NewTaskCustomFieldValueDeleteRequest(taskID, valueID int64) CustomFieldValueDeleteRequest {
	return CustomFieldValueDeleteRequest{
		Path: CustomFieldValueDeleteRequestPath{
			Owner:   taskCustomFieldValueOwner{taskID: taskID},
			ValueID: valueID,
		},
	}
}

// NewProjectCustomFieldValueDeleteRequest creates a new
// CustomFieldValueDeleteRequest for a project.
func NewProjectCustomFieldValueDeleteRequest(projectID, valueID int64) CustomFieldValueDeleteRequest {
	return CustomFieldValueDeleteRequest{
		Path: CustomFieldValueDeleteRequestPath{
			Owner:   projectCustomFieldValueOwner{projectID: projectID},
			ValueID: valueID,
		},
	}
}

// NewCompanyCustomFieldValueDeleteRequest creates a new
// CustomFieldValueDeleteRequest for a company.
func NewCompanyCustomFieldValueDeleteRequest(companyID, valueID int64) CustomFieldValueDeleteRequest {
	return CustomFieldValueDeleteRequest{
		Path: CustomFieldValueDeleteRequestPath{
			Owner:   companyCustomFieldValueOwner{companyID: companyID},
			ValueID: valueID,
		},
	}
}

// HTTPRequest creates an HTTP request for the CustomFieldValueDeleteRequest.
func (c CustomFieldValueDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if c.Path.Owner == nil {
		return nil, fmt.Errorf("a task, project or company owner is required for a custom field value")
	}
	if c.Path.ValueID == 0 {
		return nil, fmt.Errorf("custom field value ID is required to delete a custom field value")
	}
	uri := fmt.Sprintf("%s/projects/api/v3/%s/%d/customfields/%d.json",
		server, c.Path.Owner.label(), c.Path.Owner.id(), c.Path.ValueID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// CustomFieldValueDeleteResponse represents the response body for clearing a
// custom field value.
type CustomFieldValueDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the
// CustomFieldValueDeleteResponse. If some unexpected HTTP status code is
// returned by the API, a twapi.HTTPError is returned.
func (c *CustomFieldValueDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusNoContent {
		return twapi.NewHTTPError(resp, "failed to delete custom field value")
	}
	return nil
}

// CustomFieldValueDelete clears a custom field value from a project, task or
// company.
func CustomFieldValueDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomFieldValueDeleteRequest,
) (*CustomFieldValueDeleteResponse, error) {
	return twapi.Execute[CustomFieldValueDeleteRequest, *CustomFieldValueDeleteResponse](ctx, engine, req)
}

// CustomFieldValueGetRequestPath contains the path parameters for loading a
// single custom field value. The Owner identifies which entity (task, project
// or company) the value belongs to.
type CustomFieldValueGetRequestPath struct {
	// Owner is the entity (task, project or company) the value belongs to.
	// It is populated by the NewTask…/NewProject…/NewCompany… request
	// constructors.
	Owner CustomFieldValueOwner

	// ValueID is the unique identifier of the custom field value to retrieve.
	ValueID int64
}

// CustomFieldValueGetRequest represents the request body for loading a single
// custom field value from a project, task or company.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/custom-fields/get-projects-api-v3-tasks-task-id-customfields-custom-field-id-json
// https://apidocs.teamwork.com/docs/teamwork/v3/custom-fields/get-projects-api-v3-projects-project-id-customfields-custom-field-id-json
// https://apidocs.teamwork.com/docs/teamwork/v3/custom-fields/get-projects-api-v3-companies-company-id-customfields-custom-field-id-json
//
//nolint:lll
type CustomFieldValueGetRequest struct {
	// Path contains the path parameters for the request.
	Path CustomFieldValueGetRequestPath
}

// NewTaskCustomFieldValueGetRequest creates a new CustomFieldValueGetRequest
// for a task.
func NewTaskCustomFieldValueGetRequest(taskID, valueID int64) CustomFieldValueGetRequest {
	return CustomFieldValueGetRequest{
		Path: CustomFieldValueGetRequestPath{
			Owner:   taskCustomFieldValueOwner{taskID: taskID},
			ValueID: valueID,
		},
	}
}

// NewProjectCustomFieldValueGetRequest creates a new CustomFieldValueGetRequest
// for a project.
func NewProjectCustomFieldValueGetRequest(projectID, valueID int64) CustomFieldValueGetRequest {
	return CustomFieldValueGetRequest{
		Path: CustomFieldValueGetRequestPath{
			Owner:   projectCustomFieldValueOwner{projectID: projectID},
			ValueID: valueID,
		},
	}
}

// NewCompanyCustomFieldValueGetRequest creates a new CustomFieldValueGetRequest
// for a company.
func NewCompanyCustomFieldValueGetRequest(companyID, valueID int64) CustomFieldValueGetRequest {
	return CustomFieldValueGetRequest{
		Path: CustomFieldValueGetRequestPath{
			Owner:   companyCustomFieldValueOwner{companyID: companyID},
			ValueID: valueID,
		},
	}
}

// HTTPRequest creates an HTTP request for the CustomFieldValueGetRequest.
func (c CustomFieldValueGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if c.Path.Owner == nil {
		return nil, fmt.Errorf("a task, project or company owner is required for a custom field value")
	}
	if c.Path.ValueID == 0 {
		return nil, fmt.Errorf("custom field value ID is required to retrieve a custom field value")
	}
	uri := fmt.Sprintf("%s/projects/api/v3/%s/%d/customfields/%d.json",
		server, c.Path.Owner.label(), c.Path.Owner.id(), c.Path.ValueID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// CustomFieldValueGetResponse contains the information related to a single
// custom field value.
type CustomFieldValueGetResponse struct {
	// CustomFieldValue is the retrieved custom field value.
	CustomFieldValue CustomFieldValue `json:"customfieldValue"`
}

// UnmarshalJSON decodes the response into the CustomFieldValue field regardless
// of the entity-specific wrapper key returned by the API.
func (c *CustomFieldValueGetResponse) UnmarshalJSON(data []byte) error {
	return decodeCustomFieldValue(data, &c.CustomFieldValue)
}

// HandleHTTPResponse handles the HTTP response for the
// CustomFieldValueGetResponse. If some unexpected HTTP status code is returned
// by the API, a twapi.HTTPError is returned.
func (c *CustomFieldValueGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve custom field value")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode retrieve custom field value response: %w", err)
	}
	return nil
}

// CustomFieldValueGet retrieves a single custom field value from a project,
// task or company.
func CustomFieldValueGet(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomFieldValueGetRequest,
) (*CustomFieldValueGetResponse, error) {
	return twapi.Execute[CustomFieldValueGetRequest, *CustomFieldValueGetResponse](ctx, engine, req)
}

// CustomFieldValueListRequestPath contains the path parameters for listing the
// custom field values of an entity. The Owner identifies which entity (task,
// project or company) to list values for.
type CustomFieldValueListRequestPath struct {
	// Owner is the entity (task, project or company) to list values for. It
	// is populated by the NewTask…/NewProject…/NewCompany… request
	// constructors.
	Owner CustomFieldValueOwner
}

// CustomFieldValueListRequestFilters contains the filters for listing the
// custom field values of an entity.
type CustomFieldValueListRequestFilters struct {
	// CustomFieldIDs is an optional list of custom field IDs to filter values by.
	CustomFieldIDs []int64

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of custom field values to retrieve per page.
	// Defaults to 50.
	PageSize int64
}

func (c CustomFieldValueListRequestFilters) apply(req *http.Request) {
	query := req.URL.Query()
	if len(c.CustomFieldIDs) > 0 {
		ids := make([]string, len(c.CustomFieldIDs))
		for i, id := range c.CustomFieldIDs {
			ids[i] = strconv.FormatInt(id, 10)
		}
		query.Set("customFieldIds", strings.Join(ids, ","))
	}
	if c.Page > 0 {
		query.Set("page", strconv.FormatInt(c.Page, 10))
	}
	if c.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(c.PageSize, 10))
	}
	req.URL.RawQuery = query.Encode()
}

// CustomFieldValueListRequest represents the request body for listing the
// custom field values of a project, task or company.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/custom-fields/get-projects-api-v3-tasks-task-id-customfields-json
// https://apidocs.teamwork.com/docs/teamwork/v3/custom-fields/get-projects-api-v3-projects-project-id-customfields-json
// https://apidocs.teamwork.com/docs/teamwork/v3/custom-fields/get-projects-api-v3-companies-company-id-customfields-json
//
//nolint:lll
type CustomFieldValueListRequest struct {
	// Path contains the path parameters for the request.
	Path CustomFieldValueListRequestPath

	// Filters contains the filters for listing custom field values.
	Filters CustomFieldValueListRequestFilters
}

// NewTaskCustomFieldValueListRequest creates a new CustomFieldValueListRequest
// for a task.
func NewTaskCustomFieldValueListRequest(taskID int64) CustomFieldValueListRequest {
	return CustomFieldValueListRequest{
		Path: CustomFieldValueListRequestPath{Owner: taskCustomFieldValueOwner{taskID: taskID}},
		Filters: CustomFieldValueListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// NewProjectCustomFieldValueListRequest creates a new
// CustomFieldValueListRequest for a project.
func NewProjectCustomFieldValueListRequest(projectID int64) CustomFieldValueListRequest {
	return CustomFieldValueListRequest{
		Path: CustomFieldValueListRequestPath{Owner: projectCustomFieldValueOwner{projectID: projectID}},
		Filters: CustomFieldValueListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// NewCompanyCustomFieldValueListRequest creates a new
// CustomFieldValueListRequest for a company.
func NewCompanyCustomFieldValueListRequest(companyID int64) CustomFieldValueListRequest {
	return CustomFieldValueListRequest{
		Path: CustomFieldValueListRequestPath{Owner: companyCustomFieldValueOwner{companyID: companyID}},
		Filters: CustomFieldValueListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the CustomFieldValueListRequest.
func (c CustomFieldValueListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if c.Path.Owner == nil {
		return nil, fmt.Errorf("a task, project or company owner is required for a custom field value")
	}
	uri := fmt.Sprintf("%s/projects/api/v3/%s/%d/customfields.json",
		server, c.Path.Owner.label(), c.Path.Owner.id())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	c.Filters.apply(req)

	return req, nil
}

// CustomFieldValueListResponse contains a list of custom field values for a
// project, task or company.
type CustomFieldValueListResponse struct {
	request CustomFieldValueListRequest

	// Meta contains the pagination information for the response.
	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`

	// CustomFieldValues is the list of custom field values returned by the
	// request, regardless of the underlying entity type.
	CustomFieldValues []CustomFieldValue `json:"customfieldValues"`
}

// UnmarshalJSON decodes the response into the CustomFieldValues slice
// regardless of the entity-specific wrapper key returned by the API.
func (c *CustomFieldValueListResponse) UnmarshalJSON(data []byte) error {
	var envelope struct {
		Meta struct {
			Page struct {
				HasMore bool `json:"hasMore"`
			} `json:"page"`
		} `json:"meta"`
		CustomFieldTasks     []CustomFieldValue `json:"customfieldTasks"`
		CustomFieldProjects  []CustomFieldValue `json:"customfieldProjects"`
		CustomFieldCompanies []CustomFieldValue `json:"customfieldCompanies"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return err
	}
	c.Meta.Page.HasMore = envelope.Meta.Page.HasMore
	switch {
	case len(envelope.CustomFieldTasks) > 0:
		c.CustomFieldValues = envelope.CustomFieldTasks
	case len(envelope.CustomFieldProjects) > 0:
		c.CustomFieldValues = envelope.CustomFieldProjects
	case len(envelope.CustomFieldCompanies) > 0:
		c.CustomFieldValues = envelope.CustomFieldCompanies
	}
	return nil
}

// HandleHTTPResponse handles the HTTP response for the
// CustomFieldValueListResponse. If some unexpected HTTP status code is returned
// by the API, a twapi.HTTPError is returned.
func (c *CustomFieldValueListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list custom field values")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode list custom field values response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (c *CustomFieldValueListResponse) SetRequest(req CustomFieldValueListRequest) {
	c.request = req
}

// Iterate returns the request set to the next page, if available. If there are
// no more pages, a nil request is returned.
func (c *CustomFieldValueListResponse) Iterate() *CustomFieldValueListRequest {
	if !c.Meta.Page.HasMore {
		return nil
	}
	req := c.request
	req.Filters.Page++
	return &req
}

// CustomFieldValueList retrieves the custom field values for a project, task or
// company.
func CustomFieldValueList(
	ctx context.Context,
	engine *twapi.Engine,
	req CustomFieldValueListRequest,
) (*CustomFieldValueListResponse, error) {
	return twapi.Execute[CustomFieldValueListRequest, *CustomFieldValueListResponse](ctx, engine, req)
}

// CustomFieldValueOwner identifies the entity (task, project or company) a
// custom field value is associated with. The interface is sealed — only the
// owner types defined in this package satisfy it. Callers obtain an owner
// indirectly by using one of the NewTask…/NewProject…/NewCompany… request
// constructors.
type CustomFieldValueOwner interface {
	id() int64
	label() string
	responseField() string
}

type taskCustomFieldValueOwner struct{ taskID int64 }

func (o taskCustomFieldValueOwner) id() int64             { return o.taskID }
func (o taskCustomFieldValueOwner) label() string         { return "tasks" }
func (o taskCustomFieldValueOwner) responseField() string { return "customfieldTask" }

type projectCustomFieldValueOwner struct{ projectID int64 }

func (o projectCustomFieldValueOwner) id() int64             { return o.projectID }
func (o projectCustomFieldValueOwner) label() string         { return "projects" }
func (o projectCustomFieldValueOwner) responseField() string { return "customfieldProject" }

type companyCustomFieldValueOwner struct{ companyID int64 }

func (o companyCustomFieldValueOwner) id() int64             { return o.companyID }
func (o companyCustomFieldValueOwner) label() string         { return "companies" }
func (o companyCustomFieldValueOwner) responseField() string { return "customfieldCompany" }

// decodeCustomFieldValue decodes a single custom field value response into the
// provided CustomFieldValue, looking for the entity-specific wrapper key
// returned by the API.
func decodeCustomFieldValue(data []byte, dst *CustomFieldValue) error {
	var envelope struct {
		CustomFieldTask    *CustomFieldValue `json:"customfieldTask"`
		CustomFieldProject *CustomFieldValue `json:"customfieldProject"`
		CustomFieldCompany *CustomFieldValue `json:"customfieldCompany"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return err
	}
	switch {
	case envelope.CustomFieldTask != nil:
		*dst = *envelope.CustomFieldTask
	case envelope.CustomFieldProject != nil:
		*dst = *envelope.CustomFieldProject
	case envelope.CustomFieldCompany != nil:
		*dst = *envelope.CustomFieldCompany
	}
	return nil
}
