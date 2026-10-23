package projects_test

import (
	"context"
	"fmt"
	"math/rand"
	"slices"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestCompanyCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name  string
		input projects.CompanyCreateRequest
	}{{
		name:  "only required fields",
		input: projects.NewCompanyCreateRequest(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
	}, {
		name: "all fields",
		input: projects.CompanyCreateRequest{
			Name:        fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
			AddressOne:  new("123 Main St"),
			AddressTwo:  new("Apt. 456"),
			City:        new("Cork"),
			CountryCode: new("IR"),
			EmailOne:    new("test1@company.com"),
			EmailTwo:    new("test2@company.com"),
			EmailThree:  new("test3@company.com"),
			Fax:         new("123-456-7890"),
			Phone:       new("123-456-7890"),
			Profile:     new("This is a test company profile."),
			State:       new("Cork"),
			Website:     new("https://www.example.com"),
			Zip:         new("12345"),
			ManagerID:   &testResources.UserID,
			IndustryID:  new(int64(1)), // Web Development Agency,
			TagIDs:      []int64{testResources.TagID},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			company, err := projects.CompanyCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx = context.Background() // t.Context is always canceled in cleanup
				_, err := projects.CompanyDelete(ctx, engine, projects.NewCompanyDeleteRequest(company.Company.ID))
				if err != nil {
					t.Errorf("failed to delete company after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if company.Company.ID == 0 {
				t.Error("expected a valid company ID but got 0")
			}
		})
	}
}

func TestCompanyBulkCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	t.Cleanup(cancel)

	req := projects.NewCompanyBulkCreateRequest(
		projects.NewCompanyCreateRequest(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
		projects.NewCompanyCreateRequest(""), // rejected: a company needs a name
		projects.NewCompanyCreateRequest(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
	)
	req.PartialSuccess = true

	response, err := projects.CompanyBulkCreate(ctx, engine, req)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	t.Cleanup(func() {
		ctx := context.Background() // t.Context is always canceled in cleanup
		for _, company := range response.Companies {
			_, err := projects.CompanyDelete(ctx, engine, projects.NewCompanyDeleteRequest(company.ID))
			if err != nil {
				t.Errorf("failed to delete company after test: %s", err)
			}
		}
	})

	indexes, err := response.CompanyIndexes(len(req.Companies))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !slices.Equal(indexes, []int{0, 2}) {
		t.Errorf("expected companies at positions [0 2], got %v", indexes)
	}
	for _, company := range response.Companies {
		if company.ID == 0 {
			t.Error("expected a valid company ID but got 0")
		}
	}
	if len(response.Failed) != 1 || response.Failed[0].Index != 1 || len(response.Failed[0].Errors) == 0 {
		t.Errorf("expected position 1 to fail with a reason, got %+v", response.Failed)
	}
}

func TestCompanyBulkCreateResponseCompanyIndexes(t *testing.T) {
	tests := []struct {
		name      string
		response  projects.CompanyBulkCreateResponse
		requested int
		want      []int
		wantErr   bool
	}{{
		name:      "every company created",
		response:  projects.CompanyBulkCreateResponse{Companies: make([]projects.Company, 3)},
		requested: 3,
		want:      []int{0, 1, 2},
	}, {
		name: "failures are skipped",
		response: projects.CompanyBulkCreateResponse{
			Companies: make([]projects.Company, 2),
			Failed:    []projects.CompanyBulkCreateFailure{{Index: 0}, {Index: 2}},
		},
		requested: 4,
		want:      []int{1, 3},
	}, {
		name:      "a company unaccounted for",
		response:  projects.CompanyBulkCreateResponse{Companies: make([]projects.Company, 1)},
		requested: 2,
		wantErr:   true,
	}, {
		name: "a failure outside the request",
		response: projects.CompanyBulkCreateResponse{
			Companies: make([]projects.Company, 1),
			Failed:    []projects.CompanyBulkCreateFailure{{Index: 5}},
		},
		requested: 2,
		wantErr:   true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.response.CompanyIndexes(tt.requested)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected an error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestCompanyUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	companyID, companyCleanup, err := createCompany(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(companyCleanup)

	tests := []struct {
		name  string
		input projects.CompanyUpdateRequest
	}{{
		name: "all fields",
		input: projects.CompanyUpdateRequest{
			Path: projects.CompanyUpdateRequestPath{
				ID: companyID,
			},
			Name:        new(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
			AddressOne:  new("123 Main St"),
			AddressTwo:  new("Apt. 456"),
			City:        new("Cork"),
			CountryCode: new("IR"),
			EmailOne:    new("test1@company.com"),
			EmailTwo:    new("test2@company.com"),
			EmailThree:  new("test3@company.com"),
			Fax:         new("123-456-7890"),
			Phone:       new("123-456-7890"),
			Profile:     new("This is a test company profile."),
			State:       new("Cork"),
			Website:     new("https://www.example.com"),
			Zip:         new("12345"),
			ManagerID:   &testResources.UserID,
			IndustryID:  new(int64(1)), // Web Development Agency,
			TagIDs:      []int64{testResources.TagID},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.CompanyUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestCompanyDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	companyID, _, err := createCompany(t)
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.CompanyDelete(ctx, engine, projects.NewCompanyDeleteRequest(companyID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestCompanyGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	companyID, companyCleanup, err := createCompany(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(companyCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.CompanyGet(ctx, engine, projects.NewCompanyGetRequest(companyID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestCompanyList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	_, companyCleanup, err := createCompany(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(companyCleanup)

	tests := []struct {
		name  string
		input projects.CompanyListRequest
	}{{
		name: "all companies",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.CompanyList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
