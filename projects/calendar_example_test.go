package projects_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
	"github.com/teamwork/twapi-go-sdk/session"
)

func ExampleCalendarList() {
	address, stop, err := startCalendarServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	calendarRequest := projects.NewCalendarListRequest()

	calendarResponse, err := projects.CalendarList(ctx, engine, calendarRequest)
	if err != nil {
		fmt.Printf("failed to list calendars: %s", err)
	} else {
		for _, calendar := range calendarResponse.Calendars {
			fmt.Printf("retrieved calendar with identifier %d and name %s\n", calendar.ID, calendar.Name)
		}
	}

	// Output: retrieved calendar with identifier 301 and name blocked_time
	// retrieved calendar with identifier 281 and name brandon.hansen@teamwork.com
}

func ExampleCalendarList_pagination() {
	address, stop, err := startCalendarServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	// Create an iterator to fetch all calendars across multiple pages
	iterator, err := twapi.Iterate[projects.CalendarListRequest, *projects.CalendarListResponse](
		ctx, engine, projects.NewCalendarListRequest(),
	)
	if err != nil {
		fmt.Printf("failed to create iterator: %s", err)
		return
	}

	// Iterate through all pages
	for {
		response, hasMore, err := iterator()
		if err != nil {
			fmt.Printf("failed to get calendars: %s", err)
			return
		}

		for _, calendar := range response.Calendars {
			fmt.Printf("calendar: %s (type: %s)\n", calendar.Name, calendar.Type)
		}

		if !hasMore {
			break
		}
	}

	// Output: calendar: blocked_time (type: blocked_time)
	// calendar: brandon.hansen@teamwork.com (type: google)
}

func ExampleCalendarEventList() {
	address, stop, err := startCalendarServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	eventRequest := projects.NewCalendarEventListRequest(281)
	eventRequest.Filters.StartedAfterDate = "2026-01-11"
	eventRequest.Filters.EndedBeforeDate = "2026-01-19"
	eventRequest.Filters.Include = "users,timelogs"

	eventResponse, err := projects.CalendarEventList(ctx, engine, eventRequest)
	if err != nil {
		fmt.Printf("failed to list calendar events: %s", err)
	} else {
		fmt.Printf("retrieved %d calendar events\n", len(eventResponse.Events))
		for _, event := range eventResponse.Events {
			summary := "(no summary)"
			if event.Summary != nil {
				summary = *event.Summary
			}
			fmt.Printf("event: %s (start: %s, end: %s)\n", summary, event.Start.DateTime.Format(time.RFC3339), event.End.DateTime.Format(time.RFC3339))
		}
	}

	// Output: retrieved 2 calendar events
	// event: Planning (start: 2025-06-13T10:00:00Z, end: 2025-06-13T11:00:00Z)
	// event: Development (start: 2025-06-14T12:00:00Z, end: 2025-06-15T12:00:00Z)
}

func ExampleCalendarEventList_withFilters() {
	address, stop, err := startCalendarServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	eventRequest := projects.NewCalendarEventListRequest(281)
	eventRequest.Filters.StartedAfterDate = "2026-01-11"
	eventRequest.Filters.EndedBeforeDate = "2026-01-19"

	// Enable various filters
	skipCounts := true
	eventRequest.Filters.SkipCounts = &skipCounts

	includeTimelogs := true
	eventRequest.Filters.IncludeTimelogs = &includeTimelogs

	eventRequest.Filters.Include = "users,masterInstances,timelogs"

	eventResponse, err := projects.CalendarEventList(ctx, engine, eventRequest)
	if err != nil {
		fmt.Printf("failed to list calendar events: %s", err)
	} else {
		fmt.Printf("retrieved %d calendar events with filters\n", len(eventResponse.Events))
	}

	// Output: retrieved 2 calendar events with filters
}

func startCalendarServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()

	// Calendar list endpoint
	mux.HandleFunc("GET /projects/api/v3/calendars", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{
			"calendars": [
				{
					"id": 301,
					"name": "blocked_time",
					"type": "blocked_time",
					"primary": false,
					"createdAt": "2023-10-23T17:16:50Z",
					"updatedAt": "2023-10-23T17:16:50Z"
				},
				{
					"id": 281,
					"name": "brandon.hansen@teamwork.com",
					"type": "google",
					"primary": true,
					"createdAt": "2023-10-23T13:53:44Z",
					"updatedAt": "2023-10-23T13:53:44Z"
				}
			],
			"meta": {
				"page": {
					"pageOffset": 0,
					"pageSize": 50,
					"count": 2,
					"hasMore": false
				}
			},
			"included": {}
		}`)
	})

	// Calendar events endpoint
	mux.HandleFunc("GET /projects/api/v3/calendars/{calendarId}/events", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("calendarId") != "281" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{
			"STATUS": "OK",
			"events": [
				{
					"id": "24413522",
					"iCalUID": "uid-24413522",
					"status": "confirmed",
					"htmlLink": "https://calendar.test/events/24413522",
					"createdAt": "2025-03-27T15:06:32Z",
					"updatedAt": "2025-04-03T13:50:00Z",
					"summary": "Planning",
					"description": "",
					"color": "#ff0000",
					"calendar": {"id": 281, "type": "calendar"},
					"createdBy": {"id": 1, "type": "user"},
					"organizer": {"user": {"id": 2, "type": "user"}, "email": "organizer@example.com", "fullName": "Organizer"},
					"eventCreator": {"user": {"id": 2, "type": "user"}, "email": "organizer@example.com", "fullName": "Organizer"},
					"start": {"dateTime": "2025-06-13T10:00:00Z", "timeZone": "UTC"},
					"end": {"dateTime": "2025-06-13T11:00:00Z", "timeZone": "UTC"},
					"allDay": false,
					"attendees": [],
					"attendeesOmitted": false,
					"location": "Board Room",
					"type": "default",
					"isModified": false,
					"guestsCanInviteOthers": true,
					"guestsCanModify": true,
					"guestsCanSeeOtherGuests": true,
					"calOwnerCanEdit": true
				},
				{
					"id": "24413523",
					"iCalUID": "uid-24413523",
					"status": "confirmed",
					"htmlLink": "https://calendar.test/events/24413523",
					"createdAt": "2025-03-27T15:06:32Z",
					"updatedAt": "2025-04-03T13:50:00Z",
					"summary": "Development",
					"description": "Dev work",
					"color": "#00ff00",
					"calendar": {"id": 281, "type": "calendar"},
					"createdBy": {"id": 1, "type": "user"},
					"organizer": {"user": {"id": 3, "type": "user"}, "email": "dev@example.com", "fullName": "Dev"},
					"eventCreator": {"user": {"id": 3, "type": "user"}, "email": "dev@example.com", "fullName": "Dev"},
					"start": {"dateTime": "2025-06-14T12:00:00Z", "timeZone": "UTC"},
					"end": {"dateTime": "2025-06-15T12:00:00Z", "timeZone": "UTC"},
					"allDay": false,
					"attendees": [],
					"attendeesOmitted": false,
					"location": "Main Office",
					"type": "default",
					"isModified": false,
					"guestsCanInviteOthers": true,
					"guestsCanModify": true,
					"guestsCanSeeOtherGuests": true,
					"calOwnerCanEdit": true
				}
			]
		}`)
	})

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer your_token" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			r.URL.Path = strings.TrimSuffix(r.URL.Path, ".json")
			mux.ServeHTTP(w, r)
		}),
	}

	stop := make(chan struct{})
	go func() {
		_ = server.Serve(ln)
	}()
	go func() {
		<-stop
		_ = server.Shutdown(context.Background())
	}()

	return ln.Addr().String(), func() {
		close(stop)
	}, nil
}
