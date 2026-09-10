package projects_test

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"testing"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
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

func TestFileGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	fileID, fileCleanup, err := createFile(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(fileCleanup)

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	t.Cleanup(cancel)

	req := projects.NewFileGetRequest(fileID)
	req.IncludeVersions = true
	req.Include = []projects.FileRequestSideload{projects.FileRequestSideloadUsers}

	fileResponse, err := projects.FileGet(ctx, engine, req)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if fileResponse.File.ID != fileID {
		t.Errorf("expected file ID %d but got %d", fileID, fileResponse.File.ID)
	}
	if fileResponse.File.DownloadURL == nil || *fileResponse.File.DownloadURL == "" {
		t.Error("expected an active file to carry a download URL")
	}
	if len(fileResponse.File.Versions) != 1 {
		t.Errorf("expected one version but got %d", len(fileResponse.File.Versions))
	}
	if _, ok := fileResponse.Included.Users[strconv.FormatInt(fileResponse.File.UploadedBy, 10)]; !ok {
		t.Errorf("expected the uploader %d to be sideloaded", fileResponse.File.UploadedBy)
	}
}

func TestFileList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	fileID, fileCleanup, err := createFile(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(fileCleanup)

	tests := []struct {
		name  string
		input projects.FileListRequest
	}{{
		// Unscoped, so every file on the account is a candidate for the page;
		// newest first keeps the fixture on it whatever the others are named.
		name: "all files",
		input: func() projects.FileListRequest {
			req := projects.NewFileListRequest()
			req.Filters.OrderBy = projects.FileOrderByDateUploaded
			req.Filters.OrderMode = twapi.OrderModeDescending
			return req
		}(),
	}, {
		name: "project files",
		input: projects.FileListRequest{
			Path: projects.FileListRequestPath{
				ProjectID: testResources.ProjectID,
			},
		},
	}, {
		name: "by identifier",
		input: projects.FileListRequest{
			Filters: projects.FileListRequestFilters{
				IDs:     []int64{fileID},
				Include: []projects.FileRequestSideload{projects.FileRequestSideloadProjects},
				Fields: projects.FileListFields{
					Files: []projects.FileField{projects.FileFieldID, projects.FileFieldDisplayName},
				},
			},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			t.Cleanup(cancel)

			filesResponse, err := projects.FileList(ctx, engine, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			var found bool
			for _, file := range filesResponse.Files {
				if file.ID == fileID {
					found = true
				}
			}
			if !found {
				t.Errorf("expected file %d in the list", fileID)
			}
		})
	}
}

func TestFileDownload(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	fileID, fileCleanup, err := createFile(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(fileCleanup)

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	t.Cleanup(cancel)

	download, err := projects.FileDownload(ctx, engine, projects.NewFileDownloadRequest(fileID))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	t.Cleanup(func() { _ = download.Body.Close() })

	content, err := io.ReadAll(download.Body)
	if err != nil {
		t.Fatalf("failed to read the download: %s", err)
	}
	// createPendingFile uploads this exact content.
	if want := "This is a test file"; string(content) != want {
		t.Errorf("expected content %q but got %q", want, content)
	}
	if download.Name == "" {
		t.Error("expected the download to carry a file name")
	}
}
