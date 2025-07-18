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

func ExampleTagCreate() {
	address, stop, err := startTagServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	tagResponse, err := projects.TagCreate(ctx, engine, projects.NewTagCreateRequest("Test Tag"))
	if err != nil {
		fmt.Printf("failed to create tag: %s", err)
	} else {
		fmt.Printf("created tag with identifier %d\n", tagResponse.Tag.ID)
	}

	// Output: created tag with identifier 12345
}

func ExampleTagUpdate() {
	address, stop, err := startTagServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.TagUpdate(ctx, engine, projects.TagUpdateRequest{
		Path: projects.TagUpdateRequestPath{
			ID: 12345,
		},
		Name: twapi.Ptr("Updated tag"),
	})
	if err != nil {
		fmt.Printf("failed to update tag: %s", err)
	} else {
		fmt.Println("tag updated!")
	}

	// Output: tag updated!
}

func ExampleTagDelete() {
	address, stop, err := startTagServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.TagDelete(ctx, engine, projects.NewTagDeleteRequest(12345))
	if err != nil {
		fmt.Printf("failed to delete tag: %s", err)
	} else {
		fmt.Println("tag deleted!")
	}

	// Output: tag deleted!
}

func ExampleTagGet() {
	address, stop, err := startTagServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	tagResponse, err := projects.TagGet(ctx, engine, projects.NewTagGetRequest(12345))
	if err != nil {
		fmt.Printf("failed to retrieve tag: %s", err)
	} else {
		fmt.Printf("retrieved tag with identifier %d\n", tagResponse.Tag.ID)
	}

	// Output: retrieved tag with identifier 12345
}

func ExampleTagList() {
	address, stop, err := startTagServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	tagsResponse, err := projects.TagList(ctx, engine, projects.TagListRequest{
		Filters: projects.TagListRequestFilters{
			SearchTerm: "Q&A",
		},
	})
	if err != nil {
		fmt.Printf("failed to list tags: %s", err)
	} else {
		for _, tag := range tagsResponse.Tags {
			fmt.Printf("retrieved tag with identifier %d\n", tag.ID)
		}
	}

	// Output: retrieved tag with identifier 12345
	// retrieved tag with identifier 12346
}

func startTagServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /projects/api/v3/tags", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"tag":{"id":12345}}`)
	})
	mux.HandleFunc("PATCH /projects/api/v3/tags/{id}", func(w http.ResponseWriter, r *http.Request) {
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
		_, _ = fmt.Fprintln(w, `{"tag":{"id":12345}}`)
	})
	mux.HandleFunc("DELETE /projects/api/v3/tags/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /projects/api/v3/tags/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"tag":{"id":12345}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/tags", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"tags":[{"id":12345},{"id":12346}]}`)
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
