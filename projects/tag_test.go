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

func TestTagCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name  string
		input projects.TagCreateRequest
	}{{
		name:  "only required fields",
		input: projects.NewTagCreateRequest(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
	}, {
		name: "all fields",
		input: projects.TagCreateRequest{
			Name:      fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
			ProjectID: &testResources.ProjectID,
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			t.Cleanup(cancel)

			tagResponse, err := projects.TagCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx = context.Background() // t.Context is always canceled in cleanup
				_, err := projects.TagDelete(ctx, engine, projects.NewTagDeleteRequest(tagResponse.Tag.ID))
				if err != nil {
					t.Errorf("failed to delete tag after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if tagResponse.Tag.ID == 0 {
				t.Error("expected a valid tag ID but got 0")
			}
		})
	}
}

func TestTagUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tagID, tagCleanup, err := createTag(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(tagCleanup)

	tests := []struct {
		name  string
		input projects.TagUpdateRequest
	}{{
		name: "all fields",
		input: projects.TagUpdateRequest{
			Path: projects.TagUpdateRequestPath{
				ID: tagID,
			},
			Name:      twapi.Ptr(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
			ProjectID: &testResources.ProjectID,
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.TagUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestTagDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tagID, _, err := createTag(t)
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	t.Cleanup(cancel)

	_, err = projects.TagDelete(ctx, engine, projects.NewTagDeleteRequest(tagID))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestTagGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tagID, tagCleanup, err := createTag(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(tagCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.TagGet(ctx, engine, projects.NewTagGetRequest(tagID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestTagList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	_, tagCleanup, err := createTag(t)
	if err != nil {
		t.Fatal(err)
	}
	defer tagCleanup()

	tests := []struct {
		name          string
		input         projects.TagListRequest
		expectedError bool
	}{{
		name: "it should list tags",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := projects.TagList(ctx, engine, tt.input)
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
