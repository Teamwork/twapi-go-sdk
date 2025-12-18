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

func ExampleUserAssignJobRole() {
	address, stop, err := starUserJobRoleServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	userAssignJobRoleRequest := projects.NewUserAssignJobRoleRequest(12345)
	userAssignJobRoleRequest.IDs = []int64{1, 2, 3}

	userAssignJobRoleResponse, err := projects.UserAssignJobRole(ctx, engine, userAssignJobRoleRequest)
	if err != nil {
		fmt.Printf("failed to assign job role to users: %s", err)
	} else {
		fmt.Printf("assigned job role with identifier %d\n", userAssignJobRoleResponse.JobRole.ID)
	}

	// Output: assigned job role with identifier 12345
}

func ExampleUserUnassignJobRole() {
	address, stop, err := starUserJobRoleServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	userUnassignJobRoleRequest := projects.NewUserUnassignJobRoleRequest(12345)
	userUnassignJobRoleRequest.IDs = []int64{1, 2, 3}

	_, err = projects.UserUnassignJobRole(ctx, engine, userUnassignJobRoleRequest)
	if err != nil {
		fmt.Printf("failed to unassign job role to users: %s", err)
	} else {
		fmt.Println("unassigned job role successfully")
	}

	// Output: unassigned job role successfully
}

func starUserJobRoleServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /projects/api/v3/jobroles/{id}/people", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"jobrole":{"id":12345}}`)
	})
	mux.HandleFunc("DELETE /projects/api/v3/jobroles/{id}/people", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		w.WriteHeader(http.StatusNoContent)
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
