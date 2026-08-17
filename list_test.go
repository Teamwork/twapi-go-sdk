package twapi_test

import (
	"encoding/json"
	"net/url"
	"testing"

	twapi "github.com/teamwork/twapi-go-sdk"
)

func TestListCountModeApply(t *testing.T) {
	tests := []struct {
		name string
		mode twapi.ListCountMode
		want string
	}{{
		name: "default leaves the decision to the API",
		mode: twapi.ListCountModeDefault,
		want: "",
	}, {
		name: "exact asks for the count query",
		mode: twapi.ListCountModeExact,
		want: "false",
	}, {
		name: "skip opts out of the count query",
		mode: twapi.ListCountModeSkip,
		want: "true",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := make(url.Values)
			tt.mode.Apply(query)

			if got := query.Get("skipCounts"); got != tt.want {
				t.Errorf("expected skipCounts=%q but got %q", tt.want, got)
			}
			if tt.want == "" && len(query) != 0 {
				t.Errorf("expected an untouched query but got %q", query.Encode())
			}
		})
	}
}

func TestListMetaResolveCount(t *testing.T) {
	// the API reports a count even when it skips the count query, but that value
	// is a lower bound derived from the page rather than a total.
	const payload = `{"page":{"count":51,"hasMore":true}}`

	count := int64(51)

	tests := []struct {
		name string
		mode twapi.ListCountMode
		want *int64
	}{{
		name: "default keeps the count",
		mode: twapi.ListCountModeDefault,
		want: &count,
	}, {
		name: "exact keeps the count",
		mode: twapi.ListCountModeExact,
		want: &count,
	}, {
		name: "skip drops the lower bound",
		mode: twapi.ListCountModeSkip,
		want: nil,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var meta twapi.ListMeta
			if err := json.Unmarshal([]byte(payload), &meta); err != nil {
				t.Fatalf("failed to decode meta: %s", err)
			}
			meta.ResolveCount(tt.mode)

			switch got := meta.Page.Count; {
			case got == nil && tt.want != nil:
				t.Errorf("expected count %d but got nil", *tt.want)
			case got != nil && tt.want == nil:
				t.Errorf("expected no count but got %d", *got)
			case got != nil && *got != *tt.want:
				t.Errorf("expected count %d but got %d", *tt.want, *got)
			}

			if !meta.Page.HasMore {
				t.Error("expected HasMore to survive resolving the count")
			}
		})
	}
}
