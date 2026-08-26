package projects_test

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
	"github.com/teamwork/twapi-go-sdk/session"
)

func TestPendingFileCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	name := fmt.Sprintf("test%d%d.txt", time.Now().UnixNano(), rand.Intn(100))

	// There is nothing to clean up: a pending file that is never attached is not
	// visible anywhere, and the API offers no way to discard one.
	pendingFile, err := projects.PendingFileCreate(ctx, engine,
		projects.NewPendingFileCreateRequest(name, []byte("This is a test file")))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if pendingFile.Ref == "" {
		t.Error("expected a valid pending file reference")
	}
}

// TestPendingFileCreateUpload pins how the contents reach the storage service.
// The signature covers the headers it lists, and only the storage service would
// notice one sent that was not signed, or a signed one left out.
func TestPendingFileCreateUpload(t *testing.T) {
	const contents = "# Plan\n"

	tests := []struct {
		name          string
		signedHeaders string
		contentType   string
		wantACL       string
	}{{
		// Signed into the URL unless the installation's bucket sets its own ACL.
		name:          "canned ACL signed",
		signedHeaders: "content-length;host;x-amz-acl",
		wantACL:       "public-read",
	}, {
		name:          "canned ACL not signed",
		signedHeaders: "content-length;host",
		wantACL:       "",
	}, {
		name:          "explicit content type",
		signedHeaders: "host",
		contentType:   "text/markdown",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var presignQuery url.Values
			var upload *http.Request
			var uploaded []byte

			mux := http.NewServeMux()
			mux.HandleFunc("GET /projects/api/v1/pendingfiles/presignedurl.json",
				func(w http.ResponseWriter, r *http.Request) {
					if r.Header.Get("Authorization") != "Bearer your_token" {
						http.Error(w, "Unauthorized", http.StatusUnauthorized)
						return
					}
					presignQuery = r.URL.Query()

					uploadURL := url.URL{
						Scheme: "http",
						Host:   r.Host,
						Path:   "/storage/tf_12345.md",
						RawQuery: url.Values{
							"X-Amz-Algorithm":     []string{"AWS4-HMAC-SHA256"},
							"X-Amz-SignedHeaders": []string{tt.signedHeaders},
							"X-Amz-Signature":     []string{"c0ffee"},
						}.Encode(),
					}
					w.Header().Set("Content-Type", "application/json")
					_, _ = fmt.Fprintf(w, `{"ref":"tf_12345.md","url":%q}`, uploadURL.String())
				})
			mux.HandleFunc("PUT /storage/tf_12345.md", func(w http.ResponseWriter, r *http.Request) {
				upload = r
				var err error
				if uploaded, err = io.ReadAll(r.Body); err != nil {
					http.Error(w, "Bad Request", http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
			})

			server := httptest.NewServer(mux)
			t.Cleanup(server.Close)

			testEngine := twapi.NewEngine(session.NewBearerToken("your_token", server.URL))

			request := projects.NewPendingFileCreateRequest("plan.md", []byte(contents))
			request.ContentType = tt.contentType

			response, err := projects.PendingFileCreate(t.Context(), testEngine, request)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if response.Ref != "tf_12345.md" {
				t.Errorf("expected the reference from the reservation, got %q", response.Ref)
			}

			if got := presignQuery.Get("fileName"); got != "plan.md" {
				t.Errorf("expected the file name in the reservation, got %q", got)
			}
			if got := presignQuery.Get("fileSize"); got != fmt.Sprint(len(contents)) {
				t.Errorf("expected the size in the reservation, got %q", got)
			}

			if upload == nil {
				t.Fatal("expected the contents to be uploaded")
			}
			if string(uploaded) != contents {
				t.Errorf("expected the file contents in the upload, got %q", uploaded)
			}
			// A chunked upload, which an unmeasured reader produces, is rejected.
			if upload.ContentLength != int64(len(contents)) {
				t.Errorf("expected a content length of %d, got %d", len(contents), upload.ContentLength)
			}
			// The URL carries its own credentials; a second mechanism is refused.
			if got := upload.Header.Get("Authorization"); got != "" {
				t.Errorf("expected no authorization on the upload, got %q", got)
			}
			if got := upload.Header.Get("X-Amz-Acl"); got != tt.wantACL {
				t.Errorf("expected the canned ACL %q, got %q", tt.wantACL, got)
			}
			switch {
			case tt.contentType != "":
				if got := upload.Header.Get("Content-Type"); got != tt.contentType {
					t.Errorf("expected the content type %q, got %q", tt.contentType, got)
				}
			default:
				// The extension's type comes from the host's table, so only assert it is set.
				if upload.Header.Get("Content-Type") == "" {
					t.Error("expected a content type on the upload")
				}
			}
		})
	}
}

func TestPendingFileCreateRejectsMissingFields(t *testing.T) {
	tests := []struct {
		name    string
		request projects.PendingFileCreateRequest
	}{{
		name:    "no file name",
		request: projects.NewPendingFileCreateRequest("", []byte("contents")),
	}, {
		name:    "no contents",
		request: projects.NewPendingFileCreateRequestFromReader("plan.md", nil, 7),
	}, {
		name:    "empty contents",
		request: projects.NewPendingFileCreateRequest("plan.md", nil),
	}, {
		name:    "no size",
		request: projects.NewPendingFileCreateRequestFromReader("plan.md", http.NoBody, 0),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Nothing should be sent: this client fails the test if it is used.
			testEngine := twapi.NewEngine(
				session.NewBearerToken("your_token", "http://example.com"),
				twapi.WithHTTPClient(twapi.HTTPClientFunc(func(*http.Request) (*http.Response, error) {
					t.Error("expected no request to be sent")
					return nil, fmt.Errorf("unexpected request")
				})),
			)

			if _, err := projects.PendingFileCreate(t.Context(), testEngine, tt.request); err == nil {
				t.Error("expected an error, got none")
			}
		})
	}
}

// TestNewPendingFileUploadPlan pins the rules a caller sending the contents
// itself has to follow. They cannot be worked out from the URL by inspection,
// and the storage service is the only thing that notices one broken, so the
// plan is the only place they are written down.
func TestNewPendingFileUploadPlan(t *testing.T) {
	// signedURL builds a pre-signed URL carrying the signature parameters the
	// plan reads, in the form the storage service produces them.
	signedURL := func(path, signedHeaders string, extra url.Values) string {
		query := url.Values{
			"X-Amz-Algorithm":     []string{"AWS4-HMAC-SHA256"},
			"X-Amz-SignedHeaders": []string{signedHeaders},
			"X-Amz-Signature":     []string{"c0ffee"},
		}
		for key, values := range extra {
			query[key] = values
		}
		return (&url.URL{
			Scheme:   "https",
			Host:     "storage.example.com",
			Path:     path,
			RawQuery: query.Encode(),
		}).String()
	}

	tests := []struct {
		name            string
		uploadURL       string
		fileName        string
		contentType     string
		wantACL         string
		wantContentType string
		wantExpiresAt   time.Time
	}{{
		// Signed into the URL unless the installation's bucket sets its own ACL.
		name:            "canned ACL signed",
		uploadURL:       signedURL("/tf_12345.txt", "content-length;host;x-amz-acl", nil),
		fileName:        "notes.txt",
		wantACL:         "public-read",
		wantContentType: "text/plain",
	}, {
		name:            "canned ACL not signed",
		uploadURL:       signedURL("/tf_12345.txt", "content-length;host", nil),
		fileName:        "notes.txt",
		wantACL:         "",
		wantContentType: "text/plain",
	}, {
		name:            "explicit content type wins",
		uploadURL:       signedURL("/tf_12345.txt", "host", nil),
		fileName:        "notes.txt",
		contentType:     "text/markdown",
		wantContentType: "text/markdown",
	}, {
		// The caller knows the name; the reference only carries the extension.
		name:            "file name decides the type",
		uploadURL:       signedURL("/tf_12345.twtest", "host", nil),
		fileName:        "notes.txt",
		wantContentType: "text/plain",
	}, {
		name:            "falls back to the reference extension",
		uploadURL:       signedURL("/tf_12345.txt", "host", nil),
		fileName:        "",
		wantContentType: "text/plain",
	}, {
		name:            "unknown extension is binary",
		uploadURL:       signedURL("/tf_12345.twtest", "host", nil),
		fileName:        "archive.twtest",
		wantContentType: "application/octet-stream",
	}, {
		name: "expiry read from the signature",
		uploadURL: signedURL("/tf_12345.txt", "host", url.Values{
			"X-Amz-Date":    []string{"20260826T120000Z"},
			"X-Amz-Expires": []string{"600"},
		}),
		fileName:        "notes.txt",
		wantContentType: "text/plain",
		wantExpiresAt:   time.Date(2026, 8, 26, 12, 10, 0, 0, time.UTC),
	}, {
		// Reported, never enforced: a deadline that cannot be read is better
		// left unsaid than guessed at.
		name: "unreadable expiry is not guessed",
		uploadURL: signedURL("/tf_12345.txt", "host", url.Values{
			"X-Amz-Date":    []string{"not-a-date"},
			"X-Amz-Expires": []string{"600"},
		}),
		fileName:        "notes.txt",
		wantContentType: "text/plain",
	}, {
		name:            "absent expiry is not guessed",
		uploadURL:       signedURL("/tf_12345.txt", "host", nil),
		fileName:        "notes.txt",
		wantContentType: "text/plain",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan, err := projects.NewPendingFileUploadPlan(tt.uploadURL, tt.fileName, tt.contentType)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if plan.Method != http.MethodPut {
				t.Errorf("expected a PUT, got %s", plan.Method)
			}
			if plan.URL != tt.uploadURL {
				t.Errorf("expected the pre-signed URL unchanged, got %q", plan.URL)
			}
			// The type is the host table's below the top level, so compare the
			// media type and leave any charset parameter alone.
			if got, _, _ := strings.Cut(plan.Headers.Get("Content-Type"), ";"); got != tt.wantContentType {
				t.Errorf("expected the content type %q, got %q", tt.wantContentType, got)
			}
			if got := plan.Headers.Get("X-Amz-Acl"); got != tt.wantACL {
				t.Errorf("expected the canned ACL %q, got %q", tt.wantACL, got)
			}
			// Go derives it from the request, so offering one would be a header
			// the caller must not repeat.
			if got := plan.Headers.Get("Content-Length"); got != "" {
				t.Errorf("expected no content length among the headers, got %q", got)
			}
			if !plan.ExpiresAt.Equal(tt.wantExpiresAt) {
				t.Errorf("expected an expiry of %s, got %s", tt.wantExpiresAt, plan.ExpiresAt)
			}
		})
	}
}

func TestNewPendingFileUploadPlanRejectsBadURL(t *testing.T) {
	tests := []struct {
		name      string
		uploadURL string
	}{{
		name:      "no URL",
		uploadURL: "",
	}, {
		name:      "unparseable URL",
		uploadURL: "https://storage.example.com/\x7f",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := projects.NewPendingFileUploadPlan(tt.uploadURL, "plan.md", ""); err == nil {
				t.Error("expected an error, got none")
			}
		})
	}
}
