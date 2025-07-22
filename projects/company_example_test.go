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

func ExampleCompanyCreate() {
	address, stop, err := startCompanyServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	companyRequest := projects.NewCompanyCreateRequest("Test Company")
	companyRequest.Profile = twapi.Ptr("A company created via the API.")

	companyResponse, err := projects.CompanyCreate(ctx, engine, companyRequest)
	if err != nil {
		fmt.Printf("failed to create company: %s", err)
	} else {
		fmt.Printf("created company with identifier %d\n", companyResponse.Company.ID)
	}

	// Output: created company with identifier 12345
}

func ExampleCompanyUpdate() {
	address, stop, err := startCompanyServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	companyRequest := projects.NewCompanyUpdateRequest(12345)
	companyRequest.Profile = twapi.Ptr("Updated profile")

	_, err = projects.CompanyUpdate(ctx, engine, companyRequest)
	if err != nil {
		fmt.Printf("failed to update company: %s", err)
	} else {
		fmt.Println("company updated!")
	}

	// Output: company updated!
}

func ExampleCompanyDelete() {
	address, stop, err := startCompanyServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.CompanyDelete(ctx, engine, projects.NewCompanyDeleteRequest(12345))
	if err != nil {
		fmt.Printf("failed to delete company: %s", err)
	} else {
		fmt.Println("company deleted!")
	}

	// Output: company deleted!
}

func ExampleCompanyGet() {
	address, stop, err := startCompanyServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	companyResponse, err := projects.CompanyGet(ctx, engine, projects.NewCompanyGetRequest(12345))
	if err != nil {
		fmt.Printf("failed to retrieve company: %s", err)
	} else {
		fmt.Printf("retrieved company with identifier %d\n", companyResponse.Company.ID)
	}

	// Output: retrieved company with identifier 12345
}

func ExampleCompanyList() {
	address, stop, err := startCompanyServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	companiesRequest := projects.NewCompanyListRequest()
	companiesRequest.Filters.SearchTerm = "John"

	companiesResponse, err := projects.CompanyList(ctx, engine, companiesRequest)
	if err != nil {
		fmt.Printf("failed to list companies: %s", err)
	} else {
		for _, company := range companiesResponse.Companies {
			fmt.Printf("retrieved company with identifier %d\n", company.ID)
		}
	}

	// Output: retrieved company with identifier 12345
	// retrieved company with identifier 12346
}

func startCompanyServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /projects/api/v3/companies", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"company":{"id":12345}}`)
	})
	mux.HandleFunc("PATCH /projects/api/v3/companies/{id}", func(w http.ResponseWriter, r *http.Request) {
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
		_, _ = fmt.Fprintln(w, `{"company":{"id":12345}}`)
	})
	mux.HandleFunc("DELETE /projects/api/v3/companies/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /projects/api/v3/companies/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"company":{"id":12345}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/companies", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"companies":[{"id":12345},{"id":12346}]}`)
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
