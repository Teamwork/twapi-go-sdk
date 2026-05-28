package projects_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestCustomItemCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	displayName := fmt.Sprintf("test-ci-%d%d", time.Now().UnixNano(), rand.Intn(100))
	labelSingular := "Test Item"
	labelPlural := "Test Items"

	minimal := projects.NewCustomItemCreateRequest(testResources.ProjectID, displayName)
	minimal.LabelSingular = &labelSingular
	minimal.LabelPlural = &labelPlural

	tests := []struct {
		name  string
		input projects.CustomItemCreateRequest
	}{{
		name:  "minimal request",
		input: minimal,
	}, {
		name: "all fields",
		input: projects.CustomItemCreateRequest{
			Path:          projects.CustomItemCreateRequestPath{ProjectID: testResources.ProjectID},
			DisplayName:   displayName + "-full",
			LabelSingular: &labelSingular,
			LabelPlural:   &labelPlural,
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			t.Cleanup(cancel)

			customItemResponse, err := projects.CustomItemCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx := context.Background()
				_, err := projects.CustomItemDelete(ctx, engine,
					projects.NewCustomItemDeleteRequest(customItemResponse.CustomItem.ID))
				if err != nil {
					t.Errorf("failed to delete custom item after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if customItemResponse.CustomItem.ID == 0 {
				t.Error("expected a valid custom item ID but got 0")
			}
		})
	}
}

func TestCustomItemUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	customItemID, customItemCleanup, err := createCustomItem(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(customItemCleanup)

	updatedName := fmt.Sprintf("test-ci-up-%d%d", time.Now().UnixNano(), rand.Intn(100))
	updatedDescription := "updated by integration test"

	tests := []struct {
		name  string
		input projects.CustomItemUpdateRequest
	}{{
		name: "rename and re-describe",
		input: projects.CustomItemUpdateRequest{
			Path:        projects.CustomItemUpdateRequestPath{ID: customItemID},
			DisplayName: &updatedName,
			Description: &updatedDescription,
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.CustomItemUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestCustomItemDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	customItemID, _, err := createCustomItem(t)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	if _, err := projects.CustomItemDelete(ctx, engine,
		projects.NewCustomItemDeleteRequest(customItemID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestCustomItemGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	customItemID, customItemCleanup, err := createCustomItem(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(customItemCleanup)

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	if _, err := projects.CustomItemGet(ctx, engine,
		projects.NewCustomItemGetRequest(customItemID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestCustomItemList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	_, customItemCleanup, err := createCustomItem(t)
	if err != nil {
		t.Fatal(err)
	}
	defer customItemCleanup()

	tests := []struct {
		name          string
		input         projects.CustomItemListRequest
		expectedError bool
	}{{
		name:  "all custom items on project",
		input: projects.NewCustomItemListRequest(testResources.ProjectID),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()

			_, err := projects.CustomItemList(ctx, engine, tt.input)
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

// createCustomItem is a test helper that creates a new custom item type on
// the shared test project and returns its ID along with a cleanup func that
// deletes it. LabelSingular and LabelPlural are required by the API.
func createCustomItem(t testEngine) (int64, func(), error) {
	displayName := fmt.Sprintf("test-ci-h-%d%d", time.Now().UnixNano(), rand.Intn(100))
	singular := "Test Record"
	plural := "Test Records"
	req := projects.NewCustomItemCreateRequest(testResources.ProjectID, displayName)
	req.LabelSingular = &singular
	req.LabelPlural = &plural
	customItemResponse, err := projects.CustomItemCreate(t.Context(), engine, req)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create custom item for test: %w", err)
	}
	id := customItemResponse.CustomItem.ID
	return id, func() {
		ctx := context.Background() // t.Context is always canceled in cleanup
		_, err := projects.CustomItemDelete(ctx, engine,
			projects.NewCustomItemDeleteRequest(id))
		if err != nil {
			t.Errorf("failed to delete custom item after test: %s", err)
		}
	}, nil
}
