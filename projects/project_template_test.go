package projects_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestProjectTemplateCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name  string
		input projects.ProjectTemplateCreateRequest
	}{{
		name:  "only required fields",
		input: projects.NewProjectTemplateCreateRequest(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
	}, {
		name: "all fields",
		input: projects.ProjectTemplateCreateRequest{
			ProjectCreateRequest: projects.ProjectCreateRequest{
				Name:        fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
				Description: new("This is a test project"),
				StartAt:     new(projects.NewLegacyDate(time.Now().Add(24 * time.Hour))),
				EndAt:       new(projects.NewLegacyDate(time.Now().Add(48 * time.Hour))),
				CategoryID:  &testResources.ProjectCategoryID,
				CompanyID:   testResources.CompanyID,
				OwnerID:     &testResources.UserID,
				TagIDs:      []int64{testResources.TagID},
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			project, err := projects.ProjectTemplateCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx = context.Background() // t.Context is always canceled in cleanup
				_, err := projects.ProjectDelete(ctx, engine, projects.NewProjectDeleteRequest(int64(project.ID)))
				if err != nil {
					t.Errorf("failed to delete project after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if project.ID == 0 {
				t.Error("expected a valid project ID but got 0")
			}
		})
	}
}

func TestProjectTemplateList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	_, projectCleanup, err := createProject(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(projectCleanup)

	tests := []struct {
		name  string
		input projects.ProjectTemplateListRequest
	}{{
		name: "all project templates",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.ProjectTemplateList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
