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
