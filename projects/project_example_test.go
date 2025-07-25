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

func ExampleProjectCreate() {
	address, stop, err := startProjectServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	projectRequest := projects.NewProjectCreateRequest("New Project")
	projectRequest.Description = twapi.Ptr("This is a new project created via the API.")
	projectRequest.StartAt = twapi.Ptr(projects.LegacyDate(time.Now().AddDate(0, 0, 1)))
	projectRequest.EndAt = twapi.Ptr(projects.LegacyDate(time.Now().AddDate(0, 0, 30)))
	projectRequest.CompanyID = 12345
	projectRequest.OwnerID = twapi.Ptr(int64(67890))
	projectRequest.TagIDs = []int64{11111, 22222}

	projectResponse, err := projects.ProjectCreate(ctx, engine, projectRequest)
	if err != nil {
		fmt.Printf("failed to create project: %s", err)
	} else {
		fmt.Printf("created project with identifier %d\n", projectResponse.ID)
	}

	// Output: created project with identifier 12345
}

func ExampleProjectUpdate() {
	address, stop, err := startProjectServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	projectRequest := projects.NewProjectUpdateRequest(12345)
	projectRequest.Description = twapi.Ptr("This is an updated description.")

	_, err = projects.ProjectUpdate(ctx, engine, projectRequest)
	if err != nil {
		fmt.Printf("failed to update project: %s", err)
	} else {
		fmt.Println("project updated!")
	}

	// Output: project updated!
}

func ExampleProjectDelete() {
	address, stop, err := startProjectServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.ProjectDelete(ctx, engine, projects.NewProjectDeleteRequest(12345))
	if err != nil {
		fmt.Printf("failed to delete project: %s", err)
	} else {
		fmt.Println("project deleted!")
	}

	// Output: project deleted!
}

func ExampleProjectGet() {
	address, stop, err := startProjectServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	projectResponse, err := projects.ProjectGet(ctx, engine, projects.NewProjectGetRequest(12345))
	if err != nil {
		fmt.Printf("failed to retrieve project: %s", err)
	} else {
		fmt.Printf("retrieved project with identifier %d\n", projectResponse.Project.ID)
	}

	// Output: retrieved project with identifier 12345
}

func ExampleProjectList() {
	address, stop, err := startProjectServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	projectsRequest := projects.NewProjectListRequest()
	projectsRequest.Filters.SearchTerm = "Example"

	projectsResponse, err := projects.ProjectList(ctx, engine, projectsRequest)
	if err != nil {
		fmt.Printf("failed to list projects: %s", err)
	} else {
		for _, project := range projectsResponse.Projects {
			fmt.Printf("retrieved project with identifier %d\n", project.ID)
		}
	}

	// Output: retrieved project with identifier 12345
	// retrieved project with identifier 12346
}

func startProjectServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /projects", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK","id":"12345"}`)
	})
	mux.HandleFunc("PUT /projects/{id}", func(w http.ResponseWriter, r *http.Request) {
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
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK"}`)
	})
	mux.HandleFunc("DELETE /projects/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK"}`)
	})
	mux.HandleFunc("GET /projects/api/v3/projects/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"project":{"id":12345}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/projects", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"projects":[{"id":12345},{"id":12346}]}`)
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
