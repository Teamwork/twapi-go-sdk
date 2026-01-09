package projects_test

import (
	"context"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestCalendarEventList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	calendarID, cleanup, err := createCalendar(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)

	tests := []struct {
		name          string
		input         projects.CalendarEventListRequest
		expectedError bool
	}{{
		name: "all calendar events",
		input: projects.CalendarEventListRequest{
			Path: projects.CalendarEventListRequestPath{
				CalendarID: calendarID,
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.CalendarEventList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
