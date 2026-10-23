package projects_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
	"github.com/teamwork/twapi-go-sdk/session"
)

func TestUserCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name  string
		input projects.UserCreateRequest
	}{{
		name: "only required fields",
		input: projects.NewUserCreateRequest(
			fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
			fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
			fmt.Sprintf("test%d%d@example.com", time.Now().UnixNano(), rand.Intn(100)),
		),
	}, {
		name: "all fields",
		input: projects.UserCreateRequest{
			FirstName: fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
			LastName:  fmt.Sprintf("user%d%d", time.Now().UnixNano(), rand.Intn(100)),
			Title:     new("Test User"),
			Email:     fmt.Sprintf("email%d%d@example.com", time.Now().UnixNano(), rand.Intn(100)),
			Admin:     new(true),
			Type:      new("account"),
			CompanyID: &testResources.CompanyID,
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			user, err := projects.UserCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx = context.Background() // t.Context is always canceled in cleanup
				_, err := projects.UserDelete(ctx, engine, projects.NewUserDeleteRequest(int64(user.ID)))
				if err != nil {
					t.Errorf("failed to delete user after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if user.ID == 0 {
				t.Error("expected a valid user ID but got 0")
			}
		})
	}
}

func TestUserBulkCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	users := make([]projects.UserBulkCreateRequestUser, 2)
	for i := range users {
		users[i] = projects.UserBulkCreateRequestUser{
			FirstName: fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
			LastName:  fmt.Sprintf("bulk%d%d", time.Now().UnixNano(), rand.Intn(100)),
			Email:     fmt.Sprintf("bulk%d%d@example.com", time.Now().UnixNano(), rand.Intn(100)),
		}
	}

	response, err := projects.UserBulkCreate(ctx, engine, projects.NewUserBulkCreateRequest(users))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	t.Cleanup(func() {
		for _, user := range response.Users {
			ctx := context.Background() // t.Context is always canceled in cleanup
			if _, err := projects.UserDelete(ctx, engine, projects.NewUserDeleteRequest(user.ID)); err != nil {
				t.Errorf("failed to delete user after test: %s", err)
			}
		}
	})
	if len(response.Users) != len(users) {
		t.Errorf("expected %d users to be created, got %d (failed: %+v)", len(users), len(response.Users), response.Failed)
	}
}

// bulkCreateServer answers the bulk create route with the given status and
// body, and records the request body it received.
func bulkCreateServer(t *testing.T, status int, responseBody string) (*twapi.Engine, *[]byte) {
	t.Helper()

	var received []byte
	mux := http.NewServeMux()
	mux.HandleFunc("POST /projects/api/v3/people/bulk.json", func(w http.ResponseWriter, r *http.Request) {
		received, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, responseBody)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return twapi.NewEngine(session.NewBearerToken("token", server.URL)), &received
}

func TestUserBulkCreateRequestBody(t *testing.T) {
	testEngine, received := bulkCreateServer(t, http.StatusCreated,
		`{"people":[{"id":12345,"email":"first@example.com"},{"id":12346,"email":"second@example.com"}]}`)

	userType := projects.UserTypeCollaborator
	request := projects.NewUserBulkCreateRequest([]projects.UserBulkCreateRequestUser{{
		FirstName: "First",
		LastName:  "User",
		Email:     "first@example.com",
	}, {
		FirstName: "Second",
		LastName:  "User",
		Email:     "second@example.com",
		Type:      &userType,
		CompanyID: new(int64(777)),
	}})

	response, err := projects.UserBulkCreate(t.Context(), testEngine, request)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(response.Users) != 2 || response.Users[0].ID != 12345 || response.Users[1].ID != 12346 {
		t.Errorf("unexpected users: %+v", response.Users)
	}
	if len(response.Failed) != 0 {
		t.Errorf("expected no failures, got %+v", response.Failed)
	}

	var body struct {
		People  []map[string]any `json:"people"`
		Options map[string]any   `json:"options"`
	}
	if err := json.Unmarshal(*received, &body); err != nil {
		t.Fatalf("failed to decode the request body: %s", err)
	}
	if len(body.People) != 2 {
		t.Fatalf("expected 2 people in the request, got %d", len(body.People))
	}
	for i, person := range body.People {
		// The API invites a person whose entry omits sendInvite, so the field
		// must always be present and false unless asked for.
		if sendInvite, ok := person["sendInvite"]; !ok || sendInvite != false {
			t.Errorf("person %d: expected sendInvite to be sent as false, got %v (present: %t)", i, sendInvite, ok)
		}
	}
	if body.People[1]["userType"] != "collaborator" || body.People[1]["companyId"] != float64(777) {
		t.Errorf("unexpected optional fields: %v", body.People[1])
	}
	if body.Options["partialSuccess"] != true {
		t.Errorf("expected options.partialSuccess to be true, got %v", body.Options)
	}
}

func TestUserBulkCreatePartialFailure(t *testing.T) {
	testEngine, _ := bulkCreateServer(t, http.StatusCreated, `{
		"people":[{"id":12345}],
		"failed":[{"index":1,"errors":[{"title":"invalid payload","detail":"email is already in use"}]}]
	}`)

	request := projects.NewUserBulkCreateRequest([]projects.UserBulkCreateRequestUser{
		{FirstName: "First", LastName: "User", Email: "first@example.com"},
		{FirstName: "Second", LastName: "User", Email: "taken@example.com"},
	})
	response, err := projects.UserBulkCreate(t.Context(), testEngine, request)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(response.Users) != 1 || response.Users[0].ID != 12345 {
		t.Errorf("unexpected users: %+v", response.Users)
	}
	if len(response.Failed) != 1 || response.Failed[0].Index != 1 ||
		len(response.Failed[0].Errors) != 1 || response.Failed[0].Errors[0].Detail != "email is already in use" {
		t.Errorf("unexpected failures: %+v", response.Failed)
	}
}

func TestUserBulkCreateFailures(t *testing.T) {
	testEngine, _ := bulkCreateServer(t, http.StatusBadRequest, `{"errors":[
		{"title":"invalid payload","detail":"email is required","meta":{"index":0}},
		{"title":"invalid payload","detail":"first name is required","meta":{"index":0}},
		{"title":"invalid payload","detail":"email is already in use","meta":{"index":2}}
	]}`)

	request := projects.NewUserBulkCreateRequest([]projects.UserBulkCreateRequestUser{
		{}, {FirstName: "Second", LastName: "User", Email: "second@example.com"}, {Email: "taken@example.com"},
	})
	_, err := projects.UserBulkCreate(t.Context(), testEngine, request)
	if err == nil {
		t.Fatal("expected an error")
	}

	failures := projects.UserBulkCreateFailures(err)
	if len(failures) != 2 {
		t.Fatalf("expected 2 failed entries, got %+v", failures)
	}
	if failures[0].Index != 0 || len(failures[0].Errors) != 2 {
		t.Errorf("unexpected first failure: %+v", failures[0])
	}
	if failures[1].Index != 2 || failures[1].Errors[0].Detail != "email is already in use" {
		t.Errorf("unexpected second failure: %+v", failures[1])
	}

	if projects.UserBulkCreateFailures(fmt.Errorf("not an API error")) != nil {
		t.Error("expected no failures for an error that is not an API error")
	}
}

func TestUserBulkCreateEmpty(t *testing.T) {
	testEngine, received := bulkCreateServer(t, http.StatusCreated, `{"people":[]}`)

	if _, err := projects.UserBulkCreate(t.Context(), testEngine,
		projects.NewUserBulkCreateRequest(nil)); err == nil {
		t.Error("expected an error for an empty list")
	}
	if *received != nil {
		t.Error("expected no request to be sent for an empty list")
	}
}

func TestUserUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	userID, userCleanup, err := createUser(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(userCleanup)

	tests := []struct {
		name  string
		input projects.UserUpdateRequest
	}{{
		name: "all fields",
		input: projects.UserUpdateRequest{
			Path: projects.UserUpdateRequestPath{
				ID: userID,
			},
			FirstName: new(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
			LastName:  new(fmt.Sprintf("user%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Title:     new("Test User"),
			Email:     new(fmt.Sprintf("email%d%d@example.com", time.Now().UnixNano(), rand.Intn(100))),
			Admin:     new(true),
			Type:      new("account"),
			CompanyID: &testResources.CompanyID,
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.UserUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestUserDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	userID, _, err := createUser(t)
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.UserDelete(ctx, engine, projects.NewUserDeleteRequest(userID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestUserGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	userID, userCleanup, err := createUser(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(userCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.UserGet(ctx, engine, projects.NewUserGetRequest(userID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestUserGetMe(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err := projects.UserGetMe(ctx, engine, projects.NewUserGetMeRequest()); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestUserList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	userID, userCleanup, err := createUser(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(userCleanup)

	if err = addProjectMember(t, testResources.ProjectID, userID); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		input         projects.UserListRequest
		expectedError bool
	}{{
		name: "all users",
	}, {
		name: "users for project",
		input: projects.UserListRequest{
			Path: projects.UserListRequestPath{
				ProjectID: testResources.ProjectID,
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.UserList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
