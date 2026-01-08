package projects

import (
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
	_ twapi.HTTPRequester = (*CalendarEventListRequest)(nil)
	_ twapi.HTTPResponser = (*CalendarEventListResponse)(nil)
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
	Type string `json:"type"`

	// Primary indicates whether this is the primary calendar for the user.
	Primary bool `json:"primary"`

	// CreatedAt is the date and time when the calendar was created.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt is the date and time when the calendar was last updated.
	UpdatedAt time.Time `json:"updatedAt"`
}

// CalendarEvent represents an event (task) from a calendar. Note that the API
// returns calendar events in a task-like format.
type CalendarEvent struct {
	// ID is the unique identifier of the event.
	ID int64 `json:"id"`

	// Name is the name/title of the event.
	Name string `json:"name"`

	// Description is the description of the event.
	Description string `json:"description"`

	// DescriptionContentType is the content type of the description.
	DescriptionContentType string `json:"descriptionContentType"`

	// Priority is the priority of the event.
	Priority string `json:"priority"`

	// Status is the status of the event.
	Status string `json:"status"`

	// StartDate is the start date of the event in YYYYMMDD format.
	StartDate string `json:"startDate"`

	// DueDate is the due date of the event in YYYYMMDD format.
	DueDate string `json:"dueDate"`

	// DateCreated is when the event was created.
	DateCreated time.Time `json:"dateCreated"`

	// DateChanged is when the event was changed.
	DateChanged time.Time `json:"dateChanged"`

	// DateLastModified is when the event was last modified.
	DateLastModified time.Time `json:"dateLastModified"`

	// ProjectID is the ID of the project this event belongs to.
	ProjectID int64 `json:"projectId"`

	// TaskListID is the ID of the task list this event belongs to.
	TaskListID int64 `json:"taskListId"`

	// CreatedBy contains information about the user who created the event.
	CreatedBy *EventUser `json:"createdBy,omitempty"`

	// UpdatedBy contains information about the user who last updated the event.
	UpdatedBy *EventUser `json:"updatedBy,omitempty"`

	// AssignedTo contains information about users assigned to the event.
	AssignedTo []EventUser `json:"assignedTo,omitempty"`

	// AssignedToTeams contains information about teams assigned to the event.
	AssignedToTeams []EventTeam `json:"assignedToTeams,omitempty"`

	// Tags contains tags associated with the event.
	Tags []EventTag `json:"tags,omitempty"`

	// Progress is the completion progress of the event (0-100).
	Progress int64 `json:"progress"`

	// NumComments is the number of comments on the event.
	NumComments int64 `json:"numComments"`

	// NumAttachments is the number of attachments on the event.
	NumAttachments int64 `json:"numAttachments"`

	// IsPrivate indicates if the event is private.
	IsPrivate bool `json:"isPrivate"`

	// CanEdit indicates if the current user can edit the event.
	CanEdit bool `json:"canEdit"`

	// CanComplete indicates if the current user can complete the event.
	CanComplete bool `json:"canComplete"`
}

// EventUser represents a user associated with a calendar event.
type EventUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	AvatarURL string `json:"avatarUrl"`
}

// EventTeam represents a team assigned to a calendar event.
type EventTeam struct {
	TeamID        int64  `json:"teamId"`
	TeamName      string `json:"teamName"`
	TeamLogo      string `json:"teamLogo"`
	TeamLogoIcon  string `json:"teamLogoIcon"`
	TeamLogoColor string `json:"teamLogoColor"`
}

// EventTag represents a tag associated with a calendar event.
type EventTag struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	ProjectID int64  `json:"projectId"`
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
// https://apidocs.teamwork.com/docs/teamwork/v3/calendars/get-projects-api-v3-calendars-json
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
// https://apidocs.teamwork.com/docs/teamwork/v3/calendars/get-projects-api-v3-calendars-json
type CalendarListResponse struct {
	request CalendarListRequest

	// Calendars is the list of calendars.
	Calendars []Calendar `json:"calendars"`

	// Meta contains pagination metadata.
	Meta struct {
		Page struct {
			PageOffset int64 `json:"pageOffset"`
			PageSize   int64 `json:"pageSize"`
			Count      int64 `json:"count"`
			HasMore    bool  `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`

	// Included contains any included related resources.
	Included map[string]interface{} `json:"included"`
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

// CalendarEventListRequestPath contains the path parameters for loading
// calendar events.
type CalendarEventListRequestPath struct {
	// CalendarID is the unique identifier of the calendar.
	CalendarID int64
}

// CalendarEventListRequestFilters contains filters for loading calendar events.
type CalendarEventListRequestFilters struct {
	// StartedAfterDate filters events that start after this date (YYYY-MM-DD format).
	StartedAfterDate string

	// EndedBeforeDate filters events that end before this date (YYYY-MM-DD format).
	EndedBeforeDate string

	// Include specifies related resources to include (comma-separated).
	// e.g., "users,masterInstances,timelogs,timelogs.tags,timelogs.projects.permissions,timelogs.projects"
	Include string

	// IncludeMasterInstances indicates whether to include master instances.
	IncludeMasterInstances *bool

	// ShowDeletedInstances indicates whether to show deleted instances.
	ShowDeletedInstances *bool

	// IncludeInstances indicates whether to include instances.
	IncludeInstances *bool

	// IncludeOngoingEvents indicates whether to include ongoing events.
	IncludeOngoingEvents *bool

	// IncludeDeletedInstances indicates whether to include deleted instances.
	IncludeDeletedInstances *bool

	// IncludeModifiedInstances indicates whether to include modified instances.
	IncludeModifiedInstances *bool

	// IncludeTimelogs indicates whether to include time logs.
	IncludeTimelogs *bool

	// SkipCounts indicates whether to skip counts in the response.
	SkipCounts *bool

	// Cursor is used for cursor-based pagination.
	Cursor string

	// Limit is the maximum number of items to return.
	Limit int64

	// PageSize is the number of items per page.
	PageSize int64
}

// CalendarEventListRequest represents the request for loading calendar events.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/calendar-events/get-projects-api-v3-calendars-calendar-id-events-json
type CalendarEventListRequest struct {
	// Path contains the path parameters for the request.
	Path CalendarEventListRequestPath

	// Filters contains the filters for loading calendar events.
	Filters CalendarEventListRequestFilters
}

// NewCalendarEventListRequest creates a new CalendarEventListRequest with
// default values for the specified calendar ID.
func NewCalendarEventListRequest(calendarID int64) CalendarEventListRequest {
	return CalendarEventListRequest{
		Path: CalendarEventListRequestPath{
			CalendarID: calendarID,
		},
		Filters: CalendarEventListRequestFilters{
			Limit:    100,
			PageSize: 100,
		},
	}
}

// HTTPRequest creates an HTTP request for the CalendarEventListRequest.
func (c CalendarEventListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/api/v3/calendars/%d/events.json", server, c.Path.CalendarID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()

	if c.Filters.StartedAfterDate != "" {
		query.Set("startedAfterDate", c.Filters.StartedAfterDate)
	}
	if c.Filters.EndedBeforeDate != "" {
		query.Set("endedBeforeDate", c.Filters.EndedBeforeDate)
	}
	if c.Filters.Include != "" {
		query.Set("include", c.Filters.Include)
	}
	if c.Filters.IncludeMasterInstances != nil {
		query.Set("includeMasterInstances", strconv.FormatBool(*c.Filters.IncludeMasterInstances))
	}
	if c.Filters.ShowDeletedInstances != nil {
		query.Set("showDeletedInstances", strconv.FormatBool(*c.Filters.ShowDeletedInstances))
	}
	if c.Filters.IncludeInstances != nil {
		query.Set("includeInstances", strconv.FormatBool(*c.Filters.IncludeInstances))
	}
	if c.Filters.IncludeOngoingEvents != nil {
		query.Set("includeOngoingEvents", strconv.FormatBool(*c.Filters.IncludeOngoingEvents))
	}
	if c.Filters.IncludeDeletedInstances != nil {
		query.Set("includeDeletedInstances", strconv.FormatBool(*c.Filters.IncludeDeletedInstances))
	}
	if c.Filters.IncludeModifiedInstances != nil {
		query.Set("includeModifiedInstances", strconv.FormatBool(*c.Filters.IncludeModifiedInstances))
	}
	if c.Filters.IncludeTimelogs != nil {
		query.Set("includeTimelogs", strconv.FormatBool(*c.Filters.IncludeTimelogs))
	}
	if c.Filters.SkipCounts != nil {
		query.Set("skipCounts", strconv.FormatBool(*c.Filters.SkipCounts))
	}
	if c.Filters.Cursor != "" {
		query.Set("cursor", c.Filters.Cursor)
	}
	if c.Filters.Limit > 0 {
		query.Set("limit", strconv.FormatInt(c.Filters.Limit, 10))
	}
	if c.Filters.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(c.Filters.PageSize, 10))
	}

	req.URL.RawQuery = query.Encode()

	return req, nil
}

// CalendarEventListResponse contains the response for loading calendar events.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/calendar-events/get-projects-api-v3-calendars-calendar-id-events-json
type CalendarEventListResponse struct {
	request CalendarEventListRequest

	// STATUS indicates the status of the response.
	STATUS string `json:"STATUS"`

	// Tasks contains the calendar events (returned as tasks).
	Tasks []CalendarEvent `json:"tasks"`
}

// HandleHTTPResponse handles the HTTP response for the CalendarEventListResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (c *CalendarEventListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list calendar events")
	}

	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return fmt.Errorf("failed to decode list calendar events response: %w", err)
	}

	if c.STATUS != "OK" {
		return fmt.Errorf("calendar events API returned non-OK status: %s", c.STATUS)
	}

	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (c *CalendarEventListResponse) SetRequest(req CalendarEventListRequest) {
	c.request = req
}

// Iterate returns the request set to the next page, if available. Currently,
// the calendar events API doesn't provide clear pagination metadata, so this
// returns nil. If cursor-based pagination is needed, implement based on the
// cursor field.
func (c *CalendarEventListResponse) Iterate() *CalendarEventListRequest {
	// The API doesn't seem to provide pagination metadata in the same way as other endpoints.
	// If pagination is needed, it should be implemented using the cursor field.
	return nil
}

// CalendarEventList retrieves calendar events using the provided request and
// returns the response.
func CalendarEventList(
	ctx context.Context,
	engine *twapi.Engine,
	req CalendarEventListRequest,
) (*CalendarEventListResponse, error) {
	return twapi.Execute[CalendarEventListRequest, *CalendarEventListResponse](ctx, engine, req)
}
