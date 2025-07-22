package projects_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
	"github.com/teamwork/twapi-go-sdk/session"
)

func ExampleMilestoneCreate() {
	address, stop, err := startMilestoneServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	milestoneRequest := projects.NewMilestoneCreateRequest(777, "New Milestone", projects.NewLegacyDate(time.Now()),
		projects.LegacyUserGroups{
			UserIDs: []int64{456, 789},
		},
	)
	milestoneRequest.Description = twapi.Ptr("This is a new milestone created via the API.")

	milestoneResponse, err := projects.MilestoneCreate(ctx, engine, milestoneRequest)
	if err != nil {
		fmt.Printf("failed to create milestone: %s", err)
	} else {
		fmt.Printf("created milestone with identifier %d\n", milestoneResponse.ID)
	}

	// Output: created milestone with identifier 12345
}

func ExampleMilestoneUpdate() {
	address, stop, err := startMilestoneServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	milestoneRequest := projects.NewMilestoneUpdateRequest(12345)
	milestoneRequest.Description = twapi.Ptr("This is an updated description.")

	_, err = projects.MilestoneUpdate(ctx, engine, milestoneRequest)
	if err != nil {
		fmt.Printf("failed to update milestone: %s", err)
	} else {
		fmt.Println("milestone updated!")
	}

	// Output: milestone updated!
}

func ExampleMilestoneDelete() {
	address, stop, err := startMilestoneServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.MilestoneDelete(ctx, engine, projects.NewMilestoneDeleteRequest(12345))
	if err != nil {
		fmt.Printf("failed to delete milestone: %s", err)
	} else {
		fmt.Println("milestone deleted!")
	}

	// Output: milestone deleted!
}

func ExampleMilestoneGet() {
	address, stop, err := startMilestoneServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	milestoneResponse, err := projects.MilestoneGet(ctx, engine, projects.NewMilestoneGetRequest(12345))
	if err != nil {
		fmt.Printf("failed to retrieve milestone: %s", err)
	} else {
		fmt.Printf("retrieved milestone with identifier %d\n", milestoneResponse.Milestone.ID)
	}

	// Output: retrieved milestone with identifier 12345
}

func ExampleMilestoneList() {
	address, stop, err := startMilestoneServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	milestonesRequest := projects.NewMilestoneListRequest()
	milestonesRequest.Filters.SearchTerm = "Example"

	milestonesResponse, err := projects.MilestoneList(ctx, engine, milestonesRequest)
	if err != nil {
		fmt.Printf("failed to list milestones: %s", err)
	} else {
		for _, milestone := range milestonesResponse.Milestones {
			fmt.Printf("retrieved milestone with identifier %d\n", milestone.ID)
		}
	}

	// Output: retrieved milestone with identifier 12345
	// retrieved milestone with identifier 12346
}

func startMilestoneServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /projects/{id}/milestones", func(w http.ResponseWriter, r *http.Request) {
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
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK","milestoneId":"12345"}`)
	})
	mux.HandleFunc("PUT /milestones/{id}", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("DELETE /milestones/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK"}`)
	})
	mux.HandleFunc("GET /projects/api/v3/milestones/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"milestone":{"id":12345}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/milestones", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"milestones":[{"id":12345},{"id":12346}]}`)
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
