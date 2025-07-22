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

func ExampleTaskCreate() {
	address, stop, err := startTaskServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	taskRequest := projects.NewTaskCreateRequest(777, "New Task")
	taskRequest.Description = twapi.Ptr("This is a new task created via the API.")

	taskResponse, err := projects.TaskCreate(ctx, engine, taskRequest)
	if err != nil {
		fmt.Printf("failed to create task: %s", err)
	} else {
		fmt.Printf("created task with identifier %d\n", taskResponse.Task.ID)
	}

	// Output: created task with identifier 12345
}

func ExampleTaskUpdate() {
	address, stop, err := startTaskServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	taskRequest := projects.NewTaskUpdateRequest(12345)
	taskRequest.Description = twapi.Ptr("This is an updated description.")

	_, err = projects.TaskUpdate(ctx, engine, taskRequest)
	if err != nil {
		fmt.Printf("failed to update task: %s", err)
	} else {
		fmt.Println("task updated!")
	}

	// Output: task updated!
}

func ExampleTaskDelete() {
	address, stop, err := startTaskServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.TaskDelete(ctx, engine, projects.NewTaskDeleteRequest(12345))
	if err != nil {
		fmt.Printf("failed to delete task: %s", err)
	} else {
		fmt.Println("task deleted!")
	}

	// Output: task deleted!
}

func ExampleTaskGet() {
	address, stop, err := startTaskServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	taskResponse, err := projects.TaskGet(ctx, engine, projects.NewTaskGetRequest(12345))
	if err != nil {
		fmt.Printf("failed to retrieve task: %s", err)
	} else {
		fmt.Printf("retrieved task with identifier %d\n", taskResponse.Task.ID)
	}

	// Output: retrieved task with identifier 12345
}

func ExampleTaskList() {
	address, stop, err := startTaskServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	tasksRequest := projects.NewTaskListRequest()
	tasksRequest.Filters.SearchTerm = "Example"

	tasksResponse, err := projects.TaskList(ctx, engine, tasksRequest)
	if err != nil {
		fmt.Printf("failed to list tasks: %s", err)
	} else {
		for _, task := range tasksResponse.Tasks {
			fmt.Printf("retrieved task with identifier %d\n", task.ID)
		}
	}

	// Output: retrieved task with identifier 12345
	// retrieved task with identifier 12346
}

func startTaskServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /projects/api/v3/tasklists/{id}/tasks", func(w http.ResponseWriter, r *http.Request) {
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
		_, _ = fmt.Fprintln(w, `{"task":{"id":12345}}`)
	})
	mux.HandleFunc("PUT /projects/api/v3/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
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
		_, _ = fmt.Fprintln(w, `{"task":{"id":12345}}`)
	})
	mux.HandleFunc("DELETE /projects/api/v3/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"affected":{"taskIds":[12345]}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"task":{"id":12345}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/tasks", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"tasks":[{"id":12345},{"id":12346}]}`)
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
