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

func ExampleWorkflowCreate() {
	address, stop, err := startWorkflowServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	workflowRequest := projects.NewWorkflowCreateRequest("My Workflow")

	workflowResponse, err := projects.WorkflowCreate(ctx, engine, workflowRequest)
	if err != nil {
		fmt.Printf("failed to create workflow: %s", err)
	} else {
		fmt.Printf("created workflow with identifier %d\n", workflowResponse.Workflow.ID)
	}

	// Output: created workflow with identifier 12345
}

func ExampleWorkflowUpdate() {
	address, stop, err := startWorkflowServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	workflowRequest := projects.NewWorkflowUpdateRequest(12345)
	workflowRequest.Name = new("Updated Workflow Name")

	_, err = projects.WorkflowUpdate(ctx, engine, workflowRequest)
	if err != nil {
		fmt.Printf("failed to update workflow: %s", err)
	} else {
		fmt.Println("workflow updated!")
	}

	// Output: workflow updated!
}

func ExampleWorkflowDelete() {
	address, stop, err := startWorkflowServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.WorkflowDelete(ctx, engine, projects.NewWorkflowDeleteRequest(12345))
	if err != nil {
		fmt.Printf("failed to delete workflow: %s", err)
	} else {
		fmt.Println("workflow deleted!")
	}

	// Output: workflow deleted!
}

func ExampleWorkflowGet() {
	address, stop, err := startWorkflowServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	workflowResponse, err := projects.WorkflowGet(ctx, engine, projects.NewWorkflowGetRequest(12345))
	if err != nil {
		fmt.Printf("failed to retrieve workflow: %s", err)
	} else {
		fmt.Printf("retrieved workflow with identifier %d\n", workflowResponse.Workflow.ID)
	}

	// Output: retrieved workflow with identifier 12345
}

func ExampleWorkflowList() {
	address, stop, err := startWorkflowServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	workflowsRequest := projects.NewWorkflowListRequest()
	workflowsRequest.Filters.SearchTerm = "My"

	workflowsResponse, err := projects.WorkflowList(ctx, engine, workflowsRequest)
	if err != nil {
		fmt.Printf("failed to list workflows: %s", err)
	} else {
		for _, workflow := range workflowsResponse.Workflows {
			fmt.Printf("retrieved workflow with identifier %d\n", workflow.ID)
		}
	}

	// Output: retrieved workflow with identifier 12345
	// retrieved workflow with identifier 12346
}

func startWorkflowServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /projects/api/v3/workflows", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"workflow":{"id":12345,"name":"My Workflow"}}`)
	})
	mux.HandleFunc("PATCH /projects/api/v3/workflows/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{}`)
	})
	mux.HandleFunc("DELETE /projects/api/v3/workflows/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /projects/api/v3/workflows/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"workflow":{"id":12345,"name":"My Workflow"}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/workflows", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"workflows":[{"id":12345,"name":"My Workflow"},{"id":12346,"name":"Another Workflow"}]}`)
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
