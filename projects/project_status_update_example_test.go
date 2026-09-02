//nolint:lll
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

func ExampleProjectStatusUpdateList() {
	address, stop, err := startProjectStatusUpdateServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	req := projects.NewProjectStatusUpdateListRequest()
	req.Filters.ProjectIDs = []int64{12345}
	req.Filters.Include = []projects.ProjectStatusUpdateListRequestSideload{
		projects.ProjectStatusUpdateListRequestSideloadCreatedBy,
	}

	resp, err := projects.ProjectStatusUpdateList(ctx, engine, req)
	if err != nil {
		fmt.Printf("failed to list project updates: %s", err)
		return
	}

	for _, update := range resp.ProjectUpdates {
		fmt.Printf("project update %d reports health %q\n", update.ID, update.HealthLabel)
	}

	// Output: project update 777 reports health "Needs Attention"
}

func startProjectStatusUpdateServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /projects/api/v3/projects/updates", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("projectIds") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"meta":{"page":{"hasMore":false}},"projectUpdates":[{"id":777,"text":"**Blocked** on the vendor contract.","health":1,"healthLabel":"Needs Attention","color":"#F44336","projectId":12345,"createdBy":98765,"isActive":true}],"included":{"users":{"98765":{"id":98765,"firstName":"John","lastName":"Doe"}}}}`)
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
