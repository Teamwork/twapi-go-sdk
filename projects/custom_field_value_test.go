package projects_test

import (
	"context"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestCustomFieldValueCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	customFieldID, customFieldCleanup, err := createCustomField(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(customFieldCleanup)

	tests := []struct {
		name  string
		input projects.CustomFieldValueCreateRequest
	}{{
		name:  "task value",
		input: projects.NewTaskCustomFieldValueCreateRequest(testResources.TaskID, customFieldID, "task value"),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			t.Cleanup(cancel)

			valueResponse, err := projects.CustomFieldValueCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx := context.Background()
				request := projects.NewTaskCustomFieldValueDeleteRequest(
					testResources.TaskID,
					valueResponse.CustomFieldValue.ID,
				)
				_, err := projects.CustomFieldValueDelete(ctx, engine, request)
				if err != nil {
					t.Errorf("failed to delete custom field value after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if valueResponse.CustomFieldValue.ID == 0 {
				t.Error("expected a valid custom field value ID but got 0")
			}
		})
	}
}

func TestCustomFieldValueUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	customFieldID, customFieldCleanup, err := createCustomField(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(customFieldCleanup)

	valueID, valueCleanup, err := createTaskCustomFieldValue(t, testResources.TaskID, customFieldID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(valueCleanup)

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	request := projects.NewTaskCustomFieldValueUpdateRequest(
		testResources.TaskID,
		customFieldID,
		valueID,
		"updated value",
	)
	if _, err := projects.CustomFieldValueUpdate(ctx, engine, request); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestCustomFieldValueDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	customFieldID, customFieldCleanup, err := createCustomField(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(customFieldCleanup)

	valueID, _, err := createTaskCustomFieldValue(t, testResources.TaskID, customFieldID)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	request := projects.NewTaskCustomFieldValueDeleteRequest(testResources.TaskID, valueID)
	if _, err := projects.CustomFieldValueDelete(ctx, engine, request); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestCustomFieldValueGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	customFieldID, customFieldCleanup, err := createCustomField(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(customFieldCleanup)

	valueID, valueCleanup, err := createTaskCustomFieldValue(t, testResources.TaskID, customFieldID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(valueCleanup)

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	request := projects.NewTaskCustomFieldValueGetRequest(testResources.TaskID, valueID)
	if _, err := projects.CustomFieldValueGet(ctx, engine, request); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestCustomFieldValueList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	customFieldID, customFieldCleanup, err := createCustomField(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(customFieldCleanup)

	_, valueCleanup, err := createTaskCustomFieldValue(t, testResources.TaskID, customFieldID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(valueCleanup)

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	request := projects.NewTaskCustomFieldValueListRequest(testResources.TaskID)
	if _, err := projects.CustomFieldValueList(ctx, engine, request); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}
