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

func ExampleTaskCapacityCreate() {
	address, stop, err := startTaskCapacityServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	monday := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
	response, err := projects.TaskCapacityCreate(ctx, engine, projects.NewTaskCapacityCreateRequest(
		12345, 456, []projects.TaskCapacityDate{
			{Date: twapi.Date(monday), Minutes: 90},
			{Date: twapi.Date(monday.AddDate(0, 0, 1)), Minutes: 30},
		},
	))
	if err != nil {
		fmt.Printf("failed to create task capacity: %s", err)
	} else {
		for _, row := range response.Capacities {
			fmt.Printf("%s: %d minutes\n", row.Date, row.Minutes)
		}
	}

	// Output: 2026-10-05: 90 minutes
	// 2026-10-06: 30 minutes
}

func ExampleTaskCapacityList() {
	address, stop, err := startTaskCapacityServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	response, err := projects.TaskCapacityList(ctx, engine, projects.NewTaskCapacityListRequest(12345, 456))
	if err != nil {
		fmt.Printf("failed to list task capacity: %s", err)
	} else {
		fmt.Printf("retrieved %d rows\n", len(response.Capacities))
	}

	// Output: retrieved 2 rows
}

func startTaskCapacityServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	const rows = `{"capacities":[` +
		`{"id":1,"taskId":12345,"userId":456,"date":"2026-10-05","minutes":90,"seconds":0},` +
		`{"id":2,"taskId":12345,"userId":456,"date":"2026-10-06","minutes":30,"seconds":0}]}`

	mux := http.NewServeMux()
	mux.HandleFunc("GET /projects/api/v3/tasks/capacity", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, rows)
	})
	mux.HandleFunc("POST /projects/api/v3/tasks/{id}/capacity", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = fmt.Fprintln(w, rows)
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
