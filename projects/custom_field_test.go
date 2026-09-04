package projects_test

import (
	"context"
	"encoding/json"
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
			Type:        projects.CustomFieldTypeRating,
			Entity:      projects.CustomFieldEntityProject,
			Description: new("integration test custom field"),
			Required:    new(false),
			ProjectID:   &testResources.ProjectID,
			Options: projects.CustomFieldOptionsRating{
				Icon: "star",
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

func TestCustomFieldUnmarshalJSONOptions(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		verify func(t *testing.T, cf projects.CustomField)
	}{{
		name: "dropdown choices nested under options",
		input: `{
			"id": 70375,
			"name": "TestNumbers",
			"type": "dropdown",
			"entity": "task",
			"options": {"choices": [
				{"value": "1", "color": "#4cd5e3"},
				{"value": "2", "color": "#e12d42"}
			]}
		}`,
		verify: func(t *testing.T, cf projects.CustomField) {
			options, ok := cf.Options.(projects.CustomFieldOptionsDropdown)
			if !ok {
				t.Fatalf("expected dropdown options, got %T", cf.Options)
			}
			if len(options.Choices) != 2 {
				t.Fatalf("expected 2 choices, got %d", len(options.Choices))
			}
			if options.Choices[0].Value != "1" || options.Choices[1].Value != "2" {
				t.Errorf("unexpected choice values: %+v", options.Choices)
			}
		},
	}, {
		name: "multiselect choices nested under options",
		input: `{
			"id": 71000,
			"name": "Tags",
			"type": "multiselect",
			"entity": "task",
			"options": {"choices": [
				{"value": "red", "color": "#4cd5e3"},
				{"value": "green", "color": "#e12d42"}
			]}
		}`,
		verify: func(t *testing.T, cf projects.CustomField) {
			options, ok := cf.Options.(projects.CustomFieldOptionsDropdown)
			if !ok {
				t.Fatalf("expected dropdown options for multiselect, got %T", cf.Options)
			}
			if len(options.Choices) != 2 {
				t.Fatalf("expected 2 choices, got %d", len(options.Choices))
			}
			if options.Choices[0].Value != "red" || options.Choices[1].Value != "green" {
				t.Errorf("unexpected choice values: %+v", options.Choices)
			}
		},
	}, {
		name: "status choices nested under options",
		input: `{
			"id": 71001,
			"name": "Stage",
			"type": "status",
			"entity": "task",
			"options": {"choices": [
				{"value": "open", "color": "#4cd5e3"}
			]}
		}`,
		verify: func(t *testing.T, cf projects.CustomField) {
			options, ok := cf.Options.(projects.CustomFieldOptionsDropdown)
			if !ok {
				t.Fatalf("expected dropdown options for status, got %T", cf.Options)
			}
			if len(options.Choices) != 1 || options.Choices[0].Value != "open" {
				t.Errorf("unexpected choices: %+v", options.Choices)
			}
		},
	}, {
		name: "rating options nested under options",
		input: `{
			"id": 1,
			"name": "Rating",
			"type": "rating",
			"entity": "task",
			"options": {"icon": "star", "color": "#ff0000"}
		}`,
		verify: func(t *testing.T, cf projects.CustomField) {
			options, ok := cf.Options.(projects.CustomFieldOptionsRating)
			if !ok {
				t.Fatalf("expected rating options, got %T", cf.Options)
			}
			if options.Icon != "star" {
				t.Errorf("expected icon 'star', got %q", options.Icon)
			}
		},
	}, {
		name: "number decimal options nested under options",
		input: `{
			"id": 2,
			"name": "Decimal",
			"type": "number-decimal",
			"entity": "task",
			"options": {"decimals": 3}
		}`,
		verify: func(t *testing.T, cf projects.CustomField) {
			options, ok := cf.Options.(projects.CustomFieldOptionsNumberDecimal)
			if !ok {
				t.Fatalf("expected number decimal options, got %T", cf.Options)
			}
			if options.DecimalPoints == nil || *options.DecimalPoints != 3 {
				t.Errorf("expected 3 decimal points, got %v", options.DecimalPoints)
			}
		},
	}, {
		name: "missing options is not an error",
		input: `{
			"id": 3,
			"name": "Text",
			"type": "text-short",
			"entity": "task"
		}`,
		verify: func(t *testing.T, cf projects.CustomField) {
			if cf.Options != nil {
				t.Errorf("expected nil options, got %+v", cf.Options)
			}
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cf projects.CustomField
			if err := json.Unmarshal([]byte(tt.input), &cf); err != nil {
				t.Fatalf("failed to unmarshal custom field: %s", err)
			}
			tt.verify(t, cf)
		})
	}
}

// TestCustomFieldOptionsColorIsOptional pins both halves of an optional colour
// on the choice-based options, which are the same structs on the way in and on
// the way out.
//
// A twapi.HexColor with no value encodes to the bare "#", which the endpoint
// rejects, so a request that sets no colour has to leave the key out entirely.
// The read direction has to accept that same "#", which is what the endpoint
// answers for a choice created without one.
func TestCustomFieldOptionsColorIsOptional(t *testing.T) {
	t.Run("an unset colour sends no key", func(t *testing.T) {
		encoded, err := json.Marshal(projects.CustomFieldOptionsDropdown{
			Choices: []projects.CustomFieldOptionsDropdownChoice{
				{Value: "Blocked"},
				{Value: "Shipped", Color: "8bc34a"},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error encoding the options: %s", err)
		}

		const want = `{"choices":[{"value":"Blocked"},{"value":"Shipped","color":"#8bc34a"}]}`
		if string(encoded) != want {
			t.Errorf("expected %s but got %s", want, encoded)
		}
	})

	t.Run("a colourless choice decodes", func(t *testing.T) {
		var options projects.CustomFieldOptionsDropdown
		if err := json.Unmarshal([]byte(`{"choices":[{"value":"Blocked","color":"#"}]}`), &options); err != nil {
			t.Fatalf("unexpected error decoding the options: %s", err)
		}
		if got := options.Choices[0].Color; got != "" {
			t.Errorf("expected the bare sign to decode as unset but got %q", got)
		}
	})

	t.Run("a rating colour behaves the same", func(t *testing.T) {
		encoded, err := json.Marshal(projects.CustomFieldOptionsRating{Icon: "star"})
		if err != nil {
			t.Fatalf("unexpected error encoding the options: %s", err)
		}
		if string(encoded) != `{"icon":"star"}` {
			t.Errorf(`expected {"icon":"star"} but got %s`, encoded)
		}
	})
}
