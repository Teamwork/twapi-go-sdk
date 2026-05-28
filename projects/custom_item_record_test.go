package projects_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestCustomItemRecordCreate(t *testing.T) {
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
		input projects.CustomItemRecordCreateRequest
	}{{
		name: "only required fields",
		input: projects.NewCustomItemRecordCreateRequest(
			customItemID,
			fmt.Sprintf("record-%d%d", time.Now().UnixNano(), rand.Intn(100)),
		),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			t.Cleanup(cancel)

			recordResponse, err := projects.CustomItemRecordCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx := context.Background()
				_, err := projects.CustomItemRecordDelete(ctx, engine,
					projects.NewCustomItemRecordDeleteRequest(customItemID, recordResponse.CustomItemRecord.ID))
				if err != nil {
					t.Errorf("failed to delete custom item record after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if recordResponse.CustomItemRecord.ID == 0 {
				t.Error("expected a valid record ID but got 0")
			}
		})
	}
}

func TestCustomItemRecordUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	customItemID, customItemCleanup, err := createCustomItem(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(customItemCleanup)

	recordID, recordCleanup, err := createCustomItemRecord(t, customItemID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(recordCleanup)

	updatedName := fmt.Sprintf("renamed-%d%d", time.Now().UnixNano(), rand.Intn(100))

	tests := []struct {
		name  string
		input projects.CustomItemRecordUpdateRequest
	}{{
		name: "rename",
		input: projects.CustomItemRecordUpdateRequest{
			Path: projects.CustomItemRecordUpdateRequestPath{
				CustomItemID: customItemID,
				ID:           recordID,
			},
			Name: &updatedName,
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.CustomItemRecordUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestCustomItemRecordDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	customItemID, customItemCleanup, err := createCustomItem(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(customItemCleanup)

	recordID, _, err := createCustomItemRecord(t, customItemID)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	if _, err := projects.CustomItemRecordDelete(ctx, engine,
		projects.NewCustomItemRecordDeleteRequest(customItemID, recordID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestCustomItemRecordBulkDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	customItemID, customItemCleanup, err := createCustomItem(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(customItemCleanup)

	ids := make([]int64, 0, 2)
	for i := 0; i < 2; i++ {
		recordID, _, err := createCustomItemRecord(t, customItemID)
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, recordID)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	if _, err := projects.CustomItemRecordBulkDelete(ctx, engine,
		projects.NewCustomItemRecordBulkDeleteRequest(customItemID, ids)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestCustomItemRecordGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	customItemID, customItemCleanup, err := createCustomItem(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(customItemCleanup)

	recordID, recordCleanup, err := createCustomItemRecord(t, customItemID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(recordCleanup)

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	if _, err := projects.CustomItemRecordGet(ctx, engine,
		projects.NewCustomItemRecordGetRequest(customItemID, recordID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestCustomItemRecordList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	customItemID, customItemCleanup, err := createCustomItem(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(customItemCleanup)

	_, recordCleanup, err := createCustomItemRecord(t, customItemID)
	if err != nil {
		t.Fatal(err)
	}
	defer recordCleanup()

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()

	if _, err := projects.CustomItemRecordList(ctx, engine,
		projects.NewCustomItemRecordListRequest(customItemID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

// createCustomItemRecord is a test helper that creates a new record on the
// given custom item type and returns its ID along with a cleanup func.
func createCustomItemRecord(t testEngine, customItemID int64) (int64, func(), error) {
	name := fmt.Sprintf("test-record-%d%d", time.Now().UnixNano(), rand.Intn(100))
	recordResponse, err := projects.CustomItemRecordCreate(t.Context(), engine,
		projects.NewCustomItemRecordCreateRequest(customItemID, name))
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create custom item record for test: %w", err)
	}
	id := recordResponse.CustomItemRecord.ID
	return id, func() {
		ctx := context.Background() // t.Context is always canceled in cleanup
		_, err := projects.CustomItemRecordDelete(ctx, engine,
			projects.NewCustomItemRecordDeleteRequest(customItemID, id))
		if err != nil {
			t.Errorf("failed to delete custom item record after test: %s", err)
		}
	}, nil
}
