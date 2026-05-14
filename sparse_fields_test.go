package twapi_test

import (
	"net/url"
	"testing"

	twapi "github.com/teamwork/twapi-go-sdk"
)

// testField is a stand-in for any product package's generated `<Entity>Field`
// typed-string. Using a local type keeps this test independent of any product
// package and exercises ApplySparseFields purely through its generic contract.
type testField string

const (
	testFieldID   testField = "id"
	testFieldName testField = "name"
	testFieldKind testField = "kind"
)

func TestApplySparseFields(t *testing.T) {
	tests := []struct {
		name      string
		entityKey string
		fields    []testField
		want      string
	}{
		{
			name:      "empty slice writes nothing",
			entityKey: "things",
			fields:    nil,
			want:      "",
		},
		{
			name:      "single field",
			entityKey: "things",
			fields:    []testField{testFieldID},
			want:      "id",
		},
		{
			name:      "multiple fields preserve caller order",
			entityKey: "things",
			fields:    []testField{testFieldID, testFieldName, testFieldKind},
			want:      "id,name,kind",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			query := url.Values{}
			twapi.ApplySparseFields(query, tc.entityKey, tc.fields)

			got := query.Get("fields[" + tc.entityKey + "]")
			if got != tc.want {
				t.Errorf("fields[%s] = %q, want %q", tc.entityKey, got, tc.want)
			}
			if tc.want == "" && len(query) != 0 {
				t.Errorf("expected no query params, got %v", query)
			}
		})
	}
}
