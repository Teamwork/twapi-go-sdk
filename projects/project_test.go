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
		name: "it should create a project with valid input",
		input: projects.ProjectCreateRequest{
			Name: fmt.Sprintf("Test Project %d", time.Now().UnixNano()),
		},
	}, {
		name: "it should fail to create a project with missing name",
		input: projects.ProjectCreateRequest{
			Description: twapi.Ptr("This project has no name"),
		},
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			project, err := projects.ProjectCreate(ctx, engine, tt.input)
			defer func() {
				if err != nil {
					return
				}
				var deleteRequest projects.ProjectDeleteRequest
				deleteRequest.Path.ID = int64(project.ID)
				if _, err := projects.ProjectDelete(context.Background(), engine, deleteRequest); err != nil {
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

	projectName := fmt.Sprintf("Test Project %d", time.Now().UnixNano())
	project, err := projects.ProjectCreate(context.Background(), engine, projects.ProjectCreateRequest{
		Name: projectName,
	})
	if err != nil {
		t.Fatalf("failed to create a project: %s", err)
	}
	defer func() {
		var deleteRequest projects.ProjectDeleteRequest
		deleteRequest.Path.ID = int64(project.ID)
		if _, err := projects.ProjectDelete(context.Background(), engine, deleteRequest); err != nil {
			t.Errorf("failed to delete project after test: %s", err)
		}
	}()

	tests := []struct {
		name          string
		input         projects.ProjectUpdateRequest
		expectedError bool
	}{{
		name: "it should update a project with valid input",
		input: func() projects.ProjectUpdateRequest {
			var p projects.ProjectUpdateRequest
			p.Path.ID = int64(project.ID)
			p.Name = &projectName
			p.Description = twapi.Ptr("This is a test project")
			return p
		}(),
	}, {
		name: "it should fail to update a project with missing name",
		input: func() projects.ProjectUpdateRequest {
			var p projects.ProjectUpdateRequest
			p.Path.ID = int64(project.ID)
			p.Name = twapi.Ptr("")
			return p
		}(),
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
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

	project, err := projects.ProjectCreate(context.Background(), engine, projects.ProjectCreateRequest{
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
		name: "it should delete a project with valid input",
		input: func() projects.ProjectDeleteRequest {
			var p projects.ProjectDeleteRequest
			p.Path.ID = int64(project.ID)
			return p
		}(),
	}, {
		name:          "it should fail to delete an unknown project",
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
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

func TestProjectRetrieve(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	project, err := projects.ProjectCreate(context.Background(), engine, projects.ProjectCreateRequest{
		Name: fmt.Sprintf("Test Project %d", time.Now().UnixNano()),
	})
	if err != nil {
		t.Fatalf("failed to create a project: %s", err)
	}
	defer func() {
		var deleteRequest projects.ProjectDeleteRequest
		deleteRequest.Path.ID = int64(project.ID)
		if _, err := projects.ProjectDelete(context.Background(), engine, deleteRequest); err != nil {
			t.Errorf("failed to delete project after test: %s", err)
		}
	}()

	tests := []struct {
		name          string
		input         projects.ProjectRetrieveRequest
		expectedError bool
	}{{
		name: "it should retrieve a project with valid input",
		input: func() projects.ProjectRetrieveRequest {
			var p projects.ProjectRetrieveRequest
			p.Path.ID = int64(project.ID)
			return p
		}(),
	}, {
		name:          "it should fail to retrieve an unknown project",
		expectedError: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.ProjectRetrieve(ctx, engine, tt.input)
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

func TestProjectRetrieveMany(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	project, err := projects.ProjectCreate(context.Background(), engine, projects.ProjectCreateRequest{
		Name: fmt.Sprintf("Test Project %d", time.Now().UnixNano()),
	})
	if err != nil {
		t.Fatalf("failed to create a project: %s", err)
	}
	defer func() {
		var deleteRequest projects.ProjectDeleteRequest
		deleteRequest.Path.ID = int64(project.ID)
		if _, err := projects.ProjectDelete(context.Background(), engine, deleteRequest); err != nil {
			t.Errorf("failed to delete project after test: %s", err)
		}
	}()

	tests := []struct {
		name          string
		input         projects.ProjectRetrieveManyRequest
		expectedError bool
	}{{
		name: "it should retrieve many projects",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.ProjectRetrieveMany(ctx, engine, tt.input)
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
