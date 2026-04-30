package projects

import (
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
	_ twapi.HTTPRequester = (*SearchRequest)(nil)
	_ twapi.HTTPResponser = (*SearchResponse)(nil)
)

// SearchItem is the result item of the search, a feature that allows users to
// quickly find projects, tasks, files, messages, and other items across their
// workspace by entering keywords, helping them locate information and navigate
// their work efficiently from a single place.
//
// More information can be found at:
// https://support.teamwork.com/projects/using-teamwork/search-command-center
type SearchItem twapi.Relationship

// SearchRequestPath contains the path parameters for loading multiple
// searches.
type SearchRequestPath struct{}

// SearchRequestType contains the possible types to filter searches by their
// type.
type SearchRequestType string

// List of possible types for SearchRequestType.
const (
	SearchRequestTypeCalendarEvents    SearchRequestType = "calendarevents"
	SearchRequestTypeComments          SearchRequestType = "comments"
	SearchRequestTypeTaskComments      SearchRequestType = "taskcomments"
	SearchRequestTypeMilestoneComments SearchRequestType = "milestonecomments"
	SearchRequestTypeFileComments      SearchRequestType = "filecomments"
	SearchRequestTypeLinkComments      SearchRequestType = "linkcomments"
	SearchRequestTypeNotebookComments  SearchRequestType = "notebookcomments"
	SearchRequestTypeCompanies         SearchRequestType = "companies"
	SearchRequestTypeLinks             SearchRequestType = "links"
	SearchRequestTypeMessages          SearchRequestType = "messages"
	SearchRequestTypeMilestones        SearchRequestType = "milestones"
	SearchRequestTypeNotebooks         SearchRequestType = "notebooks"
	SearchRequestTypeProjects          SearchRequestType = "projects"
	SearchRequestTypeTasklists         SearchRequestType = "tasklists"
	SearchRequestTypeTasks             SearchRequestType = "tasks"
	SearchRequestTypeTeams             SearchRequestType = "teams"
	SearchRequestTypeTimelogs          SearchRequestType = "timelogs"
	SearchRequestTypeUsers             SearchRequestType = "users"
)

// SearchRequestSideload contains the possible sideload options when searching.
type SearchRequestSideload string

// List of possible sideload options for SearchRequestSideload.
const (
	SearchRequestSideloadCalendarEvents SearchRequestSideload = "calendarevents"
	SearchRequestSideloadComments       SearchRequestSideload = "comments"
	SearchRequestSideloadCompanies      SearchRequestSideload = "companies"
	SearchRequestSideloadLinks          SearchRequestSideload = "links"
	SearchRequestSideloadMessages       SearchRequestSideload = "messages"
	SearchRequestSideloadMilestones     SearchRequestSideload = "milestones"
	SearchRequestSideloadNotebooks      SearchRequestSideload = "notebooks"
	SearchRequestSideloadProjects       SearchRequestSideload = "projects"
	SearchRequestSideloadTasklists      SearchRequestSideload = "tasklists"
	SearchRequestSideloadTasks          SearchRequestSideload = "tasks"
	SearchRequestSideloadTeams          SearchRequestSideload = "teams"
	SearchRequestSideloadTimelogs       SearchRequestSideload = "timelogs"
	SearchRequestSideloadUsers          SearchRequestSideload = "users"
)

// SearchRequestFilters contains the filters for searching.
type SearchRequestFilters struct {
	// SearchTerm contains the keywords to search for. This is a mandatory field
	// and must be at least 3 characters long.
	SearchTerm string

	// Type is an optional type to filter searches by their type. By default, all
	// types are included in the results.
	Type SearchRequestType

	// ProjectID is an optional project ID to filter searches by their associated
	// project.
	ProjectID int64

	// IncludeCompletedItems is an optional flag to include completed items in the
	// search results. The default is false, which means completed items are
	// excluded from the results.
	IncludeCompletedItems *bool

	// IncludeArchivedProjects is an optional flag to include archived projects in
	// the search results. The default is false.
	IncludeArchivedProjects *bool

	// IncludeTentativeProjects is an optional flag to include tentative projects
	// in the search results. The default is false.
	IncludeTentativeProjects *bool

	// IncludeArchivedMessages is an optional flag to include archived messages in
	// the search results. The default is false.
	IncludeArchivedMessages *bool

	// UpdatedAfter is an optional timestamp to filter searches updated after the
	// specified time.
	UpdatedAfter time.Time

	// ExtendedSearch is an optional flag to enable extended search, which allows
	// searching for items updated more than 5 years ago. The default is false.
	ExtendedSearch *bool

	// Cursor is an optional cursor to retrieve the next set of results.
	Cursor string

	// Limit is an optional limit to specify the number of results to return per
	// page. The default is 50.
	Limit int64

	// Include contains additional related information to include in the response
	// as a sideload.
	Include []SearchRequestSideload
}

func (s SearchRequestFilters) apply(req *http.Request) {
	query := req.URL.Query()
	query.Set("searchTerm", s.SearchTerm)
	if s.Type != "" {
		query.Set("type", string(s.Type))
	}
	if s.ProjectID > 0 {
		query.Set("projectId", strconv.FormatInt(s.ProjectID, 10))
	}
	if s.IncludeCompletedItems != nil {
		query.Set("includeCompletedItems", strconv.FormatBool(*s.IncludeCompletedItems))
	}
	if !s.UpdatedAfter.IsZero() {
		query.Set("updatedAfter", s.UpdatedAfter.Format(time.RFC3339))
	}
	if s.ExtendedSearch != nil {
		query.Set("extendedSearch", strconv.FormatBool(*s.ExtendedSearch))
	}
	if s.Cursor != "" {
		query.Set("cursor", s.Cursor)
	}
	if s.Limit > 0 {
		query.Set("limit", strconv.FormatInt(s.Limit, 10))
	}
	if len(s.Include) > 0 {
		var include []string
		for _, sideload := range s.Include {
			include = append(include, string(sideload))
		}
		query.Set("include", strings.Join(include, ","))
	}

	// By default the API returns results ordered by updated date. To ensure we
	// get the most relevant results, we set the orderBy parameter to relevance.
	query.Set("orderBy", "relevance")

	req.URL.RawQuery = query.Encode()
}

// SearchRequest represents the request body for searching.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/search/get-projects-api-v3-search-json
type SearchRequest struct {
	// Path contains the path parameters for the request.
	Path SearchRequestPath

	// Filters contains the filters for searching.
	Filters SearchRequestFilters
}

// NewSearchRequest creates a new SearchRequest with default values.
func NewSearchRequest(searchTerm string) SearchRequest {
	return SearchRequest{
		Filters: SearchRequestFilters{
			SearchTerm: searchTerm,
			Limit:      50,
		},
	}
}

// HTTPRequest creates an HTTP request for the SearchRequest.
func (s SearchRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/search.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	s.Filters.apply(req)

	return req, nil
}

// SearchResponse contains search results matching the request filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/search/get-projects-api-v3-search-json
type SearchResponse struct {
	request SearchRequest

	Meta struct {
		PreviousCursor *string `json:"prevCursor"`
		NextCursor     *string `json:"nextCursor"`
		Limit          *int64  `json:"limit"`
	} `json:"meta"`

	// Items is the list of search results matching the request filters. Each item
	// contains a relationship to the actual item, which can be found as a
	// sideload.
	Items []SearchItem `json:"search"`

	// Included contains related objects included in the response.
	Included struct {
		// CalendarEvents contains the calendar events associated with the search
		// results.
		//
		// The key is the string representation of the calendar event ID.
		CalendarEvents map[string]CalendarEvent `json:"calendarEvents"`
		// Comments contains the comments associated with the search results.
		//
		// The key is the string representation of the comment ID.
		Comments map[string]Comment `json:"comments"`
		// Companies contains the companies associated with the search results.
		//
		// The key is the string representation of the company ID.
		Companies map[string]Company `json:"companies"`
		// Links contains the links associated with the search results.
		//
		// The key is the string representation of the link ID.
		Links map[string]Link `json:"links"`
		// Messages contains the messages associated with the search results.
		//
		// The key is the string representation of the message ID.
		Messages map[string]Message `json:"messages"`
		// Milestones contains the milestones associated with the search results.
		//
		// The key is the string representation of the milestone ID.
		Milestones map[string]Milestone `json:"milestones"`
		// Notebooks contains the notebooks associated with the search results.
		//
		// The key is the string representation of the notebook ID.
		Notebooks map[string]Notebook `json:"notebooks"`
		// Projects contains the projects associated with the search results.
		//
		// The key is the string representation of the project ID.
		Projects map[string]Project `json:"projects"`
		// Tasklists contains the tasklists associated with the search results.
		//
		// The key is the string representation of the tasklist ID.
		Tasklists map[string]Tasklist `json:"tasklists"`
		// Tasks contains the tasks associated with the search results.
		//
		// The key is the string representation of the task ID.
		Tasks map[string]Task `json:"tasks"`
		// Teams contains the teams associated with the search results.
		//
		// The key is the string representation of the team ID.
		Teams map[string]Team `json:"teams"`
		// Timelogs contains the timelogs associated with the search results.
		//
		// The key is the string representation of the timelog ID.
		Timelogs map[string]Timelog `json:"timelogs"`
		// Users contains the users associated with the search results.
		//
		// The key is the string representation of the user ID.
		Users map[string]User `json:"users"`
	} `json:"included"`
}

// HandleHTTPResponse handles the HTTP response for the SearchResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (s *SearchResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to search")
	}

	if err := json.NewDecoder(resp.Body).Decode(s); err != nil {
		return fmt.Errorf("failed to decode search response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (s *SearchResponse) SetRequest(req SearchRequest) {
	s.request = req
}

// Iterate returns the request set to the next page, if available. If there are
// no more pages, a nil request is returned.
func (s *SearchResponse) Iterate() *SearchRequest {
	if s.Meta.NextCursor == nil {
		return nil
	}
	req := s.request
	req.Filters.Cursor = *s.Meta.NextCursor
	return &req
}

// Search searches using the provided request and returns the response.
func Search(
	ctx context.Context,
	engine *twapi.Engine,
	req SearchRequest,
) (*SearchResponse, error) {
	return twapi.Execute[SearchRequest, *SearchResponse](ctx, engine, req)
}
