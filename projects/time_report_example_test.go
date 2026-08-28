package projects_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
	"github.com/teamwork/twapi-go-sdk/session"
)

func ExampleTimeReportList() {
	address, stop, err := startTimeReportServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	timeReportRequest := projects.NewTimeReportListRequest(
		projects.TimeReportTypeUser,
		twapi.Date(time.Now().AddDate(0, 0, -7)),
		twapi.Date(time.Now()),
	)
	timeReportRequest.Filters.Include = []projects.TimeReportSideload{projects.TimeReportSideloadUsers}

	timeReportResponse, err := projects.TimeReportList(ctx, engine, timeReportRequest)
	if err != nil {
		fmt.Printf("failed to retrieve time report: %s", err)
		return
	}

	for _, row := range timeReportResponse.TimeReport.Users {
		user := timeReportResponse.Included.Users[strconv.FormatInt(row.User.ID, 10)]
		fmt.Printf("user %d (%s %s) logged %d minutes\n",
			row.User.ID, user.FirstName, user.LastName, row.LoggedTime)
	}

	// Output:
	// user 12345 (Gary Meehan) logged 810 minutes
	// user 12346 (Alex Smith) logged 750 minutes
}

func ExampleTimeReportList_groupedByTask() {
	address, stop, err := startTimeReportServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	timeReportRequest := projects.NewTimeReportListRequest(
		projects.TimeReportTypeTask,
		twapi.Date(time.Now().AddDate(0, 0, -7)),
		twapi.Date(time.Now()),
	)
	timeReportRequest.Filters.Include = []projects.TimeReportSideload{projects.TimeReportSideloadTasks}
	timeReportRequest.Filters.Fields.Tasks = []projects.TaskField{projects.TaskFieldID, projects.TaskFieldName}

	timeReportResponse, err := projects.TimeReportList(ctx, engine, timeReportRequest)
	if err != nil {
		fmt.Printf("failed to retrieve time report: %s", err)
		return
	}

	for _, row := range timeReportResponse.TimeReport.Tasks {
		task := timeReportResponse.Included.Tasks[strconv.FormatInt(row.Task.ID, 10)]
		fmt.Printf("task %d (%s) logged %d minutes\n", row.Task.ID, task.Name, row.LoggedTime)
	}

	// Output:
	// task 777 (Write the release notes) logged 120 minutes
	// task 778 (Review the release notes) logged 45 minutes
}

func ExampleTimeReportTotals() {
	address, stop, err := startTimeReportServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	totalsRequest := projects.NewTimeReportTotalsRequest(
		projects.TimeReportGroupByWeek,
		twapi.Date(time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)),
		twapi.Date(time.Date(2026, time.August, 9, 0, 0, 0, 0, time.UTC)),
	)

	totalsResponse, err := projects.TimeReportTotals(ctx, engine, totalsRequest)
	if err != nil {
		fmt.Printf("failed to retrieve time report totals: %s", err)
		return
	}

	for _, period := range totalsResponse.Dates {
		fmt.Printf("%s to %s: %d minutes logged\n", period.StartDate, period.EndDate, period.LoggedTime)
	}
	fmt.Printf("total: %d minutes logged\n", totalsResponse.LoggedTime)

	// Output:
	// 2026-08-01 to 2026-08-02: 810 minutes logged
	// 2026-08-03 to 2026-08-09: 750 minutes logged
	// total: 1560 minutes logged
}

func startTimeReportServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /projects/api/v3/time/report/user", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{
			"time": {
				"users": [
					{"loggedTime": 810, "billableTime": 600, "user": {"id": 12345, "type": "users"}},
					{"loggedTime": 750, "billableTime": 500, "user": {"id": 12346, "type": "users"}}
				]
			},
			"meta": {"page": {"hasMore": false}},
			"included": {
				"users": {
					"12345": {"id": 12345, "firstName": "Gary", "lastName": "Meehan"},
					"12346": {"id": 12346, "firstName": "Alex", "lastName": "Smith"}
				}
			}
		}`)
	})

	mux.HandleFunc("GET /projects/api/v3/time/report/task", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{
			"time": {
				"tasks": [
					{"loggedTime": 120, "billableTime": 120, "task": {"id": 777, "type": "tasks"}},
					{"loggedTime": 45, "billableTime": 0, "task": {"id": 778, "type": "tasks"}}
				]
			},
			"meta": {"page": {"hasMore": false}},
			"included": {
				"tasks": {
					"777": {"id": 777, "name": "Write the release notes"},
					"778": {"id": 778, "name": "Review the release notes"}
				}
			}
		}`)
	})

	mux.HandleFunc("GET /projects/api/v3/time/report/totals", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{
			"loggedTime": 1560, "billableTime": 1200, "nonBillableTime": 360, "billedTime": 240, "estimatedTime": 0,
			"dates": [
				{"startDate": "2026-08-01", "endDate": "2026-08-02", "loggedTime": 810, "billableTime": 600,
				 "nonBillableTime": 210, "billedTime": 120, "estimatedTime": 0},
				{"startDate": "2026-08-03", "endDate": "2026-08-09", "loggedTime": 750, "billableTime": 600,
				 "nonBillableTime": 150, "billedTime": 120, "estimatedTime": 0}
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
