package projects_test

import (
	"context"
	"fmt"
	"math/rand"
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
