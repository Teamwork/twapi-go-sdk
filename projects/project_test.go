package projects_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestProjectCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name  string
		input projects.ProjectCreateRequest
	}{{
		name:  "only required fields",
		input: projects.NewProjectCreateRequest(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
	}, {
		name: "all fields",
		input: projects.ProjectCreateRequest{
			Name:        fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
			Description: new("This is a test project"),
			StartAt:     new(projects.NewLegacyDate(time.Now().Add(24 * time.Hour))),
			EndAt:       new(projects.NewLegacyDate(time.Now().Add(48 * time.Hour))),
			CategoryID:  &testResources.ProjectCategoryID,
			CompanyID:   testResources.CompanyID,
			OwnerID:     &testResources.UserID,
			TagIDs:      []int64{testResources.TagID},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			project, err := projects.ProjectCreate(ctx, engine, tt.input)
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

func TestProjectUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	projectID, projectCleanup, err := createProject(t)
	if err != nil {
		t.Fatal(err)
	}
	defer projectCleanup()

	tests := []struct {
		name  string
		input projects.ProjectUpdateRequest
	}{{
		name: "all fields",
		input: projects.ProjectUpdateRequest{
			Path: projects.ProjectUpdateRequestPath{
				ID: projectID,
			},
			Name:        new(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Description: new("This is a test project"),
			StartAt:     new(projects.NewLegacyDate(time.Now().Add(24 * time.Hour))),
			EndAt:       new(projects.NewLegacyDate(time.Now().Add(48 * time.Hour))),
			CategoryID:  &testResources.ProjectCategoryID,
			CompanyID:   &testResources.CompanyID,
			OwnerID:     &testResources.UserID,
			TagIDs:      []int64{testResources.TagID},
			Status:      new(projects.ProjectStatusArchived),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.ProjectUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestProjectDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	projectID, _, err := createProject(t)
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.ProjectDelete(ctx, engine, projects.NewProjectDeleteRequest(projectID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestProjectClone(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	projectID, projectCleanup, err := createProject(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(projectCleanup)

	projectTemplateID, projectTemplateCleanup, err := createProjectTemplate(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(projectTemplateCleanup)

	movedProjectID, _, err := createProject(t)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		input projects.ProjectCloneRequest
	}{{
		name: "clone regular project",
		input: projects.ProjectCloneRequest{
			Path: projects.ProjectCloneRequestPath{
				ID: projectID,
			},
			Name:                   new(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Description:            new("This is a cloned test project"),
			CompanyID:              &testResources.CompanyID,
			CopyFiles:              new(true),
			CopyMessages:           new(true),
			CopyMilestones:         new(true),
			CopyTasks:              new(true),
			CopyTasklists:          new(true),
			CopyNotebooks:          new(true),
			CopyLinks:              new(true),
			CopyComments:           new(true),
			CopyFollowers:          new(true),
			CopyInvoices:           new(true),
			CopyTimelogs:           new(true),
			CopyExpenses:           new(true),
			CopyWebhooks:           new(true),
			CopyProjectRoles:       new(true),
			CopyCustomFields:       new(true),
			CopyCustomItems:        new(true),
			CopyProjectUpdates:     new(true),
			CopyRisks:              new(true),
			CopyForms:              new(true),
			CopyAutomations:        new(true),
			CopyPeople:             new(true),
			CopyProjectPrivacy:     new(true),
			CopyBudgets:            new(true),
			CopyAllocations:        new(true),
			CopyLogo:               new(true),
			CopyProjectPreferences: new(true),
		},
	}, {
		name: "clone template project",
		input: projects.ProjectCloneRequest{
			Path: projects.ProjectCloneRequestPath{
				ID: projectTemplateID,
			},
			Name: new(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
		},
	}, {
		name: "clone regular project to template",
		input: projects.ProjectCloneRequest{
			Path: projects.ProjectCloneRequestPath{
				ID: projectID,
			},
			Name:            new(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
			NewFromTemplate: new(false),
			ToTemplate:      new(true),
		},
	}, {
		name: "clone template project to regular",
		input: projects.ProjectCloneRequest{
			Path: projects.ProjectCloneRequestPath{
				ID: projectTemplateID,
			},
			Name:               new(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
			NewFromTemplate:    new(true),
			ToTemplate:         new(false),
			TemplateDateTarget: new(projects.ProjectCloneTemplateDateTargetEnd),
			TargetDate:         new(projects.NewLegacyDate(time.Now().Add(7 * 24 * time.Hour))),
			DaysOffset:         new(int64(10)),
		},
	}, {
		name: "move regular project",
		input: projects.ProjectCloneRequest{
			Path: projects.ProjectCloneRequestPath{
				ID: movedProjectID,
			},
			Name:   new(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Action: new(projects.ProjectCloneActionMove),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
			t.Cleanup(cancel)

			project, err := projects.ProjectClone(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx = context.Background() // t.Context is always canceled in cleanup
				_, err := projects.ProjectDelete(ctx, engine, projects.NewProjectDeleteRequest(int64(project.ID)))
				if err != nil {
					t.Errorf("failed to delete cloned project after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestProjectGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	projectID, projectCleanup, err := createProject(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(projectCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.ProjectGet(ctx, engine, projects.NewProjectGetRequest(projectID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestProjectList(t *testing.T) {
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
		input projects.ProjectListRequest
	}{{
		name: "all projects",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.ProjectList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
