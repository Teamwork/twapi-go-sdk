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

func ExampleLinkCreate() {
	address, stop, err := startLinkServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	linkRequest := projects.NewLinkCreateRequest(777, "http://example.com")

	linkResponse, err := projects.LinkCreate(ctx, engine, linkRequest)
	if err != nil {
		fmt.Printf("failed to create link: %s", err)
	} else {
		fmt.Printf("created link with identifier %d\n", linkResponse.ID)
	}

	// Output: created link with identifier 12345
}

func ExampleLinkUpdate() {
	address, stop, err := startLinkServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	linkRequest := projects.NewLinkUpdateRequest(12345)
	linkRequest.Title = new("This is an updated link.")

	_, err = projects.LinkUpdate(ctx, engine, linkRequest)
	if err != nil {
		fmt.Printf("failed to update link: %s", err)
	} else {
		fmt.Println("link updated!")
	}

	// Output: link updated!
}

func ExampleLinkDelete() {
	address, stop, err := startLinkServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.LinkDelete(ctx, engine, projects.NewLinkDeleteRequest(12345))
	if err != nil {
		fmt.Printf("failed to delete link: %s", err)
	} else {
		fmt.Println("link deleted!")
	}

	// Output: link deleted!
}

func ExampleLinkGet() {
	address, stop, err := startLinkServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	linkResponse, err := projects.LinkGet(ctx, engine, projects.NewLinkGetRequest(12345))
	if err != nil {
		fmt.Printf("failed to retrieve link: %s", err)
	} else {
		fmt.Printf("retrieved link with identifier %d\n", linkResponse.Link.ID)
	}

	// Output: retrieved link with identifier 12345
}

func ExampleLinkList() {
	address, stop, err := startLinkServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	linksRequest := projects.NewLinkListRequest()
	linksRequest.Filters.SearchTerm = "Example"

	linksResponse, err := projects.LinkList(ctx, engine, linksRequest)
	if err != nil {
		fmt.Printf("failed to list links: %s", err)
	} else {
		for _, link := range linksResponse.Links {
			fmt.Printf("retrieved link with identifier %d\n", link.ID)
		}
	}

	// Output: retrieved link with identifier 12345
	// retrieved link with identifier 12346
}

func startLinkServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /projects/{id}/links", func(w http.ResponseWriter, r *http.Request) {
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
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK","id":"12345"}`)
	})
	mux.HandleFunc("PUT /links/{id}", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("DELETE /links/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK"}`)
	})
	mux.HandleFunc("GET /links/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"link":{"id":"12345"}}`)
	})
	mux.HandleFunc("GET /links", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"project":{"links":[{"id":"12345"},{"id":"12346"}]}}`)
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
