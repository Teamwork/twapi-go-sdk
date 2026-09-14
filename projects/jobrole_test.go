package projects_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestJobRoleCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name  string
		input projects.JobRoleCreateRequest
	}{{
		name: "only required fields",
		input: projects.NewJobRoleCreateRequest(
			fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
		),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			jobRoleResponse, err := projects.JobRoleCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx = context.Background() // t.Context is always canceled in cleanup
				_, err := projects.JobRoleDelete(ctx, engine, projects.NewJobRoleDeleteRequest(jobRoleResponse.JobRole.ID))
				if err != nil {
					t.Errorf("failed to delete job role after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if jobRoleResponse.JobRole.ID == 0 {
				t.Error("expected a valid job role ID but got 0")
			}
		})
	}
}

func TestJobRoleUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	jobRoleID, jobRoleCleanup, err := createJobRole(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(jobRoleCleanup)

	tests := []struct {
		name  string
		input projects.JobRoleUpdateRequest
	}{{
		name: "all fields",
		input: projects.JobRoleUpdateRequest{
			Path: projects.JobRoleUpdateRequestPath{
				ID: jobRoleID,
			},
			Name: new(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.JobRoleUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestJobRoleDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	jobRoleID, _, err := createJobRole(t)
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.JobRoleDelete(ctx, engine, projects.NewJobRoleDeleteRequest(jobRoleID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestJobRoleGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	jobRoleID, jobRoleCleanup, err := createJobRole(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(jobRoleCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.JobRoleGet(ctx, engine, projects.NewJobRoleGetRequest(jobRoleID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	withUsers := projects.NewJobRoleGetRequest(jobRoleID)
	withUsers.Include = []projects.JobRoleRequestSideload{projects.JobRoleRequestSideloadUsers}

	if _, err = projects.JobRoleGet(ctx, engine, withUsers); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestJobRoleList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	_, jobRoleCleanup, err := createJobRole(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(jobRoleCleanup)

	tests := []struct {
		name          string
		input         projects.JobRoleListRequest
		expectedError bool
	}{{
		name: "all jobRoles",
	}, {
		name: "sideloading users",
		input: func() projects.JobRoleListRequest {
			req := projects.NewJobRoleListRequest()
			req.Filters.Include = []projects.JobRoleRequestSideload{projects.JobRoleRequestSideloadUsers}
			return req
		}(),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.JobRoleList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

// TestJobRoleResponsesDecodeSideloads pins that the sideloads a caller asked for
// actually survive decoding. The responses carried no Included struct at first,
// so `include=users,currencies` was honoured by the API and then silently
// discarded here.
//
// It also covers JobRole.Users and JobRole.PrimaryUsers, which the endpoint only
// populates when the users sideload is requested.
func TestJobRoleResponsesDecodeSideloads(t *testing.T) {
	jobRole := `{
		"id": 12345,
		"name": "Survey Technician",
		"users": [{"id": 456, "type": "users"}],
		"primaryUsers": [{"id": 456, "type": "users"}]
	}`
	included := `{
		"users": {"456": {"id": 456, "firstName": "John", "lastName": "Doe"}},
		"currencies": {"3": {"id": 3, "code": "EUR", "symbol": "\u20ac", "name": "Euro"}}
	}`

	var getResponse projects.JobRoleGetResponse
	if err := json.Unmarshal([]byte(`{"jobRole":`+jobRole+`,"included":`+included+`}`), &getResponse); err != nil {
		t.Fatalf("failed to decode get response: %s", err)
	}

	var listResponse projects.JobRoleListResponse
	if err := json.Unmarshal([]byte(`{"jobRoles":[`+jobRole+`],"included":`+included+`}`), &listResponse); err != nil {
		t.Fatalf("failed to decode list response: %s", err)
	}

	if len(listResponse.JobRoles) != 1 {
		t.Fatalf("expected one job role, got %d", len(listResponse.JobRoles))
	}

	for name, decoded := range map[string]struct {
		jobRole    projects.JobRole
		users      map[string]projects.User
		currencies map[string]projects.Currency
	}{
		"get":  {getResponse.JobRole, getResponse.Included.Users, getResponse.Included.Currencies},
		"list": {listResponse.JobRoles[0], listResponse.Included.Users, listResponse.Included.Currencies},
	} {
		t.Run(name, func(t *testing.T) {
			if len(decoded.jobRole.Users) != 1 || decoded.jobRole.Users[0].ID != 456 {
				t.Errorf("expected the job role membership to decode, got %v", decoded.jobRole.Users)
			}
			if len(decoded.jobRole.PrimaryUsers) != 1 || decoded.jobRole.PrimaryUsers[0].ID != 456 {
				t.Errorf("expected the primary membership to decode, got %v", decoded.jobRole.PrimaryUsers)
			}
			if got := decoded.users["456"].FirstName; got != "John" {
				t.Errorf("expected the user sideload to decode, got %q", got)
			}
			if got := decoded.currencies["3"].Code; got != "EUR" {
				t.Errorf("expected the currency sideload to decode, got %q", got)
			}
		})
	}
}

// TestJobRoleIncludesAreOneCommaSeparatedParam pins how the sideload list is
// encoded, and that the get route carries it at all. The get and the list share
// one handler, so both read the same include parameter, and a request repeating
// it instead of joining the values loses every sideload but the first.
func TestJobRoleIncludesAreOneCommaSeparatedParam(t *testing.T) {
	sideloads := []projects.JobRoleRequestSideload{
		projects.JobRoleRequestSideloadUsers,
		projects.JobRoleRequestSideloadCurrencies,
	}

	listRequest := projects.NewJobRoleListRequest()
	listRequest.Filters.Include = sideloads

	getRequest := projects.NewJobRoleGetRequest(12345)
	getRequest.Include = sideloads

	for name, request := range map[string]twapi.HTTPRequester{
		"list": listRequest,
		"get":  getRequest,
	} {
		t.Run(name, func(t *testing.T) {
			httpRequest, err := request.HTTPRequest(t.Context(), "https://test.teamwork.com")
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			raw := httpRequest.URL.RawQuery
			if got := strings.Count(raw, "include="); got != 1 {
				t.Errorf("expected exactly one include parameter, got %d in %q", got, raw)
			}
			if want := "include=users%2Ccurrencies"; !strings.Contains(raw, want) {
				t.Errorf("expected %s in the query, got %q", want, raw)
			}
		})
	}
}
