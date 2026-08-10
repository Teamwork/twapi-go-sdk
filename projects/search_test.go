package projects_test

import (
	"context"
	"encoding/json"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestSearch(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name          string
		input         projects.SearchRequest
		expectedError bool
	}{{
		name:  "all searches",
		input: projects.NewSearchRequest("example"),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.Search(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestSearchRequestGeneration(t *testing.T) {
	includeHighlights := true

	req := projects.NewSearchRequest("example")
	req.Filters.IncludeHighlights = &includeHighlights

	httpReq, err := req.HTTPRequest(context.Background(), "https://test.com")
	if err != nil {
		t.Fatalf("unexpected error creating HTTP request: %s", err)
	}

	query, err := url.ParseQuery(httpReq.URL.RawQuery)
	if err != nil {
		t.Fatalf("failed to parse query string: %s", err)
	}

	if query.Get("searchTerm") != "example" {
		t.Errorf("expected searchTerm=example but got %q", query.Get("searchTerm"))
	}
	if query.Get("includeHighlights") != "true" {
		t.Errorf("expected includeHighlights=true but got %q", query.Get("includeHighlights"))
	}
}

func TestSearchRequestGeneration_noIncludeHighlights(t *testing.T) {
	req := projects.NewSearchRequest("example")

	httpReq, err := req.HTTPRequest(context.Background(), "https://test.com")
	if err != nil {
		t.Fatalf("unexpected error creating HTTP request: %s", err)
	}

	query, err := url.ParseQuery(httpReq.URL.RawQuery)
	if err != nil {
		t.Fatalf("failed to parse query string: %s", err)
	}

	if query.Has("includeHighlights") {
		t.Errorf("expected includeHighlights to be unset but got %q", query.Get("includeHighlights"))
	}
}

func TestSearchItemHighlights(t *testing.T) {
	tests := []struct {
		name string
		body string
		want map[string][]string
	}{{
		name: "valid highlights",
		body: `{"id":15,"type":"tasks","meta":{"highlights":{` +
			`"taskName":["<em>Task</em> 1"],` +
			`"description":["something about the <em>task</em>","second <em>task</em> fragment"]}}}`,
		want: map[string][]string{
			"taskName":    {"<em>Task</em> 1"},
			"description": {"something about the <em>task</em>", "second <em>task</em> fragment"},
		},
	}, {
		name: "no meta",
		body: `{"id":15,"type":"tasks"}`,
	}, {
		name: "meta without highlights",
		body: `{"id":15,"type":"tasks","meta":{"other":true}}`,
	}, {
		name: "highlights is not an object",
		body: `{"id":15,"type":"tasks","meta":{"highlights":"broken"}}`,
	}, {
		name: "field is not an array",
		body: `{"id":15,"type":"tasks","meta":{"highlights":{"taskName":"broken"}}}`,
	}, {
		name: "malformed fields are skipped, valid fields kept",
		body: `{"id":15,"type":"tasks","meta":{"highlights":{` +
			`"taskName":["<em>Task</em> 1"],` +
			`"description":"broken",` +
			`"body":[42]}}}`,
		want: map[string][]string{
			"taskName": {"<em>Task</em> 1"},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var item projects.SearchItem
			if err := json.Unmarshal([]byte(tt.body), &item); err != nil {
				t.Fatalf("failed to decode search item: %s", err)
			}

			if got := item.Highlights(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("highlights = %v, want %v", got, tt.want)
			}
		})
	}
}
