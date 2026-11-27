package projects_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"testing"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
	"github.com/teamwork/twapi-go-sdk/session"
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

func TestFileCreateAutoNewVersionResolvesTheFile(t *testing.T) {
	// The create endpoint answers 8 in every case below: the file ID when it
	// made a file, the version ID when it added a version.
	const projectID, returnedID = 42, 8
	const createPath, versionPath, filePath = "/projects/42/files.json",
		"/projects/api/v3/fileversions/8.json", "/projects/api/v3/files/8.json"

	version := func(fileID, versionNo, projectID int64) string {
		return fmt.Sprintf(`{"fileversion":{"fileVersionId":8,"file":{"id":%d,"type":"files"},`+
			`"versionNo":%d,"project":{"id":%d,"type":"projects"},"uploadedBy":{"id":1,"type":"users"}}}`,
			fileID, versionNo, projectID)
	}
	file := func(latestVersionID int64) string {
		return fmt.Sprintf(`{"file":{"id":8,"version":{"id":%d,"type":"fileVersions"}}}`, latestVersionID)
	}

	tests := []struct {
		name           string
		autoNewVersion *bool
		versionStatus  int
		versionBody    string
		fileStatus     int
		fileBody       string
		wantID         int64
		wantVersionID  int64
		wantPaths      []string
		wantErr        bool
	}{{
		name:           "new version of another file",
		autoNewVersion: new(true),
		versionStatus:  http.StatusOK,
		versionBody:    version(7, 2, projectID),
		fileStatus:     http.StatusNotFound,
		wantID:         7,
		wantVersionID:  8,
		wantPaths:      []string{createPath, versionPath, filePath},
	}, {
		name:           "new version, and a file shares its number with an older version",
		autoNewVersion: new(true),
		versionStatus:  http.StatusOK,
		versionBody:    version(7, 2, projectID),
		fileStatus:     http.StatusOK,
		fileBody:       file(5),
		wantID:         7,
		wantVersionID:  8,
		wantPaths:      []string{createPath, versionPath, filePath},
	}, {
		name:           "new file, no version has its number",
		autoNewVersion: new(true),
		versionStatus:  http.StatusNotFound,
		wantID:         8,
		wantPaths:      []string{createPath, versionPath},
	}, {
		name:           "new file, the version with its number is a first version",
		autoNewVersion: new(true),
		versionStatus:  http.StatusOK,
		versionBody:    version(3, 1, projectID),
		wantID:         8,
		wantPaths:      []string{createPath, versionPath},
	}, {
		name:           "new file, the version with its number is in another project",
		autoNewVersion: new(true),
		versionStatus:  http.StatusOK,
		versionBody:    version(3, 4, 99),
		wantID:         8,
		wantPaths:      []string{createPath, versionPath},
	}, {
		name:           "new file, its own version is newer than the one with its number",
		autoNewVersion: new(true),
		versionStatus:  http.StatusOK,
		versionBody:    version(3, 4, projectID),
		fileStatus:     http.StatusOK,
		fileBody:       file(12),
		wantID:         8,
		wantPaths:      []string{createPath, versionPath, filePath},
	}, {
		name:           "without auto new version",
		autoNewVersion: new(false),
		wantID:         8,
		wantPaths:      []string{createPath},
	}, {
		name:      "auto new version omitted",
		wantID:    8,
		wantPaths: []string{createPath},
	}, {
		name:           "lookup fails after the upload is stored",
		autoNewVersion: new(true),
		versionStatus:  http.StatusInternalServerError,
		wantID:         8,
		wantPaths:      []string{createPath, versionPath},
		wantErr:        true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var paths []string
			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				paths = append(paths, r.URL.Path)
				switch r.URL.Path {
				case createPath:
					w.WriteHeader(http.StatusCreated)
					_, _ = fmt.Fprintf(w, `{"STATUS":"OK","id":"%d","fileId":"%d"}`, returnedID, returnedID)
				case versionPath:
					w.WriteHeader(tt.versionStatus)
					_, _ = io.WriteString(w, tt.versionBody)
				case filePath:
					w.WriteHeader(tt.fileStatus)
					_, _ = io.WriteString(w, tt.fileBody)
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			})
			server := httptest.NewServer(mux)
			t.Cleanup(server.Close)
			testEngine := twapi.NewEngine(session.NewBearerToken("your_token", server.URL))

			req := projects.NewFileCreateRequest(projectID, "tf_12345.md")
			req.AutoNewVersion = tt.autoNewVersion

			resp, err := projects.FileCreate(t.Context(), testEngine, req)
			if tt.wantErr != (err != nil) {
				t.Fatalf("expected error %t, got %v", tt.wantErr, err)
			}
			if resp == nil {
				t.Fatal("expected a response")
			}
			if int64(resp.ID) != tt.wantID || resp.VersionID != tt.wantVersionID {
				t.Errorf("expected file %d version %d, got file %d version %d",
					tt.wantID, tt.wantVersionID, resp.ID, resp.VersionID)
			}
			if !slices.Equal(paths, tt.wantPaths) {
				t.Errorf("expected requests %v, got %v", tt.wantPaths, paths)
			}
		})
	}
}

func TestFileVersionUploadedByAcceptsBothShapes(t *testing.T) {
	for _, body := range []string{`{"uploadedBy":1}`, `{"uploadedBy":{"id":1,"type":"users"}}`} {
		var version projects.FileVersion
		if err := json.Unmarshal([]byte(body), &version); err != nil {
			t.Fatalf("%s: unexpected error: %s", body, err)
		}
		if version.UploadedBy != 1 {
			t.Errorf("%s: expected uploader 1, got %d", body, version.UploadedBy)
		}
	}
}

func TestFileVersionCreate(t *testing.T) {
	var method, path string
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode the request body: %s", err)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"fileversion":{"fileVersionId":15,"fileId":10,"file":{"id":10,"type":"files"},`+
			`"versionNo":3,"uploadedBy":{"id":1,"type":"users"}},"included":{}}`)
	}))
	t.Cleanup(server.Close)
	testEngine := twapi.NewEngine(session.NewBearerToken("your_token", server.URL))

	req := projects.NewFileVersionCreateRequest(10, "tf_12345.md")
	req.Description = new("second draft")
	resp, err := projects.FileVersionCreate(t.Context(), testEngine, req)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if method != http.MethodPost || path != "/projects/api/v3/files/10.json" {
		t.Errorf("expected POST /projects/api/v3/files/10.json, got %s %s", method, path)
	}
	want := map[string]any{"fileversion": map[string]any{
		"pendingFileRef": "tf_12345.md",
		"description":    "second draft",
	}}
	if fmt.Sprint(body) != fmt.Sprint(want) {
		t.Errorf("expected body %v, got %v", want, body)
	}
	if v := resp.FileVersion; v.ID != 15 || v.File.ID != 10 || v.VersionNo != 3 || v.UploadedBy != 1 {
		t.Errorf("unexpected version %+v", v)
	}
}

func TestFileVersionCreateRejectsMissingFields(t *testing.T) {
	testEngine := twapi.NewEngine(session.NewBearerToken("your_token", "https://test.teamwork.com"))
	for name, req := range map[string]projects.FileVersionCreateRequest{
		"no file":      projects.NewFileVersionCreateRequest(0, "tf_12345.md"),
		"no reference": projects.NewFileVersionCreateRequest(10, ""),
	} {
		if _, err := projects.FileVersionCreate(t.Context(), testEngine, req); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestFileUpdate(t *testing.T) {
	tests := []struct {
		name     string
		request  func(projects.FileUpdateRequest) projects.FileUpdateRequest
		wantBody string
	}{{
		name: "rename and describe a version",
		request: func(r projects.FileUpdateRequest) projects.FileUpdateRequest {
			r.VersionID = new(int64(17))
			r.Name = new("spec.md")
			r.Description = new("second draft")
			return r
		},
		wantBody: `{"file":{"versionId":17,"name":"spec.md","description":"second draft"}}`,
	}, {
		name: "replace tags, category and privacy",
		request: func(r projects.FileUpdateRequest) projects.FileUpdateRequest {
			r.TagIDs = &[]int64{1, 2}
			r.CategoryID = new(int64(3))
			r.Private = new(false)
			return r
		},
		wantBody: `{"file":{"categoryId":3,"tagIds":[1,2],"private":false}}`,
	}, {
		name: "remove every tag",
		request: func(r projects.FileUpdateRequest) projects.FileUpdateRequest {
			r.TagIDs = &[]int64{}
			return r
		},
		wantBody: `{"file":{"tagIds":[]}}`,
	}, {
		name:     "nothing set sends no field",
		request:  func(r projects.FileUpdateRequest) projects.FileUpdateRequest { return r },
		wantBody: `{"file":{}}`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var method, path, body string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				raw, _ := io.ReadAll(r.Body)
				method, path, body = r.Method, r.URL.Path, string(raw)
				_, _ = io.WriteString(w, `{"file":{"id":10,"displayName":"spec.md","uploadedBy":1,`+
					`"version":{"id":18,"type":"fileVersions"}},"included":{}}`)
			}))
			t.Cleanup(server.Close)
			testEngine := twapi.NewEngine(session.NewBearerToken("your_token", server.URL))

			resp, err := projects.FileUpdate(t.Context(), testEngine, tt.request(projects.NewFileUpdateRequest(10)))
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if method != http.MethodPatch || path != "/projects/api/v3/files/10.json" {
				t.Errorf("expected PATCH /projects/api/v3/files/10.json, got %s %s", method, path)
			}
			if body != tt.wantBody+"\n" {
				t.Errorf("expected body %s, got %s", tt.wantBody, body)
			}
			if resp.File.ID != 10 || resp.File.Version.ID != 18 {
				t.Errorf("unexpected file %+v", resp.File)
			}
		})
	}
}

func TestFileUpdateRejectsMissingFile(t *testing.T) {
	testEngine := twapi.NewEngine(session.NewBearerToken("your_token", "https://test.teamwork.com"))
	if _, err := projects.FileUpdate(t.Context(), testEngine, projects.NewFileUpdateRequest(0)); err == nil {
		t.Error("expected an error")
	}
}
