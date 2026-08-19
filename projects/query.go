package projects

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
)

// Helpers for building list-request query strings. They follow the shape of
// twapi.ApplySparseFields — take the query, the parameter name and the value —
// and centralise the "skip unset values" rule that every filter's apply method
// would otherwise spell out inline. Each helper is a no-op when the value is
// unset, so the API's default behaviour is preserved.

// querySetString sets key to value, skipping the empty string. It is generic
// over ~string so the typed enums used by filters (order-by keys, sideloads,
// status values) can be passed without an explicit conversion.
func querySetString[T ~string](query url.Values, key string, value T) {
	if value == "" {
		return
	}
	query.Set(key, string(value))
}

// querySetInt64 sets key to value, skipping non-positive values. Every non-pointer
// int64 filter in the SDK identifies real entities or pagination bounds, where
// zero and negative values mean "unset" rather than a value to send. A filter
// that must be able to send zero takes a *int64 and querySetInt64Ptr instead.
func querySetInt64(query url.Values, key string, value int64) {
	if value <= 0 {
		return
	}
	query.Set(key, strconv.FormatInt(value, 10))
}

// querySetTimestamp sets key to the RFC3339 representation of value. Nil and zero
// timestamps are skipped.
func querySetTimestamp(query url.Values, key string, value *time.Time) {
	if value == nil || value.IsZero() {
		return
	}
	query.Set(key, value.Format(time.RFC3339))
}

// querySetDate sets key to the date-only representation of value. Nil and zero dates
// are skipped.
func querySetDate(query url.Values, key string, value *twapi.Date) {
	if value == nil || value.IsZero() {
		return
	}
	query.Set(key, value.String())
}

// querySetBool sets key to value. Only nil is skipped: an explicit false is sent, so
// callers can override an API default that is true.
func querySetBool(query url.Values, key string, value *bool) {
	if value == nil {
		return
	}
	query.Set(key, strconv.FormatBool(*value))
}

// querySetInt64s sets key to the comma-separated list of values. An empty list is
// skipped. It is generic over ~int64 so the typed numeric enums used by filters
// (project healths) can be passed without converting each element.
func querySetInt64s[T ~int64](query url.Values, key string, values []T) {
	if len(values) == 0 {
		return
	}
	formatted := make([]string, len(values))
	for i, value := range values {
		formatted[i] = strconv.FormatInt(int64(value), 10)
	}
	query.Set(key, strings.Join(formatted, ","))
}

// querySetStrings sets key to the comma-separated list of values. An empty list
// is skipped. It is generic over ~string so the typed enum slices used by filters
// (sideloads, status values, feature keys) can be passed without converting each
// element. Empty elements are kept: the API rejects an unknown value, which is a
// better outcome than silently narrowing a selection the caller built wrong.
func querySetStrings[T ~string](query url.Values, key string, values []T) {
	if len(values) == 0 {
		return
	}
	formatted := make([]string, len(values))
	for i, value := range values {
		formatted[i] = string(value)
	}
	query.Set(key, strings.Join(formatted, ","))
}
