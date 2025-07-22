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
			Title:     twapi.Ptr("Test User"),
			Email:     fmt.Sprintf("email%d%d@example.com", time.Now().UnixNano(), rand.Intn(100)),
			Admin:     twapi.Ptr(true),
			Type:      twapi.Ptr("account"),
			CompanyID: &testResources.CompanyID,
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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
			FirstName: twapi.Ptr(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
			LastName:  twapi.Ptr(fmt.Sprintf("user%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Title:     twapi.Ptr("Test User"),
			Email:     twapi.Ptr(fmt.Sprintf("email%d%d@example.com", time.Now().UnixNano(), rand.Intn(100))),
			Admin:     twapi.Ptr(true),
			Type:      twapi.Ptr("account"),
			CompanyID: &testResources.CompanyID,
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.UserList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
