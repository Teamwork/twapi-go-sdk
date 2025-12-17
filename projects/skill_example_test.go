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

func ExampleSkillCreate() {
	address, stop, err := startSkillServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	skillRequest := projects.NewSkillCreateRequest("Software Development")

	skillResponse, err := projects.SkillCreate(ctx, engine, skillRequest)
	if err != nil {
		fmt.Printf("failed to create skill: %s", err)
	} else {
		fmt.Printf("created skill with identifier %d\n", skillResponse.Skill.ID)
	}

	// Output: created skill with identifier 12345
}

func ExampleSkillUpdate() {
	address, stop, err := startSkillServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	skillRequest := projects.NewSkillUpdateRequest(12345)
	skillRequest.UserIDs = []int64{1, 2, 3}

	_, err = projects.SkillUpdate(ctx, engine, skillRequest)
	if err != nil {
		fmt.Printf("failed to update skill: %s", err)
	} else {
		fmt.Println("skill updated!")
	}

	// Output: skill updated!
}

func ExampleSkillDelete() {
	address, stop, err := startSkillServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.SkillDelete(ctx, engine, projects.NewSkillDeleteRequest(12345))
	if err != nil {
		fmt.Printf("failed to delete skill: %s", err)
	} else {
		fmt.Println("skill deleted!")
	}

	// Output: skill deleted!
}

func ExampleSkillGet() {
	address, stop, err := startSkillServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	skillResponse, err := projects.SkillGet(ctx, engine, projects.NewSkillGetRequest(12345))
	if err != nil {
		fmt.Printf("failed to retrieve skill: %s", err)
	} else {
		fmt.Printf("retrieved skill with identifier %d\n", skillResponse.Skill.ID)
	}

	// Output: retrieved skill with identifier 12345
}

func ExampleSkillList() {
	address, stop, err := startSkillServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	skillsRequest := projects.NewSkillListRequest()
	skillsRequest.Filters.SearchTerm = "John"

	skillsResponse, err := projects.SkillList(ctx, engine, skillsRequest)
	if err != nil {
		fmt.Printf("failed to list skills: %s", err)
	} else {
		for _, skill := range skillsResponse.Skills {
			fmt.Printf("retrieved skill with identifier %d\n", skill.ID)
		}
	}

	// Output: retrieved skill with identifier 12345
	// retrieved skill with identifier 12346
}

func startSkillServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /projects/api/v3/skills", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"skill":{"id":12345}}`)
	})
	mux.HandleFunc("PATCH /projects/api/v3/skills/{id}", func(w http.ResponseWriter, r *http.Request) {
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
		_, _ = fmt.Fprintln(w, `{"skill":{"id":12345}}`)
	})
	mux.HandleFunc("DELETE /projects/api/v3/skills/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /projects/api/v3/skills/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"skill":{"id":12345}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/me", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"skill":{"id":12345}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/skills", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"skills":[{"id":12345},{"id":12346}]}`)
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
