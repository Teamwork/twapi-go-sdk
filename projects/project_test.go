package projects_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestProjectCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name          string
		input         projects.ProjectCreateRequest
		expectedError bool
	}{{
		name:  "it should create a project with valid input",
		input: projects.NewProjectCreateRequest(fmt.Sprintf("Test Project %d", time.Now().UnixNano())),
	}, {
		name: "it should fail to create a project with missing name",
		input: projects.ProjectCreateRequest{
			Description: twapi.Ptr("This project has no name"),
		},
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			project, err := projects.ProjectCreate(ctx, engine, tt.input)
			defer func() {
				if err != nil {
					return
				}
				_, err := projects.ProjectDelete(ctx, engine, projects.NewProjectDeleteRequest(int64(project.ID)))
				if err != nil {
					t.Errorf("failed to delete project after test: %s", err)
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
			if project.ID == 0 {
				t.Error("expected a valid project ID but got 0")
			}
		})
	}
}

func TestProjectUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	project, err := projects.ProjectCreate(t.Context(), engine, projects.ProjectCreateRequest{
		Name: fmt.Sprintf("Test Project %d", time.Now().UnixNano()),
	})
	if err != nil {
		t.Fatalf("failed to create a project: %s", err)
	}
	defer func() {
		_, err := projects.ProjectDelete(t.Context(), engine, projects.NewProjectDeleteRequest(int64(project.ID)))
		if err != nil {
			t.Errorf("failed to delete project after test: %s", err)
		}
	}()

	tests := []struct {
		name          string
		input         projects.ProjectUpdateRequest
		expectedError bool
	}{{
		name: "it should update a project with valid input",
		input: projects.ProjectUpdateRequest{
			Path: projects.ProjectUpdateRequestPath{
				ID: int64(project.ID),
			},
			Description: twapi.Ptr("This is a test project"),
		},
	}, {
		name: "it should fail to update a project with missing name",
		input: projects.ProjectUpdateRequest{
			Path: projects.ProjectUpdateRequestPath{
				ID: int64(project.ID),
			},
			Name: twapi.Ptr(""),
		},
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.ProjectUpdate(ctx, engine, tt.input)
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

func TestProjectDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	project, err := projects.ProjectCreate(t.Context(), engine, projects.ProjectCreateRequest{
		Name: fmt.Sprintf("Test Project %d", time.Now().UnixNano()),
	})
	if err != nil {
		t.Fatalf("failed to create a project: %s", err)
	}

	tests := []struct {
		name          string
		input         projects.ProjectDeleteRequest
		expectedError bool
	}{{
		name:  "it should delete a project with valid input",
		input: projects.NewProjectDeleteRequest(int64(project.ID)),
	}, {
		name:          "it should fail to delete an unknown project",
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.ProjectDelete(ctx, engine, tt.input)
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

func TestProjectGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	project, err := projects.ProjectCreate(t.Context(), engine, projects.ProjectCreateRequest{
		Name: fmt.Sprintf("Test Project %d", time.Now().UnixNano()),
	})
	if err != nil {
		t.Fatalf("failed to create a project: %s", err)
	}
	defer func() {
		_, err := projects.ProjectDelete(t.Context(), engine, projects.NewProjectDeleteRequest(int64(project.ID)))
		if err != nil {
			t.Errorf("failed to delete project after test: %s", err)
		}
	}()

	tests := []struct {
		name          string
		input         projects.ProjectGetRequest
		expectedError bool
	}{{
		name:  "it should retrieve a project with valid input",
		input: projects.NewProjectGetRequest(int64(project.ID)),
	}, {
		name:          "it should fail to retrieve an unknown project",
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.ProjectGet(ctx, engine, tt.input)
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

func TestProjectList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	project, err := projects.ProjectCreate(t.Context(), engine, projects.ProjectCreateRequest{
		Name: fmt.Sprintf("Test Project %d", time.Now().UnixNano()),
	})
	if err != nil {
		t.Fatalf("failed to create a project: %s", err)
	}
	defer func() {
		_, err := projects.ProjectDelete(t.Context(), engine, projects.NewProjectDeleteRequest(int64(project.ID)))
		if err != nil {
			t.Errorf("failed to delete project after test: %s", err)
		}
	}()

	tests := []struct {
		name          string
		input         projects.ProjectListRequest
		expectedError bool
	}{{
		name: "it should list projects",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.ProjectList(ctx, engine, tt.input)
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
