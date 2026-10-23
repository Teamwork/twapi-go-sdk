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

func ExampleTimelogCreate() {
	address, stop, err := startTimelogServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	timelogRequest := projects.NewTimelogCreateRequestInTask(777, time.Now(), 30*time.Minute)
	timelogRequest.Description = new("This is a new timelog created via the API.")

	timelogResponse, err := projects.TimelogCreate(ctx, engine, timelogRequest)
	if err != nil {
		fmt.Printf("failed to create timelog: %s", err)
	} else {
		fmt.Printf("created timelog with identifier %d\n", timelogResponse.Timelog.ID)
	}

	// Output: created timelog with identifier 12345
}

func ExampleTimelogUpdate() {
	address, stop, err := startTimelogServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	timelogRequest := projects.NewTimelogUpdateRequest(12345)
	timelogRequest.Description = new("This is an updated description.")

	_, err = projects.TimelogUpdate(ctx, engine, timelogRequest)
	if err != nil {
		fmt.Printf("failed to update timelog: %s", err)
	} else {
		fmt.Println("timelog updated!")
	}

	// Output: timelog updated!
}

func ExampleTimelogDelete() {
	address, stop, err := startTimelogServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.TimelogDelete(ctx, engine, projects.NewTimelogDeleteRequest(12345))
	if err != nil {
		fmt.Printf("failed to delete timelog: %s", err)
	} else {
		fmt.Println("timelog deleted!")
	}

	// Output: timelog deleted!
}

func ExampleTimelogGet() {
	address, stop, err := startTimelogServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	timelogResponse, err := projects.TimelogGet(ctx, engine, projects.NewTimelogGetRequest(12345))
	if err != nil {
		fmt.Printf("failed to retrieve timelog: %s", err)
	} else {
		fmt.Printf("retrieved timelog with identifier %d\n", timelogResponse.Timelog.ID)
	}

	// Output: retrieved timelog with identifier 12345
}

func ExampleTimelogList() {
	address, stop, err := startTimelogServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	timelogsRequest := projects.NewTimelogListRequest()
	timelogsResponse, err := projects.TimelogList(ctx, engine, timelogsRequest)
	if err != nil {
		fmt.Printf("failed to list timelogs: %s", err)
	} else {
		for _, timelog := range timelogsResponse.Timelogs {
			fmt.Printf("retrieved timelog with identifier %d\n", timelog.ID)
		}
	}

	// Output: retrieved timelog with identifier 12345
	// retrieved timelog with identifier 12346
}

func ExampleTimelogBulkMove() {
	address, stop, err := startTimelogServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	moveRequest := projects.NewTimelogBulkMoveRequest(777, 778, []int64{12345, 12346})
	moveResponse, err := projects.TimelogBulkMove(ctx, engine, moveRequest)
	if err != nil {
		fmt.Printf("failed to move timelogs: %s", err)
	} else {
		fmt.Printf("moved %d timelogs\n", moveResponse.MovedCount)
	}

	// Output: moved 2 timelogs
}

func ExampleTimelogBulkBillableUpdate() {
	address, stop, err := startTimelogServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	// make the non-billable, not yet invoiced timelogs of project 777 billable
	updateRequest := projects.NewTimelogBulkBillableUpdateRequest(true, 777)
	updateRequest.BillableType = projects.TimelogBillableTypeNonBillable
	updateRequest.InvoicedType = projects.TimelogInvoicedTypeNonInvoiced

	updateResponse, err := projects.TimelogBulkBillableUpdate(ctx, engine, updateRequest)
	if err != nil {
		fmt.Printf("failed to update timelogs: %s", err)
	} else if updateResponse.Count != nil {
		fmt.Printf("%d timelogs are now billable\n", *updateResponse.Count)
	}

	// Output: 3 timelogs are now billable
}

func startTimelogServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /projects/api/v3/tasks/{id}/time", func(w http.ResponseWriter, r *http.Request) {
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
		_, _ = fmt.Fprintln(w, `{"timelog":{"id":12345}}`)
	})
	mux.HandleFunc("PATCH /projects/api/v3/time/{id}", func(w http.ResponseWriter, r *http.Request) {
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
		_, _ = fmt.Fprintln(w, `{}`)
	})
	mux.HandleFunc("DELETE /projects/api/v3/time/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /projects/api/v3/time/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"timelog":{"id":12345}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/time", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"timelogs":[{"id":12345},{"id":12346}]}`)
	})

	mux.HandleFunc("PUT /projects/{id}/time_entries/move", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		if r.PathValue("id") != "777" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK","numberOfMovedTimelogs":"2"}`)
	})
	mux.HandleFunc("PUT /time_entries", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("count", "3")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK"}`)
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
