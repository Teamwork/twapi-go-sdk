package projects_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
	"github.com/teamwork/twapi-go-sdk/session"
)

// bulkRecorder is a test server that records the one request it receives and
// answers it with a fixed status, headers and body.
type bulkRecorder struct {
	method string
	path   string
	body   map[string]any
}

func startBulkServer(
	t *testing.T,
	status int,
	headers map[string]string,
	response string,
) (*twapi.Engine, *bulkRecorder) {
	t.Helper()

	recorder := &bulkRecorder{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder.method = r.Method
		recorder.path = r.URL.Path
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %s", err)
		}
		if err := json.Unmarshal(raw, &recorder.body); err != nil {
			t.Errorf("failed to decode request body %q: %s", raw, err)
		}
		for key, value := range headers {
			w.Header().Set(key, value)
		}
		w.WriteHeader(status)
		_, _ = io.WriteString(w, response)
	}))
	t.Cleanup(server.Close)

	return twapi.NewEngine(session.NewBearerToken("token", server.URL)), recorder
}

func TestTimelogBulkMoveRequest(t *testing.T) {
	engine, recorder := startBulkServer(t, http.StatusOK, nil, `{"STATUS":"OK","numberOfMovedTimelogs":"2"}`)

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)

	response, err := projects.TimelogBulkMove(ctx, engine,
		projects.NewTimelogBulkMoveRequest(777, 778, []int64{12345, 12346}))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if response.MovedCount != 2 {
		t.Errorf("expected 2 moved timelogs, got %d", response.MovedCount)
	}

	if recorder.method != http.MethodPut {
		t.Errorf("expected method PUT, got %s", recorder.method)
	}
	if want := "/projects/777/time_entries/move.json"; recorder.path != want {
		t.Errorf("expected path %q, got %q", want, recorder.path)
	}
	want := map[string]any{"projectId": float64(778), "timelogIdList": "12345,12346"}
	if len(recorder.body) != len(want) {
		t.Errorf("expected body %v, got %v", want, recorder.body)
	}
	for key, value := range want {
		if recorder.body[key] != value {
			t.Errorf("expected %q to be %v, got %v", key, value, recorder.body[key])
		}
	}
}

// TestTimelogBulkMoveErrorBody pins that a 200 answer carrying no STATUS is not
// read as a move. The endpoint answers that way when nothing could be moved.
func TestTimelogBulkMoveErrorBody(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{{
		name:   "nothing to move",
		status: http.StatusOK,
		body:   `{"errMsg":"No valid time logs to move"}`,
	}, {
		name:   "empty body",
		status: http.StatusOK,
		body:   ``,
	}, {
		name:   "error status",
		status: http.StatusForbidden,
		body:   `{"STATUS":"Error","MESSAGE":"Forbidden"}`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, _ := startBulkServer(t, tt.status, nil, tt.body)

			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			t.Cleanup(cancel)

			_, err := projects.TimelogBulkMove(ctx, engine,
				projects.NewTimelogBulkMoveRequest(777, 778, []int64{12345}))
			var httpErr *twapi.HTTPError
			if !errors.As(err, &httpErr) {
				t.Fatalf("expected a twapi.HTTPError, got %v", err)
			}
			if httpErr.StatusCode != tt.status {
				t.Errorf("expected status %d, got %d", tt.status, httpErr.StatusCode)
			}
		})
	}
}

func TestTimelogBulkBillableUpdateRequest(t *testing.T) {
	engine, recorder := startBulkServer(t, http.StatusOK, map[string]string{"count": "3"}, `{"STATUS":"OK"}`)

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)

	request := projects.NewTimelogBulkBillableUpdateRequest(false, 777, 778)
	request.UserIDs = []int64{12345}
	request.FromDate = new(projects.NewLegacyDate(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)))
	request.ToDate = new(projects.NewLegacyDate(time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)))
	request.BillableType = projects.TimelogBillableTypeBillable
	request.InvoicedType = projects.TimelogInvoicedTypeNonInvoiced
	request.TagIDs = []int64{1, 2}
	request.MatchAllTags = new(true)

	response, err := projects.TimelogBulkBillableUpdate(ctx, engine, request)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if response.Count == nil || *response.Count != 3 {
		t.Errorf("expected a count of 3, got %v", response.Count)
	}

	if recorder.method != http.MethodPut {
		t.Errorf("expected method PUT, got %s", recorder.method)
	}
	if want := "/time_entries.json"; recorder.path != want {
		t.Errorf("expected path %q, got %q", want, recorder.path)
	}
	want := map[string]any{
		"isBillable":   false,
		"projectId":    "777,778",
		"userId":       "12345",
		"fromDate":     "20260801",
		"toDate":       "20260831",
		"billableType": "billable",
		"invoicedType": "noninvoiced",
		"tagIds":       "1,2",
		"matchAllTags": true,
	}
	if len(recorder.body) != len(want) {
		t.Errorf("expected body %v, got %v", want, recorder.body)
	}
	for key, value := range want {
		if recorder.body[key] != value {
			t.Errorf("expected %q to be %v, got %v", key, value, recorder.body[key])
		}
	}
}

// TestTimelogBulkBillableUpdateOmitsUnsetFilters pins that an unset filter
// sends no key: the endpoint reads an empty value as "no restriction", so a
// stray key could only widen the update.
func TestTimelogBulkBillableUpdateOmitsUnsetFilters(t *testing.T) {
	engine, recorder := startBulkServer(t, http.StatusOK, nil, `{"STATUS":"OK"}`)

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)

	response, err := projects.TimelogBulkBillableUpdate(ctx, engine,
		projects.NewTimelogBulkBillableUpdateRequest(true, 777))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if response.Count != nil {
		t.Errorf("expected no count without the header, got %d", *response.Count)
	}
	want := map[string]any{"isBillable": true, "projectId": "777"}
	if len(recorder.body) != len(want) {
		t.Errorf("expected body %v, got %v", want, recorder.body)
	}
	for key, value := range want {
		if recorder.body[key] != value {
			t.Errorf("expected %q to be %v, got %v", key, value, recorder.body[key])
		}
	}
}

// TestTimelogBulkBillableUpdateNothingMatched pins the answer when the filters
// match no timelog: a 422, with nothing changed.
func TestTimelogBulkBillableUpdateNothingMatched(t *testing.T) {
	engine, _ := startBulkServer(t, http.StatusUnprocessableEntity, nil,
		`{"STATUS":"Error","MESSAGE":"Nothing to modify based on selected options"}`)

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)

	_, err := projects.TimelogBulkBillableUpdate(ctx, engine,
		projects.NewTimelogBulkBillableUpdateRequest(true, 777))
	var httpErr *twapi.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected a twapi.HTTPError, got %v", err)
	}
	if httpErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", httpErr.StatusCode)
	}
}

func TestTimelogBulkCreateRequest(t *testing.T) {
	engine, recorder := startBulkServer(t, http.StatusCreated, nil, `{
		"timelogs":[{"id":12345},{"id":12346}],
		"failed":[{"index":1,"errors":[{"title":"invalid payload","detail":"task is required"}]}]
	}`)

	loggedAt := time.Date(2026, 10, 20, 9, 30, 0, 0, time.UTC)
	request := projects.NewTimelogBulkCreateRequest([]projects.TimelogBulkCreateRequestTimelog{
		{ProjectID: new(int64(777)), Date: twapi.Date(loggedAt), Time: twapi.Time(loggedAt), IsUTC: true, Hours: 1},
		{ProjectID: new(int64(777)), Date: twapi.Date(loggedAt), Time: twapi.Time(loggedAt), Minutes: 30},
		{TaskID: new(int64(12345)), Date: twapi.Date(loggedAt), Time: twapi.Time(loggedAt), Minutes: 15,
			Billable: new(false), TagIDs: []int64{1, 2}},
	})
	response, err := projects.TimelogBulkCreate(t.Context(), engine, request)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if recorder.method != http.MethodPost || recorder.path != "/projects/api/v3/time/bulk.json" {
		t.Errorf("unexpected request %s %s", recorder.method, recorder.path)
	}
	options, _ := recorder.body["timelogOptions"].(map[string]any)
	if options["partialSuccess"] != true {
		t.Errorf("expected timelogOptions.partialSuccess to be true, got %v", recorder.body["timelogOptions"])
	}
	entries, _ := recorder.body["timelogs"].([]any)
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %v", recorder.body["timelogs"])
	}
	first := entries[0].(map[string]any)["timelog"].(map[string]any)
	if first["projectId"] != float64(777) || first["date"] != "2026-10-20" || first["time"] != "09:30:00" {
		t.Errorf("unexpected first entry: %v", first)
	}
	if _, ok := first["taskId"]; ok {
		t.Errorf("expected no taskId on a project entry, got %v", first)
	}
	third := entries[2].(map[string]any)["timelog"].(map[string]any)
	if third["taskId"] != float64(12345) || third["isBillable"] != false || len(third["tagIds"].([]any)) != 2 {
		t.Errorf("unexpected third entry: %v", third)
	}
	if _, ok := third["projectId"]; ok {
		t.Errorf("expected no projectId on a task entry, got %v", third)
	}

	if len(response.Timelogs) != 2 || response.Timelogs[0].ID != 12345 || response.Timelogs[1].ID != 12346 {
		t.Errorf("unexpected timelogs: %+v", response.Timelogs)
	}
	if len(response.Failed) != 1 || response.Failed[0].Index != 1 ||
		response.Failed[0].Errors[0].Detail != "task is required" {
		t.Errorf("unexpected failures: %+v", response.Failed)
	}
}

func TestTimelogBulkCreateFailures(t *testing.T) {
	engine, _ := startBulkServer(t, http.StatusBadRequest, nil, `{"errors":[
		{"title":"invalid payload","detail":"project not found","meta":{"index":0}},
		{"title":"invalid payload","detail":"minutes must be positive","meta":{"index":0}},
		{"title":"invalid payload","detail":"task is required","meta":{"index":1}}
	]}`)

	loggedAt := time.Date(2026, 10, 20, 9, 30, 0, 0, time.UTC)
	_, err := projects.TimelogBulkCreate(t.Context(), engine, projects.NewTimelogBulkCreateRequest(
		[]projects.TimelogBulkCreateRequestTimelog{
			{ProjectID: new(int64(1)), Date: twapi.Date(loggedAt), Time: twapi.Time(loggedAt)},
			{ProjectID: new(int64(777)), Date: twapi.Date(loggedAt), Time: twapi.Time(loggedAt), Hours: 1},
		}))
	if err == nil {
		t.Fatal("expected an error")
	}

	failures := projects.TimelogBulkCreateFailures(err)
	if len(failures) != 2 || failures[0].Index != 0 || len(failures[0].Errors) != 2 ||
		failures[1].Index != 1 || failures[1].Errors[0].Detail != "task is required" {
		t.Errorf("unexpected failures: %+v", failures)
	}
	if projects.TimelogBulkCreateFailures(errors.New("not an API error")) != nil {
		t.Error("expected no failures for an error that is not an API error")
	}
}

func TestTimelogBulkCreateEmpty(t *testing.T) {
	engine := twapi.NewEngine(session.NewBearerToken("token", "https://test.teamwork.com"))
	if _, err := projects.TimelogBulkCreate(t.Context(), engine, projects.NewTimelogBulkCreateRequest(nil)); err == nil {
		t.Error("expected an error for an empty list")
	}
}
