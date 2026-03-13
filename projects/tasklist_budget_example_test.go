package projects_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
	"github.com/teamwork/twapi-go-sdk/session"
)

func ExampleProjectBudgetTasklistBudgetList() {
	address, stop, err := startTasklistBudgetServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	req := projects.NewProjectBudgetTasklistBudgetListRequest(12345)
	req.Filters.PageSize = 1
	req.Filters.Include = []projects.ProjectBudgetTasklistBudgetListRequestSideload{
		projects.ProjectBudgetTasklistBudgetListRequestSideloadTasklists,
	}

	resp, err := projects.ProjectBudgetTasklistBudgetList(ctx, engine, req)
	if err != nil {
		fmt.Printf("failed to list tasklist budgets: %s", err)
		return
	}

	for _, budget := range resp.TasklistBudgets {
		fmt.Printf("retrieved tasklist budget with identifier %d\n", budget.ID)
	}

	// Output: retrieved tasklist budget with identifier 98765
}

func startTasklistBudgetServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /projects/api/v3/projects/budgets/{id}/tasklists/budgets", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"meta":{"page":{"hasMore":false}},"tasklistBudgets":[{"id":98765,"projectBudgetId":12345,"tasklistId":4567}],"included":{"tasklists":{"4567":{"id":4567,"name":"Engineering backlog"}}}}`)
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
