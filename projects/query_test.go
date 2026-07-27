package projects

// This is the only in-package test file for the projects package — every other
// test lives in projects_test. The querySet* helpers are unexported, so they can
// only be exercised from inside the package.

import (
	"net/url"
	"testing"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
)

// testEnum is a stand-in for the typed string enums the SDK uses for filter
// values (order-by keys, sideloads, status values). It exercises querySetString
// through its generic ~string contract rather than depending on any particular
// resource's enum.
type testEnum string

const testEnumValue testEnum = "createdAt"

// assertQuery checks that query holds exactly the expected key when want is
// non-empty, and is untouched when want is empty. Every querySet* helper is a
// no-op for unset values, so "wrote nothing at all" is the assertion that
// matters most.
func assertQuery(t *testing.T, query url.Values, key, want string) {
	t.Helper()

	if want == "" {
		if len(query) != 0 {
			t.Errorf("expected no query params, got %v", query)
		}
		return
	}
	if got := query.Get(key); got != want {
		t.Errorf("%s = %q, want %q", key, got, want)
	}
	if len(query) != 1 {
		t.Errorf("expected only %s to be set, got %v", key, query)
	}
}

func TestQuerySetString(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{{
		name:  "empty string writes nothing",
		value: "",
		want:  "",
	}, {
		name:  "value is set",
		value: "urgent",
		want:  "urgent",
	}, {
		name:  "whitespace is not treated as empty",
		value: " ",
		want:  " ",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := url.Values{}
			querySetString(query, "searchTerm", tt.value)
			assertQuery(t, query, "searchTerm", tt.want)
		})
	}
}

func TestQuerySetStringTypedEnum(t *testing.T) {
	query := url.Values{}
	querySetString(query, "orderBy", testEnumValue)
	assertQuery(t, query, "orderBy", "createdAt")

	empty := url.Values{}
	querySetString(empty, "orderBy", testEnum(""))
	assertQuery(t, empty, "orderBy", "")
}

func TestQuerySetInt64(t *testing.T) {
	tests := []struct {
		name  string
		value int64
		want  string
	}{{
		name:  "zero writes nothing",
		value: 0,
		want:  "",
	}, {
		// Documented behaviour: non-positive means unset, matching the `> 0`
		// guard every int64 filter in the SDK used before these helpers existed.
		name:  "negative writes nothing",
		value: -1,
		want:  "",
	}, {
		name:  "positive is set",
		value: 42,
		want:  "42",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := url.Values{}
			querySetInt64(query, "page", tt.value)
			assertQuery(t, query, "page", tt.want)
		})
	}
}

func TestQuerySetBool(t *testing.T) {
	trueValue, falseValue := true, false

	tests := []struct {
		name  string
		value *bool
		want  string
	}{{
		name:  "nil writes nothing",
		value: nil,
		want:  "",
	}, {
		name:  "true is set",
		value: &trueValue,
		want:  "true",
	}, {
		// An explicit false must reach the API so callers can override a filter
		// whose server-side default is true.
		name:  "explicit false is set",
		value: &falseValue,
		want:  "false",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := url.Values{}
			querySetBool(query, "includeCompletedTasks", tt.value)
			assertQuery(t, query, "includeCompletedTasks", tt.want)
		})
	}
}

func TestQuerySetTimestamp(t *testing.T) {
	value := time.Date(2026, time.July, 27, 14, 30, 5, 0, time.UTC)
	var zero time.Time

	tests := []struct {
		name  string
		value *time.Time
		want  string
	}{{
		name:  "nil writes nothing",
		value: nil,
		want:  "",
	}, {
		name:  "zero timestamp writes nothing",
		value: &zero,
		want:  "",
	}, {
		name:  "value is formatted as RFC3339",
		value: &value,
		want:  "2026-07-27T14:30:05Z",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := url.Values{}
			querySetTimestamp(query, "createdAfter", tt.value)
			assertQuery(t, query, "createdAfter", tt.want)
		})
	}
}

func TestQuerySetDate(t *testing.T) {
	value := twapi.Date(time.Date(2026, time.July, 27, 14, 30, 5, 0, time.UTC))
	var zero twapi.Date

	tests := []struct {
		name  string
		value *twapi.Date
		want  string
	}{{
		name:  "nil writes nothing",
		value: nil,
		want:  "",
	}, {
		name:  "zero date writes nothing",
		value: &zero,
		want:  "",
	}, {
		name:  "value is formatted as a date, dropping the time",
		value: &value,
		want:  "2026-07-27",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := url.Values{}
			querySetDate(query, "dueAfter", tt.value)
			assertQuery(t, query, "dueAfter", tt.want)
		})
	}
}

func TestQuerySetInt64s(t *testing.T) {
	tests := []struct {
		name   string
		values []int64
		want   string
	}{{
		name:   "nil writes nothing",
		values: nil,
		want:   "",
	}, {
		name:   "empty slice writes nothing",
		values: []int64{},
		want:   "",
	}, {
		name:   "single value",
		values: []int64{1},
		want:   "1",
	}, {
		name:   "multiple values are comma separated in caller order",
		values: []int64{3, 1, 2},
		want:   "3,1,2",
	}, {
		// Unlike querySetInt64, the slice helper does not treat non-positive
		// values as unset: a caller that passes them meant to send them.
		name:   "non-positive values are kept",
		values: []int64{0, -1},
		want:   "0,-1",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := url.Values{}
			querySetInt64s(query, "tagIds", tt.values)
			assertQuery(t, query, "tagIds", tt.want)
		})
	}
}

// TestQuerySetHelpersDoNotClobber verifies the helpers accumulate into a shared
// url.Values, which is how every filter's apply method uses them.
func TestQuerySetHelpersDoNotClobber(t *testing.T) {
	completed := true
	createdAfter := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)

	query := url.Values{}
	querySetString(query, "searchTerm", "urgent")
	querySetInt64(query, "pageSize", 50)
	querySetBool(query, "includeCompletedTasks", &completed)
	querySetTimestamp(query, "createdAfter", &createdAfter)
	querySetInt64s(query, "tagIds", []int64{7, 8})

	want := url.Values{
		"searchTerm":            []string{"urgent"},
		"pageSize":              []string{"50"},
		"includeCompletedTasks": []string{"true"},
		"createdAfter":          []string{"2026-01-02T03:04:05Z"},
		"tagIds":                []string{"7,8"},
	}
	if got := query.Encode(); got != want.Encode() {
		t.Errorf("query = %q, want %q", got, want.Encode())
	}
}
