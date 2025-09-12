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

func TestNotebookCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name  string
		input projects.NotebookCreateRequest
	}{{
		name: "only required fields",
		input: projects.NewNotebookCreateRequest(
			testResources.ProjectID,
			fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
			"An amazing content",
			projects.NotebookTypeMarkdown,
		),
	}, {
		name: "all fields",
		input: projects.NotebookCreateRequest{
			Path: projects.NotebookCreateRequestPath{
				ProjectID: testResources.ProjectID,
			},
			Name:        fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
			Description: twapi.Ptr("This is a test notebook"),
			Contents:    "An amazing content",
			Type:        projects.NotebookTypeMarkdown,
			TagIDs:      []int64{testResources.TagID},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			t.Cleanup(cancel)

			notebookResponse, err := projects.NotebookCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx = context.Background() // t.Context is always canceled in cleanup
				_, err := projects.NotebookDelete(ctx, engine, projects.NewNotebookDeleteRequest(notebookResponse.Notebook.ID))
				if err != nil {
					t.Errorf("failed to delete notebook after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if notebookResponse.Notebook.ID == 0 {
				t.Error("expected a valid notebook ID but got 0")
			}
		})
	}
}

func TestNotebookUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	notebookID, notebookCleanup, err := createNotebook(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(notebookCleanup)

	tests := []struct {
		name  string
		input projects.NotebookUpdateRequest
	}{{
		name: "all fields",
		input: projects.NotebookUpdateRequest{
			Path: projects.NotebookUpdateRequestPath{
				ID: notebookID,
			},
			Name:        twapi.Ptr(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Description: twapi.Ptr("This is a test notebook"),
			Contents:    twapi.Ptr("<h1>An amazing content updated</h1>"),
			Type:        twapi.Ptr(projects.NotebookTypeHTML),
			TagIDs:      []int64{testResources.TagID},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.NotebookUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestNotebookDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	notebookID, _, err := createNotebook(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.NotebookDelete(ctx, engine, projects.NewNotebookDeleteRequest(notebookID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestNotebookGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	notebookID, notebookCleanup, err := createNotebook(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(notebookCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.NotebookGet(ctx, engine, projects.NewNotebookGetRequest(notebookID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestNotebookList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	_, notebookCleanup, err := createNotebook(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(notebookCleanup)

	tests := []struct {
		name  string
		input projects.NotebookListRequest
	}{{
		name: "all notebooks",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.NotebookList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
