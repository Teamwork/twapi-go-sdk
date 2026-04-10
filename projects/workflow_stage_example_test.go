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

func ExampleWorkflowStageCreate() {
	address, stop, err := startWorkflowStageServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	stageRequest := projects.NewWorkflowStageCreateRequest(777, "In Progress")

	stageResponse, err := projects.WorkflowStageCreate(ctx, engine, stageRequest)
	if err != nil {
		fmt.Printf("failed to create workflow stage: %s", err)
	} else {
		fmt.Printf("created workflow stage with identifier %d\n", stageResponse.Stage.ID)
	}

	// Output: created workflow stage with identifier 12345
}

func ExampleWorkflowStageUpdate() {
	address, stop, err := startWorkflowStageServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	stageRequest := projects.NewWorkflowStageUpdateRequest(777, 12345)
	stageRequest.Name = new("Updated Stage Name")

	_, err = projects.WorkflowStageUpdate(ctx, engine, stageRequest)
	if err != nil {
		fmt.Printf("failed to update workflow stage: %s", err)
	} else {
		fmt.Println("workflow stage updated!")
	}

	// Output: workflow stage updated!
}

func ExampleWorkflowStageDelete() {
	address, stop, err := startWorkflowStageServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.WorkflowStageDelete(ctx, engine, projects.NewWorkflowStageDeleteRequest(777, 12345))
	if err != nil {
		fmt.Printf("failed to delete workflow stage: %s", err)
	} else {
		fmt.Println("workflow stage deleted!")
	}

	// Output: workflow stage deleted!
}

func ExampleWorkflowStageGet() {
	address, stop, err := startWorkflowStageServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	stageResponse, err := projects.WorkflowStageGet(ctx, engine, projects.NewWorkflowStageGetRequest(777, 12345))
	if err != nil {
		fmt.Printf("failed to retrieve workflow stage: %s", err)
	} else {
		fmt.Printf("retrieved workflow stage with identifier %d\n", stageResponse.Stage.ID)
	}

	// Output: retrieved workflow stage with identifier 12345
}

func ExampleWorkflowStageList() {
	address, stop, err := startWorkflowStageServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	stagesResponse, err := projects.WorkflowStageList(ctx, engine, projects.NewWorkflowStageListRequest(777))
	if err != nil {
		fmt.Printf("failed to list workflow stages: %s", err)
	} else {
		for _, stage := range stagesResponse.Stages {
			fmt.Printf("retrieved workflow stage with identifier %d\n", stage.ID)
		}
	}

	// Output: retrieved workflow stage with identifier 12345
	// retrieved workflow stage with identifier 12346
}

func startWorkflowStageServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /projects/api/v3/workflows/{workflowId}/stages", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		if r.PathValue("workflowId") != "777" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"stage":{"id":12345,"name":"In Progress"}}`)
	})
	mux.HandleFunc("PATCH /projects/api/v3/workflows/{workflowId}/stages/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Content-Type") != "application/json" {
				http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
				return
			}
			if r.PathValue("workflowId") != "777" || r.PathValue("id") != "12345" {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintln(w, `{}`)
		},
	)
	mux.HandleFunc("DELETE /projects/api/v3/workflows/{workflowId}/stages/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			if r.PathValue("workflowId") != "777" || r.PathValue("id") != "12345" {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		},
	)
	mux.HandleFunc("GET /projects/api/v3/workflows/{workflowId}/stages/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			if r.PathValue("workflowId") != "777" || r.PathValue("id") != "12345" {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintln(w, `{"stage":{"id":12345,"name":"In Progress"}}`)
		},
	)
	mux.HandleFunc("GET /projects/api/v3/workflows/{workflowId}/stages", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("workflowId") != "777" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"stages":[{"id":12345,"name":"In Progress"},{"id":12346,"name":"Done"}]}`)
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
