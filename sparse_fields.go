package twapi

import (
	"net/url"
	"strings"
)

// ApplySparseFields adds a `fields[<entityKey>]=...` parameter to query when
// fields is non-empty. entityKey is the JSON collection name used by the v3
// response (e.g. "tasks" for the main task list, "customfields" for a sideload
// of custom fields). F is constrained to typed string aliases so callers can
// only pass values that correspond to real JSON attributes on the entity —
// concrete types are generated alongside each list response.
//
// https://apidocs.teamwork.com/guides/teamwork/sparse-fieldsets
func ApplySparseFields[F ~string](query url.Values, entityKey string, fields []F) {
	if len(fields) == 0 {
		return
	}
	parts := make([]string, len(fields))
	for i, f := range fields {
		parts[i] = string(f)
	}
	query.Set("fields["+entityKey+"]", strings.Join(parts, ","))
}
