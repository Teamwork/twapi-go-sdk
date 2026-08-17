package projects_test

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
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
