package projects_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestCustomItemFieldCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	customItemID, customItemCleanup, err := createCustomItem(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(customItemCleanup)

	tests := []struct {
		name  string
		input projects.CustomItemFieldCreateRequest
	}{{
		name: "only required fields",
		input: projects.NewCustomItemFieldCreateRequest(
			customItemID,
			fmt.Sprintf("field-%d%d", time.Now().UnixNano(), rand.Intn(100)),
			projects.CustomItemFieldTypeTextShort,
		),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			t.Cleanup(cancel)

			fieldResponse, err := projects.CustomItemFieldCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx := context.Background()
				_, err := projects.CustomItemFieldDelete(ctx, engine,
					projects.NewCustomItemFieldDeleteRequest(customItemID, fieldResponse.CustomItemField.ID))
				if err != nil {
					t.Errorf("failed to delete custom item field after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if fieldResponse.CustomItemField.ID == 0 {
				t.Error("expected a valid field ID but got 0")
			}
		})
	}
}

func TestCustomItemFieldUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	customItemID, customItemCleanup, err := createCustomItem(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(customItemCleanup)

	fieldID, fieldCleanup, err := createCustomItemField(t, customItemID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(fieldCleanup)

	updatedName := fmt.Sprintf("renamed-%d%d", time.Now().UnixNano(), rand.Intn(100))

	tests := []struct {
		name  string
		input projects.CustomItemFieldUpdateRequest
	}{{
		name: "rename",
		input: projects.CustomItemFieldUpdateRequest{
			Path: projects.CustomItemFieldUpdateRequestPath{
				CustomItemID: customItemID,
				ID:           fieldID,
			},
			DisplayName: &updatedName,
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.CustomItemFieldUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestCustomItemFieldDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	customItemID, customItemCleanup, err := createCustomItem(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(customItemCleanup)

	fieldID, _, err := createCustomItemField(t, customItemID)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	if _, err := projects.CustomItemFieldDelete(ctx, engine,
		projects.NewCustomItemFieldDeleteRequest(customItemID, fieldID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestCustomItemFieldGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	customItemID, customItemCleanup, err := createCustomItem(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(customItemCleanup)

	fieldID, fieldCleanup, err := createCustomItemField(t, customItemID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(fieldCleanup)

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	if _, err := projects.CustomItemFieldGet(ctx, engine,
		projects.NewCustomItemFieldGetRequest(customItemID, fieldID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestCustomItemFieldList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	customItemID, customItemCleanup, err := createCustomItem(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(customItemCleanup)

	_, fieldCleanup, err := createCustomItemField(t, customItemID)
	if err != nil {
		t.Fatal(err)
	}
	defer fieldCleanup()

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()

	if _, err := projects.CustomItemFieldList(ctx, engine,
		projects.NewCustomItemFieldListRequest(customItemID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

// createCustomItemField is a test helper that creates a new field on the
// given custom item type and returns its ID along with a cleanup func.
func createCustomItemField(t testEngine, customItemID int64) (int64, func(), error) {
	name := fmt.Sprintf("test-field-%d%d", time.Now().UnixNano(), rand.Intn(100))
	fieldResponse, err := projects.CustomItemFieldCreate(t.Context(), engine,
		projects.NewCustomItemFieldCreateRequest(customItemID, name, projects.CustomItemFieldTypeTextShort))
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create custom item field for test: %w", err)
	}
	id := fieldResponse.CustomItemField.ID
	return id, func() {
		ctx := context.Background() // t.Context is always canceled in cleanup
		_, err := projects.CustomItemFieldDelete(ctx, engine,
			projects.NewCustomItemFieldDeleteRequest(customItemID, id))
		if err != nil {
			t.Errorf("failed to delete custom item field after test: %s", err)
		}
	}, nil
}
