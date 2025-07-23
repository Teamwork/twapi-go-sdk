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

func ExampleCommentCreate() {
	address, stop, err := startCommentServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	commentRequest := projects.NewCommentCreateRequestInTask(777, "<h1>New Comment</h1>")
	commentRequest.ContentType = twapi.Ptr("HTML")

	commentResponse, err := projects.CommentCreate(ctx, engine, commentRequest)
	if err != nil {
		fmt.Printf("failed to create comment: %s", err)
	} else {
		fmt.Printf("created comment with identifier %d\n", commentResponse.ID)
	}

	// Output: created comment with identifier 12345
}

func ExampleCommentUpdate() {
	address, stop, err := startCommentServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	commentRequest := projects.NewCommentUpdateRequest(12345)
	commentRequest.Body = "<h1>Updated Comment</h1>"

	_, err = projects.CommentUpdate(ctx, engine, commentRequest)
	if err != nil {
		fmt.Printf("failed to update comment: %s", err)
	} else {
		fmt.Println("comment updated!")
	}

	// Output: comment updated!
}

func ExampleCommentDelete() {
	address, stop, err := startCommentServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.CommentDelete(ctx, engine, projects.NewCommentDeleteRequest(12345))
	if err != nil {
		fmt.Printf("failed to delete comment: %s", err)
	} else {
		fmt.Println("comment deleted!")
	}

	// Output: comment deleted!
}

func ExampleCommentGet() {
	address, stop, err := startCommentServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	commentResponse, err := projects.CommentGet(ctx, engine, projects.NewCommentGetRequest(12345))
	if err != nil {
		fmt.Printf("failed to retrieve comment: %s", err)
	} else {
		fmt.Printf("retrieved comment with identifier %d\n", commentResponse.Comment.ID)
	}

	// Output: retrieved comment with identifier 12345
}

func ExampleCommentList() {
	address, stop, err := startCommentServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	commentsRequest := projects.NewCommentListRequest()
	commentsRequest.Filters.SearchTerm = "Example"

	commentsResponse, err := projects.CommentList(ctx, engine, commentsRequest)
	if err != nil {
		fmt.Printf("failed to list comments: %s", err)
	} else {
		for _, comment := range commentsResponse.Comments {
			fmt.Printf("retrieved comment with identifier %d\n", comment.ID)
		}
	}

	// Output: retrieved comment with identifier 12345
	// retrieved comment with identifier 12346
}

func startCommentServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /tasks/{id}/comments", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("PUT /comments/{id}", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("DELETE /comments/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK"}`)
	})
	mux.HandleFunc("GET /projects/api/v3/comments/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"comment":{"id":12345}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/comments", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"comments":[{"id":12345},{"id":12346}]}`)
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
