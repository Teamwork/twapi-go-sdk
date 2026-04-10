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

func ExampleWorkflowStageTaskMove() {
	address, stop, err := startWorkflowStageTaskServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.WorkflowStageTaskMove(ctx, engine, projects.NewWorkflowStageTaskMoveRequest(123, 456, 789))
	if err != nil {
		fmt.Printf("failed to move workflow stage task: %s", err)
	} else {
		fmt.Println("moved workflow stage task")
	}

	// Output: moved workflow stage task
}

func startWorkflowStageTaskServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /projects/api/v3/tasks/{taskId}/workflows/{workflowId}",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Content-Type") != "application/json" {
				http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
				return
			}
			if r.PathValue("taskId") != "789" {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			if r.PathValue("workflowId") != "123" {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		},
	)

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
