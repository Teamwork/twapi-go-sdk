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

	return req, nil
}

// CompanyGetResponse contains all the information related to a client/company.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/companies/get-projects-api-v3-companies-company-id-json
type CompanyGetResponse struct {
	Company Company `json:"company"`
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

// CompanyListRequestFilters contains the filters for loading multiple
// clients/companies.
type CompanyListRequestFilters struct {
	// SearchTerm is an optional search term to filter clients/companies by name.
	SearchTerm string

	// TagIDs is an optional list of tag IDs to filter companies by tags.
	TagIDs []int64

	// MatchAllTags is an optional flag to indicate if all tags must match. If set
	// to true, only companies matching all specified tags will be returned.
	MatchAllTags *bool

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of companies to retrieve per page. Defaults to 50.
	PageSize int64
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

	query := req.URL.Query()
	if c.Filters.SearchTerm != "" {
		query.Set("searchTerm", c.Filters.SearchTerm)
	}
	if len(c.Filters.TagIDs) > 0 {
		tagIDs := make([]string, len(c.Filters.TagIDs))
		for i, id := range c.Filters.TagIDs {
			tagIDs[i] = strconv.FormatInt(id, 10)
		}
		query.Set("projectTagIds", strings.Join(tagIDs, ","))
	}
	if c.Filters.MatchAllTags != nil {
		query.Set("matchAllProjectTags", strconv.FormatBool(*c.Filters.MatchAllTags))
	}
	if c.Filters.Page > 0 {
		query.Set("page", strconv.FormatInt(c.Filters.Page, 10))
	}
	if c.Filters.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(c.Filters.PageSize, 10))
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// CompanyListResponse contains information by multiple clients/companies
// matching the request filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/companies/get-projects-api-v3-companies-json
type CompanyListResponse struct {
	request CompanyListRequest

	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`
	Companies []Company `json:"companies"`
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
