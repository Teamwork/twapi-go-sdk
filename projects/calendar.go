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
	_ twapi.HTTPRequester = (*CalendarListRequest)(nil)
	_ twapi.HTTPResponser = (*CalendarListResponse)(nil)
)

// CalendarType represents the type of calendar.
type CalendarType string

const (
	// CalendarTypeEvent represents a standard event calendar.
	CalendarTypeEvent CalendarType = "event"
	// CalendarTypeBlockedTime represents a blocked time calendar. This is used
	// for creating events that block off time without specific details. When
	// using this type the calendar name MUST be "blocked_time". There can only be
	// one blocked time calendar per account.
	CalendarTypeBlockedTime CalendarType = "blocked_time"
	// CalendarTypeGoogle represents a Google calendar. This is synchronized with
	// Google Calendar events. There are restrictions on what can be done with
	// these calendars.
	CalendarTypeGoogle CalendarType = "google"
	// CalendarTypeOutlook represents an Outlook calendar. This is synchronized
	// with Outlook Calendar events. There are restrictions on what can be done
	// with these calendars.
	CalendarTypeOutlook CalendarType = "outlook"
	// CalendarTypeHoliday represents a holiday calendar. This is used for marking
	// holidays.
	CalendarTypeHoliday CalendarType = "holiday"
)

// Calendar represents a calendar in Teamwork. Calendars can be of different
// types such as Google calendars, blocked time calendars, or other integrated
// calendar services.
type Calendar struct {
	// ID is the unique identifier of the calendar.
	ID int64 `json:"id"`

	// Name is the name of the calendar.
	Name string `json:"name"`

	// Type is the type of calendar (e.g., "google", "blocked_time").
	Type CalendarType `json:"type"`

	// CreatedAt is the date and time when the calendar was created.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt is the date and time when the calendar was last updated.
	UpdatedAt time.Time `json:"updatedAt"`
}

// CalendarCreateRequest represents the request to create a new calendar.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/calendars/post-calendars-json
type CalendarCreateRequest struct {
	// Name is the name of the calendar.
	Name string `json:"name"`

	// Type is the type of calendar.
	Type *CalendarType `json:"type,omitempty"`
}

// NewCalendarCreateRequest creates a new CalendarCreateRequest with the
// provided name.
func NewCalendarCreateRequest(name string) CalendarCreateRequest {
	return CalendarCreateRequest{
		Name: name,
	}
}

// HTTPRequest creates an HTTP request for the CalendarCreateRequest.
func (c CalendarCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/calendars.json"

	payload := struct {
		Calendar CalendarCreateRequest `json:"calendar"`
	}{Calendar: c}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create user request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// CalendarCreateResponse contains the response for creating a calendar.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/calendars/post-calendars-json
type CalendarCreateResponse struct {
	// Calendar is the created calendar.
	Calendar Calendar `json:"calendar"`
}

// HandleHTTPResponse handles the HTTP response for the CalendarCreateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (c *CalendarCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create calendar")
	}

	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode create calendar response: %w", err)
	}
	return nil
}

// CalendarCreate creates a new calendar using the provided request and returns
// the response.
func CalendarCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req CalendarCreateRequest,
) (*CalendarCreateResponse, error) {
	return twapi.Execute[CalendarCreateRequest, *CalendarCreateResponse](ctx, engine, req)
}

// CalendarDeleteRequestPath represents the path parameters for deleting a
// calendar.
type CalendarDeleteRequestPath struct {
	// ID is the unique identifier of the calendar to delete.
	ID int64
}

// CalendarDeleteRequest represents the request to delete a calendar.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/calendars/delete-calendars-id-json
type CalendarDeleteRequest struct {
	Path CalendarDeleteRequestPath
}

// NewCalendarDeleteRequest creates a new CalendarDeleteRequest with the
// provided calendar ID.
func NewCalendarDeleteRequest(id int64) CalendarDeleteRequest {
	return CalendarDeleteRequest{
		Path: CalendarDeleteRequestPath{
			ID: id,
		},
	}
}

// HTTPRequest creates an HTTP request for the CalendarDeleteRequest.
func (c CalendarDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/calendars/%d.json", server, c.Path.ID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// CalendarDeleteResponse contains the response for deleting a calendar.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/calendars/delete-calendars-id-json
type CalendarDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the CalendarDeleteResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (c *CalendarDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusNoContent {
		return twapi.NewHTTPError(resp, "failed to delete calendar")
	}
	return nil
}

// CalendarDelete deletes a calendar using the provided request and returns
// the response.
func CalendarDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req CalendarDeleteRequest,
) (*CalendarDeleteResponse, error) {
	return twapi.Execute[CalendarDeleteRequest, *CalendarDeleteResponse](ctx, engine, req)
}

// CalendarListRequestFilters contains filters for loading calendars.
type CalendarListRequestFilters struct {
	// Page is the page number for pagination.
	Page int64

	// PageSize is the number of items per page.
	PageSize int64
}

// CalendarListRequest represents the request for loading calendars.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/calendars/get-calendars-json
type CalendarListRequest struct {
	// Filters contains the filters for loading calendars.
	Filters CalendarListRequestFilters
}

// NewCalendarListRequest creates a new CalendarListRequest with default values.
func NewCalendarListRequest() CalendarListRequest {
	return CalendarListRequest{
		Filters: CalendarListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the CalendarListRequest.
func (c CalendarListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/calendars.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	if c.Filters.Page > 0 {
		query.Set("page", strconv.FormatInt(c.Filters.Page, 10))
	}
	if c.Filters.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(c.Filters.PageSize, 10))
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// CalendarListResponse contains the response for loading calendars.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/calendars/get-calendars-json
type CalendarListResponse struct {
	request CalendarListRequest

	// Calendars is the list of calendars.
	Calendars []Calendar `json:"calendars"`

	// Meta contains pagination metadata.
	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`
}

// HandleHTTPResponse handles the HTTP response for the CalendarListResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (c *CalendarListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list calendars")
	}

	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode list calendars response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (c *CalendarListResponse) SetRequest(req CalendarListRequest) {
	c.request = req
}

// Iterate returns the request set to the next page, if available. If there
// are no more pages, a nil request is returned.
func (c *CalendarListResponse) Iterate() *CalendarListRequest {
	if !c.Meta.Page.HasMore {
		return nil
	}
	req := c.request
	req.Filters.Page++
	return &req
}

// CalendarList retrieves calendars using the provided request and returns the
// response.
func CalendarList(
	ctx context.Context,
	engine *twapi.Engine,
	req CalendarListRequest,
) (*CalendarListResponse, error) {
	return twapi.Execute[CalendarListRequest, *CalendarListResponse](ctx, engine, req)
}
