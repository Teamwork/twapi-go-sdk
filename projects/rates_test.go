package projects_test

import (
	"context"
	"math/rand"
	"net/http"
	"strings"
	"testing"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestRateUserGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name   string
		userID int64
		req    func(userID int64) projects.RateUserGetRequest
	}{{
		name:   "get user rates for test user",
		userID: testResources.UserID,
		req:    projects.NewRateUserGetRequest,
	}, {
		name:   "get user rates with include installation rate",
		userID: testResources.UserID,
		req: func(userID int64) projects.RateUserGetRequest {
			req := projects.NewRateUserGetRequest(userID)
			req.Filters.IncludeInstallationRate = true
			return req
		},
	}, {
		name:   "get user rates with include user cost",
		userID: testResources.UserID,
		req: func(userID int64) projects.RateUserGetRequest {
			req := projects.NewRateUserGetRequest(userID)
			req.Filters.IncludeUserCost = true
			return req
		},
	}, {
		name:   "get user rates with include archived projects",
		userID: testResources.UserID,
		req: func(userID int64) projects.RateUserGetRequest {
			req := projects.NewRateUserGetRequest(userID)
			req.Filters.IncludeArchivedProjects = true
			return req
		},
	}, {
		name:   "get user rates with include deleted projects",
		userID: testResources.UserID,
		req: func(userID int64) projects.RateUserGetRequest {
			req := projects.NewRateUserGetRequest(userID)
			req.Filters.IncludeDeletedProjects = true
			return req
		},
	}, {
		name:   "get user rates with projects sideload",
		userID: testResources.UserID,
		req: func(userID int64) projects.RateUserGetRequest {
			req := projects.NewRateUserGetRequest(userID)
			req.Filters.Include = []projects.RateUserGetRequestSideload{
				projects.RateSideloadProjects,
			}
			return req
		},
	}, {
		name:   "get user rates with pagination",
		userID: testResources.UserID,
		req: func(userID int64) projects.RateUserGetRequest {
			req := projects.NewRateUserGetRequest(userID)
			req.Filters.Page = 1
			req.Filters.PageSize = 20
			return req
		},
	}, {
		name:   "get user rates with all filters",
		userID: testResources.UserID,
		req: func(userID int64) projects.RateUserGetRequest {
			req := projects.NewRateUserGetRequest(userID)
			req.Filters.IncludeInstallationRate = true
			req.Filters.IncludeUserCost = true
			req.Filters.IncludeArchivedProjects = true
			req.Filters.IncludeDeletedProjects = true
			req.Filters.Include = []projects.RateUserGetRequestSideload{
				projects.RateSideloadProjects,
			}
			req.Filters.Page = 1
			req.Filters.PageSize = 10
			return req
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			req := tt.req(tt.userID)
			resp, err := projects.RateUserGet(ctx, engine, req)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if resp == nil {
				t.Error("expected a response but got nil")
			}
		})
	}
}

func TestRateInstallationUserList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name string
		req  projects.RateInstallationUserListRequest
	}{{
		name: "default request",
		req:  projects.NewRateInstallationUserListRequest(),
	}, {
		name: "with pagination",
		req: projects.RateInstallationUserListRequest{
			Filters: projects.RateInstallationUserListRequestFilters{
				Page:     1,
				PageSize: 10,
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			resp, err := projects.RateInstallationUserList(ctx, engine, tt.req)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if resp == nil {
				t.Error("expected a response but got nil")
			}
		})
	}
}

func TestRateInstallationUserGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name   string
		userID int64
		req    func(userID int64) projects.RateInstallationUserGetRequest
	}{{
		name:   "get installation user rate for test user",
		userID: testResources.UserID,
		req:    projects.NewRateInstallationUserGetRequest,
	}, {
		name:   "get installation user rate with currencies sideload",
		userID: testResources.UserID,
		req: func(userID int64) projects.RateInstallationUserGetRequest {
			req := projects.NewRateInstallationUserGetRequest(userID)
			req.Filters.Include = []projects.RateInstallationUserGetRequestSideload{
				projects.RateInstallationUserGetRequestSideloadCurrencies,
			}
			return req
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			req := tt.req(tt.userID)
			resp, err := projects.RateInstallationUserGet(ctx, engine, req)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if resp == nil {
				t.Error("expected a response but got nil")
			}
		})
	}
}

func TestRateInstallationUserUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name   string
		userID int64
		rate   int64
	}{{
		name:   "update installation user rate",
		userID: testResources.UserID,
		rate:   int64(rand.Intn(10000) + 1000), // cents
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			req := projects.NewRateInstallationUserUpdateRequest(tt.userID, &tt.rate)
			resp, err := projects.RateInstallationUserUpdate(ctx, engine, req)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if resp == nil {
				t.Error("expected a response but got nil")
			}
		})
	}
}

func TestRateInstallationUserBulkUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name string
		req  projects.RateInstallationUserBulkUpdateRequest
	}{{
		name: "update specific users",
		req: func() projects.RateInstallationUserBulkUpdateRequest {
			rate := int64(rand.Intn(10000) + 1000)
			return projects.RateInstallationUserBulkUpdateRequest{
				IDs:      []int64{testResources.UserID},
				UserRate: &rate,
			}
		}(),
	}, {
		name: "update all users",
		req: func() projects.RateInstallationUserBulkUpdateRequest {
			rate := int64(rand.Intn(10000) + 1000)
			return projects.RateInstallationUserBulkUpdateRequest{
				All:      true,
				UserRate: &rate,
			}
		}(),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			resp, err := projects.RateInstallationUserBulkUpdate(ctx, engine, tt.req)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if resp == nil {
				t.Error("expected a response but got nil")
			}
		})
	}
}

func TestRateProjectGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name      string
		projectID int64
		req       func(projectID int64) projects.RateProjectGetRequest
	}{{
		name:      "get project rate for test project",
		projectID: testResources.ProjectID,
		req:       projects.NewRateProjectGetRequest,
	}, {
		name:      "get project rate with currencies sideload",
		projectID: testResources.ProjectID,
		req: func(projectID int64) projects.RateProjectGetRequest {
			req := projects.NewRateProjectGetRequest(projectID)
			req.Filters.Include = []projects.RateProjectGetRequestSideload{
				projects.RateProjectGetRequestSideloadCurrencies,
			}
			return req
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			req := tt.req(tt.projectID)
			resp, err := projects.RateProjectGet(ctx, engine, req)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if resp == nil {
				t.Error("expected a response but got nil")
			}
		})
	}
}

func TestRateProjectUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name      string
		projectID int64
		rate      int64
	}{{
		name:      "update project rate",
		projectID: testResources.ProjectID,
		rate:      int64(rand.Intn(10000) + 1000),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			req := projects.NewRateProjectUpdateRequest(tt.projectID, &tt.rate)
			resp, err := projects.RateProjectUpdate(ctx, engine, req)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if resp == nil {
				t.Error("expected a response but got nil")
			}
		})
	}
}

func TestRateProjectAndUsersUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name      string
		projectID int64
		rate      int64
	}{{
		name:      "update project and users rate",
		projectID: testResources.ProjectID,
		rate:      int64(rand.Intn(10000) + 1000),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			req := projects.NewRateProjectAndUsersUpdateRequest(tt.projectID, tt.rate)
			// Add user rate exceptions if needed
			req.UserRates = []projects.ProjectUserRateRequest{
				{
					User: twapi.Relationship{
						ID:   testResources.UserID,
						Type: "users",
					},
					UserRate: tt.rate + 500, // Different rate for specific user
				},
			}

			resp, err := projects.RateProjectAndUsersUpdate(ctx, engine, req)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if resp == nil {
				t.Error("expected a response but got nil")
			}
		})
	}
}

func TestRateProjectUserList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name      string
		projectID int64
		req       projects.RateProjectUserListRequest
	}{{
		name:      "default request",
		projectID: testResources.ProjectID,
		req:       projects.NewRateProjectUserListRequest(testResources.ProjectID),
	}, {
		name:      "with filters",
		projectID: testResources.ProjectID,
		req: projects.RateProjectUserListRequest{
			Path: projects.RateProjectUserListRequestPath{
				ProjectID: testResources.ProjectID,
			},
			Filters: projects.RateProjectUserListRequestFilters{
				SearchTerm: "test",
				OrderBy:    projects.RateProjectUserListRequestOrderByUsername,
				OrderMode:  twapi.OrderModeAscending,
				Page:       1,
				PageSize:   10,
			},
		},
	}, {
		name:      "with descending order",
		projectID: testResources.ProjectID,
		req: projects.RateProjectUserListRequest{
			Path: projects.RateProjectUserListRequestPath{
				ProjectID: testResources.ProjectID,
			},
			Filters: projects.RateProjectUserListRequestFilters{
				OrderBy:   projects.RateProjectUserListRequestOrderByUsername,
				OrderMode: twapi.OrderModeDescending,
				Page:      1,
				PageSize:  10,
			},
		},
	}, {
		name:      "without search term",
		projectID: testResources.ProjectID,
		req: projects.RateProjectUserListRequest{
			Path: projects.RateProjectUserListRequestPath{
				ProjectID: testResources.ProjectID,
			},
			Filters: projects.RateProjectUserListRequestFilters{
				OrderBy:   projects.RateProjectUserListRequestOrderByUsername,
				OrderMode: twapi.OrderModeAscending,
				Page:      1,
				PageSize:  20,
			},
		},
	}, {
		name:      "pagination only",
		projectID: testResources.ProjectID,
		req: projects.RateProjectUserListRequest{
			Path: projects.RateProjectUserListRequestPath{
				ProjectID: testResources.ProjectID,
			},
			Filters: projects.RateProjectUserListRequestFilters{
				Page:     2,
				PageSize: 15,
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			resp, err := projects.RateProjectUserList(ctx, engine, tt.req)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if resp == nil {
				t.Error("expected a response but got nil")
			}
		})
	}
}

func TestRateProjectUserGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name      string
		projectID int64
		userID    int64
		req       func(projectID int64, userID int64) projects.RateProjectUserGetRequest
	}{{
		name:      "get project user rate",
		projectID: testResources.ProjectID,
		userID:    testResources.UserID,
		req:       projects.NewRateProjectUserGetRequest,
	}, {
		name:      "get project user rate with currencies sideload",
		projectID: testResources.ProjectID,
		userID:    testResources.UserID,
		req: func(projectID int64, userID int64) projects.RateProjectUserGetRequest {
			req := projects.NewRateProjectUserGetRequest(projectID, userID)
			req.Filters.Include = []projects.RateProjectUserGetRequestSideload{
				projects.RateProjectUserGetRequestSideloadCurrencies,
			}
			return req
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			req := tt.req(tt.projectID, tt.userID)
			resp, err := projects.RateProjectUserGet(ctx, engine, req)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if resp == nil {
				t.Error("expected a response but got nil")
			}
		})
	}
}

func TestRateProjectUserUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name      string
		projectID int64
		userID    int64
		rate      int64
	}{{
		name:      "update project user rate",
		projectID: testResources.ProjectID,
		userID:    testResources.UserID,
		rate:      int64(rand.Intn(10000) + 1000),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			req := projects.NewRateProjectUserUpdateRequest(tt.projectID, tt.userID, &tt.rate)
			resp, err := projects.RateProjectUserUpdate(ctx, engine, req)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if resp == nil {
				t.Error("expected a response but got nil")
			}
		})
	}
}

func TestRateProjectUserHistoryGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name      string
		projectID int64
		userID    int64
		req       func(projectID int64, userID int64) projects.RateProjectUserHistoryGetRequest
	}{{
		name:      "get project user rate history",
		projectID: testResources.ProjectID,
		userID:    testResources.UserID,
		req:       projects.NewRateProjectUserHistoryGetRequest,
	}, {
		name:      "get project user rate history with search term",
		projectID: testResources.ProjectID,
		userID:    testResources.UserID,
		req: func(projectID int64, userID int64) projects.RateProjectUserHistoryGetRequest {
			req := projects.NewRateProjectUserHistoryGetRequest(projectID, userID)
			req.Filters.SearchTerm = "test"
			return req
		},
	}, {
		name:      "get project user rate history with ordering",
		projectID: testResources.ProjectID,
		userID:    testResources.UserID,
		req: func(projectID int64, userID int64) projects.RateProjectUserHistoryGetRequest {
			req := projects.NewRateProjectUserHistoryGetRequest(projectID, userID)
			req.Filters.OrderBy = projects.RateProjectUserHistoryGetRequestOrderByUsername
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		},
	}, {
		name:      "get project user rate history with users sideload",
		projectID: testResources.ProjectID,
		userID:    testResources.UserID,
		req: func(projectID int64, userID int64) projects.RateProjectUserHistoryGetRequest {
			req := projects.NewRateProjectUserHistoryGetRequest(projectID, userID)
			req.Filters.Include = []projects.RateProjectUserHistoryGetRequestSideload{
				projects.RateProjectUserHistoryGetRequestSideloadUsers,
			}
			return req
		},
	}, {
		name:      "get project user rate history with pagination",
		projectID: testResources.ProjectID,
		userID:    testResources.UserID,
		req: func(projectID int64, userID int64) projects.RateProjectUserHistoryGetRequest {
			req := projects.NewRateProjectUserHistoryGetRequest(projectID, userID)
			req.Filters.Page = 1
			req.Filters.PageSize = 20
			return req
		},
	}, {
		name:      "get project user rate history with all filters",
		projectID: testResources.ProjectID,
		userID:    testResources.UserID,
		req: func(projectID int64, userID int64) projects.RateProjectUserHistoryGetRequest {
			req := projects.NewRateProjectUserHistoryGetRequest(projectID, userID)
			req.Filters.SearchTerm = "test"
			req.Filters.OrderBy = projects.RateProjectUserHistoryGetRequestOrderByUsername
			req.Filters.OrderMode = twapi.OrderModeDescending
			req.Filters.Include = []projects.RateProjectUserHistoryGetRequestSideload{
				projects.RateProjectUserHistoryGetRequestSideloadUsers,
			}
			req.Filters.Page = 1
			req.Filters.PageSize = 10
			return req
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			req := tt.req(tt.projectID, tt.userID)
			resp, err := projects.RateProjectUserHistoryGet(ctx, engine, req)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if resp == nil {
				t.Error("expected a response but got nil")
			}
		})
	}
}

// Test HTTP Request generation
func TestRateRequestGeneration(t *testing.T) {
	tests := []struct {
		name        string
		requestFunc func() interface {
			HTTPRequest(context.Context, string) (*http.Request, error)
		}
		expectedURI string
	}{{
		name: "RateUserGetRequest",
		requestFunc: func() interface {
			HTTPRequest(context.Context, string) (*http.Request, error)
		} {
			return projects.NewRateUserGetRequest(123)
		},
		expectedURI: "/projects/api/v3/people/123/rates",
	}, {
		name: "RateInstallationUserListRequest",
		requestFunc: func() interface {
			HTTPRequest(context.Context, string) (*http.Request, error)
		} {
			return projects.NewRateInstallationUserListRequest()
		},
		expectedURI: "/projects/api/v3/rates/installation/users.json",
	}, {
		name: "RateInstallationUserGetRequest",
		requestFunc: func() interface {
			HTTPRequest(context.Context, string) (*http.Request, error)
		} {
			return projects.NewRateInstallationUserGetRequest(456)
		},
		expectedURI: "/projects/api/v3/rates/installation/users/456.json",
	}, {
		name: "RateProjectGetRequest",
		requestFunc: func() interface {
			HTTPRequest(context.Context, string) (*http.Request, error)
		} {
			return projects.NewRateProjectGetRequest(789)
		},
		expectedURI: "/projects/api/v3/rates/projects/789.json",
	}, {
		name: "RateProjectUserListRequest",
		requestFunc: func() interface {
			HTTPRequest(context.Context, string) (*http.Request, error)
		} {
			return projects.NewRateProjectUserListRequest(789)
		},
		expectedURI: "/projects/api/v3/rates/projects/789/users",
	}, {
		name: "RateProjectUserGetRequest",
		requestFunc: func() interface {
			HTTPRequest(context.Context, string) (*http.Request, error)
		} {
			return projects.NewRateProjectUserGetRequest(789, 123)
		},
		expectedURI: "/projects/api/v3/rates/projects/789/users/123.json",
	}, {
		name: "RateProjectUserHistoryGetRequest",
		requestFunc: func() interface {
			HTTPRequest(context.Context, string) (*http.Request, error)
		} {
			return projects.NewRateProjectUserHistoryGetRequest(789, 123)
		},
		expectedURI: "/projects/api/v3/rates/projects/789/users/123/history",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.requestFunc()
			httpReq, err := req.HTTPRequest(context.Background(), "https://test.com")
			if err != nil {
				t.Fatalf("unexpected error creating HTTP request: %s", err)
			}

			if !strings.HasSuffix(httpReq.URL.Path, tt.expectedURI) {
				t.Errorf("expected URI to end with %q but got %q", tt.expectedURI, httpReq.URL.Path)
			}
		})
	}
}

// Test pagination functionality
func TestRatePagination(t *testing.T) {
	// Test RateInstallationUserListResponse pagination
	t.Run("RateInstallationUserListResponse pagination", func(t *testing.T) {
		resp := &projects.RateInstallationUserListResponse{}
		req := projects.NewRateInstallationUserListRequest()
		req.Filters.Page = 1

		resp.SetRequest(req)
		resp.Meta.Page.HasMore = true

		nextReq := resp.Iterate()
		if nextReq == nil {
			t.Error("expected next request but got nil")
		} else if nextReq.Filters.Page != 2 {
			t.Errorf("expected page 2 but got %d", nextReq.Filters.Page)
		}

		// Test when no more pages
		resp.Meta.Page.HasMore = false
		nextReq = resp.Iterate()
		if nextReq != nil {
			t.Error("expected nil but got next request")
		}
	})

	// Test RateProjectUserListResponse pagination
	t.Run("RateProjectUserListResponse pagination", func(t *testing.T) {
		resp := &projects.RateProjectUserListResponse{}
		req := projects.NewRateProjectUserListRequest(123)
		req.Filters.Page = 1

		resp.SetRequest(req)
		resp.Meta.Page.HasMore = true

		nextReq := resp.Iterate()
		if nextReq == nil {
			t.Error("expected next request but got nil")
		} else if nextReq.Filters.Page != 2 {
			t.Errorf("expected page 2 but got %d", nextReq.Filters.Page)
		}
	})

	// Test RateProjectUserHistoryGetResponse pagination
	t.Run("RateProjectUserHistoryGetResponse pagination", func(t *testing.T) {
		resp := &projects.RateProjectUserHistoryGetResponse{}
		req := projects.NewRateProjectUserHistoryGetRequest(123, 456)
		req.Filters.Page = 1

		resp.SetRequest(req)
		resp.Meta.Page.HasMore = true

		nextReq := resp.Iterate()
		if nextReq == nil {
			t.Error("expected next request but got nil")
		} else if nextReq.Filters.Page != 2 {
			t.Errorf("expected page 2 but got %d", nextReq.Filters.Page)
		}
	})
}

// Test constructor functions
func TestRateConstructors(t *testing.T) {
	t.Run("NewRateUserGetRequest", func(t *testing.T) {
		req := projects.NewRateUserGetRequest(123)
		if req.Path.ID != 123 {
			t.Errorf("expected ID 123 but got %d", req.Path.ID)
		}
	})

	t.Run("NewRateInstallationUserGetRequest", func(t *testing.T) {
		req := projects.NewRateInstallationUserGetRequest(456)
		if req.Path.UserID != 456 {
			t.Errorf("expected UserID 456 but got %d", req.Path.UserID)
		}
	})

	t.Run("NewRateInstallationUserUpdateRequest", func(t *testing.T) {
		var rate int64 = 5000
		req := projects.NewRateInstallationUserUpdateRequest(123, &rate)
		if req.Path.UserID != 123 {
			t.Errorf("expected UserID 123 but got %d", req.Path.UserID)
		}
		if req.CurrencyID != nil {
			t.Errorf("expected CurrencyID to be nil but got %v", req.CurrencyID)
		}
		if req.UserRate == nil || *req.UserRate != 5000 {
			t.Errorf("expected UserRate 5000 but got %v", req.UserRate)
		}
	})

	t.Run("NewRateProjectGetRequest", func(t *testing.T) {
		req := projects.NewRateProjectGetRequest(789)
		if req.Path.ProjectID != 789 {
			t.Errorf("expected ProjectID 789 but got %d", req.Path.ProjectID)
		}
	})

	t.Run("NewRateProjectUserGetRequest", func(t *testing.T) {
		req := projects.NewRateProjectUserGetRequest(789, 123)
		if req.Path.ProjectID != 789 {
			t.Errorf("expected ProjectID 789 but got %d", req.Path.ProjectID)
		}
		if req.Path.UserID != 123 {
			t.Errorf("expected UserID 123 but got %d", req.Path.UserID)
		}
	})
}
