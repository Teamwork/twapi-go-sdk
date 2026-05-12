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

func ExampleCustomFieldCreate() {
	address, stop, err := startCustomFieldServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	customFieldResponse, err := projects.CustomFieldCreate(ctx, engine, projects.NewCustomFieldCreateRequest(
		"Priority Score", projects.CustomFieldTypeNumberInteger, projects.CustomFieldEntityTask,
	))
	if err != nil {
		fmt.Printf("failed to create custom field: %s", err)
	} else {
		fmt.Printf("created custom field with identifier %d\n", customFieldResponse.CustomField.ID)
	}

	// Output: created custom field with identifier 12345
}

func ExampleCustomFieldUpdate() {
	address, stop, err := startCustomFieldServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	customFieldRequest := projects.NewCustomFieldUpdateRequest(12345)
	customFieldRequest.Name = new("Updated name")

	_, err = projects.CustomFieldUpdate(ctx, engine, customFieldRequest)
	if err != nil {
		fmt.Printf("failed to update custom field: %s", err)
	} else {
		fmt.Println("custom field updated!")
	}

	// Output: custom field updated!
}

func ExampleCustomFieldDelete() {
	address, stop, err := startCustomFieldServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.CustomFieldDelete(ctx, engine, projects.NewCustomFieldDeleteRequest(12345))
	if err != nil {
		fmt.Printf("failed to delete custom field: %s", err)
	} else {
		fmt.Println("custom field deleted!")
	}

	// Output: custom field deleted!
}

func ExampleCustomFieldGet() {
	address, stop, err := startCustomFieldServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	customFieldResponse, err := projects.CustomFieldGet(ctx, engine, projects.NewCustomFieldGetRequest(12345))
	if err != nil {
		fmt.Printf("failed to retrieve custom field: %s", err)
	} else {
		fmt.Printf("retrieved custom field with identifier %d\n", customFieldResponse.CustomField.ID)
	}

	// Output: retrieved custom field with identifier 12345
}

func ExampleCustomFieldList() {
	address, stop, err := startCustomFieldServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	customFieldsRequest := projects.NewCustomFieldListRequest()
	customFieldsRequest.Filters.SearchTerm = "priority"

	customFieldsResponse, err := projects.CustomFieldList(ctx, engine, customFieldsRequest)
	if err != nil {
		fmt.Printf("failed to list custom fields: %s", err)
	} else {
		for _, customField := range customFieldsResponse.CustomFields {
			fmt.Printf("retrieved custom field with identifier %d\n", customField.ID)
		}
	}

	// Output: retrieved custom field with identifier 12345
	// retrieved custom field with identifier 12346
}

func startCustomFieldServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /projects/api/v3/customfields", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"customfield":{"id":12345}}`)
	})
	mux.HandleFunc("PATCH /projects/api/v3/customfields/{id}", func(w http.ResponseWriter, r *http.Request) {
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
		_, _ = fmt.Fprintln(w, `{"customfield":{"id":12345}}`)
	})
	mux.HandleFunc("DELETE /projects/api/v3/customfields/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /projects/api/v3/customfields/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"customfield":{"id":12345}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/customfields", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"customfields":[{"id":12345},{"id":12346}]}`)
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
