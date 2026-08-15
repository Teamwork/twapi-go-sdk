package projects_test

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestPendingFileCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name  string
		input projects.PendingFileCreateRequest
	}{{
		name: "from bytes",
		input: projects.NewPendingFileCreateRequestFromBytes(
			fmt.Sprintf("test%d%d.txt", time.Now().UnixNano(), rand.Intn(100)),
			[]byte("This is a test file"),
		),
	}, {
		name: "from reader",
		input: projects.NewPendingFileCreateRequest(
			fmt.Sprintf("test%d%d.txt", time.Now().UnixNano(), rand.Intn(100)),
			strings.NewReader("This is a test file"),
		),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			// There is nothing to clean up: a pending file that is never attached
			// is not visible anywhere, and the API offers no way to discard one.
			pendingFile, err := projects.PendingFileCreate(ctx, engine, tt.input)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if pendingFile.PendingFile.Ref == "" {
				t.Error("expected a valid pending file reference")
			}
		})
	}
}

func TestPendingFileCreateMissingFields(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name  string
		input projects.PendingFileCreateRequest
	}{{
		name:  "no file name",
		input: projects.NewPendingFileCreateRequestFromBytes("", []byte("This is a test file")),
	}, {
		name:  "no contents",
		input: projects.NewPendingFileCreateRequest("test.txt", nil),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.PendingFileCreate(ctx, engine, tt.input); err == nil {
				t.Error("expected an error, got none")
			}
		})
	}
}
