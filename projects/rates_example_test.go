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

func ExampleRateUserGet() {
	address, stop, err := startRatesServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	req := projects.NewRateUserGetRequest(12345) // User ID
	// Configure optional filters
	req.Filters.IncludeUserCost = true
	// Include supported related resources via enum
	req.Filters.Include = []projects.RateUserGetRequestSideload{
		projects.RateSideloadProjects,
	}

	_, err = projects.RateUserGet(ctx, engine, req)
	if err != nil {
		fmt.Printf("failed to get user rates: %s", err)
	} else {
		fmt.Printf("retrieved user rates with identifier %d\n", 12345)
	}

	// Output: retrieved user rates with identifier 12345
}

func ExampleRateInstallationUserGet() {
	address, stop, err := startRatesServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	req := projects.NewRateInstallationUserGetRequest(12345) // User ID
	req.Filters.Include = []projects.RateInstallationUserGetRequestSideload{
		projects.RateInstallationUserGetRequestSideloadCurrencies,
	}

	_, err = projects.RateInstallationUserGet(ctx, engine, req)
	if err != nil {
		fmt.Printf("failed to get installation user rate: %s", err)
	} else {
		fmt.Printf("retrieved installation user rate with identifier %d\n", 12345)
	}

	// Output: retrieved installation user rate with identifier 12345
}

func ExampleRateInstallationUserUpdate() {
	address, stop, err := startRatesServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	var rate int64 = 5000
	req := projects.NewRateInstallationUserUpdateRequest(12345, &rate) // User ID, Rate (cents)
	_, err = projects.RateInstallationUserUpdate(ctx, engine, req)
	if err != nil {
		fmt.Printf("failed to update user rate: %s", err)
	} else {
		fmt.Printf("updated installation user rate with identifier %d\n", 12345)
	}

	// Output: updated installation user rate with identifier 12345
}

func ExampleRateProjectGet() {
	address, stop, err := startRatesServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	req := projects.NewRateProjectGetRequest(67890) // Project ID
	req.Filters.Include = []projects.RateProjectGetRequestSideload{
		projects.RateProjectGetRequestSideloadCurrencies,
	}

	_, err = projects.RateProjectGet(ctx, engine, req)
	if err != nil {
		fmt.Printf("failed to get project rate: %s", err)
	} else {
		fmt.Printf("retrieved project rate with identifier %d\n", 67890)
	}

	// Output: retrieved project rate with identifier 67890
}

func ExampleRateProjectUpdate() {
	address, stop, err := startRatesServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	var rate int64 = 7500
	req := projects.NewRateProjectUpdateRequest(67890, &rate) // Project ID, Rate (cents)
	_, err = projects.RateProjectUpdate(ctx, engine, req)
	if err != nil {
		fmt.Printf("failed to update project rate: %s", err)
	} else {
		fmt.Printf("updated project rate with identifier %d\n", 67890)
	}

	// Output: updated project rate with identifier 67890
}

func ExampleRateProjectUserGet() {
	address, stop, err := startRatesServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	req := projects.NewRateProjectUserGetRequest(67890, 12345) // Project ID, User ID
	req.Filters.Include = []projects.RateProjectUserGetRequestSideload{
		projects.RateProjectUserGetRequestSideloadCurrencies,
	}

	_, err = projects.RateProjectUserGet(ctx, engine, req)
	if err != nil {
		fmt.Printf("failed to get project user rate: %s", err)
	} else {
		fmt.Printf("retrieved project user rate with identifier %d\n", 12345)
	}

	// Output: retrieved project user rate with identifier 12345
}

func ExampleRateProjectUserUpdate() {
	address, stop, err := startRatesServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	var rate int64 = 8500
	req := projects.NewRateProjectUserUpdateRequest(67890, 12345, &rate) // Project ID, User ID, Rate (cents)
	_, err = projects.RateProjectUserUpdate(ctx, engine, req)
	if err != nil {
		fmt.Printf("failed to update user rate: %s", err)
	} else {
		fmt.Printf("updated project user rate with identifier %d\n", 12345)
	}

	// Output: updated project user rate with identifier 12345
}

func startRatesServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()

	// GET /projects/api/v3/people/{id}/rates
	mux.HandleFunc("GET /projects/api/v3/people/{id}/rates", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"installationRate":5000,"projectRates":[{"id":123,"rate":7500}],"installationRates":[{"rate":5000,"currency":{"id":1,"code":"USD"}}],"userCost":4000}`)
	})

	// GET /projects/api/v3/rates/installation/users.json
	mux.HandleFunc("GET /projects/api/v3/rates/installation/users", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"userRates":[{"user":{"id":12345},"rate":5000}],"meta":{"page":{"count":1,"hasMore":false}}}`)
	})

	// GET /projects/api/v3/rates/installation/users/{id}.json
	mux.HandleFunc("GET /projects/api/v3/rates/installation/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"userRate":5000,"userRates":[{"rate":5000,"currency":{"id":1,"code":"USD"}}]}`)
	})

	// PUT /projects/api/v3/rates/installation/users/{id}.json
	mux.HandleFunc("PUT /projects/api/v3/rates/installation/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK"}`)
	})

	// GET /projects/api/v3/rates/projects/{id}.json
	mux.HandleFunc("GET /projects/api/v3/rates/projects/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "67890" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"projectRate":7500,"rate":{"amount":75.00,"currency":{"id":1,"code":"USD"}}}`)
	})

	// PUT /projects/api/v3/rates/projects/{id}.json
	mux.HandleFunc("PUT /projects/api/v3/rates/projects/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "67890" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK"}`)
	})

	// GET /projects/api/v3/rates/projects/{id}/users/{userId}.json
	mux.HandleFunc("GET /projects/api/v3/rates/projects/{projectId}/users/{userId}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("projectId") != "67890" || r.PathValue("userId") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"userRate":{"amount":85.00,"currency":{"id":1,"code":"USD"}},"rate":8500}`)
	})

	// PUT /projects/api/v3/rates/projects/{id}/users/{userId}.json
	mux.HandleFunc("PUT /projects/api/v3/rates/projects/{projectId}/users/{userId}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("projectId") != "67890" || r.PathValue("userId") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"userRate":8500}`)
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
