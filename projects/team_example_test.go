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

func ExampleTeamCreate() {
	address, stop, err := startTeamServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	teamRequest := projects.NewTeamCreateRequest("My Team")

	teamResponse, err := projects.TeamCreate(ctx, engine, teamRequest)
	if err != nil {
		fmt.Printf("failed to create team: %s", err)
	} else {
		fmt.Printf("created team with identifier %d\n", teamResponse.ID)
	}

	// Output: created team with identifier 12345
}

func ExampleTeamUpdate() {
	address, stop, err := startTeamServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	teamRequest := projects.NewTeamUpdateRequest(12345)
	teamRequest.Description = twapi.Ptr("This is an updated team description.")

	_, err = projects.TeamUpdate(ctx, engine, teamRequest)
	if err != nil {
		fmt.Printf("failed to update team: %s", err)
	} else {
		fmt.Println("team updated!")
	}

	// Output: team updated!
}

func ExampleTeamDelete() {
	address, stop, err := startTeamServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.TeamDelete(ctx, engine, projects.NewTeamDeleteRequest(12345))
	if err != nil {
		fmt.Printf("failed to delete team: %s", err)
	} else {
		fmt.Println("team deleted!")
	}

	// Output: team deleted!
}

func ExampleTeamGet() {
	address, stop, err := startTeamServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	teamResponse, err := projects.TeamGet(ctx, engine, projects.NewTeamGetRequest(12345))
	if err != nil {
		fmt.Printf("failed to retrieve team: %s", err)
	} else {
		fmt.Printf("retrieved team with identifier %d\n", teamResponse.Team.ID)
	}

	// Output: retrieved team with identifier 12345
}

func ExampleTeamList() {
	address, stop, err := startTeamServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	teamsRequest := projects.NewTeamListRequest()
	teamsRequest.Filters.SearchTerm = "Team A"

	teamsResponse, err := projects.TeamList(ctx, engine, teamsRequest)
	if err != nil {
		fmt.Printf("failed to list teams: %s", err)
	} else {
		for _, team := range teamsResponse.Teams {
			fmt.Printf("retrieved team with identifier %d\n", team.ID)
		}
	}

	// Output: retrieved team with identifier 12345
	// retrieved team with identifier 12346
}

func startTeamServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /teams", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK","id":"12345"}`)
	})
	mux.HandleFunc("PUT /teams/{id}", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("DELETE /teams/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK"}`)
	})
	mux.HandleFunc("GET /teams/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"team":{"id":"12345"}}`)
	})
	mux.HandleFunc("GET /teams", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"teams":[{"id":"12345"},{"id":"12346"}]}`)
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
