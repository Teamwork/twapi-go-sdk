package projects_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
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

// TestCustomFieldValueListResponseCount covers the decoder rather than the
// query wiring: this response owns an UnmarshalJSON that has to pick the right
// entity-specific wrapper key, and the count has to survive that detour for
// every one of them.
func TestCustomFieldValueListResponseCount(t *testing.T) {
	count := int64(51)

	tests := []struct {
		name    string
		payload string
	}{{
		name:    "task values",
		payload: `{"meta":{"page":{"count":51,"hasMore":true}},"customfieldTasks":[{"id":12345}]}`,
	}, {
		name:    "project values",
		payload: `{"meta":{"page":{"count":51,"hasMore":true}},"customfieldProjects":[{"id":12345}]}`,
	}, {
		name:    "company values",
		payload: `{"meta":{"page":{"count":51,"hasMore":true}},"customfieldCompanies":[{"id":12345}]}`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response projects.CustomFieldValueListResponse
			httpResp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(tt.payload)),
			}
			if err := response.HandleHTTPResponse(httpResp); err != nil {
				t.Fatalf("unexpected error handling response: %s", err)
			}

			req := projects.NewTaskCustomFieldValueListRequest(777)
			req.Filters.CountMode = twapi.ListCountModeExact
			response.SetRequest(req)

			if response.Meta.Page.Count == nil {
				t.Fatal("expected the count to survive the custom decoder but got nil")
			}
			if got := *response.Meta.Page.Count; got != count {
				t.Errorf("expected count %d but got %d", count, got)
			}
			if !response.Meta.Page.HasMore {
				t.Error("expected hasMore to be true")
			}
			if len(response.CustomFieldValues) != 1 {
				t.Errorf("expected 1 custom field value but got %d", len(response.CustomFieldValues))
			}
		})
	}
}

// TestCustomFieldValueListResponseSkippedCount asserts the skipped-count
// reconciliation still applies on this response, since its custom decoder
// bypasses the shared struct tags.
func TestCustomFieldValueListResponseSkippedCount(t *testing.T) {
	const payload = `{"meta":{"page":{"count":51,"hasMore":true}},"customfieldTasks":[{"id":12345}]}`

	var response projects.CustomFieldValueListResponse
	httpResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(payload)),
	}
	if err := response.HandleHTTPResponse(httpResp); err != nil {
		t.Fatalf("unexpected error handling response: %s", err)
	}

	req := projects.NewTaskCustomFieldValueListRequest(777)
	req.Filters.CountMode = twapi.ListCountModeSkip
	response.SetRequest(req)

	if response.Meta.Page.Count != nil {
		t.Errorf("expected no count but got %d", *response.Meta.Page.Count)
	}
}
