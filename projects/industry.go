package projects

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var (
	_ twapi.HTTPRequester = (*IndustryListRequest)(nil)
	_ twapi.HTTPResponser = (*IndustryListResponse)(nil)
)

// Industry refers to the business sector or market category that a company
// belongs to, such as technology, healthcare, finance, or education. It helps
// provide context about the nature of a company’s work and can be used to
// better organize and filter data across the platform. By associating companies
// and projects with specific industries, Teamwork.com allows teams to gain
// clearer insights, tailor communication, and segment information in ways that
// make it easier to manage relationships and understand the broader business
// landscape in which their clients and partners operate.
type Industry struct {
	// ID is the unique identifier of the industry.
	ID LegacyNumber `json:"id"`

	// Name is the name of the industry.
	Name string `json:"name"`
}

// IndustryListRequest represents the request body for loading multiple industries.
//
// Not documented.
type IndustryListRequest struct{}

// NewIndustryListRequest creates a new IndustryListRequest with default values.
func NewIndustryListRequest() IndustryListRequest {
	return IndustryListRequest{}
}

// HTTPRequest creates an HTTP request for the IndustryListRequest.
func (p IndustryListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/industries.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// IndustryListResponse contains information by multiple industries matching the
// request filters.
//
// Not documented.
type IndustryListResponse struct {
	Industries []Industry `json:"industries"`
}

// HandleHTTPResponse handles the HTTP response for the IndustryListResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (p *IndustryListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list industries")
	}

	if err := json.NewDecoder(resp.Body).Decode(p); err != nil {
		return fmt.Errorf("failed to decode list industries response: %w", err)
	}
	return nil
}

// IndustryList retrieves multiple industries using the provided request
// and returns the response.
func IndustryList(
	ctx context.Context,
	engine *twapi.Engine,
	req IndustryListRequest,
) (*IndustryListResponse, error) {
	return twapi.Execute[IndustryListRequest, *IndustryListResponse](ctx, engine, req)
}
