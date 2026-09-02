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

// TestDateTimeAsMapKey pins the text round trip for the types used as JSON map
// keys. encoding/json routes a map key through encoding.TextUnmarshaler with
// the quotes already stripped, so a Date or Time reaching UnmarshalText as a
// bare 2006-01-02 or 15:04:05 must be parsed directly — handing it back to the
// JSON decoder reads the leading digits as a number and then fails on the
// first hyphen or colon. The workload endpoint keys its per-user dates this
// way, so the whole response fails to decode when this regresses.
func TestDateTimeAsMapKey(t *testing.T) {
	t.Run("date", func(t *testing.T) {
		var decoded map[twapi.Date]int64
		if err := json.Unmarshal([]byte(`{"2026-01-02":5}`), &decoded); err != nil {
			t.Fatalf("unexpected error decoding date map key: %s", err)
		}
		want := twapi.Date(time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC))
		if got, ok := decoded[want]; !ok || got != 5 {
			t.Errorf("expected key 2026-01-02 to hold 5 but got %v", decoded)
		}

		encoded, err := json.Marshal(decoded)
		if err != nil {
			t.Fatalf("unexpected error encoding date map key: %s", err)
		}
		if string(encoded) != `{"2026-01-02":5}` {
			t.Errorf(`expected {"2026-01-02":5} but got %s`, encoded)
		}
	})

	t.Run("time", func(t *testing.T) {
		var decoded map[twapi.Time]int64
		if err := json.Unmarshal([]byte(`{"15:04:05":5}`), &decoded); err != nil {
			t.Fatalf("unexpected error decoding time map key: %s", err)
		}
		want := twapi.Time(time.Date(0, time.January, 1, 15, 4, 5, 0, time.UTC))
		if got, ok := decoded[want]; !ok || got != 5 {
			t.Errorf("expected key 15:04:05 to hold 5 but got %v", decoded)
		}

		encoded, err := json.Marshal(decoded)
		if err != nil {
			t.Fatalf("unexpected error encoding time map key: %s", err)
		}
		if string(encoded) != `{"15:04:05":5}` {
			t.Errorf(`expected {"15:04:05":5} but got %s`, encoded)
		}
	})

	// A date-time value still decodes through UnmarshalJSON, including the
	// timestamp form the API returns for some date-only fields.
	t.Run("date value keeps accepting a timestamp", func(t *testing.T) {
		var decoded struct {
			At twapi.Date `json:"at"`
		}
		if err := json.Unmarshal([]byte(`{"at":"2026-01-02T03:04:05Z"}`), &decoded); err != nil {
			t.Fatalf("unexpected error decoding date value: %s", err)
		}
		if got := decoded.At.String(); got != "2026-01-02" {
			t.Errorf("expected 2026-01-02 but got %s", got)
		}
	})
}

// TestHexColorRoundTrip covers the spellings an endpoint uses for a colour that
// is always set. The leading "#" is optional on the way in — the allocation
// endpoints send six bare digits, which this type used to reject outright — and
// always present on the way out, whichever form arrived.
func TestHexColorRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "with the leading sign", in: `"#8bc34a"`, want: `"#8bc34a"`},
		{name: "without the leading sign", in: `"8bc34a"`, want: `"#8bc34a"`},
		{name: "upper case", in: `"#8BC34A"`, want: `"#8bc34a"`},
		{name: "empty", in: `""`, wantErr: true},
		{name: "null", in: `null`, wantErr: true},
		{name: "not a colour", in: `"green"`, wantErr: true},
		{name: "too many digits", in: `"#8bc34ab"`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var color twapi.HexColor
			err := json.Unmarshal([]byte(tt.in), &color)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error for %s but got %q", tt.in, color)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %s: %s", tt.in, err)
			}

			encoded, err := json.Marshal(color)
			if err != nil {
				t.Fatalf("unexpected error encoding %s: %s", tt.in, err)
			}
			if string(encoded) != tt.want {
				t.Errorf("expected %s to encode as %s but got %s", tt.in, tt.want, encoded)
			}
		})
	}
}

// TestNewHexColorNormalises pins the constructor against the "##rrggbb" a
// caller passing the form it reads off a design tool would otherwise produce:
// the type stores the digits alone and String adds the sign back.
func TestNewHexColorNormalises(t *testing.T) {
	for _, in := range []string{"#8BC34A", "8bc34a"} {
		if got := twapi.NewHexColor(in).String(); got != "#8bc34a" {
			t.Errorf("expected NewHexColor(%q) to be #8bc34a but got %s", in, got)
		}
	}
}

// TestOptionalHexColorRoundTrip covers the shapes an endpoint uses for a colour
// that may be unset. HexColor takes every one of them but the blank, which is
// the whole difference between the two types.
//
// Unset must come back out as the empty string rather than as null or as the
// bare "#" HexColor's zero value produces: the type is declared as a string, and
// consumers deriving a JSON Schema from these models validate against that.
func TestOptionalHexColorRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "with the leading sign", in: `"#8BC34A"`, want: `"#8bc34a"`},
		{name: "without the leading sign", in: `"8bc34a"`, want: `"#8bc34a"`},
		{name: "empty", in: `""`, want: `""`},
		{name: "null", in: `null`, want: `""`},
		{name: "the bare sign", in: `"#"`, want: `""`},
		{name: "not a colour", in: `"green"`, wantErr: true},
		{name: "too short", in: `"#8bc34"`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var color twapi.OptionalHexColor
			err := json.Unmarshal([]byte(tt.in), &color)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error for %s but got %q", tt.in, color)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %s: %s", tt.in, err)
			}

			encoded, err := json.Marshal(color)
			if err != nil {
				t.Fatalf("unexpected error encoding %s: %s", tt.in, err)
			}
			if string(encoded) != tt.want {
				t.Errorf("expected %s to encode as %s but got %s", tt.in, tt.want, encoded)
			}
		})
	}
}
