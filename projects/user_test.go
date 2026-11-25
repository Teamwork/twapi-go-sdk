package projects_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
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
			FirstName:    new(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
			LastName:     new(fmt.Sprintf("user%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Title:        new("Test User"),
			Email:        new(fmt.Sprintf("email%d%d@example.com", time.Now().UnixNano(), rand.Intn(100))),
			Admin:        new(true),
			Type:         new("account"),
			CompanyID:    &testResources.CompanyID,
			Profile:      new("Public notes"),
			PrivateNotes: new("Private notes"),
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

func TestUserGetFullProfile(t *testing.T) {
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

	update := projects.NewUserUpdateRequest(userID)
	update.Profile = new("Public notes")
	if _, err = projects.UserUpdate(ctx, engine, update); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	request := projects.NewUserGetRequest(userID)
	request.Filters.FullProfile = true
	response, err := projects.UserGet(ctx, engine, request)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if response.User.Profile == nil || *response.User.Profile != "Public notes" {
		t.Errorf("expected the public notes to be read back, got %v", response.User.Profile)
	}
}

// userServer answers one route with the given body and records the request
// body it received.
func userServer(t *testing.T, pattern, responseBody string) (*twapi.Engine, *[]byte) {
	t.Helper()

	var received []byte
	mux := http.NewServeMux()
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		received, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, responseBody)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return twapi.NewEngine(session.NewBearerToken("token", server.URL)), &received
}

func TestUserNotesReachTheWire(t *testing.T) {
	t.Run("update", func(t *testing.T) {
		testEngine, received := userServer(t, "PUT /people/777.json", `{"STATUS":"OK"}`)

		request := projects.NewUserUpdateRequest(777)
		request.Profile = new("Public")
		request.PrivateNotes = new("")
		if _, err := projects.UserUpdate(t.Context(), testEngine, request); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		var body struct {
			Person map[string]any `json:"person"`
		}
		if err := json.Unmarshal(*received, &body); err != nil {
			t.Fatalf("failed to decode the request body: %s", err)
		}
		// An empty string clears the notes, so it must be sent.
		if body.Person["profile"] != "Public" || body.Person["privateNotes"] != "" {
			t.Errorf("unexpected notes in the request: %v", body.Person)
		}
	})

	t.Run("update without notes", func(t *testing.T) {
		testEngine, received := userServer(t, "PUT /people/777.json", `{"STATUS":"OK"}`)

		request := projects.NewUserUpdateRequest(777)
		request.Title = new("Developer")
		if _, err := projects.UserUpdate(t.Context(), testEngine, request); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		var body struct {
			Person map[string]any `json:"person"`
		}
		if err := json.Unmarshal(*received, &body); err != nil {
			t.Fatalf("failed to decode the request body: %s", err)
		}
		for _, key := range []string{"profile", "privateNotes"} {
			if _, ok := body.Person[key]; ok {
				t.Errorf("expected %q to be omitted, got %v", key, body.Person)
			}
		}
	})

	t.Run("create", func(t *testing.T) {
		testEngine, received := userServer(t, "POST /people.json", `{"STATUS":"OK","id":"12345"}`)

		request := projects.NewUserCreateRequest("First", "User", "john@example.com")
		request.Profile = new("Public")
		request.PrivateNotes = new("Private")
		// The mock answers 200, which the create rejects; only the body matters.
		_, _ = projects.UserCreate(t.Context(), testEngine, request)

		var body struct {
			Person map[string]any `json:"person"`
		}
		if err := json.Unmarshal(*received, &body); err != nil {
			t.Fatalf("failed to decode the request body: %s", err)
		}
		if body.Person["profile"] != "Public" || body.Person["privateNotes"] != "Private" {
			t.Errorf("unexpected notes in the request: %v", body.Person)
		}
	})

	t.Run("bulk create", func(t *testing.T) {
		testEngine, received := bulkCreateServer(t, http.StatusCreated, `{"people":[{"id":12345}]}`)

		request := projects.NewUserBulkCreateRequest([]projects.UserBulkCreateRequestUser{{
			FirstName:    "First",
			LastName:     "User",
			Email:        "john@example.com",
			Profile:      new("Public"),
			PrivateNotes: new("Private"),
		}})
		if _, err := projects.UserBulkCreate(t.Context(), testEngine, request); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		var body struct {
			People []map[string]any `json:"people"`
		}
		if err := json.Unmarshal(*received, &body); err != nil {
			t.Fatalf("failed to decode the request body: %s", err)
		}
		if body.People[0]["profile"] != "Public" || body.People[0]["privateNotes"] != "Private" {
			t.Errorf("unexpected notes in the request: %v", body.People[0])
		}
	})
}

func TestUserGetFullProfileReachesTheWire(t *testing.T) {
	tests := []struct {
		name        string
		fullProfile bool
		body        string
		wantQuery   string
		wantPrivate *string
	}{{
		name:        "owner company",
		fullProfile: true,
		body:        `{"person":{"id":777,"profileText":"Public","privateNotesText":""}}`,
		wantQuery:   "fullProfile=true",
		wantPrivate: new(""),
	}, {
		// The API leaves the private notes out for anyone outside the owner
		// company.
		name:        "other company",
		fullProfile: true,
		body:        `{"person":{"id":777,"profileText":"Public"}}`,
		wantQuery:   "fullProfile=true",
	}, {
		name: "not asked for",
		body: `{"person":{"id":777}}`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var query string
			mux := http.NewServeMux()
			mux.HandleFunc("GET /projects/api/v3/people/777.json", func(w http.ResponseWriter, r *http.Request) {
				query = r.URL.RawQuery
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, tt.body)
			})
			server := httptest.NewServer(mux)
			t.Cleanup(server.Close)
			testEngine := twapi.NewEngine(session.NewBearerToken("token", server.URL))

			request := projects.NewUserGetRequest(777)
			request.Filters.FullProfile = tt.fullProfile
			response, err := projects.UserGet(t.Context(), testEngine, request)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if query != tt.wantQuery {
				t.Errorf("expected query %q, got %q", tt.wantQuery, query)
			}
			if tt.fullProfile && (response.User.Profile == nil || *response.User.Profile != "Public") {
				t.Errorf("expected the profile, got %v", response.User.Profile)
			}
			switch {
			case tt.wantPrivate == nil && response.User.PrivateNotes != nil:
				t.Errorf("expected no private notes, got %q", *response.User.PrivateNotes)
			case tt.wantPrivate != nil && (response.User.PrivateNotes == nil ||
				*response.User.PrivateNotes != *tt.wantPrivate):
				t.Errorf("expected private notes %q, got %v", *tt.wantPrivate, response.User.PrivateNotes)
			}
		})
	}
}

func TestUserListFullProfileReachesTheWire(t *testing.T) {
	for _, fullProfile := range []bool{true, false} {
		t.Run(strconv.FormatBool(fullProfile), func(t *testing.T) {
			var query url.Values
			mux := http.NewServeMux()
			mux.HandleFunc("GET /projects/api/v3/people.json", func(w http.ResponseWriter, r *http.Request) {
				query = r.URL.Query()
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{"people":[{"id":777,"profileText":"Public"}]}`)
			})
			server := httptest.NewServer(mux)
			t.Cleanup(server.Close)
			testEngine := twapi.NewEngine(session.NewBearerToken("token", server.URL))

			request := projects.NewUserListRequest()
			request.Filters.FullProfile = fullProfile
			if _, err := projects.UserList(t.Context(), testEngine, request); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if got := query.Has("fullProfile"); got != fullProfile {
				t.Errorf("expected fullProfile sent %t, got %t: %s", fullProfile, got, query.Encode())
			}
		})
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
