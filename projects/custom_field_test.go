package projects_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestCustomFieldCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name  string
		input projects.CustomFieldCreateRequest
	}{{
		name: "only required fields",
		input: projects.NewCustomFieldCreateRequest(
			fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
			projects.CustomFieldTypeTextLong,
			projects.CustomFieldEntityTask,
		),
	}, {
		name: "all fields",
		input: projects.CustomFieldCreateRequest{
			Name:        fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
			Type:        projects.CustomFieldTypeNumberDecimal,
			Entity:      projects.CustomFieldEntityProject,
			Description: new("integration test custom field"),
			Required:    new(false),
			ProjectID:   &testResources.ProjectID,
			Options: projects.CustomFieldOptionsNumberDecimal{
				DecimalPoints: new(2),
			},
			CurrencyCode: new("USD"),
			Unit:         new(projects.CustomFieldUnitPercent),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			t.Cleanup(cancel)

			customFieldResponse, err := projects.CustomFieldCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx := context.Background()
				_, err := projects.CustomFieldDelete(ctx, engine,
					projects.NewCustomFieldDeleteRequest(customFieldResponse.CustomField.ID))
				if err != nil {
					t.Errorf("failed to delete custom field after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if customFieldResponse.CustomField.ID == 0 {
				t.Error("expected a valid custom field ID but got 0")
			}
		})
	}
}

func TestCustomFieldUpdate(t *testing.T) {
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
		input projects.CustomFieldUpdateRequest
	}{{
		name: "all fields",
		input: projects.CustomFieldUpdateRequest{
			Path: projects.CustomFieldUpdateRequestPath{
				ID: customFieldID,
			},
			Name:        new(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Description: new("updated description"),
			Required:    new(true),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.CustomFieldUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestCustomFieldDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	customFieldID, _, err := createCustomField(t)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.CustomFieldDelete(ctx, engine, projects.NewCustomFieldDeleteRequest(customFieldID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestCustomFieldGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	customFieldID, customFieldCleanup, err := createCustomField(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(customFieldCleanup)

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.CustomFieldGet(ctx, engine, projects.NewCustomFieldGetRequest(customFieldID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestCustomFieldList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	_, customFieldCleanup, err := createCustomField(t)
	if err != nil {
		t.Fatal(err)
	}
	defer customFieldCleanup()

	tests := []struct {
		name          string
		input         projects.CustomFieldListRequest
		expectedError bool
	}{{
		name: "all custom fields",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()

			_, err := projects.CustomFieldList(ctx, engine, tt.input)
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
