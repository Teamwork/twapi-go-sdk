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

func ExampleCustomFieldValueCreate() {
	address, stop, err := startCustomFieldValueServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	customFieldValueResponse, err := projects.CustomFieldValueCreate(ctx, engine,
		projects.NewTaskCustomFieldValueCreateRequest(777, 555, "in progress"))
	if err != nil {
		fmt.Printf("failed to create custom field value: %s", err)
	} else {
		fmt.Printf("created custom field value with identifier %d\n", customFieldValueResponse.CustomFieldValue.ID)
	}

	// Output: created custom field value with identifier 12345
}

func ExampleCustomFieldValueUpdate() {
	address, stop, err := startCustomFieldValueServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.CustomFieldValueUpdate(ctx, engine,
		projects.NewTaskCustomFieldValueUpdateRequest(777, 666, 555, "done"))
	if err != nil {
		fmt.Printf("failed to update custom field value: %s", err)
	} else {
		fmt.Println("custom field value updated!")
	}

	// Output: custom field value updated!
}

func ExampleCustomFieldValueDelete() {
	address, stop, err := startCustomFieldValueServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.CustomFieldValueDelete(ctx, engine,
		projects.NewTaskCustomFieldValueDeleteRequest(777, 555))
	if err != nil {
		fmt.Printf("failed to delete custom field value: %s", err)
	} else {
		fmt.Println("custom field value deleted!")
	}

	// Output: custom field value deleted!
}

func ExampleCustomFieldValueGet() {
	address, stop, err := startCustomFieldValueServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	customFieldValueResponse, err := projects.CustomFieldValueGet(ctx, engine,
		projects.NewTaskCustomFieldValueGetRequest(777, 555))
	if err != nil {
		fmt.Printf("failed to retrieve custom field value: %s", err)
	} else {
		fmt.Printf("retrieved custom field value with identifier %d\n", customFieldValueResponse.CustomFieldValue.ID)
	}

	// Output: retrieved custom field value with identifier 12345
}

func ExampleCustomFieldValueList() {
	address, stop, err := startCustomFieldValueServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	customFieldValuesResponse, err := projects.CustomFieldValueList(ctx, engine,
		projects.NewTaskCustomFieldValueListRequest(777))
	if err != nil {
		fmt.Printf("failed to list custom field values: %s", err)
	} else {
		for _, customFieldValue := range customFieldValuesResponse.CustomFieldValues {
			fmt.Printf("retrieved custom field value with identifier %d\n", customFieldValue.ID)
		}
	}

	// Output: retrieved custom field value with identifier 12345
	// retrieved custom field value with identifier 12346
}

func startCustomFieldValueServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /projects/api/v3/tasks/{taskId}/customfields", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		if r.PathValue("taskId") != "777" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"customfieldTask":{"id":12345,"customfieldId":555}}`)
	})
	mux.HandleFunc("PATCH /projects/api/v3/tasks/{taskId}/customfields/{customFieldId}",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Content-Type") != "application/json" {
				http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
				return
			}
			if r.PathValue("taskId") != "777" || r.PathValue("customFieldId") != "555" {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintln(w, `{"customfieldTask":{"id":12345,"customfieldId":555}}`)
		},
	)
	mux.HandleFunc("DELETE /projects/api/v3/tasks/{taskId}/customfields/{customFieldId}",
		func(w http.ResponseWriter, r *http.Request) {
			if r.PathValue("taskId") != "777" || r.PathValue("customFieldId") != "555" {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		},
	)
	mux.HandleFunc("GET /projects/api/v3/tasks/{taskId}/customfields/{customFieldId}",
		func(w http.ResponseWriter, r *http.Request) {
			if r.PathValue("taskId") != "777" || r.PathValue("customFieldId") != "555" {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintln(w, `{"customfieldTask":{"id":12345,"customfieldId":555}}`)
		},
	)
	mux.HandleFunc("GET /projects/api/v3/tasks/{taskId}/customfields",
		func(w http.ResponseWriter, r *http.Request) {
			if r.PathValue("taskId") != "777" {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintln(w, `{"customfieldTasks":[{"id":12345},{"id":12346}]}`)
		},
	)

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
