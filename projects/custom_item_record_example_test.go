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

func ExampleCustomItemRecordCreate() {
	address, stop, err := startCustomItemRecordServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	recordRequest := projects.NewCustomItemRecordCreateRequest(12345, "Acme Inc Contract")
	recordRequest.FieldValues = projects.CustomItemRecordFieldValues{
		"f_67890": "Acme Inc",
	}

	recordResponse, err := projects.CustomItemRecordCreate(ctx, engine, recordRequest)
	if err != nil {
		fmt.Printf("failed to create custom item record: %s", err)
	} else {
		fmt.Printf("created custom item record with identifier %d\n", recordResponse.CustomItemRecord.ID)
	}

	// Output: created custom item record with identifier 99001
}

func ExampleCustomItemRecordUpdate() {
	address, stop, err := startCustomItemRecordServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	updatedName := "Acme Inc Contract (renewed)"
	recordRequest := projects.NewCustomItemRecordUpdateRequest(12345, 99001)
	recordRequest.Name = &updatedName

	_, err = projects.CustomItemRecordUpdate(ctx, engine, recordRequest)
	if err != nil {
		fmt.Printf("failed to update custom item record: %s", err)
	} else {
		fmt.Println("custom item record updated!")
	}

	// Output: custom item record updated!
}

func ExampleCustomItemRecordDelete() {
	address, stop, err := startCustomItemRecordServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.CustomItemRecordDelete(ctx, engine,
		projects.NewCustomItemRecordDeleteRequest(12345, 99001))
	if err != nil {
		fmt.Printf("failed to delete custom item record: %s", err)
	} else {
		fmt.Println("custom item record deleted!")
	}

	// Output: custom item record deleted!
}

func ExampleCustomItemRecordBulkDelete() {
	address, stop, err := startCustomItemRecordServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.CustomItemRecordBulkDelete(ctx, engine,
		projects.NewCustomItemRecordBulkDeleteRequest(12345, []int64{99001, 99002, 99003}))
	if err != nil {
		fmt.Printf("failed to bulk delete custom item records: %s", err)
	} else {
		fmt.Println("custom item records deleted!")
	}

	// Output: custom item records deleted!
}

func ExampleCustomItemRecordGet() {
	address, stop, err := startCustomItemRecordServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	recordResponse, err := projects.CustomItemRecordGet(ctx, engine,
		projects.NewCustomItemRecordGetRequest(12345, 99001))
	if err != nil {
		fmt.Printf("failed to retrieve custom item record: %s", err)
	} else {
		fmt.Printf("retrieved custom item record with identifier %d\n", recordResponse.CustomItemRecord.ID)
	}

	// Output: retrieved custom item record with identifier 99001
}

func ExampleCustomItemRecordList() {
	address, stop, err := startCustomItemRecordServer()
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	recordsRequest := projects.NewCustomItemRecordListRequest(12345)
	recordsRequest.Filters.SearchTerm = "acme"

	recordsResponse, err := projects.CustomItemRecordList(ctx, engine, recordsRequest)
	if err != nil {
		fmt.Printf("failed to list custom item records: %s", err)
	} else {
		for _, record := range recordsResponse.CustomItemRecords {
			fmt.Printf("retrieved custom item record with identifier %d\n", record.ID)
		}
	}

	// Output: retrieved custom item record with identifier 99001
	// retrieved custom item record with identifier 99002
}

func startCustomItemRecordServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /projects/api/v3/customitems/{customItemId}/records/bulk/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /projects/api/v3/customitems/{customItemId}/records", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"customItemRecord":{"id":99001,"name":"Acme Inc Contract","state":"active","fieldValues":{"f_67890":"Acme Inc"}}}`)
	})
	mux.HandleFunc("PATCH /projects/api/v3/customitems/{customItemId}/records/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		if r.PathValue("id") != "99001" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"customItemRecord":{"id":99001,"name":"Acme Inc Contract (renewed)","state":"active"}}`)
	})
	mux.HandleFunc("DELETE /projects/api/v3/customitems/{customItemId}/records/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "99001" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /projects/api/v3/customitems/{customItemId}/records/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "99001" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"customItemRecord":{"id":99001,"name":"Acme Inc Contract","state":"active","fieldValues":{"f_67890":"Acme Inc"}}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/customitems/{customItemId}/records", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"customItemRecords":[{"id":99001,"name":"Acme Inc Contract","state":"active"},{"id":99002,"name":"Acme Subsidiary Contract","state":"active"}],"meta":{"page":{"offset":0,"size":50,"count":2,"hasMore":false}}}`)
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
