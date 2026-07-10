package projects_test

import (
	"context"
	"testing"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestTimeReportList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	newRequest := func(reportType projects.TimeReportType) projects.TimeReportListRequest {
		req := projects.NewTimeReportListRequest(
			reportType,
			twapi.Date(time.Now().AddDate(0, 0, -7)),
			twapi.Date(time.Now()),
		)
		req.Filters.Include = []projects.TimeReportSideload{
			projects.TimeReportSideloadUsers,
			projects.TimeReportSideloadProjects,
		}
		return req
	}

	tests := []struct {
		name  string
		input projects.TimeReportListRequest
	}{{
		name:  "grouped by user",
		input: newRequest(projects.TimeReportTypeUser),
	}, {
		name:  "grouped by project",
		input: newRequest(projects.TimeReportTypeProject),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.TimeReportList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
