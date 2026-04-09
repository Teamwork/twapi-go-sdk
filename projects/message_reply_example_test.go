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

func ExampleMessageReplyCreate() {
	address, stop, err := startMessageReplyServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	messageRequest := projects.NewMessageReplyCreateRequest(777, "This is a test reply.")

	messageResponse, err := projects.MessageReplyCreate(ctx, engine, messageRequest)
	if err != nil {
		fmt.Printf("failed to create message reply: %s", err)
	} else {
		fmt.Printf("created message reply with identifier %d\n", messageResponse.ID)
	}

	// Output: created message reply with identifier 12345
}

func ExampleMessageReplyUpdate() {
	address, stop, err := startMessageReplyServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	messageRequest := projects.NewMessageReplyUpdateRequest(12345)
	messageRequest.Body = new("This is an updated reply.")

	_, err = projects.MessageReplyUpdate(ctx, engine, messageRequest)
	if err != nil {
		fmt.Printf("failed to update message reply: %s", err)
	} else {
		fmt.Println("message reply updated!")
	}

	// Output: message reply updated!
}

func ExampleMessageReplyDelete() {
	address, stop, err := startMessageReplyServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.MessageReplyDelete(ctx, engine, projects.NewMessageReplyDeleteRequest(12345))
	if err != nil {
		fmt.Printf("failed to delete message reply: %s", err)
	} else {
		fmt.Println("message reply deleted!")
	}

	// Output: message reply deleted!
}

func ExampleMessageReplyGet() {
	address, stop, err := startMessageReplyServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	messageResponse, err := projects.MessageReplyGet(ctx, engine, projects.NewMessageReplyGetRequest(12345))
	if err != nil {
		fmt.Printf("failed to retrieve message reply: %s", err)
	} else {
		fmt.Printf("retrieved message reply with identifier %d\n", messageResponse.MessageReply.ID)
	}

	// Output: retrieved message reply with identifier 12345
}

func ExampleMessageReplyList() {
	address, stop, err := startMessageReplyServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	messagesRequest := projects.NewMessageReplyListRequest()
	messagesRequest.Filters.MessageIDs = []int64{10}

	messagesResponse, err := projects.MessageReplyList(ctx, engine, messagesRequest)
	if err != nil {
		fmt.Printf("failed to list message replies: %s", err)
	} else {
		for _, message := range messagesResponse.MessageReplies {
			fmt.Printf("retrieved message reply with identifier %d\n", message.ID)
		}
	}

	// Output: retrieved message reply with identifier 12345
	// retrieved message reply with identifier 12346
}

func startMessageReplyServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /messages/{id}/replies", func(w http.ResponseWriter, r *http.Request) {
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
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK","postId":"12345"}`)
	})
	mux.HandleFunc("PUT /messageReplies/{id}", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("DELETE /messageReplies/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK"}`)
	})
	mux.HandleFunc("GET /projects/api/v3/messagereplies/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"messageReply":{"id":12345}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/messagereplies", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"messageReplies":[{"id":12345},{"id":12346}]}`)
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
