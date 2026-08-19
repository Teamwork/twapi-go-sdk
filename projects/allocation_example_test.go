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

func ExampleAllocationCreate() {
	address, stop, err := startAllocationServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	start := twapi.Date(time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC))
	end := twapi.Date(time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC))

	allocationRequest := projects.NewAllocationCreateRequest(777, 456, "Design phase", start, end, 4, "#3c8f7c")
	allocationRequest.Allocation.Description = new("Four hours a day for the design phase.")
	allocationRequest.Allocation.IsBillable = new(true)

	allocationResponse, err := projects.AllocationCreate(ctx, engine, allocationRequest)
	if err != nil {
		fmt.Printf("failed to create allocation: %s", err)
	} else {
		fmt.Printf("created allocation with identifier %d\n", allocationResponse.Allocation.ID)
	}

	// Output: created allocation with identifier 12345
}

func ExampleAllocationUpdate() {
	address, stop, err := startAllocationServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	allocationRequest := projects.NewAllocationUpdateRequest(12345)
	allocationRequest.Allocation.HoursPerDay = new(6.0)

	_, err = projects.AllocationUpdate(ctx, engine, allocationRequest)
	if err != nil {
		fmt.Printf("failed to update allocation: %s", err)
	} else {
		fmt.Println("allocation updated!")
	}

	// Output: allocation updated!
}

func ExampleAllocationDelete() {
	address, stop, err := startAllocationServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	// the delete is a soft delete unless HardDelete is set, so the allocation can
	// still be brought back with AllocationRestore
	_, err = projects.AllocationDelete(ctx, engine, projects.NewAllocationDeleteRequest(12345))
	if err != nil {
		fmt.Printf("failed to delete allocation: %s", err)
	} else {
		fmt.Println("allocation deleted!")
	}

	// Output: allocation deleted!
}

func ExampleAllocationRestore() {
	address, stop, err := startAllocationServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	allocationResponse, err := projects.AllocationRestore(ctx, engine, projects.NewAllocationRestoreRequest(12345))
	if err != nil {
		fmt.Printf("failed to restore allocation: %s", err)
	} else {
		fmt.Printf("restored allocation is %s\n", allocationResponse.Allocation.Status)
	}

	// Output: restored allocation is active
}

func ExampleAllocationTaskLink() {
	address, stop, err := startAllocationServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	// the link is incremental: it leaves the allocation's other links alone,
	// unlike AllocationUpsert.LinkedTaskIDs, which replaces the whole set
	_, err = projects.AllocationTaskLink(ctx, engine, projects.NewAllocationTaskLinkRequest(12345, 12346))
	if err != nil {
		fmt.Printf("failed to link task to allocation: %s", err)
	} else {
		fmt.Println("task linked to allocation!")
	}

	// Output: task linked to allocation!
}

func ExampleAllocationTaskUnlink() {
	address, stop, err := startAllocationServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.AllocationTaskUnlink(ctx, engine, projects.NewAllocationTaskUnlinkRequest(12345, 12346))
	if err != nil {
		fmt.Printf("failed to unlink task from allocation: %s", err)
	} else {
		fmt.Println("task unlinked from allocation!")
	}

	// Output: task unlinked from allocation!
}

func ExampleAllocationGet() {
	address, stop, err := startAllocationServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	allocationResponse, err := projects.AllocationGet(ctx, engine, projects.NewAllocationGetRequest(12345))
	if err != nil {
		fmt.Printf("failed to retrieve allocation: %s", err)
	} else {
		fmt.Printf("retrieved allocation with identifier %d\n", allocationResponse.Allocation.ID)
	}

	// Output: retrieved allocation with identifier 12345
}

func ExampleAllocationList() {
	address, stop, err := startAllocationServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	allocationsRequest := projects.NewAllocationListRequest()

	// the endpoint windows the results to today through thirty days from today
	// when neither bound is set, and says nothing about having done so, so set
	// both whenever a period is meant
	start := twapi.Date(time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC))
	end := twapi.Date(time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC))
	allocationsRequest.Filters.StartDate = &start
	allocationsRequest.Filters.EndDate = &end
	allocationsRequest.Filters.ProjectIDs = []int64{777}

	allocationsResponse, err := projects.AllocationList(ctx, engine, allocationsRequest)
	if err != nil {
		fmt.Printf("failed to list allocations: %s", err)
	} else {
		for _, allocation := range allocationsResponse.Allocations {
			fmt.Printf("retrieved allocation with identifier %d\n", allocation.ID)
		}
	}

	// Output:
	// retrieved allocation with identifier 12345
	// retrieved allocation with identifier 12346
}

func startAllocationServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /projects/api/v3/allocations", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"allocation":{"id":12345}}`)
	})
	mux.HandleFunc("PATCH /projects/api/v3/allocations/{id}", func(w http.ResponseWriter, r *http.Request) {
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
		_, _ = fmt.Fprintln(w, `{"allocation":{"id":12345}}`)
	})
	mux.HandleFunc("DELETE /projects/api/v3/allocations/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("PUT /projects/api/v3/allocations/{id}/restore", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"allocation":{"id":12345,"status":"active"}}`)
	})
	mux.HandleFunc("POST /projects/api/v3/allocations/{id}/link/{taskId}",
		func(w http.ResponseWriter, r *http.Request) {
			if r.PathValue("id") != "12345" || r.PathValue("taskId") != "12346" {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	mux.HandleFunc("POST /projects/api/v3/allocations/{id}/unlink/{taskId}",
		func(w http.ResponseWriter, r *http.Request) {
			if r.PathValue("id") != "12345" || r.PathValue("taskId") != "12346" {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	mux.HandleFunc("GET /projects/api/v3/allocations/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"allocation":{"id":12345}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/allocations", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"allocations":[{"id":12345},{"id":12346}]}`)
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
