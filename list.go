package twapi

import "net/url"

// ListCountMode selects whether a v3 list endpoint computes the exact number of
// items matching the request filters, exposed as ListMeta.Page.Count in the
// response.
//
// The total is derived from a dedicated count query. The API caches it per
// filter combination, independently of the requested page, so paginating over a
// result set only pays for it once. The cache is invalidated by any change to
// the counted entities though, and the narrower the filters the longer it
// survives: a count scoped to one project outlives one scoped to the whole
// installation, which any change to any of those entities invalidates. An
// unscoped exact count is therefore the expensive one to ask for.
//
// https://apidocs.teamwork.com/guides/teamwork/pagination
type ListCountMode string

// List of possible count modes for ListCountMode.
const (
	// ListCountModeDefault leaves the decision to the API, which currently
	// computes the count. ListMeta.Page.Count is populated, and the API is free
	// to use the total to optimise how it loads large result sets.
	ListCountModeDefault ListCountMode = ""

	// ListCountModeExact asks the API for the exact total, populating
	// ListMeta.Page.Count.
	ListCountModeExact ListCountMode = "exact"

	// ListCountModeSkip asks the API to skip the count query.
	// ListMeta.Page.Count is nil and pagination relies on ListMeta.Page.HasMore
	// alone.
	ListCountModeSkip ListCountMode = "skip"
)

// Apply adds the `skipCounts` parameter to query, unless the mode leaves the
// decision to the API.
func (l ListCountMode) Apply(query url.Values) {
	switch l {
	case ListCountModeExact:
		query.Set("skipCounts", "false")
	case ListCountModeSkip:
		query.Set("skipCounts", "true")
	case ListCountModeDefault:
		// leave skipCounts unset so the API applies its own default
	}
}

// ListMeta contains the metadata a v3 list endpoint reports alongside the items
// matching the request filters. List responses embed it as their Meta field.
type ListMeta struct {
	// Page contains the pagination details of the response.
	Page ListMetaPage `json:"page"`
}

// ResolveCount reconciles the count reported by the API with what the request
// asked for, clearing Page.Count when the API answered with a lower bound
// instead of a total. List responses call it from SetRequest, which is the
// first point at which the request that produced the response is known.
func (l *ListMeta) ResolveCount(mode ListCountMode) {
	if mode == ListCountModeSkip {
		l.Page.Count = nil
	}
}

// ListMetaPage contains the pagination details a v3 list endpoint reports.
type ListMetaPage struct {
	// Count is the exact number of items matching the request filters, across
	// every page. It is nil when the request asked the API to skip the count
	// query with ListCountModeSkip, because the API answers those requests with
	// a lower bound derived from the page rather than a total. Use HasMore to
	// paginate in that case.
	Count *int64 `json:"count"`

	// HasMore indicates whether more items match the request filters beyond the
	// current page.
	HasMore bool `json:"hasMore"`
}
