package projects_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestCalendarCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	// Only one blocked time calendar is allowed per user, so a leftover from an
	// interrupted run makes the "all fields" case fail until it is removed.
	deleteBlockedTimeCalendar(t)

	tests := []struct {
		name  string
		input projects.CalendarCreateRequest
	}{{
		name:  "only required fields",
		input: projects.NewCalendarCreateRequest(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
	}, {
		name: "all fields",
		input: projects.CalendarCreateRequest{
			Name: "blocked_time", // blocked time calendar must have this name
			Type: new(projects.CalendarTypeBlockedTime),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			calendarResponse, err := projects.CalendarCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx = context.Background() // t.Context is always canceled in cleanup
				_, err := projects.CalendarDelete(ctx, engine, projects.NewCalendarDeleteRequest(calendarResponse.Calendar.ID))
				if err != nil {
					t.Errorf("failed to delete calendar after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if calendarResponse.Calendar.ID == 0 {
				t.Error("expected a valid calendar ID but got 0")
			}
		})
	}
}

// deleteBlockedTimeCalendar removes the user's blocked time calendar if one
// exists, so tests that create one start from a clean slate.
func deleteBlockedTimeCalendar(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	for request := projects.NewCalendarListRequest(); ; {
		response, err := projects.CalendarList(ctx, engine, request)
		if err != nil {
			t.Fatalf("failed to list calendars: %s", err)
		}
		for _, calendar := range response.Calendars {
			if calendar.Type != projects.CalendarTypeBlockedTime {
				continue
			}
			if _, err := projects.CalendarDelete(ctx, engine, projects.NewCalendarDeleteRequest(calendar.ID)); err != nil {
				t.Fatalf("failed to delete leftover blocked time calendar: %s", err)
			}
		}
		next := response.Iterate()
		if next == nil {
			return
		}
		request = *next
	}
}

func TestCalendarDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	calendarID, _, err := createCalendar(t)
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.CalendarDelete(ctx, engine, projects.NewCalendarDeleteRequest(calendarID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestCalendarList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	_, calendarCleanup, err := createCalendar(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(calendarCleanup)

	tests := []struct {
		name          string
		input         projects.CalendarListRequest
		expectedError bool
	}{{
		name: "all calendars",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.CalendarList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
