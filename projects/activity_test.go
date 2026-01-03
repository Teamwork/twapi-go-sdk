package projects_test

import (
	"context"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestActivityList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name  string
		input projects.ActivityListRequest
	}{{
		name: "all activities",
	}, {
		name: "activities for project",
		input: projects.ActivityListRequest{
			Path: projects.ActivityListRequestPath{
				ProjectID: testResources.ProjectID,
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.ActivityList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
func TestLogItemType_UnmarshalText(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		want        projects.LogItemType
		expectError bool
	}{
		{
			name:  "valid lowercase",
			input: []byte("task"),
			want:  projects.LogItemTypeTask,
		},
		{
			name:  "valid uppercase",
			input: []byte("TASK"),
			want:  projects.LogItemTypeTask,
		},
		{
			name:  "valid mixed case with spaces",
			input: []byte("  TaskList "),
			want:  projects.LogItemTypeTasklist,
		},
		{
			name:  "underscore type",
			input: []byte("task_comment"),
			want:  projects.LogItemTypeTaskComment,
		},
		{
			name:  "camelcase type",
			input: []byte("billingInvoice"),
			want:  projects.LogItemTypeBillingInvoice,
		},
		{
			name:        "invalid type",
			input:       []byte("unknown"),
			expectError: true,
		},
		{
			name:        "empty string",
			input:       []byte(""),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var l projects.LogItemType
			err := l.UnmarshalText(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for input %q, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %q: %v", tt.input, err)
				}
				if l != tt.want {
					t.Errorf("got %q, want %q", l, tt.want)
				}
			}
		})
	}

	t.Run("nil receiver panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic on nil receiver, got none")
			}
		}()
		var l *projects.LogItemType
		_ = l.UnmarshalText([]byte("task"))
	})
}
