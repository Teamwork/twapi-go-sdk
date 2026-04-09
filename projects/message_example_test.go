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

func ExampleMessageCreate() {
	address, stop, err := startMessageServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	messageRequest := projects.NewMessageCreateRequest(777, "New Message", "This is a test message.")

	messageResponse, err := projects.MessageCreate(ctx, engine, messageRequest)
	if err != nil {
		fmt.Printf("failed to create message: %s", err)
	} else {
		fmt.Printf("created message with identifier %d\n", messageResponse.ID)
	}

	// Output: created message with identifier 12345
}

func ExampleMessageUpdate() {
	address, stop, err := startMessageServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	messageRequest := projects.NewMessageUpdateRequest(12345)
	messageRequest.Body = new("This is an updated message.")

	_, err = projects.MessageUpdate(ctx, engine, messageRequest)
	if err != nil {
		fmt.Printf("failed to update message: %s", err)
	} else {
		fmt.Println("message updated!")
	}

	// Output: message updated!
}

func ExampleMessageDelete() {
	address, stop, err := startMessageServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.MessageDelete(ctx, engine, projects.NewMessageDeleteRequest(12345))
	if err != nil {
		fmt.Printf("failed to delete message: %s", err)
	} else {
		fmt.Println("message deleted!")
	}

	// Output: message deleted!
}

func ExampleMessageGet() {
	address, stop, err := startMessageServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	messageResponse, err := projects.MessageGet(ctx, engine, projects.NewMessageGetRequest(12345))
	if err != nil {
		fmt.Printf("failed to retrieve message: %s", err)
	} else {
		fmt.Printf("retrieved message with identifier %d\n", messageResponse.Message.ID)
	}

	// Output: retrieved message with identifier 12345
}

func ExampleMessageList() {
	address, stop, err := startMessageServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	messagesRequest := projects.NewMessageListRequest()
	messagesRequest.Filters.SearchTerm = "Example"

	messagesResponse, err := projects.MessageList(ctx, engine, messagesRequest)
	if err != nil {
		fmt.Printf("failed to list messages: %s", err)
	} else {
		for _, message := range messagesResponse.Messages {
			fmt.Printf("retrieved message with identifier %d\n", message.ID)
		}
	}

	// Output: retrieved message with identifier 12345
	// retrieved message with identifier 12346
}

func startMessageServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /projects/{id}/messages", func(w http.ResponseWriter, r *http.Request) {
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
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK","messageId":"12345"}`)
	})
	mux.HandleFunc("PUT /messages/{id}", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("DELETE /messages/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK"}`)
	})
	mux.HandleFunc("GET /projects/api/v3/messages/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"message":{"id":12345}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/messages", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"messages":[{"id":12345},{"id":12346}]}`)
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
