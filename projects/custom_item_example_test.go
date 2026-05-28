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

func ExampleCustomItemCreate() {
	address, stop, err := startCustomItemServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	customItemResponse, err := projects.CustomItemCreate(ctx, engine,
		projects.NewCustomItemCreateRequest(777, "Contracts"))
	if err != nil {
		fmt.Printf("failed to create custom item: %s", err)
	} else {
		fmt.Printf("created custom item with identifier %d\n", customItemResponse.CustomItem.ID)
	}

	// Output: created custom item with identifier 12345
}

func ExampleCustomItemUpdate() {
	address, stop, err := startCustomItemServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	updatedName := "Active Contracts"
	customItemRequest := projects.NewCustomItemUpdateRequest(12345)
	customItemRequest.DisplayName = &updatedName

	_, err = projects.CustomItemUpdate(ctx, engine, customItemRequest)
	if err != nil {
		fmt.Printf("failed to update custom item: %s", err)
	} else {
		fmt.Println("custom item updated!")
	}

	// Output: custom item updated!
}

func ExampleCustomItemDelete() {
	address, stop, err := startCustomItemServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.CustomItemDelete(ctx, engine, projects.NewCustomItemDeleteRequest(12345))
	if err != nil {
		fmt.Printf("failed to delete custom item: %s", err)
	} else {
		fmt.Println("custom item deleted!")
	}

	// Output: custom item deleted!
}

func ExampleCustomItemGet() {
	address, stop, err := startCustomItemServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	customItemResponse, err := projects.CustomItemGet(ctx, engine, projects.NewCustomItemGetRequest(12345))
	if err != nil {
		fmt.Printf("failed to retrieve custom item: %s", err)
	} else {
		fmt.Printf("retrieved custom item with identifier %d\n", customItemResponse.CustomItem.ID)
	}

	// Output: retrieved custom item with identifier 12345
}

func ExampleCustomItemList() {
	address, stop, err := startCustomItemServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	customItemsRequest := projects.NewCustomItemListRequest(777)
	customItemsRequest.Filters.SearchTerm = "contract"

	customItemsResponse, err := projects.CustomItemList(ctx, engine, customItemsRequest)
	if err != nil {
		fmt.Printf("failed to list custom items: %s", err)
	} else {
		for _, customItem := range customItemsResponse.CustomItems {
			fmt.Printf("retrieved custom item with identifier %d\n", customItem.ID)
		}
	}

	// Output: retrieved custom item with identifier 12345
	// retrieved custom item with identifier 12346
}

func startCustomItemServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /projects/api/v3/projects/{projectId}/customitems", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"customItem":{"id":12345,"displayName":"Contracts","labelSingular":"Contract","labelPlural":"Contracts","state":"active"}}`)
	})
	mux.HandleFunc("PATCH /projects/api/v3/customitems/{id}", func(w http.ResponseWriter, r *http.Request) {
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
		_, _ = fmt.Fprintln(w, `{"customItem":{"id":12345,"displayName":"Active Contracts","state":"active"}}`)
	})
	mux.HandleFunc("DELETE /projects/api/v3/customitems/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /projects/api/v3/customitems/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"customItem":{"id":12345,"displayName":"Contracts","state":"active"}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/projects/{projectId}/customitems", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"customItems":[{"id":12345,"displayName":"Contracts","state":"active"},{"id":12346,"displayName":"Leads","state":"active"}],"meta":{"page":{"offset":0,"size":50,"count":2,"hasMore":false}}}`)
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
