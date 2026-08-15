package projects_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestFileCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	onlyRequiredRef, err := createPendingFile(t)
	if err != nil {
		t.Fatal(err)
	}
	allFieldsRef, err := createPendingFile(t)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		input projects.FileCreateRequest
	}{{
		name:  "only required fields",
		input: projects.NewFileCreateRequest(testResources.ProjectID, onlyRequiredRef),
	}, {
		name: "all fields",
		input: projects.FileCreateRequest{
			Path: projects.FileCreateRequestPath{
				ProjectID: testResources.ProjectID,
			},
			PendingFileRef:    allFieldsRef,
			Name:              new(fmt.Sprintf("test%d%d.txt", time.Now().UnixNano(), rand.Intn(100))),
			Description:       new("This is a test file"),
			Private:           new(false),
			CategoryName:      new("Test Files"),
			TagIDs:            projects.LegacyNumericList{testResources.TagID},
			AutoNewVersion:    new(false),
			NotifyCurrentUser: new(false),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			file, err := projects.FileCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx := context.Background() // t.Context is always canceled in cleanup
				if _, err := projects.FileDelete(ctx, engine,
					projects.NewFileDeleteRequest(int64(file.ID))); err != nil {
					t.Errorf("failed to delete file after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if file.ID == 0 {
				t.Error("expected a valid file ID but got 0")
			}
		})
	}
}

func TestFileDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	// createFile uploads and files the file; its cleanup is discarded because
	// deleting it is what this test does.
	fileID, _, err := createFile(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err := projects.FileDelete(ctx, engine, projects.NewFileDeleteRequest(fileID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}
