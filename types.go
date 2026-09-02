package twapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var reHexColor = regexp.MustCompile(`^#?([0-9a-f]{6})$`)

// HTTPError represents an error response from the API.
type HTTPError struct {
	StatusCode int
	Headers    http.Header
	Message    string
	Details    string
}

// NewHTTPError creates a new HTTPError from an http.Response.
func NewHTTPError(resp *http.Response, message string) *HTTPError {
	body := "no response body"
	if b, err := io.ReadAll(resp.Body); err == nil && len(b) > 0 {
		body = string(b)
	}
	return &HTTPError{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Message:    message,
		Details:    body,
	}
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("%s (%d): %s", e.Message, e.StatusCode, e.Details)
}

// Relationship describes the relation between the main entity and a sideload type.
type Relationship struct {
	ID   int64          `json:"id"`
	Type string         `json:"type"`
	Meta map[string]any `json:"meta,omitempty"`
}

// OptionalDateTime is a type alias for time.Time, used to represent date and
// time values in the API. The difference is that it will accept empty strings
// as valid values.
type OptionalDateTime time.Time

// MarshalJSON encodes the OptionalDateTime as a string in the format
// "2006-01-02T15:04:05Z07:00", or as null when unset.
//
// The zero value must round-trip back to null rather than to the year-1
// timestamp time.Time would produce. The API spells "unset" as an empty string
// on some fields, and encoding/json allocates the pointer before calling
// UnmarshalJSON, so an unset value reaches this method as a non-nil pointer to
// the zero time. Emitting the year-1 timestamp there would report a date the
// API never sent — and would break consumers that derive a JSON Schema from
// these models, since the value no longer matches the field's declared shape.
func (d OptionalDateTime) MarshalJSON() ([]byte, error) {
	if time.Time(d).IsZero() {
		return []byte("null"), nil
	}
	return time.Time(d).MarshalJSON()
}

// UnmarshalJSON decodes a JSON string into an OptionalDateTime type.
func (d *OptionalDateTime) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == `""` || string(data) == "null" {
		return nil
	}
	return (*time.Time)(d).UnmarshalJSON(data)
}

// Date is a type alias for time.Time, used to represent date values in the API.
type Date time.Time

// MarshalJSON encodes the Date as a string in the format "2006-01-02".
func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(d).Format("2006-01-02") + `"`), nil
}

// UnmarshalJSON decodes a JSON string into a Date type.
func (d *Date) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	return d.UnmarshalText([]byte(str))
}

// MarshalText encodes the Date as a string in the format "2006-01-02".
func (d Date) MarshalText() ([]byte, error) {
	quotedValue, err := d.MarshalJSON()
	if err != nil {
		return nil, err
	}
	// it is expected that encoding.TextMarshaler does not return quotes.
	value, err := strconv.Unquote(string(quotedValue))
	if err != nil {
		return nil, err
	}
	return []byte(value), nil
}

// UnmarshalText decodes a text string into a Date type. This is required when
// using Date type as a map key.
//
// The parsing lives here rather than in UnmarshalJSON because the text form is
// unquoted: feeding it to UnmarshalJSON would ask the JSON decoder to parse
// 2006-01-02 as a bare token, which fails on the first hyphen.
func (d *Date) UnmarshalText(text []byte) error {
	str := string(text)
	if strings.Contains(str, "T") {
		str, _, _ = strings.Cut(str, "T")
	}
	parsedTime, err := time.Parse("2006-01-02", str)
	if err != nil {
		return err
	}
	*d = Date(parsedTime)
	return nil
}

// IsZero reports whether the Date is zero.
func (d Date) IsZero() bool {
	return time.Time(d).IsZero()
}

// String returns the string representation of the Date in the format
// "2006-01-02".
func (d Date) String() string {
	return time.Time(d).Format("2006-01-02")
}

// Time is a type alias for time.Time, used to represent time values in the API.
type Time time.Time

// MarshalJSON encodes the Time as a string in the format "15:04:05".
func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(t).Format("15:04:05") + `"`), nil
}

// UnmarshalJSON decodes a JSON string into a Time type.
func (t *Time) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	return t.UnmarshalText([]byte(str))
}

// MarshalText encodes the Time as a string in the format "15:04:05".
func (t Time) MarshalText() ([]byte, error) {
	quotedValue, err := t.MarshalJSON()
	if err != nil {
		return nil, err
	}
	// it is expected that encoding.TextMarshaler does not return quotes.
	value, err := strconv.Unquote(string(quotedValue))
	if err != nil {
		return nil, err
	}
	return []byte(value), nil
}

// UnmarshalText decodes a text string into a Time type. This is required when
// using Time type as a map key.
//
// The parsing lives here rather than in UnmarshalJSON because the text form is
// unquoted: feeding it to UnmarshalJSON would ask the JSON decoder to parse
// 15:04:05 as a bare token, which fails on the first colon.
func (t *Time) UnmarshalText(text []byte) error {
	parsedTime, err := time.Parse("15:04:05", string(text))
	if err != nil {
		return err
	}
	*t = Time(parsedTime)
	return nil
}

// String returns the string representation of the Time in the format
// "15:04:05".
func (t Time) String() string {
	return time.Time(t).Format("15:04:05")
}

// Money represents a monetary value in the API.
type Money int64

// Set sets the value of Money from a float64.
func (m *Money) Set(value float64) {
	*m = Money(value * 100)
}

// Value returns the value of Money as a float64.
func (m Money) Value() float64 {
	return float64(m) / 100
}

// NewMoney creates a Money from a float64 value (in major units).
func NewMoney(value float64) Money {
	return Money(value * 100)
}

// OrderMode specifies the order direction of a list request. Ascending and
// descending are the only directions the API recognises, so the two constants
// below are the full set of usable values; the type exists to keep any other
// string from reaching a request by mistake.
type OrderMode string

// Supported order modes.
const (
	OrderModeAscending  OrderMode = "asc"
	OrderModeDescending OrderMode = "desc"
)

// NullableInt64 is a tri-state integer used on writes where the API
// distinguishes between an unset field, a field set to null, and a field set
// to a concrete value. Use the helper constructors NewNullableInt64 and
// NullInt64 instead of building it by hand.
type NullableInt64 struct {
	// Value is the integer value, ignored when Null is true.
	Value int64
	// Null indicates the field is explicitly set to null.
	Null bool
	// Set indicates the field is present in the payload. When false, the
	// field is omitted entirely.
	Set bool
}

// NewNullableInt64 returns a NullableInt64 set to the given value.
func NewNullableInt64(value int64) NullableInt64 {
	return NullableInt64{Value: value, Set: true}
}

// NullInt64 returns a NullableInt64 set to null.
func NullInt64() NullableInt64 {
	return NullableInt64{Null: true, Set: true}
}

// MarshalJSON implements json.Marshaler. Unset values are encoded as a JSON
// null only when the surrounding tag does not include omitempty, so callers
// embed NullableInt64 alongside the omitempty tag on the parent field for
// "omit when not set" semantics.
func (n NullableInt64) MarshalJSON() ([]byte, error) {
	if !n.Set || n.Null {
		return []byte("null"), nil
	}
	return json.Marshal(n.Value)
}

// HexColor defines a hexadecimal color.
//
// The leading "#" is optional on the way in and always present on the way out:
// endpoints are not consistent about sending it, and a colour that fails to
// decode takes the whole response down with it. The six digits are what the
// type stores, lower-cased.
type HexColor string

// MarshalJSON encodes the type to JSON format, always with the leading "#".
func (h HexColor) MarshalJSON() ([]byte, error) {
	return fmt.Appendf(nil, `"%s"`, h.String()), nil
}

// UnmarshalJSON validate and parse v into a hexadecimal type. A value with no
// leading "#" is accepted and normalised; the empty string is not — use
// OptionalHexColor for a field an endpoint can leave blank.
func (h *HexColor) UnmarshalJSON(v []byte) error {
	color, ok := parseHexColor(v)
	if !ok {
		return fmt.Errorf("invalid hexadecimal color: %s", unquote(v))
	}

	*h = HexColor(color)
	return nil
}

// String returns the hexadecimal color as a string, with its leading "#".
func (h HexColor) String() string {
	return "#" + string(h)
}

// NewHexColor creates a pointer to new hexadecimal color. The value is
// normalised, so "#8BC34A" and "8bc34a" produce the same colour. A value that
// is not a hexadecimal colour is stored as given: the signature has nowhere to
// report the error, so the API rejects it instead.
func NewHexColor(color string) *HexColor {
	if normalized, ok := parseHexColor([]byte(color)); ok {
		color = normalized
	}
	return new(HexColor(color))
}

// OptionalHexColor is HexColor for the fields an endpoint can leave blank.
// HexColor rejects the empty string, so a model whose colour is only set
// sometimes cannot use it — the project update endpoints report no colour for a
// project nobody has rated, and every such row would fail to decode.
//
// Unset round-trips back to the empty string rather than to null, which is how
// these endpoints spell it. That also keeps the declared string type accurate
// for consumers deriving a JSON Schema from these models, which is the trap
// OptionalDateTime cannot avoid, being defined over time.Time.
type OptionalHexColor string

// MarshalJSON encodes the type to JSON format, as "#rrggbb" or as the empty
// string when unset.
func (h OptionalHexColor) MarshalJSON() ([]byte, error) {
	if h == "" {
		return []byte(`""`), nil
	}
	return HexColor(h).MarshalJSON()
}

// UnmarshalJSON validates and parses v into an optional hexadecimal type,
// reading null, the empty string and a bare "#" as unset, and everything else
// as a HexColor.
//
// The bare "#" is what a HexColor with no value encodes to, on either side of
// the wire: an endpoint modelling an optional colour with its own non-optional
// type answers a record that has none with the sign alone.
func (h *OptionalHexColor) UnmarshalJSON(v []byte) error {
	if bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
		*h = ""
		return nil
	}
	if unquoted := unquote(v); len(unquoted) == 0 || bytes.Equal(unquoted, []byte("#")) {
		*h = ""
		return nil
	}

	var color HexColor
	if err := color.UnmarshalJSON(v); err != nil {
		return err
	}

	*h = OptionalHexColor(color)
	return nil
}

// String returns the hexadecimal color with its leading "#", or the empty
// string when unset.
func (h OptionalHexColor) String() string {
	if h == "" {
		return ""
	}
	return HexColor(h).String()
}

// parseHexColor extracts the six digits a HexColor stores out of a JSON string
// or a bare value, tolerating surrounding quotes, whitespace, upper case and a
// missing "#". ok is false when v is not a hexadecimal colour.
func parseHexColor(v []byte) (string, bool) {
	match := reHexColor.FindSubmatch(bytes.ToLower(unquote(v)))
	if match == nil {
		return "", false
	}
	return string(match[1]), true
}

// unquote trims the whitespace and the surrounding quotes of a JSON value.
func unquote(v []byte) []byte {
	return bytes.Trim(bytes.TrimSpace(v), `"`)
}
