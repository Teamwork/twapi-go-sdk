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
	_ twapi.HTTPRequester = (*CompanyCreateRequest)(nil)
	_ twapi.HTTPResponser = (*CompanyCreateResponse)(nil)
	_ twapi.HTTPRequester = (*CompanyBulkCreateRequest)(nil)
	_ twapi.HTTPResponser = (*CompanyBulkCreateResponse)(nil)
	_ twapi.HTTPRequester = (*CompanyUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*CompanyUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*CompanyDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*CompanyDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*CompanyGetRequest)(nil)
	_ twapi.HTTPResponser = (*CompanyGetResponse)(nil)
	_ twapi.HTTPRequester = (*CompanyListRequest)(nil)
	_ twapi.HTTPResponser = (*CompanyListResponse)(nil)
)

// Company represents an organization or business entity that can be associated
// with users, projects, and tasks within the platform, and it is often referred
// to as a “client.” It serves as a way to group related users and projects
// under a single organizational umbrella, making it easier to manage
// permissions, assign responsibilities, and organize work. Companies (or
// clients) are frequently used to distinguish between internal teams and
// external collaborators, enabling teams to work efficiently while maintaining
// clear boundaries around ownership, visibility, and access levels across
// different projects.
//
// More information can be found at:
// https://support.teamwork.com/projects/getting-started/companies-owner-and-external
//
// sparsefields:gen
type Company struct {
	// ID is the unique identifier of the company.
	ID int64 `json:"id"`

	// AddressOne is the first line of the company's address.
	AddressOne string `json:"addressOne"`

	// AddressTwo is the second line of the company's address.
	AddressTwo string `json:"addressTwo"`

	// City is the city where the company is located.
	City string `json:"city"`

	// CountryCode is the ISO 3166-1 alpha-2 country code where the company is
	// located.
	CountryCode string `json:"countryCode"`

	// EmailOne is the primary email address of the company.
	EmailOne string `json:"emailOne"`

	// EmailTwo is the secondary email address of the company.
	EmailTwo string `json:"emailTwo"`

	// EmailThree is the tertiary email address of the company.
	EmailThree string `json:"emailThree"`

	// Fax is the fax number of the company.
	Fax string `json:"fax"`

	// Name is the name of the company.
	Name string `json:"name"`

	// Phone is the phone number of the company.
	Phone string `json:"phone"`

	// Profile is the profile text of the company.
	Profile *string `json:"profileText"`

	// State is the state or province where the company is located.
	State string `json:"state"`

	// Website is the website URL of the company.
	Website string `json:"website"`

	// Zip is the ZIP or postal code where the company is located.
	Zip string `json:"zip"`

	// ManagedBy is the user managing the company.
	ManagedBy *twapi.Relationship `json:"clientManagedBy"`

	// Industry is the industry the company belongs to.
	Industry *twapi.Relationship `json:"industry"`

	// Tags is a list of tags associated with the company.
	Tags []twapi.Relationship `json:"tags"`

	// CreatedAt is the date and time when the company was created.
	CreatedAt *time.Time `json:"createdAt"`

	// UpdatedAt is the date and time when the company was last updated.
	UpdatedAt *time.Time `json:"updatedAt"`

	// Status is the status of the company. Possible values are "active" or
	// "deleted".
	Status string `json:"status"`
}

// CompanyCreateRequest represents the request body for creating a new
// client/company.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/companies/post-projects-api-v3-companies-json
type CompanyCreateRequest struct {
	// AddressOne is the first line of the company's address.
	AddressOne *string `json:"addressOne,omitempty"`

	// AddressTwo is the second line of the company's address.
	AddressTwo *string `json:"addressTwo,omitempty"`

	// City is the city where the company is located.
	City *string `json:"city,omitempty"`

	// CountryCode is the ISO 3166-1 alpha-2 country code where the company is
	// located.
	CountryCode *string `json:"countrycode,omitempty"`

	// EmailOne is the primary email address of the company.
	EmailOne *string `json:"emailOne,omitempty"`

	// EmailTwo is the secondary email address of the company.
	EmailTwo *string `json:"emailTwo,omitempty"`

	// EmailThree is the tertiary email address of the company.
	EmailThree *string `json:"emailThree,omitempty"`

	// Fax is the fax number of the company.
	Fax *string `json:"fax,omitempty"`

	// Name is the name of the company. This field is required.
	Name string `json:"name"`

	// Phone is the phone number of the company.
	Phone *string `json:"phone,omitempty"`

	// Profile is the profile text of the company.
	Profile *string `json:"profile,omitempty"`

	// State is the state or province where the company is located.
	State *string `json:"state,omitempty"`

	// Website is the website URL of the company.
	Website *string `json:"website,omitempty"`

	// Zip is the ZIP or postal code where the company is located.
	Zip *string `json:"zip,omitempty"`

	// ManagerID is the user ID of the user managing the company.
	ManagerID *int64 `json:"clientManagedBy"`

	// IndustryID is the industry ID the company belongs to.
	IndustryID *int64 `json:"industryCatId,omitempty"`

	// TagIDs is a list of tag IDs to associate with the company.
	TagIDs []int64 `json:"tagIds,omitempty"`
}

// NewCompanyCreateRequest creates a new CompanyCreateRequest with the
// provided name in a specific project.
func NewCompanyCreateRequest(name string) CompanyCreateRequest {
	return CompanyCreateRequest{
		Name: name,
	}
}

// HTTPRequest creates an HTTP request for the CompanyCreateRequest.
func (c CompanyCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/companies.json"

	payload := struct {
		Company CompanyCreateRequest `json:"company"`
	}{Company: c}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create company request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// CompanyCreateResponse represents the response body for creating a new
// client/company.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/companies/post-projects-api-v3-companies-json
type CompanyCreateResponse struct {
	// Company is the created company.
	Company Company `json:"company"`
}

// HandleHTTPResponse handles the HTTP response for the CompanyCreateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (c *CompanyCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create company")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode create company response: %w", err)
	}
	if c.Company.ID == 0 {
		return fmt.Errorf("create company response does not contain a valid identifier")
	}
	return nil
}

// CompanyCreate creates a new client/company using the provided request and
// returns the response.
func CompanyCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req CompanyCreateRequest,
) (*CompanyCreateResponse, error) {
	return twapi.Execute[CompanyCreateRequest, *CompanyCreateResponse](ctx, engine, req)
}

// CompanyBulkCreateRequest represents the request body for creating many
// clients/companies in a single call, through
// POST /projects/api/v3/companies/bulk.json.
//
// Each company takes the same fields as a single create. A company whose name
// is already used by an existing company (compared case-insensitively, ignoring
// surrounding spaces) is not created again: the endpoint answers with the
// existing company in its place, and the same applies to a name repeated within
// one request.
type CompanyBulkCreateRequest struct {
	// Companies are the companies to create, in order. At least one is
	// required.
	Companies []CompanyCreateRequest `json:"companies"`

	// PartialSuccess selects how a company that cannot be created is handled.
	//
	// When false, the endpoint's default, the request is all-or-nothing: the
	// first company that fails aborts it, nothing is created, and the endpoint
	// answers with an error status (400, 403, 404, 409 or 422) whose errors
	// carry the failing company's position under meta.index.
	//
	// When true, every company that can be created is, and each one that cannot
	// is reported in CompanyBulkCreateResponse.Failed by its position in
	// Companies. If none can be created the endpoint answers 403 when every
	// company was refused for permissions, and 400 otherwise, each error again
	// carrying its position under meta.index.
	PartialSuccess bool `json:"-"`
}

// NewCompanyBulkCreateRequest creates a new CompanyBulkCreateRequest with the
// provided companies.
func NewCompanyBulkCreateRequest(companies ...CompanyCreateRequest) CompanyBulkCreateRequest {
	return CompanyBulkCreateRequest{
		Companies: companies,
	}
}

// HTTPRequest creates an HTTP request for the CompanyBulkCreateRequest.
func (c CompanyBulkCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if len(c.Companies) == 0 {
		return nil, fmt.Errorf("at least one company is required to bulk create companies")
	}
	uri := server + "/projects/api/v3/companies/bulk.json"

	type options struct {
		PartialSuccess bool `json:"partialSuccess"`
	}
	payload := struct {
		Companies []CompanyCreateRequest `json:"companies"`
		Options   *options               `json:"options,omitempty"`
	}{Companies: c.Companies}
	if c.PartialSuccess {
		payload.Options = &options{PartialSuccess: true}
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode bulk create companies request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// CompanyBulkCreateError describes one reason a company of a bulk create could
// not be created.
type CompanyBulkCreateError struct {
	// Title is a short summary of the problem, such as "invalid payload",
	// "forbidden action" or "conflict".
	Title string `json:"title"`

	// Detail explains this occurrence of the problem.
	Detail string `json:"detail"`

	// Code is an application-specific error code, when there is one.
	Code string `json:"code,omitempty"`

	// Meta carries extra context about the error, when there is any.
	Meta map[string]any `json:"meta,omitempty"`
}

// CompanyBulkCreateFailure reports a company of a bulk create that could not be
// created.
type CompanyBulkCreateFailure struct {
	// Index is the company's position in CompanyBulkCreateRequest.Companies,
	// starting at zero.
	Index int `json:"index"`

	// Errors are the reasons the company could not be created.
	Errors []CompanyBulkCreateError `json:"errors"`
}

// CompanyBulkCreateResponse represents the response body for creating many
// clients/companies in a single call.
type CompanyBulkCreateResponse struct {
	// Companies holds one entry for every company of the request that did not
	// fail, in request order. An entry may be a company that existed before
	// the call, when the request named it (see CompanyBulkCreateRequest). Use
	// CompanyIndexes to pair each entry with its position in the request.
	Companies []Company `json:"companies"`

	// Failed lists the companies that could not be created, by their position
	// in the request. It is only ever populated when the request set
	// PartialSuccess.
	Failed []CompanyBulkCreateFailure `json:"failed"`
}

// HandleHTTPResponse handles the HTTP response for the
// CompanyBulkCreateResponse. If some unexpected HTTP status code is returned by
// the API, a twapi.HTTPError is returned.
func (c *CompanyBulkCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to bulk create companies")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode bulk create companies response: %w", err)
	}
	return nil
}

// CompanyIndexes returns, for each entry of Companies, its position in the
// request's Companies. requested is the number of companies the request sent.
// The endpoint answers the companies in request order and leaves out the
// failed ones, so the positions are those not listed in Failed. An error is
// returned when the response does not account for every requested company.
func (c CompanyBulkCreateResponse) CompanyIndexes(requested int) ([]int, error) {
	failed := make(map[int]struct{}, len(c.Failed))
	for _, failure := range c.Failed {
		if failure.Index < 0 || failure.Index >= requested {
			return nil, fmt.Errorf("bulk create companies response reports a failure at position %d of %d",
				failure.Index, requested)
		}
		failed[failure.Index] = struct{}{}
	}
	if len(c.Companies)+len(failed) != requested {
		return nil, fmt.Errorf("bulk create companies response holds %d companies and %d failures for %d requested",
			len(c.Companies), len(failed), requested)
	}
	indexes := make([]int, 0, len(c.Companies))
	for i := range requested {
		if _, ok := failed[i]; !ok {
			indexes = append(indexes, i)
		}
	}
	return indexes, nil
}

// CompanyBulkCreate creates many clients/companies in a single call using the
// provided request and returns the response.
func CompanyBulkCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req CompanyBulkCreateRequest,
) (*CompanyBulkCreateResponse, error) {
	return twapi.Execute[CompanyBulkCreateRequest, *CompanyBulkCreateResponse](ctx, engine, req)
}

// CompanyUpdateRequestPath contains the path parameters for updating a
// client/company.
type CompanyUpdateRequestPath struct {
	// ID is the unique identifier of the company to be updated.
	ID int64
}

// CompanyUpdateRequest represents the request body for updating a
// client/company. Besides the identifier, all other fields are optional. When a
// field is not provided, it will not be modified.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/companies/patch-projects-api-v3-companies-company-id-json
type CompanyUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path CompanyUpdateRequestPath `json:"-"`

	// AddressOne is the first line of the company's address.
	AddressOne *string `json:"addressOne,omitempty"`

	// AddressTwo is the second line of the company's address.
	AddressTwo *string `json:"addressTwo,omitempty"`

	// City is the city where the company is located.
	City *string `json:"city,omitempty"`

	// CountryCode is the ISO 3166-1 alpha-2 country code where the company is
	// located.
	CountryCode *string `json:"countrycode,omitempty"`

	// EmailOne is the primary email address of the company.
	EmailOne *string `json:"emailOne,omitempty"`

	// EmailTwo is the secondary email address of the company.
	EmailTwo *string `json:"emailTwo,omitempty"`

	// EmailThree is the tertiary email address of the company.
	EmailThree *string `json:"emailThree,omitempty"`

	// Fax is the fax number of the company.
	Fax *string `json:"fax,omitempty"`

	// Name is the name of the company.
	Name *string `json:"name"`

	// Phone is the phone number of the company.
	Phone *string `json:"phone,omitempty"`

	// Profile is the profile text of the company.
	Profile *string `json:"profile,omitempty"`

	// State is the state or province where the company is located.
	State *string `json:"state,omitempty"`

	// Website is the website URL of the company.
	Website *string `json:"website,omitempty"`

	// Zip is the ZIP or postal code where the company is located.
	Zip *string `json:"zip,omitempty"`

	// ManagerID is the user ID of the user managing the company.
	ManagerID *int64 `json:"clientManagedBy"`

	// IndustryID is the industry ID the company belongs to.
	IndustryID *int64 `json:"industryCatId,omitempty"`

	// TagIDs is a list of tag IDs to associate with the company.
	TagIDs []int64 `json:"tagIds,omitempty"`
}

// NewCompanyUpdateRequest creates a new CompanyUpdateRequest with the provided
// client/company ID. The ID is required to update a company.
func NewCompanyUpdateRequest(companyID int64) CompanyUpdateRequest {
	return CompanyUpdateRequest{
		Path: CompanyUpdateRequestPath{
			ID: companyID,
		},
	}
}

// HTTPRequest creates an HTTP request for the CompanyUpdateRequest.
func (c CompanyUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/companies/" + strconv.FormatInt(c.Path.ID, 10) + ".json"

	payload := struct {
		Company CompanyUpdateRequest `json:"company"`
	}{Company: c}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update company request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// CompanyUpdateResponse represents the response body for updating a
// client/company.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/companies/patch-projects-api-v3-companies-company-id-json
type CompanyUpdateResponse struct {
	// Company is the updated company.
	Company Company `json:"company"`
}

// HandleHTTPResponse handles the HTTP response for the CompanyUpdateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (c *CompanyUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update company")
	}
	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode update company response: %w", err)
	}
	return nil
}

// CompanyUpdate updates a new client/company using the provided request and
// returns the response.
func CompanyUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req CompanyUpdateRequest,
) (*CompanyUpdateResponse, error) {
	return twapi.Execute[CompanyUpdateRequest, *CompanyUpdateResponse](ctx, engine, req)
}

// CompanyDeleteRequestPath contains the path parameters for deleting a
// client/company.
type CompanyDeleteRequestPath struct {
	// ID is the unique identifier of the company to be deleted.
	ID int64
}

// CompanyDeleteRequest represents the request body for deleting a
// client/company.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/companies/delete-projects-api-v3-companies-company-id-json
type CompanyDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path CompanyDeleteRequestPath
}

// NewCompanyDeleteRequest creates a new CompanyDeleteRequest with the
// provided company ID.
func NewCompanyDeleteRequest(companyID int64) CompanyDeleteRequest {
	return CompanyDeleteRequest{
		Path: CompanyDeleteRequestPath{
			ID: companyID,
		},
	}
}

// HTTPRequest creates an HTTP request for the CompanyDeleteRequest.
func (c CompanyDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/companies/" + strconv.FormatInt(c.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// CompanyDeleteResponse represents the response body for deleting a
// client/company.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/companies/delete-projects-api-v3-companies-company-id-json
type CompanyDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the CompanyDeleteResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (c *CompanyDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusNoContent {
		return twapi.NewHTTPError(resp, "failed to delete company")
	}
	return nil
}

// CompanyDelete deletes a client/company using the provided request and returns
// the response.
func CompanyDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req CompanyDeleteRequest,
) (*CompanyDeleteResponse, error) {
	return twapi.Execute[CompanyDeleteRequest, *CompanyDeleteResponse](ctx, engine, req)
}

// CompanyRequestSideload contains the possible sideload options when loading
// companies.
type CompanyRequestSideload string

// List of possible sideload options for CompanyRequestSideload.
const (
	CompanyRequestSideloadCustomFields      CompanyRequestSideload = "customfields"
	CompanyRequestSideloadCustomFieldValues CompanyRequestSideload = "customfieldcompanies"
)

// CompanyRequestFilters contains the filters for loading companies.
type CompanyRequestFilters struct {
	// Include specifies related resources to include.
	Include []CompanyRequestSideload
}

func (p CompanyRequestFilters) apply(req *http.Request) {
	query := req.URL.Query()
	if len(p.Include) > 0 {
		includes := make([]string, len(p.Include))
		for i, include := range p.Include {
			includes[i] = string(include)
		}
		query.Set("include", strings.Join(includes, ","))
	}
	req.URL.RawQuery = query.Encode()
}

// CompanyGetRequestPath contains the path parameters for loading a single company.
type CompanyGetRequestPath struct {
	// ID is the unique identifier of the company to be retrieved.
	ID int64 `json:"id"`
}

// CompanyGetRequest represents the request body for loading a single
// client/company.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/companies/get-projects-api-v3-companies-company-id-json
type CompanyGetRequest struct {
	// Path contains the path parameters for the request.
	Path CompanyGetRequestPath

	// Filters contains the filters for loading a single company.
	Filters CompanyRequestFilters

	// Fields restricts the attributes returned for the company and each of its
	// sideloads. Each slot of CompanyGetFields is a separate
	// `fields[entity]=…` selection; populated slots restrict the response,
	// empty slots return the API default. Use the generated CompanyField
	// constants to ensure values match real attributes.
	Fields CompanyGetFields
}

// NewCompanyGetRequest creates a new CompanyGetRequest with the provided
// company ID. The ID is required to load a company.
func NewCompanyGetRequest(companyID int64) CompanyGetRequest {
	return CompanyGetRequest{
		Path: CompanyGetRequestPath{
			ID: companyID,
		},
	}
}

// HTTPRequest creates an HTTP request for the CompanyGetRequest.
func (c CompanyGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/companies/" + strconv.FormatInt(c.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	c.Filters.apply(req)

	query := req.URL.Query()
	c.Fields.apply(query)
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// CompanyGetResponse contains all the information related to a client/company.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/companies/get-projects-api-v3-companies-company-id-json
//
// sparsefields:get
type CompanyGetResponse struct {
	Company Company `json:"company"`

	// Included contains related objects included in the response.
	Included struct {
		// CustomFields contains the custom fields associated with the company.
		//
		// The key is the string representation of the custom field ID.
		CustomFields map[string]CustomField `json:"customfields,omitempty"`
		// CustomFieldValues contains the values of the custom fields associated
		// with the company.
		//
		// The key is the string representation of the custom field value ID.
		CustomFieldValues map[string]CustomFieldValue `json:"customfieldCompanies,omitempty"`
	} `json:"included"`
}

// HandleHTTPResponse handles the HTTP response for the CompanyGetResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (c *CompanyGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve company")
	}

	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode retrieve company response: %w", err)
	}
	return nil
}

// CompanyGet retrieves a single client/company using the provided request and
// returns the response.
func CompanyGet(
	ctx context.Context,
	engine *twapi.Engine,
	req CompanyGetRequest,
) (*CompanyGetResponse, error) {
	return twapi.Execute[CompanyGetRequest, *CompanyGetResponse](ctx, engine, req)
}

// CompanyOrderBy identifies the attributes a client/company list can be ordered
// by.
type CompanyOrderBy string

// Supported client/company order-by values.
const (
	CompanyOrderByName            CompanyOrderBy = "name"
	CompanyOrderByAccounts        CompanyOrderBy = "accounts"
	CompanyOrderByClients         CompanyOrderBy = "clients"
	CompanyOrderByCollaborators   CompanyOrderBy = "collaborators"
	CompanyOrderByContacts        CompanyOrderBy = "contacts"
	CompanyOrderByProjects        CompanyOrderBy = "projects"
	CompanyOrderByTasks           CompanyOrderBy = "tasks"
	CompanyOrderByCountry         CompanyOrderBy = "country"
	CompanyOrderByHealth          CompanyOrderBy = "health"
	CompanyOrderByWebsite         CompanyOrderBy = "website"
	CompanyOrderByEmail           CompanyOrderBy = "email"
	CompanyOrderByPhone           CompanyOrderBy = "phone"
	CompanyOrderByFax             CompanyOrderBy = "fax"
	CompanyOrderByIndustry        CompanyOrderBy = "industry"
	CompanyOrderByDateAdded       CompanyOrderBy = "dateadded"
	CompanyOrderByOwnerCompany    CompanyOrderBy = "ownercompany"
	CompanyOrderByOwnerName       CompanyOrderBy = "ownername"
	CompanyOrderByTasksCompletion CompanyOrderBy = "taskscompletion"
	CompanyOrderByCustomField     CompanyOrderBy = "customfield"
	CompanyOrderByID              CompanyOrderBy = "id"
)

// CompanyListRequestFilters contains the filters for loading multiple
// clients/companies.
type CompanyListRequestFilters struct {
	CompanyRequestFilters

	// SearchTerm is an optional search term to filter clients/companies by name.
	SearchTerm string

	// OrderBy is the field to sort the results by. Use the CompanyOrderBy
	// constants. The endpoint defaults to ownercompany.
	OrderBy CompanyOrderBy

	// OrderMode is the direction to sort the results in. See twapi.OrderMode for
	// the supported values. The endpoint defaults to ascending.
	OrderMode twapi.OrderMode

	// OrderByCustomFieldID selects the custom field to sort by. It is only used
	// when OrderBy is CompanyOrderByCustomField.
	OrderByCustomFieldID int64

	// TagIDs is an optional list of tag IDs to filter companies by tags.
	TagIDs []int64

	// MatchAllTags is an optional flag to indicate if all tags must match. If set
	// to true, only companies matching all specified tags will be returned.
	MatchAllTags *bool

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of companies to retrieve per page. Defaults to 50.
	PageSize int64

	// CountMode selects whether the API computes the exact number of companies
	// matching the filters, reported in Meta.Page.Count. Defaults to
	// twapi.ListCountModeDefault, which leaves the decision to the API.
	CountMode twapi.ListCountMode

	// Fields restricts the attributes returned for the company and each of its
	// sideloads. Each slot of CompanyListFields is a separate `fields[entity]=…`
	// selection; populated slots restrict the response, empty slots return the
	// API default. Use the generated CompanyField / CustomFieldField /
	// CustomFieldValueField constants to ensure values match real attributes.
	Fields CompanyListFields
}

func (c CompanyListRequestFilters) apply(req *http.Request) {
	c.CompanyRequestFilters.apply(req)

	query := req.URL.Query()
	if c.SearchTerm != "" {
		query.Set("searchTerm", c.SearchTerm)
	}
	if len(c.TagIDs) > 0 {
		tagIDs := make([]string, len(c.TagIDs))
		for i, id := range c.TagIDs {
			tagIDs[i] = strconv.FormatInt(id, 10)
		}
		query.Set("projectTagIds", strings.Join(tagIDs, ","))
	}
	if c.MatchAllTags != nil {
		query.Set("matchAllProjectTags", strconv.FormatBool(*c.MatchAllTags))
	}
	if c.OrderBy != "" {
		query.Set("orderBy", string(c.OrderBy))
	}
	if c.OrderMode != "" {
		query.Set("orderMode", string(c.OrderMode))
	}
	if c.OrderByCustomFieldID > 0 {
		query.Set("orderByCustomFieldId", strconv.FormatInt(c.OrderByCustomFieldID, 10))
	}
	if c.Page > 0 {
		query.Set("page", strconv.FormatInt(c.Page, 10))
	}
	if c.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(c.PageSize, 10))
	}
	c.Fields.apply(query)
	c.CountMode.Apply(query)
	req.URL.RawQuery = query.Encode()
}

// CompanyListRequest represents the request body for loading multiple
// clients/companies.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/companies/get-projects-api-v3-companies-json
type CompanyListRequest struct {
	// Filters contains the filters for loading multiple companies.
	Filters CompanyListRequestFilters
}

// NewCompanyListRequest creates a new CompanyListRequest with default values.
func NewCompanyListRequest() CompanyListRequest {
	return CompanyListRequest{
		Filters: CompanyListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the CompanyListRequest.
func (c CompanyListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/companies.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	c.Filters.apply(req)

	return req, nil
}

// CompanyListResponse contains information by multiple clients/companies
// matching the request filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/companies/get-projects-api-v3-companies-json
//
// sparsefields:list
type CompanyListResponse struct {
	request CompanyListRequest

	// Meta contains metadata about the response, including pagination details.
	Meta twapi.ListMeta `json:"meta"`

	// Companies is the list of companies matching the request filters.
	Companies []Company `json:"companies"`

	// Included contains related objects included in the response.
	Included struct {
		// CustomFields contains the custom fields associated with the company.
		//
		// The key is the string representation of the custom field ID.
		CustomFields map[string]CustomField `json:"customfields,omitempty"`
		// CustomFieldValues contains the values of the custom fields associated
		// with the company.
		//
		// The key is the string representation of the custom field value ID.
		CustomFieldValues map[string]CustomFieldValue `json:"customfieldCompanies,omitempty"`
	} `json:"included"`
}

// HandleHTTPResponse handles the HTTP response for the CompanyListResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (c *CompanyListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list companies")
	}

	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode list companies response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (c *CompanyListResponse) SetRequest(req CompanyListRequest) {
	c.request = req
	c.Meta.ResolveCount(req.Filters.CountMode)
}

// Iterate returns the request set to the next page, if available. If there are
// no more pages, a nil request is returned.
func (c *CompanyListResponse) Iterate() *CompanyListRequest {
	if !c.Meta.Page.HasMore {
		return nil
	}
	req := c.request
	req.Filters.Page++
	return &req
}

// CompanyList retrieves multiple clients/companies using the provided request
// and returns the response.
func CompanyList(
	ctx context.Context,
	engine *twapi.Engine,
	req CompanyListRequest,
) (*CompanyListResponse, error) {
	return twapi.Execute[CompanyListRequest, *CompanyListResponse](ctx, engine, req)
}
