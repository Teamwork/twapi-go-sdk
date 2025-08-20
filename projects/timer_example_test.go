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

func ExampleTimerCreate() {
	address, stop, err := startTimerServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	timerRequest := projects.NewTimerCreateRequest(777)
	timerRequest.Description = twapi.Ptr("This is a new timer created via the API.")

	timerResponse, err := projects.TimerCreate(ctx, engine, timerRequest)
	if err != nil {
		fmt.Printf("failed to create timer: %s", err)
	} else {
		fmt.Printf("created timer with identifier %d\n", timerResponse.Timer.ID)
	}

	// Output: created timer with identifier 12345
}

func ExampleTimerUpdate() {
	address, stop, err := startTimerServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	timerRequest := projects.NewTimerUpdateRequest(12345)
	timerRequest.Description = twapi.Ptr("This is an updated description.")

	_, err = projects.TimerUpdate(ctx, engine, timerRequest)
	if err != nil {
		fmt.Printf("failed to update timer: %s", err)
	} else {
		fmt.Println("timer updated!")
	}

	// Output: timer updated!
}

func ExampleTimerPause() {
	address, stop, err := startTimerServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	timerRequest := projects.NewTimerPauseRequest(12345)

	_, err = projects.TimerPause(ctx, engine, timerRequest)
	if err != nil {
		fmt.Printf("failed to pause timer: %s", err)
	} else {
		fmt.Println("timer paused!")
	}

	// Output: timer paused!
}

func ExampleTimerResume() {
	address, stop, err := startTimerServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	timerRequest := projects.NewTimerResumeRequest(12345)

	_, err = projects.TimerResume(ctx, engine, timerRequest)
	if err != nil {
		fmt.Printf("failed to resume timer: %s", err)
	} else {
		fmt.Println("timer resumed!")
	}

	// Output: timer resumed!
}

func ExampleTimerComplete() {
	address, stop, err := startTimerServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	timerRequest := projects.NewTimerCompleteRequest(12345)

	_, err = projects.TimerComplete(ctx, engine, timerRequest)
	if err != nil {
		fmt.Printf("failed to complete timer: %s", err)
	} else {
		fmt.Println("timer completed!")
	}

	// Output: timer completed!
}

func ExampleTimerDelete() {
	address, stop, err := startTimerServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.TimerDelete(ctx, engine, projects.NewTimerDeleteRequest(12345))
	if err != nil {
		fmt.Printf("failed to delete timer: %s", err)
	} else {
		fmt.Println("timer deleted!")
	}

	// Output: timer deleted!
}

func ExampleTimerGet() {
	address, stop, err := startTimerServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	timerResponse, err := projects.TimerGet(ctx, engine, projects.NewTimerGetRequest(12345))
	if err != nil {
		fmt.Printf("failed to retrieve timer: %s", err)
	} else {
		fmt.Printf("retrieved timer with identifier %d\n", timerResponse.Timer.ID)
	}

	// Output: retrieved timer with identifier 12345
}

func ExampleTimerList() {
	address, stop, err := startTimerServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	timersRequest := projects.NewTimerListRequest()
	timersResponse, err := projects.TimerList(ctx, engine, timersRequest)
	if err != nil {
		fmt.Printf("failed to list timers: %s", err)
	} else {
		for _, timer := range timersResponse.Timers {
			fmt.Printf("retrieved timer with identifier %d\n", timer.ID)
		}
	}

	// Output: retrieved timer with identifier 12345
	// retrieved timer with identifier 12346
}

func startTimerServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /projects/api/v3/me/timers", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"timer":{"id":12345}}`)
	})
	mux.HandleFunc("PUT /projects/api/v3/me/timers/{id}", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("PUT /projects/api/v3/me/timers/{id}/pause", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{}`)
	})
	mux.HandleFunc("PUT /projects/api/v3/me/timers/{id}/resume", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{}`)
	})
	mux.HandleFunc("PUT /projects/api/v3/me/timers/{id}/complete", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{}`)
	})
	mux.HandleFunc("DELETE /projects/api/v3/me/timers/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /projects/api/v3/timers/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"timer":{"id":12345}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/timers", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"timers":[{"id":12345},{"id":12346}]}`)
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
