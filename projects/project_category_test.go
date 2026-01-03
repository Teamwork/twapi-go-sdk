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

func TestProjectCategoryCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	parentProjectCategoryID, parentProjectCategoryCleanup, err := createProjectCategory(t)
	if err != nil {
		t.Fatal(err)
	}
	defer parentProjectCategoryCleanup()

	tests := []struct {
		name  string
		input projects.ProjectCategoryCreateRequest
	}{{
		name:  "only required fields",
		input: projects.NewProjectCategoryCreateRequest(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
	}, {
		name: "all fields",
		input: projects.ProjectCategoryCreateRequest{
			Name:     fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
			ParentID: &parentProjectCategoryID,
			Color:    twapi.Ptr("#ff0000"),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			projectCategory, err := projects.ProjectCategoryCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx = context.Background() // t.Context is always canceled in cleanup
				_, err := projects.ProjectCategoryDelete(ctx, engine,
					projects.NewProjectCategoryDeleteRequest(int64(projectCategory.ID)))
				if err != nil {
					t.Errorf("failed to delete project category after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if projectCategory.ID == 0 {
				t.Error("expected a valid project category ID but got 0")
			}
		})
	}
}

func TestProjectCategoryUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	parentProjectCategoryID, parentProjectCategoryCleanup, err := createProjectCategory(t)
	if err != nil {
		t.Fatal(err)
	}
	defer parentProjectCategoryCleanup()

	projectCategoryID, projectCategoryCleanup, err := createProjectCategory(t)
	if err != nil {
		t.Fatal(err)
	}
	defer projectCategoryCleanup()

	tests := []struct {
		name  string
		input projects.ProjectCategoryUpdateRequest
	}{{
		name: "all fields",
		input: projects.ProjectCategoryUpdateRequest{
			Path: projects.ProjectCategoryUpdateRequestPath{
				ID: projectCategoryID,
			},
			Name:     twapi.Ptr(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
			ParentID: &parentProjectCategoryID,
			Color:    twapi.Ptr("#aaaaaa"),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.ProjectCategoryUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestProjectCategoryDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	projectCategoryID, _, err := createProjectCategory(t)
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	_, err = projects.ProjectCategoryDelete(ctx, engine, projects.NewProjectCategoryDeleteRequest(projectCategoryID))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestProjectCategoryGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	projectCategoryID, projectCategoryCleanup, err := createProjectCategory(t)
	if err != nil {
		t.Fatal(err)
	}
	defer projectCategoryCleanup()

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	_, err = projects.ProjectCategoryGet(ctx, engine, projects.NewProjectCategoryGetRequest(projectCategoryID))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestProjectCategoryList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	_, projectCategoryCleanup, err := createProjectCategory(t)
	if err != nil {
		t.Fatal(err)
	}
	defer projectCategoryCleanup()

	tests := []struct {
		name  string
		input projects.ProjectCategoryListRequest
	}{{
		name: "all project categories",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.ProjectCategoryList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
