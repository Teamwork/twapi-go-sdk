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

func ExampleProjectCategoryCreate() {
	address, stop, err := startProjectCategoryServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	projectCategoryRequest := projects.NewProjectCategoryCreateRequest("New Project Category")
	projectCategoryRequest.ParentID = twapi.Ptr(int64(12345))
	projectCategoryRequest.Color = twapi.Ptr("#ff0000")

	projectCategoryResponse, err := projects.ProjectCategoryCreate(ctx, engine, projectCategoryRequest)
	if err != nil {
		fmt.Printf("failed to create project category: %s", err)
	} else {
		fmt.Printf("created project category with identifier %d\n", projectCategoryResponse.ID)
	}

	// Output: created project category with identifier 12345
}

func ExampleProjectCategoryUpdate() {
	address, stop, err := startProjectCategoryServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	projectCategoryRequest := projects.NewProjectCategoryUpdateRequest(12345)
	projectCategoryRequest.Color = twapi.Ptr("#aaaaaa")

	_, err = projects.ProjectCategoryUpdate(ctx, engine, projectCategoryRequest)
	if err != nil {
		fmt.Printf("failed to update project category: %s", err)
	} else {
		fmt.Println("project category updated!")
	}

	// Output: project category updated!
}

func ExampleProjectCategoryDelete() {
	address, stop, err := startProjectCategoryServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.ProjectCategoryDelete(ctx, engine, projects.NewProjectCategoryDeleteRequest(12345))
	if err != nil {
		fmt.Printf("failed to delete project category: %s", err)
	} else {
		fmt.Println("project category deleted!")
	}

	// Output: project category deleted!
}

func ExampleProjectCategoryGet() {
	address, stop, err := startProjectCategoryServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	projectCategoryResponse, err := projects.ProjectCategoryGet(ctx, engine, projects.NewProjectCategoryGetRequest(12345))
	if err != nil {
		fmt.Printf("failed to retrieve project category: %s", err)
	} else {
		fmt.Printf("retrieved project category with identifier %d\n", projectCategoryResponse.ProjectCategory.ID)
	}

	// Output: retrieved project category with identifier 12345
}

func ExampleProjectCategoryList() {
	address, stop, err := startProjectCategoryServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	projectCategoriesRequest := projects.NewProjectCategoryListRequest()
	projectCategoriesRequest.Filters.SearchTerm = "Example"

	projectCategoriesResponse, err := projects.ProjectCategoryList(ctx, engine, projectCategoriesRequest)
	if err != nil {
		fmt.Printf("failed to list project categories: %s", err)
	} else {
		for _, projectCategory := range projectCategoriesResponse.ProjectCategories {
			fmt.Printf("retrieved project category with identifier %d\n", projectCategory.ID)
		}
	}

	// Output: retrieved project category with identifier 12345
	// retrieved project category with identifier 12346
}

func startProjectCategoryServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /projectcategories", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK","categoryId":"12345"}`)
	})
	mux.HandleFunc("PUT /projectcategories/{id}", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("DELETE /projectcategories/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK"}`)
	})
	mux.HandleFunc("GET /projects/api/v3/projectcategories/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"projectCategory":{"id":12345}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/projectcategories", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"projectCategories":[{"id":12345},{"id":12346}]}`)
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
