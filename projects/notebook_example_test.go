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

func ExampleNotebookCreate() {
	address, stop, err := startNotebookServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	notebookRequest :=
		projects.NewNotebookCreateRequest(777, "New Notebook", "An amazing content", projects.NotebookTypeMarkdown)
	notebookRequest.Description = twapi.Ptr("This is a new notebook created via the API.")

	notebookResponse, err := projects.NotebookCreate(ctx, engine, notebookRequest)
	if err != nil {
		fmt.Printf("failed to create notebook: %s", err)
	} else {
		fmt.Printf("created notebook with identifier %d\n", notebookResponse.Notebook.ID)
	}

	// Output: created notebook with identifier 12345
}

func ExampleNotebookUpdate() {
	address, stop, err := startNotebookServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	notebookRequest := projects.NewNotebookUpdateRequest(12345)
	notebookRequest.Description = twapi.Ptr("This is an updated description.")

	_, err = projects.NotebookUpdate(ctx, engine, notebookRequest)
	if err != nil {
		fmt.Printf("failed to update notebook: %s", err)
	} else {
		fmt.Println("notebook updated!")
	}

	// Output: notebook updated!
}

func ExampleNotebookDelete() {
	address, stop, err := startNotebookServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.NotebookDelete(ctx, engine, projects.NewNotebookDeleteRequest(12345))
	if err != nil {
		fmt.Printf("failed to delete notebook: %s", err)
	} else {
		fmt.Println("notebook deleted!")
	}

	// Output: notebook deleted!
}

func ExampleNotebookGet() {
	address, stop, err := startNotebookServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	notebookResponse, err := projects.NotebookGet(ctx, engine, projects.NewNotebookGetRequest(12345))
	if err != nil {
		fmt.Printf("failed to retrieve notebook: %s", err)
	} else {
		fmt.Printf("retrieved notebook with identifier %d\n", notebookResponse.Notebook.ID)
	}

	// Output: retrieved notebook with identifier 12345
}

func ExampleNotebookList() {
	address, stop, err := startNotebookServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	notebooksRequest := projects.NewNotebookListRequest()
	notebooksRequest.Filters.SearchTerm = "Example"

	notebooksResponse, err := projects.NotebookList(ctx, engine, notebooksRequest)
	if err != nil {
		fmt.Printf("failed to list notebooks: %s", err)
	} else {
		for _, notebook := range notebooksResponse.Notebooks {
			fmt.Printf("retrieved notebook with identifier %d\n", notebook.ID)
		}
	}

	// Output: retrieved notebook with identifier 12345
	// retrieved notebook with identifier 12346
}

func startNotebookServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /projects/api/v3/projects/{id}/notebooks", func(w http.ResponseWriter, r *http.Request) {
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
		_, _ = fmt.Fprintln(w, `{"notebook":{"id":12345}}`)
	})
	mux.HandleFunc("PATCH /projects/api/v3/notebooks/{id}", func(w http.ResponseWriter, r *http.Request) {
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
		_, _ = fmt.Fprintln(w, `{"notebook":{"id":12345}}`)
	})
	mux.HandleFunc("DELETE /projects/api/v3/notebooks/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /projects/api/v3/notebooks/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"notebook":{"id":12345}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/notebooks", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"notebooks":[{"id":12345},{"id":12346}]}`)
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
