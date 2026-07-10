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
