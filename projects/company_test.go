package projects_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestCompanyCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	userID, userCleanup, err := createUser(t)
	if err != nil {
		t.Fatal(err)
	}
	defer userCleanup()

	tagID, tagCleanup, err := createTag(t)
	if err != nil {
		t.Fatal(err)
	}
	defer tagCleanup()

	tests := []struct {
		name  string
		input projects.CompanyCreateRequest
	}{{
		name:  "only required fields",
		input: projects.NewCompanyCreateRequest(fmt.Sprintf("Test Company %d%d", time.Now().UnixNano(), rand.Intn(100))),
	}, {
		name: "all fields",
		input: projects.CompanyCreateRequest{
			Name:        fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
			AddressOne:  twapi.Ptr("123 Main St"),
			AddressTwo:  twapi.Ptr("Apt. 456"),
			City:        twapi.Ptr("Cork"),
			CountryCode: twapi.Ptr("IR"),
			EmailOne:    twapi.Ptr("test1@company.com"),
			EmailTwo:    twapi.Ptr("test2@company.com"),
			EmailThree:  twapi.Ptr("test3@company.com"),
			Fax:         twapi.Ptr("123-456-7890"),
			Phone:       twapi.Ptr("123-456-7890"),
			Profile:     twapi.Ptr("This is a test company profile."),
			State:       twapi.Ptr("Cork"),
			Website:     twapi.Ptr("https://www.example.com"),
			Zip:         twapi.Ptr("12345"),
			ManagerID:   &userID,
			IndustryID:  twapi.Ptr(int64(1)), // Web Development Agency,
			TagIDs:      []int64{tagID},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			company, err := projects.CompanyCreate(ctx, engine, tt.input)
			defer func() {
				if err != nil {
					return
				}
				_, err := projects.CompanyDelete(ctx, engine, projects.NewCompanyDeleteRequest(company.Company.ID))
				if err != nil {
					t.Errorf("failed to delete company after test: %s", err)
				}
			}()

			if err != nil {
				t.Errorf("unexpected error: %s", err)
				return
			}
			if company.Company.ID == 0 {
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
	defer companyCleanup()

	userID, userCleanup, err := createUser(t)
	if err != nil {
		t.Fatal(err)
	}
	defer userCleanup()

	tagID, tagCleanup, err := createTag(t)
	if err != nil {
		t.Fatal(err)
	}
	defer tagCleanup()

	tests := []struct {
		name  string
		input projects.CompanyUpdateRequest
	}{{
		name: "all fields",
		input: projects.CompanyUpdateRequest{
			Path: projects.CompanyUpdateRequestPath{
				ID: companyID,
			},
			Name:        twapi.Ptr(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
			AddressOne:  twapi.Ptr("123 Main St"),
			AddressTwo:  twapi.Ptr("Apt. 456"),
			City:        twapi.Ptr("Cork"),
			CountryCode: twapi.Ptr("IR"),
			EmailOne:    twapi.Ptr("test1@company.com"),
			EmailTwo:    twapi.Ptr("test2@company.com"),
			EmailThree:  twapi.Ptr("test3@company.com"),
			Fax:         twapi.Ptr("123-456-7890"),
			Phone:       twapi.Ptr("123-456-7890"),
			Profile:     twapi.Ptr("This is a test company profile."),
			State:       twapi.Ptr("Cork"),
			Website:     twapi.Ptr("https://www.example.com"),
			Zip:         twapi.Ptr("12345"),
			ManagerID:   &userID,
			IndustryID:  twapi.Ptr(int64(1)), // Web Development Agency,
			TagIDs:      []int64{tagID},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			if _, err := projects.CompanyUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
				return
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

	tests := []struct {
		name          string
		input         projects.CompanyDeleteRequest
		expectedError bool
	}{{
		name:  "it should delete a company with valid input",
		input: projects.NewCompanyDeleteRequest(companyID),
	}, {
		name:          "it should fail to delete an unknown company",
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.CompanyDelete(ctx, engine, tt.input)
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %s", err)
				return
			}
		})
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
	defer companyCleanup()

	tests := []struct {
		name          string
		input         projects.CompanyGetRequest
		expectedError bool
	}{{
		name:  "it should retrieve a company with valid input",
		input: projects.NewCompanyGetRequest(companyID),
	}, {
		name:          "it should fail to retrieve an unknown company",
		input:         projects.NewCompanyGetRequest(999999999), // assuming this ID does not exist
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.CompanyGet(ctx, engine, tt.input)
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %s", err)
				return
			}
		})
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
	defer companyCleanup()

	tests := []struct {
		name          string
		input         projects.CompanyListRequest
		expectedError bool
	}{{
		name: "it should list companies",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.CompanyList(ctx, engine, tt.input)
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %s", err)
				return
			}
		})
	}
}
