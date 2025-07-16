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

func ExampleTasklistCreate() {
	address, stop, err := startTasklistServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	tasklist, err := projects.TasklistCreate(ctx, engine, projects.TasklistCreateRequest{
		Path: projects.TasklistCreateRequestPath{
			ProjectID: 777,
		},
		Name:        "New Tasklist",
		Description: twapi.Ptr("This is a new tasklist created via the API."),
	})
	if err != nil {
		fmt.Printf("failed to create tasklist: %s", err)
	} else {
		fmt.Printf("created tasklist with identifier %d\n", tasklist.ID)
	}

	// Output: created tasklist with identifier 12345
}

func ExampleTasklistUpdate() {
	address, stop, err := startTasklistServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.TasklistUpdate(ctx, engine, projects.TasklistUpdateRequest{
		Path: projects.TasklistUpdateRequestPath{
			ID: 12345,
		},
		Description: twapi.Ptr("This is an updated description."),
	})
	if err != nil {
		fmt.Printf("failed to update tasklist: %s", err)
	} else {
		fmt.Println("tasklist updated!")
	}

	// Output: tasklist updated!
}

func ExampleTasklistDelete() {
	address, stop, err := startTasklistServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.TasklistDelete(ctx, engine, projects.NewTasklistDeleteRequest(12345))
	if err != nil {
		fmt.Printf("failed to delete tasklist: %s", err)
	} else {
		fmt.Println("tasklist deleted!")
	}

	// Output: tasklist deleted!
}

func ExampleTasklistGet() {
	address, stop, err := startTasklistServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	tasklistResponse, err := projects.TasklistGet(ctx, engine, projects.NewTasklistGetRequest(12345))
	if err != nil {
		fmt.Printf("failed to retrieve tasklist: %s", err)
	} else {
		fmt.Printf("retrieved tasklist with identifier %d\n", tasklistResponse.Tasklist.ID)
	}

	// Output: retrieved tasklist with identifier 12345
}

func ExampleTasklistList() {
	address, stop, err := startTasklistServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	tasklistsResponse, err := projects.TasklistList(ctx, engine, projects.TasklistListRequest{
		Filters: projects.TasklistListRequestFilters{
			SearchTerm: "Example",
		},
	})
	if err != nil {
		fmt.Printf("failed to list tasklists: %s", err)
	} else {
		for _, tasklist := range tasklistsResponse.Tasklists {
			fmt.Printf("retrieved tasklist with identifier %d\n", tasklist.ID)
		}
	}

	// Output: retrieved tasklist with identifier 12345
	// retrieved tasklist with identifier 12346
}

func startTasklistServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /projects/{id}/tasklists", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		if r.PathValue("id") != "777" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK","tasklistId":"12345"}`)
	})
	mux.HandleFunc("PUT /tasklists/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK"}`)
	})
	mux.HandleFunc("DELETE /tasklists/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK"}`)
	})
	mux.HandleFunc("GET /projects/api/v3/tasklists/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"tasklist":{"id":12345}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/tasklists", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"tasklists":[{"id":12345},{"id":12346}]}`)
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
