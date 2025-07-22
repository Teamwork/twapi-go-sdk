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

func ExampleUserCreate() {
	address, stop, err := startUserServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	userRequest := projects.NewUserCreateRequest("John", "Doe", "johndoe@example.com")

	userResponse, err := projects.UserCreate(ctx, engine, userRequest)
	if err != nil {
		fmt.Printf("failed to create user: %s", err)
	} else {
		fmt.Printf("created user with identifier %d\n", userResponse.ID)
	}

	// Output: created user with identifier 12345
}

func ExampleUserUpdate() {
	address, stop, err := startUserServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	userRequest := projects.NewUserUpdateRequest(12345)
	userRequest.Title = twapi.Ptr("Software Engineer")

	_, err = projects.UserUpdate(ctx, engine, userRequest)
	if err != nil {
		fmt.Printf("failed to update user: %s", err)
	} else {
		fmt.Println("user updated!")
	}

	// Output: user updated!
}

func ExampleUserDelete() {
	address, stop, err := startUserServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.UserDelete(ctx, engine, projects.NewUserDeleteRequest(12345))
	if err != nil {
		fmt.Printf("failed to delete user: %s", err)
	} else {
		fmt.Println("user deleted!")
	}

	// Output: user deleted!
}

func ExampleUserGet() {
	address, stop, err := startUserServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	userResponse, err := projects.UserGet(ctx, engine, projects.NewUserGetRequest(12345))
	if err != nil {
		fmt.Printf("failed to retrieve user: %s", err)
	} else {
		fmt.Printf("retrieved user with identifier %d\n", userResponse.User.ID)
	}

	// Output: retrieved user with identifier 12345
}

func ExampleUserGetMe() {
	address, stop, err := startUserServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	userResponse, err := projects.UserGetMe(ctx, engine, projects.NewUserGetMeRequest())
	if err != nil {
		fmt.Printf("failed to retrieve logged user: %s", err)
	} else {
		fmt.Printf("retrieved logged user with identifier %d\n", userResponse.User.ID)
	}

	// Output: retrieved logged user with identifier 12345
}

func ExampleUserList() {
	address, stop, err := startUserServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	usersRequest := projects.NewUserListRequest()
	usersRequest.Filters.SearchTerm = "John"

	usersResponse, err := projects.UserList(ctx, engine, usersRequest)
	if err != nil {
		fmt.Printf("failed to list users: %s", err)
	} else {
		for _, user := range usersResponse.Users {
			fmt.Printf("retrieved user with identifier %d\n", user.ID)
		}
	}

	// Output: retrieved user with identifier 12345
	// retrieved user with identifier 12346
}

func startUserServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /people", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK","id":"12345"}`)
	})
	mux.HandleFunc("PUT /people/{id}", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("DELETE /people/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK"}`)
	})
	mux.HandleFunc("GET /projects/api/v3/people/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"person":{"id":12345}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/me", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"person":{"id":12345}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/people", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"people":[{"id":12345},{"id":12346}]}`)
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
