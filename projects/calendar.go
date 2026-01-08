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

// HexColor defines a hexadecimal color (e.g., "#ff0000").
type HexColor string

// CalendarAttendeeReminderMethod represents the reminder delivery method.
type CalendarAttendeeReminderMethod string

// CalendarAttendeeStatus represents an attendee's status.
type CalendarAttendeeStatus string

// CalendarEventStatus represents the event status.
type CalendarEventStatus string

// CalendarEventType represents the event type.
type CalendarEventType string

// CalendarEventTransparency represents visibility in calendars.
type CalendarEventTransparency string

// CalendarEventVisibility represents visibility restrictions.
type CalendarEventVisibility string

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

// CalendarUser contains information returned for a calendar user.
type CalendarUser struct {
	User     *twapi.Relationship `json:"user"`
	Email    *string             `json:"email"`
	FullName *string             `json:"fullName"`
}

// CalendarAttendeeReminder contains reminder details for an attendee.
type CalendarAttendeeReminder struct {
	Method  CalendarAttendeeReminderMethod `json:"method"`
	Minutes int64                          `json:"minute"`
}

// CalendarAttendeeReminders is a collection of attendee reminders.
type CalendarAttendeeReminders []CalendarAttendeeReminder

// CalendarAttendee contains all information about an event attendee.
type CalendarAttendee struct {
	User      CalendarUser               `json:"user"`
	Status    CalendarAttendeeStatus     `json:"status"`
	CanEdit   bool                       `json:"canEdit"`
	IsSelf    bool                       `json:"isSelf"`
	Reminders []CalendarAttendeeReminder `json:"reminders"`
}

// CalendarEventDate represents a date/time with timezone for an event.
type CalendarEventDate struct {
	DateTime time.Time `json:"dateTime"`
	TimeZone string    `json:"timeZone"`
}

// TimeblockSequence holds ordering information for a timeblock.
type TimeblockSequence struct {
	Position int64 `json:"position"`
	Total    int64 `json:"total"`
}

// Timeblock holds details about a timeblock linked to an event.
type Timeblock struct {
	Project  twapi.Relationship  `json:"project"`
	Task     *twapi.Relationship `json:"task"`
	Sequence *TimeblockSequence  `json:"sequence,omitempty"`
	Timelog  *twapi.Relationship `json:"timelog"`
}

// CalendarEvent contains all the information returned from an event.
type CalendarEvent struct {
	ID                      string                     `json:"id"`
	ICalUID                 string                     `json:"iCalUID"`
	Status                  CalendarEventStatus        `json:"status"`
	HTMLLink                *string                    `json:"htmlLink"`
	CreatedAt               time.Time                  `json:"createdAt"`
	UpdatedAt               *time.Time                 `json:"updatedAt"`
	Summary                 *string                    `json:"summary"`
	Description             *string                    `json:"description"`
	Color                   *HexColor                  `json:"color"`
	Calendar                *twapi.Relationship        `json:"calendar"`
	CreatedBy               twapi.Relationship         `json:"createdBy"`
	Organizer               CalendarUser               `json:"organizer"`
	EventCreator            CalendarUser               `json:"eventCreator"`
	Start                   CalendarEventDate          `json:"start"`
	End                     CalendarEventDate          `json:"end"`
	AllDay                  bool                       `json:"allDay"`
	Attendees               []CalendarAttendee         `json:"attendees"`
	AttendeesOmitted        bool                       `json:"attendeesOmitted"`
	Location                *string                    `json:"location"`
	Type                    *CalendarEventType         `json:"type"`
	IsModified              *bool                      `json:"isModified"`
	Position                *int64                     `json:"position"`
	RecurringEventID        *int64                     `json:"recurringEventId"`
	OriginalStartTime       *CalendarEventDate         `json:"originalStartTime"`
	Recurrence              *string                    `json:"recurrence"`
	GuestsCanInviteOthers   bool                       `json:"guestsCanInviteOthers"`
	GuestsCanModify         bool                       `json:"guestsCanModify"`
	GuestsCanSeeOtherGuests bool                       `json:"guestsCanSeeOtherGuests"`
	Transparency            *CalendarEventTransparency `json:"transparency"`
	Visibility              *CalendarEventVisibility   `json:"visibility"`
	VideoCallLink           *string                    `json:"videoCallLink"`
	CalOwnerCanEdit         bool                       `json:"calOwnerCanEdit"`
	Timeblock               *Timeblock                 `json:"timeblock,omitempty"`
	ExDate                  *string                    `json:"exDate,omitempty"`
	Timelog                 *twapi.Relationship        `json:"timelog,omitempty"`
}

// SlimCalendarEvent contains a reduced set of event fields.
type SlimCalendarEvent struct {
	ID               string             `json:"id"`
	Start            CalendarEventDate  `json:"start"`
	End              CalendarEventDate  `json:"end"`
	HoursPerDay      float64            `json:"hoursPerDay,omitempty"`
	TotalTime        float64            `json:"totalTime,omitempty"`
	AllDay           bool               `json:"allDay"`
	Attendee         twapi.Relationship `json:"attendee"`
	CalendarID       *int64             `json:"calendarId"`
	CalendarSyncName string             `json:"syncName,omitempty"`
	EventTypeName    string             `json:"eventTypeName,omitempty"`
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

// CalendarEventListRequestSideload contains the possible sideload options when
// loading calendar events.
type CalendarEventListRequestSideload string

// List of possible sideload options for CalendarEventListRequestSideload.
//
//nolint:lll
const (
	CalendarEventListRequestSideloadUsers                        CalendarEventListRequestSideload = "users"
	CalendarEventListRequestSideloadMasterInstances              CalendarEventListRequestSideload = "masterInstances"
	CalendarEventListRequestSideloadTimeblocks                   CalendarEventListRequestSideload = "timeblocks"
	CalendarEventListRequestSideloadTimeblocksMasterInstances    CalendarEventListRequestSideload = "timeblocks.masterInstances"
	CalendarEventListRequestSideloadTimeblockProjects            CalendarEventListRequestSideload = "timeblocks.projects"
	CalendarEventListRequestSideloadTimeblockProjectsPermissions CalendarEventListRequestSideload = "timeblocks.projects.permissions"
	CalendarEventListRequestSideloadTimeblocksTasks              CalendarEventListRequestSideload = "timeblocks.tasks"
	CalendarEventListRequestSideloadTimeblockTasksTasklists      CalendarEventListRequestSideload = "timeblocks.tasks.tasklists"
	CalendarEventListRequestSideloadTimeblockProjectsCompanies   CalendarEventListRequestSideload = "timeblocks.projects.companies"
	CalendarEventListRequestSideloadTimeblockTimelogs            CalendarEventListRequestSideload = "timeblocks.timelogs"
	CalendarEventListRequestSideloadTimeblockTimelogsTags        CalendarEventListRequestSideload = "timeblocks.timelogs.tags"
	CalendarEventListRequestSideloadTimelogs                     CalendarEventListRequestSideload = "timelogs"
	CalendarEventListRequestSideloadTimelogsTags                 CalendarEventListRequestSideload = "timelogs.tags"
	CalendarEventListRequestSideloadTimelogsTasks                CalendarEventListRequestSideload = "timelogs.tasks"
	CalendarEventListRequestSideloadTimelogsTasksTasklists       CalendarEventListRequestSideload = "timelogs.tasks.tasklists"
	CalendarEventListRequestSideloadTimelogsProjects             CalendarEventListRequestSideload = "timelogs.projects"
	CalendarEventListRequestSideloadTimelogsProjectsPermissions  CalendarEventListRequestSideload = "timelogs.projects.prermissions"
	CalendarEventListRequestSideloadTimelogsProjectsCompanies    CalendarEventListRequestSideload = "timelogs.projects.companies"
)

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
	// Use CalendarEventListRequestSideload constants to build this value.
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

	// Events contains the calendar events.
	Events []CalendarEvent `json:"events"`

	// SlimEvents contains reduced event data when requested.
	SlimEvents []SlimCalendarEvent `json:"slimEvents,omitempty"`

	Included struct {
		Users           map[string]User          `json:"users,omitempty"`
		MasterInstances map[string]CalendarEvent `json:"masterInstances,omitempty"`
	} `json:"included"`
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
