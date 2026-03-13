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

func ExampleProjectBudgetList() {
	address, stop, err := startProjectBudgetServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	req := projects.NewProjectBudgetListRequest()
	req.Filters.ProjectIDs = []int64{1215814}
	req.Filters.Status = projects.ProjectBudgetStatusUpcoming
	req.Filters.Limit = 1
	req.Filters.PageSize = 1

	resp, err := projects.ProjectBudgetList(ctx, engine, req)
	if err != nil {
		fmt.Printf("failed to list project budgets: %s", err)
		return
	}

	for _, budget := range resp.Budgets {
		fmt.Printf("retrieved project budget with identifier %d\n", budget.ID)
	}

	// Output: retrieved project budget with identifier 431426
}

func startProjectBudgetServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /projects/api/v3/projects/budgets", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("projectIds") != "1215814" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"budgets":[{"id":431426,"projectId":1215814,"status":"ACTIVE","type":"FINANCIAL","capacityUsed":40417,"capacity":100000}],"meta":{"page":{"pageOffset":0,"pageSize":1,"count":1,"hasMore":false}},"included":{}}`)
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
