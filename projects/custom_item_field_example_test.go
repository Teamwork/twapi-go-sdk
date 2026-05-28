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

func ExampleCustomItemFieldCreate() {
	address, stop, err := startCustomItemFieldServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	fieldResponse, err := projects.CustomItemFieldCreate(ctx, engine,
		projects.NewCustomItemFieldCreateRequest(12345, "Customer", projects.CustomItemFieldTypeTextShort))
	if err != nil {
		fmt.Printf("failed to create custom item field: %s", err)
	} else {
		fmt.Printf("created custom item field with identifier %d\n", fieldResponse.CustomItemField.ID)
	}

	// Output: created custom item field with identifier 67890
}

func ExampleCustomItemFieldUpdate() {
	address, stop, err := startCustomItemFieldServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	updatedName := "Primary Customer"
	fieldRequest := projects.NewCustomItemFieldUpdateRequest(12345, 67890)
	fieldRequest.DisplayName = &updatedName

	_, err = projects.CustomItemFieldUpdate(ctx, engine, fieldRequest)
	if err != nil {
		fmt.Printf("failed to update custom item field: %s", err)
	} else {
		fmt.Println("custom item field updated!")
	}

	// Output: custom item field updated!
}

func ExampleCustomItemFieldDelete() {
	address, stop, err := startCustomItemFieldServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.CustomItemFieldDelete(ctx, engine,
		projects.NewCustomItemFieldDeleteRequest(12345, 67890))
	if err != nil {
		fmt.Printf("failed to delete custom item field: %s", err)
	} else {
		fmt.Println("custom item field deleted!")
	}

	// Output: custom item field deleted!
}

func ExampleCustomItemFieldGet() {
	address, stop, err := startCustomItemFieldServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	fieldResponse, err := projects.CustomItemFieldGet(ctx, engine,
		projects.NewCustomItemFieldGetRequest(12345, 67890))
	if err != nil {
		fmt.Printf("failed to retrieve custom item field: %s", err)
	} else {
		fmt.Printf("retrieved custom item field with identifier %d\n", fieldResponse.CustomItemField.ID)
	}

	// Output: retrieved custom item field with identifier 67890
}

func ExampleCustomItemFieldList() {
	address, stop, err := startCustomItemFieldServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	fieldsResponse, err := projects.CustomItemFieldList(ctx, engine,
		projects.NewCustomItemFieldListRequest(12345))
	if err != nil {
		fmt.Printf("failed to list custom item fields: %s", err)
	} else {
		for _, field := range fieldsResponse.CustomItemFields {
			fmt.Printf("retrieved custom item field with identifier %d\n", field.ID)
		}
	}

	// Output: retrieved custom item field with identifier 67890
	// retrieved custom item field with identifier 67891
}

func startCustomItemFieldServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /projects/api/v3/customitems/{customItemId}/fields", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"customItemField":{"id":67890,"displayName":"Customer","type":"text-short","twId":"f_67890","state":"active"}}`)
	})
	mux.HandleFunc("PATCH /projects/api/v3/customitems/{customItemId}/fields/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		if r.PathValue("id") != "67890" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"customItemField":{"id":67890,"displayName":"Primary Customer","type":"text-short","twId":"f_67890","state":"active"}}`)
	})
	mux.HandleFunc("DELETE /projects/api/v3/customitems/{customItemId}/fields/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "67890" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /projects/api/v3/customitems/{customItemId}/fields/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "67890" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"customItemField":{"id":67890,"displayName":"Customer","type":"text-short","twId":"f_67890","state":"active"}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/customitems/{customItemId}/fields", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"customItemFields":[{"id":67890,"displayName":"Customer","type":"text-short","twId":"f_67890","state":"active"},{"id":67891,"displayName":"Value","type":"number-decimal","twId":"f_67891","state":"active"}],"meta":{"page":{"offset":0,"size":50,"count":2,"hasMore":false}}}`)
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
