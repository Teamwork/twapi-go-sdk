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

func ExampleJobRoleCreate() {
	address, stop, err := startJobRoleServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	jobRoleRequest := projects.NewJobRoleCreateRequest("Project Manager")

	jobRoleResponse, err := projects.JobRoleCreate(ctx, engine, jobRoleRequest)
	if err != nil {
		fmt.Printf("failed to create job role: %s", err)
	} else {
		fmt.Printf("created job role with identifier %d\n", jobRoleResponse.JobRole.ID)
	}

	// Output: created job role with identifier 12345
}

func ExampleJobRoleUpdate() {
	address, stop, err := startJobRoleServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	jobRoleRequest := projects.NewJobRoleUpdateRequest(12345)
	jobRoleRequest.Name = twapi.Ptr("Senior Project Manager")

	_, err = projects.JobRoleUpdate(ctx, engine, jobRoleRequest)
	if err != nil {
		fmt.Printf("failed to update job role: %s", err)
	} else {
		fmt.Println("job role updated!")
	}

	// Output: job role updated!
}

func ExampleJobRoleDelete() {
	address, stop, err := startJobRoleServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.JobRoleDelete(ctx, engine, projects.NewJobRoleDeleteRequest(12345))
	if err != nil {
		fmt.Printf("failed to delete job role: %s", err)
	} else {
		fmt.Println("job role deleted!")
	}

	// Output: job role deleted!
}

func ExampleJobRoleGet() {
	address, stop, err := startJobRoleServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	jobRoleResponse, err := projects.JobRoleGet(ctx, engine, projects.NewJobRoleGetRequest(12345))
	if err != nil {
		fmt.Printf("failed to retrieve job role: %s", err)
	} else {
		fmt.Printf("retrieved job role with identifier %d\n", jobRoleResponse.JobRole.ID)
	}

	// Output: retrieved job role with identifier 12345
}

func ExampleJobRoleList() {
	address, stop, err := startJobRoleServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	jobRolesRequest := projects.NewJobRoleListRequest()
	jobRolesRequest.Filters.SearchTerm = "Project Manager"

	jobRolesResponse, err := projects.JobRoleList(ctx, engine, jobRolesRequest)
	if err != nil {
		fmt.Printf("failed to list job roles: %s", err)
	} else {
		for _, jobRole := range jobRolesResponse.JobRoles {
			fmt.Printf("retrieved job role with identifier %d\n", jobRole.ID)
		}
	}

	// Output: retrieved job role with identifier 12345
	// retrieved job role with identifier 12346
}

func startJobRoleServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /projects/api/v3/jobroles", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"jobrole":{"id":12345}}`)
	})
	mux.HandleFunc("PATCH /projects/api/v3/jobroles/{id}", func(w http.ResponseWriter, r *http.Request) {
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
		_, _ = fmt.Fprintln(w, `{"jobrole":{"id":12345}}`)
	})
	mux.HandleFunc("DELETE /projects/api/v3/jobroles/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /projects/api/v3/jobroles/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"jobrole":{"id":12345}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/jobroles", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"jobroles":[{"id":12345},{"id":12346}]}`)
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
