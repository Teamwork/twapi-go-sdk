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

func ExampleRateInstallationUserList() {
	address, stop, err := startRatesServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	req := projects.NewRateInstallationUserListRequest()
	// Configure pagination
	req.Filters.Page = 1
	req.Filters.PageSize = 10

	resp, err := projects.RateInstallationUserList(ctx, engine, req)
	if err != nil {
		fmt.Printf("failed to list installation user rates: %s", err)
	} else {
		fmt.Printf("retrieved %d installation user rate(s)\n", len(resp.UserRates))
	}

	// Output: retrieved 1 installation user rate(s)
}

func ExampleRateInstallationUserBulkUpdate() {
	address, stop, err := startRatesServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	var rate int64 = 6000
	req := projects.NewRateInstallationUserBulkUpdateRequest(&rate) // Rate (cents)
	req.IDs = []int64{12345}                                        // Update specific user IDs

	_, err = projects.RateInstallationUserBulkUpdate(ctx, engine, req)
	if err != nil {
		fmt.Printf("failed to bulk update installation user rates: %s", err)
	} else {
		fmt.Printf("bulk updated installation user rates\n")
	}

	// Output: bulk updated installation user rates
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

func ExampleRateProjectAndUsersUpdate() {
	address, stop, err := startRatesServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	var projectRate int64 = 8000
	req := projects.NewRateProjectAndUsersUpdateRequest(67890, projectRate) // Project ID, Rate (cents)
	// Add user rates as exceptions
	var userRate int64 = 9000
	req.UserRates = []projects.ProjectUserRateRequest{
		{User: twapi.Relationship{ID: 12345}, UserRate: userRate},
	}

	_, err = projects.RateProjectAndUsersUpdate(ctx, engine, req)
	if err != nil {
		fmt.Printf("failed to update project and users rates: %s", err)
	} else {
		fmt.Printf("updated project and users rates with identifier %d\n", 67890)
	}

	// Output: updated project and users rates with identifier 67890
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

func ExampleRateProjectUserList() {
	address, stop, err := startRatesServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	req := projects.NewRateProjectUserListRequest(67890) // Project ID
	// Configure pagination
	req.Filters.Page = 1
	req.Filters.PageSize = 10

	resp, err := projects.RateProjectUserList(ctx, engine, req)
	if err != nil {
		fmt.Printf("failed to list project user rates: %s", err)
	} else {
		fmt.Printf("retrieved %d project user rate(s)\n", len(resp.UserRates))
	}

	// Output: retrieved 2 project user rate(s)
}

func ExampleRateProjectUserHistoryGet() {
	address, stop, err := startRatesServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	req := projects.NewRateProjectUserHistoryGetRequest(67890, 12345) // Project ID, User ID
	// Configure optional filters
	req.Filters.SearchTerm = "rate"

	resp, err := projects.RateProjectUserHistoryGet(ctx, engine, req)
	if err != nil {
		fmt.Printf("failed to get project user rate history: %s", err)
	} else {
		fmt.Printf("retrieved %d rate history entries for user %d\n", len(resp.UserRateHistory), 12345)
	}

	// Output: retrieved 2 rate history entries for user 12345
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
		_, _ = fmt.Fprintln(w, `{"installationRate":5000,"projectRates":[{"id":123,"rate":7500}],`+
			`"installationRates":{"1":{"rate":5000,"currency":{"id":1,"code":"USD"}}},"userCost":4000}`)
	})

	// GET /projects/api/v3/rates/installation/users.json
	mux.HandleFunc("GET /projects/api/v3/rates/installation/users", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"userRates":[{"user":{"id":12345},"rate":5000}],`+
			`"meta":{"page":{"count":1,"hasMore":false}}}`)
	})

	// GET /projects/api/v3/rates/installation/users/{id}.json
	mux.HandleFunc("GET /projects/api/v3/rates/installation/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"userRate":5000,"userRates":{"1":{"rate":5000,"currency":{"id":1,"code":"USD"}}}}`)
	})

	// PUT /projects/api/v3/rates/installation/users/{id}.json
	mux.HandleFunc("PUT /projects/api/v3/rates/installation/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	// PUT /projects/api/v3/rates/installation/users/bulk/update.json
	mux.HandleFunc("PUT /projects/api/v3/rates/installation/users/bulk/update",
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintln(w, `{"all":false,"ids":[12345],"excludeIds":[],"rate":6000}`)
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
		w.WriteHeader(http.StatusNoContent)
	})

	// PUT /projects/api/v3/rates/projects/{id}/actions/update
	mux.HandleFunc("PUT /projects/api/v3/rates/projects/{id}/actions/update",
		func(w http.ResponseWriter, r *http.Request) {
			if r.PathValue("id") != "67890" {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})

	// GET /projects/api/v3/rates/projects/{id}/users
	mux.HandleFunc("GET /projects/api/v3/rates/projects/{id}/users", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "67890" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"userRates":[{"user":{"id":12345},"effectiveRate":8500},`+
			`{"user":{"id":67891},"effectiveRate":9000}],"meta":{"page":{"count":2,"hasMore":false}}}`)
	})

	// GET /projects/api/v3/rates/projects/{id}/users/{userId}.json
	mux.HandleFunc("GET /projects/api/v3/rates/projects/{projectId}/users/{userId}",
		func(w http.ResponseWriter, r *http.Request) {
			if r.PathValue("projectId") != "67890" || r.PathValue("userId") != "12345" {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintln(w, `{"rate":{"amount":85.00,"currency":{"id":1,"code":"USD"}},"userRate":8500}`)
		})

	// PUT /projects/api/v3/rates/projects/{id}/users/{userId}.json
	mux.HandleFunc("PUT /projects/api/v3/rates/projects/{projectId}/users/{userId}",
		func(w http.ResponseWriter, r *http.Request) {
			if r.PathValue("projectId") != "67890" || r.PathValue("userId") != "12345" {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusCreated)
		})

	// GET /projects/api/v3/rates/projects/{projectId}/users/{userId}/history
	mux.HandleFunc("GET /projects/api/v3/rates/projects/{projectId}/users/{userId}/history",
		func(w http.ResponseWriter, r *http.Request) {
			if r.PathValue("projectId") != "67890" || r.PathValue("userId") != "12345" {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintln(w, `{"userRateHistory":[{"rate":8500,"fromDate":"2023-01-01T00:00:00Z"},`+
				`{"rate":9000,"fromDate":"2023-06-01T00:00:00Z"}],"meta":{"page":{"hasMore":false}}}`)
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
