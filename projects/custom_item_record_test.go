package projects_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
	"github.com/teamwork/twapi-go-sdk/session"
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

func TestCustomItemRecordBulkCreate(t *testing.T) {
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

	suffix := fmt.Sprintf("%d%d", time.Now().UnixNano(), rand.Intn(100))
	req := projects.NewCustomItemRecordBulkCreateRequest(customItemID, []projects.CustomItemRecordBulkCreateRecord{
		{Name: "test-record-a-" + suffix},
		{Name: "test-record-b-" + suffix},
	})
	req.Options.PartialSuccess = new(true)

	response, err := projects.CustomItemRecordBulkCreate(ctx, engine, req)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	ids := make([]int64, 0, len(response.CustomItemRecords))
	for _, record := range response.CustomItemRecords {
		ids = append(ids, record.ID)
	}
	t.Cleanup(func() {
		if len(ids) == 0 {
			return
		}
		_, err := projects.CustomItemRecordBulkDelete(context.Background(), engine,
			projects.NewCustomItemRecordBulkDeleteRequest(customItemID, ids))
		if err != nil {
			t.Errorf("failed to delete custom item records after test: %s", err)
		}
	})

	if len(response.CustomItemRecords) != 2 {
		t.Errorf("expected 2 created records, got %d", len(response.CustomItemRecords))
	}
	if len(response.Failed) != 0 {
		t.Errorf("expected no failures, got %+v", response.Failed)
	}
}

// TestCustomItemRecordBulkCreateWire pins the route, the body and the decoding
// of a partial success against a local server, so it runs without a live site.
func TestCustomItemRecordBulkCreateWire(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	mux := http.NewServeMux()
	mux.HandleFunc("/projects/api/v3/customitems/777/records/bulk.json", func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &gotBody); err != nil {
			t.Errorf("request body is not JSON: %s", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"customItemRecords":[{"id":12345,"name":"First"},{"id":12346,"name":"Third"}],`+
			`"failed":[{"index":1,"errors":[{"title":"invalid payload","detail":"name is required",`+
			`"meta":{"field":"name"}}]}],"meta":{}}`)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	testEngine := twapi.NewEngine(session.NewBearerToken("token", server.URL))
	req := projects.NewCustomItemRecordBulkCreateRequest(777, []projects.CustomItemRecordBulkCreateRecord{
		{Name: "First", SectionID: new(int64(12)), FieldValues: projects.CustomItemRecordFieldValues{"f_1": "a"}},
		{Name: ""},
		{Name: "Third", SectionName: new("Backlog")},
	})
	req.Options.PartialSuccess = new(true)
	req.Options.FireWebhook = new(false)

	response, err := projects.CustomItemRecordBulkCreate(t.Context(), testEngine, req)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", gotMethod)
	}
	if gotPath != "/projects/api/v3/customitems/777/records/bulk.json" {
		t.Errorf("unexpected path %s", gotPath)
	}
	options, _ := gotBody["options"].(map[string]any)
	if options["partialSuccess"] != true || options["fireWebhook"] != false {
		t.Errorf("unexpected options %v", options)
	}
	if _, ok := options["useNotifyViaTWIM"]; ok {
		t.Errorf("an unset option must not be sent, got %v", options)
	}
	records, _ := gotBody["records"].([]any)
	if len(records) != 3 {
		t.Fatalf("expected 3 records on the wire, got %v", gotBody["records"])
	}
	first, _ := records[0].(map[string]any)
	if first["name"] != "First" || first["sectionId"] != float64(12) {
		t.Errorf("unexpected first record %v", first)
	}
	if values, _ := first["fieldValues"].(map[string]any); values["f_1"] != "a" {
		t.Errorf("unexpected field values %v", first["fieldValues"])
	}
	third, _ := records[2].(map[string]any)
	if third["sectionName"] != "Backlog" {
		t.Errorf("unexpected third record %v", third)
	}
	if _, ok := third["sectionId"]; ok {
		t.Errorf("an unset section ID must not be sent, got %v", third)
	}

	if len(response.CustomItemRecords) != 2 || response.CustomItemRecords[1].ID != 12346 {
		t.Errorf("unexpected created records %+v", response.CustomItemRecords)
	}
	if len(response.Failed) != 1 || response.Failed[0].Index != 1 ||
		len(response.Failed[0].Errors) != 1 || response.Failed[0].Errors[0].Detail != "name is required" {
		t.Errorf("unexpected failures %+v", response.Failed)
	}
}

// TestCustomItemRecordBulkCreateAllFailed checks that a batch with nothing
// created surfaces as an HTTP error carrying the API's body.
func TestCustomItemRecordBulkCreateAllFailed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"errors":[{"title":"invalid payload","detail":"name is required","meta":{"index":0}}]}`)
	}))
	t.Cleanup(server.Close)

	testEngine := twapi.NewEngine(session.NewBearerToken("token", server.URL))
	_, err := projects.CustomItemRecordBulkCreate(t.Context(), testEngine,
		projects.NewCustomItemRecordBulkCreateRequest(777, []projects.CustomItemRecordBulkCreateRecord{{Name: ""}}))

	var httpErr *twapi.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected an HTTP error, got %v", err)
	}
	if httpErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", httpErr.StatusCode)
	}
}

func TestCustomItemRecordBulkCreateRejectsAnEmptyBatch(t *testing.T) {
	_, err := projects.NewCustomItemRecordBulkCreateRequest(777, nil).HTTPRequest(t.Context(), "http://localhost")
	if err == nil {
		t.Error("expected an error for a batch with no records")
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
