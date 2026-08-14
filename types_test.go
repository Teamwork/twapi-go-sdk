package twapi_test

import (
	"encoding/json"
	"testing"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
)

// TestOptionalDateTimeRoundTrip covers the shapes the API uses to spell "unset"
// on an optional date-time. The empty string is the interesting one: encoding/json
// allocates the pointer before calling UnmarshalJSON, so the value comes back out
// through MarshalJSON as a non-nil pointer to the zero time and must still encode
// as null.
func TestOptionalDateTimeRoundTrip(t *testing.T) {
	type payload struct {
		DeletedAt *twapi.OptionalDateTime `json:"deletedDate"`
	}

	tests := []struct {
		name string
		in   string
		want string
	}{{
		name: "empty string",
		in:   `{"deletedDate":""}`,
		want: `{"deletedDate":null}`,
	}, {
		name: "null",
		in:   `{"deletedDate":null}`,
		want: `{"deletedDate":null}`,
	}, {
		name: "absent",
		in:   `{}`,
		want: `{"deletedDate":null}`,
	}, {
		name: "timestamp",
		in:   `{"deletedDate":"2026-01-02T03:04:05Z"}`,
		want: `{"deletedDate":"2026-01-02T03:04:05Z"}`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var decoded payload
			if err := json.Unmarshal([]byte(tt.in), &decoded); err != nil {
				t.Fatalf("unexpected error decoding %s: %s", tt.in, err)
			}
			encoded, err := json.Marshal(decoded)
			if err != nil {
				t.Fatalf("unexpected error encoding: %s", err)
			}
			if string(encoded) != tt.want {
				t.Errorf("expected %s but got %s", tt.want, encoded)
			}
		})
	}
}

func TestOptionalDateTimeMarshalJSON(t *testing.T) {
	tests := []struct {
		name  string
		input twapi.OptionalDateTime
		want  string
	}{{
		name:  "zero value",
		input: twapi.OptionalDateTime{},
		want:  "null",
	}, {
		name:  "value",
		input: twapi.OptionalDateTime(time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)),
		want:  `"2026-01-02T03:04:05Z"`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if string(encoded) != tt.want {
				t.Errorf("expected %s but got %s", tt.want, encoded)
			}
		})
	}
}
