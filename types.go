package twapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

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
// "2006-01-02T15:04:05Z07:00".
func (d OptionalDateTime) MarshalJSON() ([]byte, error) {
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
	parsedTime, err := time.Parse("2006-01-02", str)
	if err != nil {
		return err
	}
	*d = Date(parsedTime)
	return nil
}

// MarshalText encodes the Date as a string in the format "2006-01-02".
func (d Date) MarshalText() ([]byte, error) {
	return d.MarshalJSON()
}

// UnmarshalText decodes a text string into a Date type. This is required when
// using Date type as a map key.
func (d *Date) UnmarshalText(text []byte) error {
	return d.UnmarshalJSON(text)
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

// UnmarshalJSON decodes a JSON string into a Date type.
func (t *Time) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	parsedTime, err := time.Parse("15:04:05", str)
	if err != nil {
		return err
	}
	*t = Time(parsedTime)
	return nil
}

// MarshalText encodes the Time as a string in the format "15:04:05".
func (t Time) MarshalText() ([]byte, error) {
	return t.MarshalJSON()
}

// UnmarshalText decodes a text string into a Time type. This is required when
// using Time type as a map key.
func (t *Time) UnmarshalText(text []byte) error {
	return t.UnmarshalJSON(text)
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
