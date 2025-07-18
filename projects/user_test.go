package projects_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestUserCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	epoch := time.Now().UnixNano()

	tests := []struct {
		name          string
		input         projects.UserCreateRequest
		expectedError bool
	}{{
		name: "it should create a user with valid input",
		input: projects.NewUserCreateRequest(
			fmt.Sprintf("Test User %d", epoch),
			"LastName",
			fmt.Sprintf("testuser%d@example.com", epoch),
		),
	}, {
		name: "it should fail to create a user with missing names",
		input: projects.UserCreateRequest{
			Email: fmt.Sprintf("testuser%d@example.com", epoch),
		},
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			user, err := projects.UserCreate(ctx, engine, tt.input)
			defer func() {
				if err != nil {
					return
				}
				_, err := projects.UserDelete(ctx, engine, projects.NewUserDeleteRequest(int64(user.ID)))
				if err != nil {
					t.Errorf("failed to delete user after test: %s", err)
				}
			}()

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
			if user.ID == 0 {
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
	defer userCleanup()

	tests := []struct {
		name          string
		input         projects.UserUpdateRequest
		expectedError bool
	}{{
		name: "it should update a user with valid input",
		input: projects.UserUpdateRequest{
			Path: projects.UserUpdateRequestPath{
				ID: userID,
			},
			Title: twapi.Ptr("Updated Title"),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.UserUpdate(ctx, engine, tt.input)
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

func TestUserDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	userID, _, err := createUser(t)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		input         projects.UserDeleteRequest
		expectedError bool
	}{{
		name:  "it should delete a user with valid input",
		input: projects.NewUserDeleteRequest(userID),
	}, {
		name:          "it should fail to delete an unknown user",
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.UserDelete(ctx, engine, tt.input)
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

func TestUserGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	userID, userCleanup, err := createUser(t)
	if err != nil {
		t.Fatal(err)
	}
	defer userCleanup()

	tests := []struct {
		name          string
		input         projects.UserGetRequest
		expectedError bool
	}{{
		name:  "it should retrieve a user with valid input",
		input: projects.NewUserGetRequest(userID),
	}, {
		name:          "it should fail to retrieve an unknown user",
		input:         projects.NewUserGetRequest(999999999), // assuming this ID does not exist
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.UserGet(ctx, engine, tt.input)
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

func TestUserGetMe(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name          string
		input         projects.UserGetMeRequest
		expectedError bool
	}{{
		name:  "it should retrieve the logged user",
		input: projects.NewUserGetMeRequest(),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.UserGetMe(ctx, engine, tt.input)
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

func TestUserList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	projectID, projectCleanup, err := createProject(t)
	if err != nil {
		t.Fatal(err)
	}
	defer projectCleanup()

	userID, userCleanup, err := createUser(t)
	if err != nil {
		t.Fatal(err)
	}
	defer userCleanup()

	if err = addProjectMember(t, projectID, userID); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		input         projects.UserListRequest
		expectedError bool
	}{{
		name: "it should list users",
	}, {
		name: "it should list users for project",
		input: projects.UserListRequest{
			Path: projects.UserListRequestPath{
				ProjectID: projectID,
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.UserList(ctx, engine, tt.input)
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
