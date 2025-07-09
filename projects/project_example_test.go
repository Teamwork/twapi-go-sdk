package projects_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
	"github.com/teamwork/twapi-go-sdk/session"
)

func ExampleProjectCreate() {
	address, stop, err := startProjectCreateServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	project, err := projects.ProjectCreate(ctx, engine, projects.ProjectCreateRequest{
		Name:        "New Project",
		Description: twapi.Ptr("This is a new project created via the API."),
		StartAt:     twapi.Ptr(projects.LegacyDate(time.Now().AddDate(0, 0, 1))),  // Start tomorrow
		EndAt:       twapi.Ptr(projects.LegacyDate(time.Now().AddDate(0, 0, 30))), // End in 30 days
		CompanyID:   12345,                                                        // Replace with your company ID
		OwnerID:     twapi.Ptr(int64(67890)),                                      // Replace with the owner user ID
		Tags:        []int64{11111, 22222},                                        // Replace with your tag IDs
	})
	if err != nil {
		fmt.Printf("failed to create project: %s", err)
	} else {
		fmt.Printf("created project with identifier %d\n", project.ID)
	}

	// Output: created project with identifier 12345
}

func startProjectCreateServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/projects.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		if r.Header.Get("Authorization") != "Bearer your_token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"id":"12345"}`)
	})

	server := &http.Server{
		Handler: mux,
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
