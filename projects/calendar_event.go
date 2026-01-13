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
	_ twapi.HTTPRequester = (*CalendarEventListRequest)(nil)
	_ twapi.HTTPResponser = (*CalendarEventListResponse)(nil)
)

// CalendarEventType represents the event type.
type CalendarEventType string

const (
	// CalendarEventTypeDefault represents a default event.
	CalendarEventTypeDefault CalendarEventType = "default"
	// CalendarEventTypeOutOfOffice represents an out-of-office event.
	CalendarEventTypeOutOfOffice CalendarEventType = "outOfOffice"
	// CalendarEventTypeFocusTime represents a focus time event.
	CalendarEventTypeFocusTime CalendarEventType = "focusTime"
	// CalendarEventTypeHoliday represents a holiday event.
	CalendarEventTypeHoliday CalendarEventType = "holiday"
)

// CalendarEventTransparency represents visibility in calendars.
type CalendarEventTransparency string

const (
	// CalendarEventTransparencyOpaque indicates the event is opaque (blocks time).
	CalendarEventTransparencyOpaque CalendarEventTransparency = "opaque"
	// CalendarEventTransparencyTransparent indicates the event is transparent
	// (does not block time).
	CalendarEventTransparencyTransparent CalendarEventTransparency = "transparent"
)

// CalendarEventVisibility represents visibility restrictions.
type CalendarEventVisibility string

const (
	// CalendarEventVisibilityPublic indicates the event is public.
	CalendarEventVisibilityPublic CalendarEventVisibility = "public"
	// CalendarEventVisibilityPrivate indicates the event is private.
	CalendarEventVisibilityPrivate CalendarEventVisibility = "private"
)

// CalendarEventStatus represents the event status.
type CalendarEventStatus string

const (
	// CalendarEventStatusConfirmed represents a confirmed event.
	CalendarEventStatusConfirmed CalendarEventStatus = "confirmed"
	// CalendarEventStatusTentative represents a tentative event.
	CalendarEventStatusTentative CalendarEventStatus = "tentative"
	// CalendarEventStatusCancelled represents a cancelled event.
	CalendarEventStatusCancelled CalendarEventStatus = "cancelled"
	// CalendarEventStatusDeleted represents a deleted event.
	CalendarEventStatusDeleted CalendarEventStatus = "deleted"
)

// CalendarAttendeeReminderMethod represents the reminder delivery method.
type CalendarAttendeeReminderMethod string

const (
	// CalendarAttendeeReminderEmail represents an email reminder.
	CalendarAttendeeReminderEmail CalendarAttendeeReminderMethod = "email"
	// CalendarAttendeeReminderPush represents a push notification reminder.
	CalendarAttendeeReminderPush CalendarAttendeeReminderMethod = "push"
	// CalendarAttendeeReminderSMS represents an SMS reminder.
	CalendarAttendeeReminderSMS CalendarAttendeeReminderMethod = "sms"
)

// CalendarAttendeeStatus represents an attendee's status.
type CalendarAttendeeStatus string

const (
	// CalendarAttendeeStatusNeedsAction represents an attendee who needs to take
	// action.
	CalendarAttendeeStatusNeedsAction CalendarAttendeeStatus = "needsAction"
	// CalendarAttendeeStatusTentative represents an attendee who has tentatively
	// accepted.
	CalendarAttendeeStatusTentative CalendarAttendeeStatus = "tentative"
	// CalendarAttendeeStatusAccepted represents an attendee who has accepted.
	CalendarAttendeeStatusAccepted CalendarAttendeeStatus = "accepted"
	// CalendarAttendeeStatusDeclined represents an attendee who has declined.
	CalendarAttendeeStatusDeclined CalendarAttendeeStatus = "declined"
	// CalendarAttendeeStatusDeleted represents an attendee who has been deleted.
	CalendarAttendeeStatusDeleted CalendarAttendeeStatus = "deleted"
	// CalendarAttendeeStatusPublic represents an attendee who is public.
	CalendarAttendeeStatusPublic CalendarAttendeeStatus = "public"
	// CalendarAttendeeStatusPrivate represents an attendee who is private.
	CalendarAttendeeStatusPrivate CalendarAttendeeStatus = "private"
)

// CalendarEvent contains all the information returned from an event.
type CalendarEvent struct {
	// ID is the unique identifier for the event.
	ID string `json:"id"`

	// Summary is a short description of the event.
	Summary *string `json:"summary"`

	// Description is a detailed description of the event.
	Description *string `json:"description"`

	// Organizer is the user who organized the event.
	Organizer CalendarUser `json:"organizer"`

	// EventCreator is the user who created the event. It differs from CreatedBy
	// as it may contain non-Teamwork users while the CreatedBy field may store
	// the synchronization user.
	EventCreator CalendarUser `json:"eventCreator"`

	// Start is the start date and time of the event.
	Start CalendarEventDate `json:"start"`

	// End is the end date and time of the event.
	End CalendarEventDate `json:"end"`

	// AllDay indicates whether the event is an all-day event.
	AllDay bool `json:"allDay"`

	// Location is the location of the event.
	Location *string `json:"location"`

	// Type is the type of the event.
	Type *CalendarEventType `json:"type"`

	// Recurrence is the recurrence rule for the event as specified in RFC5545.
	Recurrence *string `json:"recurrence"`

	// GuestsCanInviteOthers indicates whether guests can invite others.
	GuestsCanInviteOthers bool `json:"guestsCanInviteOthers"`

	// GuestsCanModify indicates whether guests can modify the event.
	GuestsCanModify bool `json:"guestsCanModify"`

	// GuestsCanSeeOtherGuests indicates whether guests can see other guests.
	GuestsCanSeeOtherGuests bool `json:"guestsCanSeeOtherGuests"`

	// Transparency indicates the event's transparency in calendars.
	Transparency *CalendarEventTransparency `json:"transparency"`

	// Visibility indicates the event's visibility restrictions.
	Visibility *CalendarEventVisibility `json:"visibility"`

	// VideoCallLink is the link to a video call for the event.
	VideoCallLink *string `json:"videoCallLink"`

	// CalOwnerCanEdit indicates whether the calendar owner can edit the event.
	CalOwnerCanEdit bool `json:"calOwnerCanEdit"`

	// Timeblock link the event Teamwork.com entities for timeblocking.
	Timeblock *Timeblock `json:"timeblock,omitempty"`

	// Attendees contains the list of event attendees.
	Attendees []CalendarAttendee `json:"attendees"`

	// AttendeesOmitted indicates whether some attendees are omitted from the
	// response.
	AttendeesOmitted bool `json:"attendeesOmitted"`

	// Calendar is the calendar to which the event belongs.
	Calendar *twapi.Relationship `json:"calendar"`

	// Status is the status of the event.
	Status CalendarEventStatus `json:"status"`

	// CreatedBy is the user who created the event.
	CreatedBy twapi.Relationship `json:"createdBy"`

	// CreatedAt is the timestamp when the event was created.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt is the timestamp when the event was last updated.
	UpdatedAt *time.Time `json:"updatedAt"`
}

// CalendarUser contains information returned for a calendar user.
type CalendarUser struct {
	// User is the relationship to the user. This is populated when the user
	// exists in Teamwork.com.
	User *twapi.Relationship `json:"user"`
	// Email is the email address of the user.
	Email *string `json:"email"`
	// FullName is the full name of the user.
	FullName *string `json:"fullName"`
}

// CalendarAttendeeReminder contains reminder details for an attendee.
type CalendarAttendeeReminder struct {
	// Method is the delivery method for the reminder.
	Method CalendarAttendeeReminderMethod `json:"method"`
	// Minutes is the number of minutes before the event when the reminder should
	// be sent.
	Minutes int64 `json:"minute"`
}

// CalendarAttendee contains all information about an event attendee.
type CalendarAttendee struct {
	// User is the attendee information.
	User CalendarUser `json:"user"`
	// Status is the attendee's status for the event.
	Status CalendarAttendeeStatus `json:"status"`
	// CanEdit indicates whether the attendee can edit the event.
	CanEdit bool `json:"canEdit"`
	// Reminders contains the reminders set for the attendee.
	Reminders []CalendarAttendeeReminder `json:"reminders"`
}

// CalendarEventDate represents a date/time with timezone for an event.
type CalendarEventDate struct {
	// DateTime is the date and time of the event.
	DateTime time.Time `json:"dateTime"`
	// TimeZone the time zone in which the time is specified. (Formatted as an
	// IANA Time Zone Database name, e.g. "Europe/Zurich".) For recurring events
	// this field is required and specifies the time zone in which the recurrence
	// is expanded. For single events this field is optional and indicates a
	// custom time zone for the event start/end.
	TimeZone string `json:"timeZone"`
}

// Timeblock holds details about a timeblock linked to an event.
type Timeblock struct {
	// Project is the project related to the timeblock. This is mandatory.
	Project twapi.Relationship `json:"project"`
	// Task is the task related to the timeblock.
	Task *twapi.Relationship `json:"task"`
	// Timelog is the timelog related to the timeblock.
	Timelog *twapi.Relationship `json:"timelog"`
}

// CalendarEventListRequestPath contains the path parameters for loading
// calendar events.
type CalendarEventListRequestPath struct {
	// CalendarID is the unique identifier of the calendar.
	CalendarID int64
}

// CalendarEventListRequestSideload contains the possible sideload options when
// loading calendar events.
type CalendarEventListRequestSideload string

// List of possible sideload options for CalendarEventListRequestSideload.
const (
	CalendarEventListRequestSideloadUsers     CalendarEventListRequestSideload = "users"
	CalendarEventListRequestSideloadProjects  CalendarEventListRequestSideload = "projects"
	CalendarEventListRequestSideloadTasks     CalendarEventListRequestSideload = "tasks"
	CalendarEventListRequestSideloadTasklists CalendarEventListRequestSideload = "tasklists"
	CalendarEventListRequestSideloadCompanies CalendarEventListRequestSideload = "companies"
	CalendarEventListRequestSideloadTimelogs  CalendarEventListRequestSideload = "timelogs"
)

// CalendarEventListRequestFilters contains filters for loading calendar events.
type CalendarEventListRequestFilters struct {
	// StartedAfterDate filters events that start after this date (YYYY-MM-DD
	// format).
	StartedAfterDate twapi.Date

	// EndedBeforeDate filters events that end before this date (YYYY-MM-DD
	// format).
	EndedBeforeDate twapi.Date

	// Include specifies related resources to include.
	Include []CalendarEventListRequestSideload

	// Cursor is used for cursor-based pagination.
	Cursor string

	// Limit is the maximum number of items to return.
	Limit int64
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
			Limit: 50,
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
	if !c.Filters.StartedAfterDate.IsZero() {
		query.Set("startedAfterDate", c.Filters.StartedAfterDate.String())
	}
	if !c.Filters.EndedBeforeDate.IsZero() {
		query.Set("endedBeforeDate", c.Filters.EndedBeforeDate.String())
	}
	if len(c.Filters.Include) > 0 {
		for _, include := range c.Filters.Include {
			query.Add("include", string(include))
		}
	}
	if c.Filters.Cursor != "" {
		query.Set("cursor", c.Filters.Cursor)
	}
	if c.Filters.Limit > 0 {
		query.Set("limit", strconv.FormatInt(c.Filters.Limit, 10))
	}
	// hardcoded filters to simplify API usage
	query.Set("includeMasterInstances", "false")
	query.Set("showDeletedInstances", "false")
	query.Set("includeInstances", "true")
	query.Set("includeOngoingEvents", "true")
	query.Set("includeDeletedInstances", "false")
	query.Set("includeModifiedInstances", "true")
	query.Set("includeTimelogs", "false")

	req.URL.RawQuery = query.Encode()
	return req, nil
}

// CalendarEventListResponse contains the response for loading calendar events.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/calendar-events/get-projects-api-v3-calendars-calendar-id-events-json
type CalendarEventListResponse struct {
	request CalendarEventListRequest

	// Meta contains metadata about the response, including pagination details.
	Meta struct {
		NextCursor *string `json:"nextCursor,omitempty"`
	} `json:"meta"`

	// Events contains the calendar events.
	Events []CalendarEvent `json:"events"`

	// Included contains related objects included in the response.
	Included struct {
		// Users is a map of user IDs to User objects.
		//
		// The key is the string representation of the user ID.
		Users map[string]User `json:"users,omitempty"`
		// Projects is a map of project IDs to Project objects.
		//
		// The key is the string representation of the project ID.
		Projects map[string]Project `json:"projects,omitempty"`
		// Tasks is a map of task IDs to Task objects.
		//
		// The key is the string representation of the task ID.
		Tasks map[string]Task `json:"tasks,omitempty"`
		// Tasklists is a map of tasklist IDs to Tasklist objects.
		//
		// The key is the string representation of the tasklist ID.
		Tasklists map[string]Tasklist `json:"tasklists,omitempty"`
		// Companies is a map of company IDs to Company objects.
		//
		// The key is the string representation of the company ID.
		Companies map[string]Company `json:"companies,omitempty"`
		// Timelogs is a map of timelog IDs to Timelog objects.
		//
		// The key is the string representation of the timelog ID.
		Timelogs map[string]Timelog `json:"timelogs,omitempty"`
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
	if c.Meta.NextCursor == nil {
		return nil
	}
	req := c.request
	req.Filters.Cursor = *c.Meta.NextCursor
	return &req
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
