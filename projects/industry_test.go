package projects_test

import (
	"context"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestIndustryList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name  string
		input projects.IndustryListRequest
	}{{
		name: "all industries",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.IndustryList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
