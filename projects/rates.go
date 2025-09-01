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
	_ twapi.HTTPRequester = (*RateUserGetRequest)(nil)
	_ twapi.HTTPResponser = (*RateUserGetResponse)(nil)
	_ twapi.HTTPRequester = (*RateInstallationUserListRequest)(nil)
	_ twapi.HTTPResponser = (*RateInstallationUserListResponse)(nil)
	_ twapi.HTTPRequester = (*RateInstallationUserGetRequest)(nil)
	_ twapi.HTTPResponser = (*RateInstallationUserGetResponse)(nil)
	_ twapi.HTTPRequester = (*RateInstallationUserUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*RateInstallationUserUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*RateInstallationUserBulkUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*RateInstallationUserBulkUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*RateProjectGetRequest)(nil)
	_ twapi.HTTPResponser = (*RateProjectGetResponse)(nil)
	_ twapi.HTTPRequester = (*RateProjectUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*RateProjectUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*RateProjectAndUsersUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*RateProjectAndUsersUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*RateProjectUserListRequest)(nil)
	_ twapi.HTTPResponser = (*RateProjectUserListResponse)(nil)
	_ twapi.HTTPRequester = (*RateProjectUserGetRequest)(nil)
	_ twapi.HTTPResponser = (*RateProjectUserGetResponse)(nil)
	_ twapi.HTTPRequester = (*RateProjectUserUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*RateProjectUserUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*RateProjectUserHistoryGetRequest)(nil)
	_ twapi.HTTPResponser = (*RateProjectUserHistoryGetResponse)(nil)
)

// Currency represents a currency in the rates system.
type Currency struct {
	// ID is the unique identifier of the currency.
	ID int64 `json:"id"`

	// Code is the currency code (e.g., "USD", "EUR").
	Code string `json:"code"`

	// Symbol is the currency symbol (e.g., "$", "€").
	Symbol string `json:"symbol"`

	// Name is the currency name.
	Name string `json:"name"`
}

// EffectiveRateSource represents the source of an effective rate.
type EffectiveRateSource string

const (
	// EffectiveRateSourceInstallationRate represents a rate derived from the user's installation rate.
	EffectiveRateSourceInstallationRate EffectiveRateSource = "installationrate"

	// EffectiveRateSourceProjectRate represents a rate derived from the project's default rate.
	EffectiveRateSourceProjectRate EffectiveRateSource = "projectrate"

	// EffectiveRateSourceUserProjectRate represents a rate derived from a user's project-specific rate.
	EffectiveRateSourceUserProjectRate EffectiveRateSource = "userprojectrate"
)

// BillableRate contains the rate and currency information for billable amounts.
type BillableRate struct {
	// Rate is the billable rate amount.
	Rate float64 `json:"rate"`

	// Currency is the currency information.
	Currency twapi.Relationship `json:"currency"`
}

// ProjectRate represents a project's rate information.
type ProjectRate struct {
	// ProjectID is the ID of the project.
	ProjectID int64 `json:"projectId"`

	// Rate is the rate amount.
	Rate int64 `json:"rate"`

	// Currency is the currency information.
	Currency Currency `json:"currency"`
}

// UserProjectRate represents a user's project rate (used in list responses).
type UserProjectRate struct {
	// Project is the relationship to the project.
	Project twapi.Relationship `json:"project"`

	// UserRate is the rate amount.
	UserRate int64 `json:"userRate"`
}

// RateUserGetRequestPath contains the path parameters for getting a user's rates.
type RateUserGetRequestPath struct {
	// ID is the unique identifier of the user whose rates are to be retrieved.
	ID int64
}

// RateUserGetRequestFilters contains the filters for getting a user's rates.
type RateUserGetRequestFilters struct {
	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of rates to retrieve per page. Defaults to 50.
	PageSize int64

	// IncludeInstallationRate includes the installation rate in the response.
	IncludeInstallationRate bool

	// IncludeUserCost includes the user cost in the response.
	IncludeUserCost bool

	// IncludeArchivedProjects includes archived projects in the response.
	IncludeArchivedProjects bool

	// IncludeDeletedProjects includes deleted projects in the response.
	IncludeDeletedProjects bool

	// Include specifies which related data to include.
	Include []string
}

// RateUserGetRequest represents the request for getting a user's rates.
type RateUserGetRequest struct {
	// Path contains the path parameters for the request.
	Path RateUserGetRequestPath

	// Filters contains the filters for the request.
	Filters RateUserGetRequestFilters
}

// NewRateUserGetRequest creates a new RateUserGetRequest with the provided user ID and default values.
func NewRateUserGetRequest(userID int64) RateUserGetRequest {
	return RateUserGetRequest{
		Path: RateUserGetRequestPath{
			ID: userID,
		},
		Filters: RateUserGetRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the RateUserGetRequest.
func (r RateUserGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/people/%d/rates", server, r.Path.ID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	if r.Filters.Page > 0 {
		query.Set("page", strconv.FormatInt(r.Filters.Page, 10))
	}
	if r.Filters.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(r.Filters.PageSize, 10))
	}
	if r.Filters.IncludeInstallationRate {
		query.Set("includeInstallationRate", "true")
	}
	if r.Filters.IncludeUserCost {
		query.Set("includeUserCost", "true")
	}
	if r.Filters.IncludeArchivedProjects {
		query.Set("includeArchivedProjects", "true")
	}
	if r.Filters.IncludeDeletedProjects {
		query.Set("includeDeletedProjects", "true")
	}
	if len(r.Filters.Include) > 0 {
		for _, include := range r.Filters.Include {
			query.Add("include", include)
		}
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// RateUserGetResponse represents the response for getting a user's rates.
type RateUserGetResponse struct {
	// ProjectRates contains project-specific rates.
	ProjectRates []UserProjectRate `json:"projectRates"`

	// InstallationRate is the user's installation rate (optional).
	InstallationRate *int64 `json:"installationRate,omitempty"`

	// InstallationRates contains rates in different currencies (optional).
	InstallationRates map[int64]twapi.Money `json:"installationRates,omitempty"`

	// UserCost is the user's cost (optional).
	UserCost *int64 `json:"userCost,omitempty"`

	// Meta contains pagination information.
	Meta struct {
		Page struct {
			PageOffset int64 `json:"pageOffset"`
			PageSize   int64 `json:"pageSize"`
			Count      int64 `json:"count"`
			HasMore    bool  `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`

	// Included contains related data.
	Included struct {
		Currencies map[string]Currency           `json:"currencies,omitempty"`
		Projects   map[string]twapi.Relationship `json:"projects,omitempty"`
	} `json:"included"`
}

// HandleHTTPResponse handles the HTTP response for the RateUserGetResponse.
func (r *RateUserGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to get user rates")
	}

	if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
		return fmt.Errorf("failed to decode get user rates response: %w", err)
	}
	return nil
}

// RateUserGet retrieves a user's rates using the provided request and returns the response.
func RateUserGet(
	ctx context.Context,
	engine *twapi.Engine,
	req RateUserGetRequest,
) (*RateUserGetResponse, error) {
	return twapi.Execute[RateUserGetRequest, *RateUserGetResponse](ctx, engine, req)
}

// RateInstallationUserListRequestFilters contains the filters for listing installation user rates.
type RateInstallationUserListRequestFilters struct {
	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of rates to retrieve per page. Defaults to 50.
	PageSize int64
}

// RateInstallationUserListRequest represents the request for listing installation user rates.
type RateInstallationUserListRequest struct {
	// Filters contains the filters for the request.
	Filters RateInstallationUserListRequestFilters
}

// NewRateInstallationUserListRequest creates a new RateInstallationUserListRequest with default values.
func NewRateInstallationUserListRequest() RateInstallationUserListRequest {
	return RateInstallationUserListRequest{
		Filters: RateInstallationUserListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the RateInstallationUserListRequest.
func (r RateInstallationUserListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/rates/installation/users.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	if r.Filters.Page > 0 {
		query.Set("page", strconv.FormatInt(r.Filters.Page, 10))
	}
	if r.Filters.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(r.Filters.PageSize, 10))
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// RateInstallationUserListResponse represents the response for listing installation user rates.
type RateInstallationUserListResponse struct {
	request RateInstallationUserListRequest

	// Meta contains pagination information.
	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`

	// UserRates contains the list of user rates.
	UserRates []struct {
		User twapi.Relationship `json:"user"`
		Rate int64              `json:"rate"`
	} `json:"userRates"`

	// Included contains related data.
	Included struct {
		Currencies map[string]Currency           `json:"currencies"`
		Users      map[string]twapi.Relationship `json:"users"`
	} `json:"included"`
}

// HandleHTTPResponse handles the HTTP response for the RateInstallationUserListResponse.
func (r *RateInstallationUserListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list installation user rates")
	}

	if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
		return fmt.Errorf("failed to decode list installation user rates response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response.
func (r *RateInstallationUserListResponse) SetRequest(req RateInstallationUserListRequest) {
	r.request = req
}

// Iterate returns the request set to the next page, if available.
func (r *RateInstallationUserListResponse) Iterate() *RateInstallationUserListRequest {
	if !r.Meta.Page.HasMore {
		return nil
	}
	req := r.request
	req.Filters.Page++
	return &req
}

// RateInstallationUserList retrieves installation user rates using the provided request and returns the response.
func RateInstallationUserList(
	ctx context.Context,
	engine *twapi.Engine,
	req RateInstallationUserListRequest,
) (*RateInstallationUserListResponse, error) {
	return twapi.Execute[RateInstallationUserListRequest, *RateInstallationUserListResponse](ctx, engine, req)
}

// RateInstallationUserGetRequestPath contains the path parameters for getting an installation user rate.
type RateInstallationUserGetRequestPath struct {
	// UserID is the unique identifier of the user whose installation rate is to be retrieved.
	UserID int64
}

// RateInstallationUserGetRequestFilters contains the filters for getting an installation user rate.
type RateInstallationUserGetRequestFilters struct {
	// Include specifies which related data to include.
	Include []string
}

// RateInstallationUserGetRequest represents the request for getting an installation user rate.
type RateInstallationUserGetRequest struct {
	// Path contains the path parameters for the request.
	Path RateInstallationUserGetRequestPath

	// Filters contains the filters for the request.
	Filters RateInstallationUserGetRequestFilters
}

// NewRateInstallationUserGetRequest creates a new RateInstallationUserGetRequest with the provided user ID.
func NewRateInstallationUserGetRequest(userID int64) RateInstallationUserGetRequest {
	return RateInstallationUserGetRequest{
		Path: RateInstallationUserGetRequestPath{
			UserID: userID,
		},
	}
}

// HTTPRequest creates an HTTP request for the RateInstallationUserGetRequest.
func (r RateInstallationUserGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/rates/installation/users/%d.json", server, r.Path.UserID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	if len(r.Filters.Include) > 0 {
		for _, include := range r.Filters.Include {
			query.Add("include", include)
		}
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// RateInstallationUserGetResponse represents the response for getting an installation user rate.
type RateInstallationUserGetResponse struct {
	// UserRate is the user's rate (legacy field).
	UserRate int64 `json:"userRate"`

	// UserRates contains rates in different currencies (key is currency ID as string for JSON compatibility).
	UserRates map[string]twapi.Money `json:"userRates"`

	// Included contains related data.
	Included struct {
		Currencies map[string]Currency `json:"currencies"`
	} `json:"included"`
}

// HandleHTTPResponse handles the HTTP response for the RateInstallationUserGetResponse.
func (r *RateInstallationUserGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to get installation user rate")
	}

	if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
		return fmt.Errorf("failed to decode get installation user rate response: %w", err)
	}
	return nil
}

// RateInstallationUserGet retrieves an installation user rate using the provided request and returns the response.
func RateInstallationUserGet(
	ctx context.Context,
	engine *twapi.Engine,
	req RateInstallationUserGetRequest,
) (*RateInstallationUserGetResponse, error) {
	return twapi.Execute[RateInstallationUserGetRequest, *RateInstallationUserGetResponse](ctx, engine, req)
}

// RateInstallationUserUpdateRequestPath contains the path parameters for updating an installation user rate.
type RateInstallationUserUpdateRequestPath struct {
	// UserID is the unique identifier of the user whose rate is to be updated.
	UserID int64
}

// RateInstallationUserUpdateRequest represents the request for updating an installation user rate.
type RateInstallationUserUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path RateInstallationUserUpdateRequestPath `json:"-"`

	// CurrencyID is the ID of the currency for the rate (optional, only used in multi-currency mode).
	CurrencyID *int64 `json:"currencyId,omitempty"`

	// UserRate is the new rate for the user. Use nil to clear/remove the rate.
	UserRate *int64 `json:"userRate"`
}

// NewRateInstallationUserUpdateRequest creates a new RateInstallationUserUpdateRequest.
func NewRateInstallationUserUpdateRequest(userID int64, rate *int64) RateInstallationUserUpdateRequest {
	return RateInstallationUserUpdateRequest{
		Path: RateInstallationUserUpdateRequestPath{
			UserID: userID,
		},
		UserRate: rate,
	}
}

// HTTPRequest creates an HTTP request for the RateInstallationUserUpdateRequest.
func (r RateInstallationUserUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/rates/installation/users/%d.json", server, r.Path.UserID)

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(r); err != nil {
		return nil, fmt.Errorf("failed to encode update installation user rate request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// RateInstallationUserUpdateResponse represents the response for updating an installation user rate.
type RateInstallationUserUpdateResponse struct{}

// HandleHTTPResponse handles the HTTP response for the RateInstallationUserUpdateResponse.
func (r *RateInstallationUserUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to update installation user rate")
	}
	return nil
}

// RateInstallationUserUpdate updates an installation user rate using the provided request and returns the response.
func RateInstallationUserUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req RateInstallationUserUpdateRequest,
) (*RateInstallationUserUpdateResponse, error) {
	return twapi.Execute[RateInstallationUserUpdateRequest, *RateInstallationUserUpdateResponse](ctx, engine, req)
}

// RateInstallationUserBulkUpdateRequest represents the request for bulk updating installation user rates.
type RateInstallationUserBulkUpdateRequest struct {
	// All indicates whether to update all users.
	All bool `json:"all,omitempty"`

	// IDs contains the user IDs to update (if All is false).
	IDs []int64 `json:"ids,omitempty"`

	// ExcludeIDs contains user IDs to exclude (if All is true).
	ExcludeIDs []int64 `json:"excludeIds,omitempty"`

	// CurrencyID is the ID of the currency for the rate (optional, only used in multi-currency mode).
	CurrencyID *int64 `json:"currencyId,omitempty"`

	// UserRate is the new rate for the users. Use nil to clear/remove the rate.
	UserRate *int64 `json:"userRate"`
}

// NewRateInstallationUserBulkUpdateRequest creates a new RateInstallationUserBulkUpdateRequest.
func NewRateInstallationUserBulkUpdateRequest(rate *int64) RateInstallationUserBulkUpdateRequest {
	return RateInstallationUserBulkUpdateRequest{
		UserRate: rate,
	}
}

// HTTPRequest creates an HTTP request for the RateInstallationUserBulkUpdateRequest.
func (r RateInstallationUserBulkUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/rates/installation/users/bulk/update.json"

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(r); err != nil {
		return nil, fmt.Errorf("failed to encode bulk update installation user rates request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// RateInstallationUserBulkUpdateResponse represents the response for bulk updating installation user rates.
type RateInstallationUserBulkUpdateResponse struct {
	// All indicates whether all users were updated.
	All bool `json:"all"`

	// IDs contains the user IDs that were updated.
	IDs []int64 `json:"ids"`

	// ExcludeIDs contains user IDs that were excluded.
	ExcludeIDs []int64 `json:"excludeIds"`

	// Rate is the rate that was set.
	Rate int64 `json:"rate"`
}

// HandleHTTPResponse handles the HTTP response for the RateInstallationUserBulkUpdateResponse.
func (r *RateInstallationUserBulkUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to bulk update installation user rates")
	}

	if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
		return fmt.Errorf("failed to decode bulk update installation user rates response: %w", err)
	}
	return nil
}

// RateInstallationUserBulkUpdate bulk updates installation user
// rates using the provided request and returns the response.
func RateInstallationUserBulkUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req RateInstallationUserBulkUpdateRequest,
) (*RateInstallationUserBulkUpdateResponse, error) {
	return twapi.Execute[RateInstallationUserBulkUpdateRequest, *RateInstallationUserBulkUpdateResponse](ctx, engine, req)
}

// RateProjectGetRequestPath contains the path parameters for getting a project rate.
type RateProjectGetRequestPath struct {
	// ProjectID is the unique identifier of the project whose rate is to be retrieved.
	ProjectID int64
}

// RateProjectGetRequestFilters contains the filters for getting a project rate.
type RateProjectGetRequestFilters struct {
	// Include specifies which related data to include.
	Include []string
}

// RateProjectGetRequest represents the request for getting a project rate.
type RateProjectGetRequest struct {
	// Path contains the path parameters for the request.
	Path RateProjectGetRequestPath

	// Filters contains the filters for the request.
	Filters RateProjectGetRequestFilters
}

// NewRateProjectGetRequest creates a new RateProjectGetRequest with the provided project ID.
func NewRateProjectGetRequest(projectID int64) RateProjectGetRequest {
	return RateProjectGetRequest{
		Path: RateProjectGetRequestPath{
			ProjectID: projectID,
		},
	}
}

// HTTPRequest creates an HTTP request for the RateProjectGetRequest.
func (r RateProjectGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/rates/projects/%d.json", server, r.Path.ProjectID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	if len(r.Filters.Include) > 0 {
		for _, include := range r.Filters.Include {
			query.Add("include", include)
		}
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// RateProjectGetResponse represents the response for getting a project rate.
type RateProjectGetResponse struct {
	// ProjectRate is the project's rate.
	ProjectRate int64 `json:"projectRate"`

	// Rate is the rate in money format.
	Rate twapi.Money `json:"rate"`

	// Included contains related data.
	Included struct {
		Currencies map[string]Currency `json:"currencies"`
	} `json:"included"`
}

// HandleHTTPResponse handles the HTTP response for the RateProjectGetResponse.
func (r *RateProjectGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to get project rate")
	}

	if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
		return fmt.Errorf("failed to decode get project rate response: %w", err)
	}
	return nil
}

// RateProjectGet retrieves a project rate using the provided request and returns the response.
func RateProjectGet(
	ctx context.Context,
	engine *twapi.Engine,
	req RateProjectGetRequest,
) (*RateProjectGetResponse, error) {
	return twapi.Execute[RateProjectGetRequest, *RateProjectGetResponse](ctx, engine, req)
}

// RateProjectUpdateRequestPath contains the path parameters for updating a project rate.
type RateProjectUpdateRequestPath struct {
	// ProjectID is the unique identifier of the project whose rate is to be updated.
	ProjectID int64
}

// RateProjectUpdateRequest represents the request for updating a project rate.
type RateProjectUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path RateProjectUpdateRequestPath `json:"-"`

	// ProjectRate is the new rate for the project. Use nil to clear/remove the rate.
	ProjectRate *int64 `json:"projectRate"`
}

// NewRateProjectUpdateRequest creates a new RateProjectUpdateRequest.
func NewRateProjectUpdateRequest(projectID int64, rate *int64) RateProjectUpdateRequest {
	return RateProjectUpdateRequest{
		Path: RateProjectUpdateRequestPath{
			ProjectID: projectID,
		},
		ProjectRate: rate,
	}
}

// HTTPRequest creates an HTTP request for the RateProjectUpdateRequest.
func (r RateProjectUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/rates/projects/%d.json", server, r.Path.ProjectID)

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(r); err != nil {
		return nil, fmt.Errorf("failed to encode update project rate request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// RateProjectUpdateResponse represents the response for updating a project rate.
type RateProjectUpdateResponse struct{}

// HandleHTTPResponse handles the HTTP response for the RateProjectUpdateResponse.
func (r *RateProjectUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusNoContent {
		return twapi.NewHTTPError(resp, "failed to update project rate")
	}
	return nil
}

// RateProjectUpdate updates a project rate using the provided request and returns the response.
func RateProjectUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req RateProjectUpdateRequest,
) (*RateProjectUpdateResponse, error) {
	return twapi.Execute[RateProjectUpdateRequest, *RateProjectUpdateResponse](ctx, engine, req)
}

// ProjectUserRateRequest represents a user rate update within a project.
type ProjectUserRateRequest struct {
	// FromDate is the date from which this rate is effective.
	FromDate *time.Time `json:"fromDate,omitempty"`

	// User is the user relationship.
	User twapi.Relationship `json:"user"`

	// UserRate is the rate for the user.
	UserRate int64 `json:"userRate"`
}

// RateProjectAndUsersUpdateRequestPath contains the path parameters for updating a project and user rates.
type RateProjectAndUsersUpdateRequestPath struct {
	// ProjectID is the unique identifier of the project.
	ProjectID int64
}

// RateProjectAndUsersUpdateRequest represents the request for updating a project rate and user rates.
type RateProjectAndUsersUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path RateProjectAndUsersUpdateRequestPath `json:"-"`

	// ProjectRate is the new rate for the project.
	ProjectRate int64 `json:"projectRate"`

	// UserRates contains the user rates to set as exceptions.
	UserRates []ProjectUserRateRequest `json:"userRates,omitempty"`
}

// NewRateProjectAndUsersUpdateRequest creates a new RateProjectAndUsersUpdateRequest.
func NewRateProjectAndUsersUpdateRequest(projectID int64, projectRate int64) RateProjectAndUsersUpdateRequest {
	return RateProjectAndUsersUpdateRequest{
		Path: RateProjectAndUsersUpdateRequestPath{
			ProjectID: projectID,
		},
		ProjectRate: projectRate,
	}
}

// HTTPRequest creates an HTTP request for the RateProjectAndUsersUpdateRequest.
func (r RateProjectAndUsersUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/rates/projects/%d/actions/update", server, r.Path.ProjectID)

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(r); err != nil {
		return nil, fmt.Errorf("failed to encode update project and users rate request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// RateProjectAndUsersUpdateResponse represents the response for updating a project rate and user rates.
type RateProjectAndUsersUpdateResponse struct{}

// HandleHTTPResponse handles the HTTP response for the RateProjectAndUsersUpdateResponse.
func (r *RateProjectAndUsersUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusNoContent {
		return twapi.NewHTTPError(resp, "failed to update project and users rates")
	}
	return nil
}

// RateProjectAndUsersUpdate updates a project rate and user rates using the provided request and returns the response.
func RateProjectAndUsersUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req RateProjectAndUsersUpdateRequest,
) (*RateProjectAndUsersUpdateResponse, error) {
	return twapi.Execute[RateProjectAndUsersUpdateRequest, *RateProjectAndUsersUpdateResponse](ctx, engine, req)
}

// RateProjectUserListRequestPath contains the path parameters for listing project user rates.
type RateProjectUserListRequestPath struct {
	// ProjectID is the unique identifier of the project.
	ProjectID int64
}

// RateProjectUserListRequestFilters contains the filters for listing project user rates.
type RateProjectUserListRequestFilters struct {
	// SearchTerm is an optional search term to filter by first name or last name.
	SearchTerm string

	// OrderBy specifies the ordering of results.
	OrderBy string

	// OrderMode specifies the order direction (asc, desc).
	OrderMode string

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of rates to retrieve per page. Defaults to 50.
	PageSize int64
}

// RateProjectUserListRequest represents the request for listing project user rates.
type RateProjectUserListRequest struct {
	// Path contains the path parameters for the request.
	Path RateProjectUserListRequestPath

	// Filters contains the filters for the request.
	Filters RateProjectUserListRequestFilters
}

// NewRateProjectUserListRequest creates a new RateProjectUserListRequest.
func NewRateProjectUserListRequest(projectID int64) RateProjectUserListRequest {
	return RateProjectUserListRequest{
		Path: RateProjectUserListRequestPath{
			ProjectID: projectID,
		},
		Filters: RateProjectUserListRequestFilters{
			Page:      1,
			PageSize:  50,
			OrderMode: "asc",
		},
	}
}

// HTTPRequest creates an HTTP request for the RateProjectUserListRequest.
func (r RateProjectUserListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/rates/projects/%d/users", server, r.Path.ProjectID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	if r.Filters.SearchTerm != "" {
		query.Set("searchTerm", r.Filters.SearchTerm)
	}
	if r.Filters.OrderBy != "" {
		query.Set("orderBy", r.Filters.OrderBy)
	}
	if r.Filters.OrderMode != "" {
		query.Set("orderMode", r.Filters.OrderMode)
	}
	if r.Filters.Page > 0 {
		query.Set("page", strconv.FormatInt(r.Filters.Page, 10))
	}
	if r.Filters.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(r.Filters.PageSize, 10))
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// EffectiveUserProjectRate represents an effective user project rate.
type EffectiveUserProjectRate struct {
	// User is the user relationship.
	User twapi.Relationship `json:"user"`

	// EffectiveRate is the effective rate.
	EffectiveRate twapi.Money `json:"effectiveRate"`

	// UserProjectRate is the user's project-specific rate.
	UserProjectRate *twapi.Money `json:"userProjectRate,omitempty"`

	// UserInstallationRate is the user's installation rate.
	UserInstallationRate *twapi.Money `json:"userInstallationRate,omitempty"`

	// ProjectRate is the project's default rate.
	ProjectRate *twapi.Money `json:"projectRate,omitempty"`

	// Source indicates the source of the effective rate.
	Source *EffectiveRateSource `json:"source,omitempty"`

	// FromDate is when this rate became effective.
	FromDate *time.Time `json:"fromDate,omitempty"`

	// ToDate is when this rate stops being effective.
	ToDate *time.Time `json:"toDate,omitempty"`

	// UpdatedAt is when this rate was last updated.
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`

	// UpdatedBy is who last updated this rate.
	UpdatedBy *twapi.Relationship `json:"updatedBy,omitempty"`

	// BillableRate contains the rate and currency for billing.
	BillableRate *BillableRate `json:"billableRate,omitempty"`
}

// RateProjectUserListResponse represents the response for listing project user rates.
type RateProjectUserListResponse struct {
	request RateProjectUserListRequest

	// Meta contains pagination information.
	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`

	// UserRates contains the list of effective user project rates.
	UserRates []EffectiveUserProjectRate `json:"userRates"`

	// Included contains related data.
	Included struct {
		CostRates  map[string]any                `json:"costRates"`
		Currencies map[string]Currency           `json:"currencies"`
		Users      map[string]twapi.Relationship `json:"users"`
	} `json:"included"`
}

// HandleHTTPResponse handles the HTTP response for the RateProjectUserListResponse.
func (r *RateProjectUserListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list project user rates")
	}

	if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
		return fmt.Errorf("failed to decode list project user rates response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response.
func (r *RateProjectUserListResponse) SetRequest(req RateProjectUserListRequest) {
	r.request = req
}

// Iterate returns the request set to the next page, if available.
func (r *RateProjectUserListResponse) Iterate() *RateProjectUserListRequest {
	if !r.Meta.Page.HasMore {
		return nil
	}
	req := r.request
	req.Filters.Page++
	return &req
}

// RateProjectUserList retrieves project user rates using the provided request and returns the response.
func RateProjectUserList(
	ctx context.Context,
	engine *twapi.Engine,
	req RateProjectUserListRequest,
) (*RateProjectUserListResponse, error) {
	return twapi.Execute[RateProjectUserListRequest, *RateProjectUserListResponse](ctx, engine, req)
}

// RateProjectUserGetRequestPath contains the path parameters for getting a project user rate.
type RateProjectUserGetRequestPath struct {
	// ProjectID is the unique identifier of the project.
	ProjectID int64

	// UserID is the unique identifier of the user.
	UserID int64
}

// RateProjectUserGetRequestFilters contains the filters for getting a project user rate.
type RateProjectUserGetRequestFilters struct {
	// Include specifies which related data to include.
	Include []string
}

// RateProjectUserGetRequest represents the request for getting a project user rate.
type RateProjectUserGetRequest struct {
	// Path contains the path parameters for the request.
	Path RateProjectUserGetRequestPath

	// Filters contains the filters for the request.
	Filters RateProjectUserGetRequestFilters
}

// NewRateProjectUserGetRequest creates a new RateProjectUserGetRequest.
func NewRateProjectUserGetRequest(projectID int64, userID int64) RateProjectUserGetRequest {
	return RateProjectUserGetRequest{
		Path: RateProjectUserGetRequestPath{
			ProjectID: projectID,
			UserID:    userID,
		},
	}
}

// HTTPRequest creates an HTTP request for the RateProjectUserGetRequest.
func (r RateProjectUserGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/rates/projects/%d/users/%d.json", server, r.Path.ProjectID, r.Path.UserID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	if len(r.Filters.Include) > 0 {
		for _, include := range r.Filters.Include {
			query.Add("include", include)
		}
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// RateProjectUserGetResponse represents the response for getting a project user rate.
type RateProjectUserGetResponse struct {
	// UserRate is the user's rate.
	UserRate int64 `json:"userRate"`

	// Rate is the rate in money format.
	Rate twapi.Money `json:"rate"`

	// Included contains related data.
	Included struct {
		Currencies map[string]Currency `json:"currencies"`
	} `json:"included"`
}

// HandleHTTPResponse handles the HTTP response for the RateProjectUserGetResponse.
func (r *RateProjectUserGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to get project user rate")
	}

	if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
		return fmt.Errorf("failed to decode get project user rate response: %w", err)
	}
	return nil
}

// RateProjectUserGet retrieves a project user rate using the provided request and returns the response.
func RateProjectUserGet(
	ctx context.Context,
	engine *twapi.Engine,
	req RateProjectUserGetRequest,
) (*RateProjectUserGetResponse, error) {
	return twapi.Execute[RateProjectUserGetRequest, *RateProjectUserGetResponse](ctx, engine, req)
}

// RateProjectUserUpdateRequestPath contains the path parameters for updating a project user rate.
type RateProjectUserUpdateRequestPath struct {
	// ProjectID is the unique identifier of the project.
	ProjectID int64

	// UserID is the unique identifier of the user.
	UserID int64
}

// RateProjectUserUpdateRequest represents the request for updating a project user rate.
type RateProjectUserUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path RateProjectUserUpdateRequestPath `json:"-"`

	// CurrencyID is the ID of the currency for the rate (optional, only used in multi-currency mode).
	CurrencyID *int64 `json:"currencyId,omitempty"`

	// UserRate is the new rate for the user. Use nil to clear/remove the rate.
	UserRate *int64 `json:"userRate"`
}

// NewRateProjectUserUpdateRequest creates a new RateProjectUserUpdateRequest.
func NewRateProjectUserUpdateRequest(projectID int64, userID int64, rate *int64) RateProjectUserUpdateRequest {
	return RateProjectUserUpdateRequest{
		Path: RateProjectUserUpdateRequestPath{
			ProjectID: projectID,
			UserID:    userID,
		},
		UserRate: rate,
	}
}

// HTTPRequest creates an HTTP request for the RateProjectUserUpdateRequest.
func (r RateProjectUserUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/rates/projects/%d/users/%d.json", server, r.Path.ProjectID, r.Path.UserID)

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(r); err != nil {
		return nil, fmt.Errorf("failed to encode update project user rate request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// RateProjectUserUpdateResponse represents the response for updating a project user rate.
type RateProjectUserUpdateResponse struct {
	// UserRate is the user's updated rate.
	UserRate int64 `json:"userRate"`

	// Rate is the rate in money format.
	Rate twapi.Money `json:"rate"`

	// Included contains related data.
	Included struct {
		Currencies map[string]Currency `json:"currencies"`
	} `json:"included"`
}

// HandleHTTPResponse handles the HTTP response for the RateProjectUserUpdateResponse.
func (r *RateProjectUserUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to update project user rate")
	}

	// API returns 201 with no content, so no need to decode response body
	return nil
}

// RateProjectUserUpdate updates a project user rate using the provided request and returns the response.
func RateProjectUserUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req RateProjectUserUpdateRequest,
) (*RateProjectUserUpdateResponse, error) {
	return twapi.Execute[RateProjectUserUpdateRequest, *RateProjectUserUpdateResponse](ctx, engine, req)
}

// RateProjectUserHistoryGetRequestPath contains the path parameters for getting project user rate history.
type RateProjectUserHistoryGetRequestPath struct {
	// ProjectID is the unique identifier of the project.
	ProjectID int64

	// UserID is the unique identifier of the user.
	UserID int64
}

// RateProjectUserHistoryGetRequestFilters contains the filters for getting project user rate history.
type RateProjectUserHistoryGetRequestFilters struct {
	// SearchTerm is an optional search term to filter by first name or last name.
	SearchTerm string

	// OrderBy specifies the ordering of results.
	OrderBy string

	// OrderMode specifies the order direction (asc, desc).
	OrderMode string

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of rates to retrieve per page. Defaults to 50.
	PageSize int64
}

// RateProjectUserHistoryGetRequest represents the request for getting project user rate history.
type RateProjectUserHistoryGetRequest struct {
	// Path contains the path parameters for the request.
	Path RateProjectUserHistoryGetRequestPath

	// Filters contains the filters for the request.
	Filters RateProjectUserHistoryGetRequestFilters
}

// NewRateProjectUserHistoryGetRequest creates a new RateProjectUserHistoryGetRequest.
func NewRateProjectUserHistoryGetRequest(projectID int64, userID int64) RateProjectUserHistoryGetRequest {
	return RateProjectUserHistoryGetRequest{
		Path: RateProjectUserHistoryGetRequestPath{
			ProjectID: projectID,
			UserID:    userID,
		},
		Filters: RateProjectUserHistoryGetRequestFilters{
			Page:      1,
			PageSize:  50,
			OrderMode: "asc",
		},
	}
}

// HTTPRequest creates an HTTP request for the RateProjectUserHistoryGetRequest.
func (r RateProjectUserHistoryGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/rates/projects/%d/users/%d/history", server, r.Path.ProjectID, r.Path.UserID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	if r.Filters.SearchTerm != "" {
		query.Set("searchTerm", r.Filters.SearchTerm)
	}
	if r.Filters.OrderBy != "" {
		query.Set("orderBy", r.Filters.OrderBy)
	}
	if r.Filters.OrderMode != "" {
		query.Set("orderMode", r.Filters.OrderMode)
	}
	if r.Filters.Page > 0 {
		query.Set("page", strconv.FormatInt(r.Filters.Page, 10))
	}
	if r.Filters.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(r.Filters.PageSize, 10))
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// UserRateHistory represents a historical rate entry for a user.
type UserRateHistory struct {
	// Rate is the rate amount.
	Rate twapi.Money `json:"rate"`

	// FromDate is the date from which this rate was effective.
	FromDate *time.Time `json:"fromDate"`

	// ToDate is the date until which this rate was effective.
	ToDate *time.Time `json:"toDate,omitempty"`

	// CreatedAt is the date when this rate was created.
	CreatedAt *time.Time `json:"createdAt"`

	// UpdatedAt is the date when this rate was last updated.
	UpdatedAt *time.Time `json:"updatedAt"`
}

// RateProjectUserHistoryGetResponse represents the response for getting project user rate history.
type RateProjectUserHistoryGetResponse struct {
	request RateProjectUserHistoryGetRequest

	// Meta contains pagination information.
	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`

	// UserRateHistory contains the list of historical rates.
	UserRateHistory []UserRateHistory `json:"userRateHistory"`

	// Included contains related data.
	Included struct {
		Currencies map[string]Currency           `json:"currencies"`
		Users      map[string]twapi.Relationship `json:"users"`
	} `json:"included"`
}

// HandleHTTPResponse handles the HTTP response for the RateProjectUserHistoryGetResponse.
func (r *RateProjectUserHistoryGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to get project user rate history")
	}

	if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
		return fmt.Errorf("failed to decode get project user rate history response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response.
func (r *RateProjectUserHistoryGetResponse) SetRequest(req RateProjectUserHistoryGetRequest) {
	r.request = req
}

// Iterate returns the request set to the next page, if available.
func (r *RateProjectUserHistoryGetResponse) Iterate() *RateProjectUserHistoryGetRequest {
	if !r.Meta.Page.HasMore {
		return nil
	}
	req := r.request
	req.Filters.Page++
	return &req
}

// RateProjectUserHistoryGet retrieves project user rate history using the provided request and returns the response.
func RateProjectUserHistoryGet(
	ctx context.Context,
	engine *twapi.Engine,
	req RateProjectUserHistoryGetRequest,
) (*RateProjectUserHistoryGetResponse, error) {
	return twapi.Execute[RateProjectUserHistoryGetRequest, *RateProjectUserHistoryGetResponse](ctx, engine, req)
}
