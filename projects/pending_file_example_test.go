package projects_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
	"github.com/teamwork/twapi-go-sdk/session"
)

func ExamplePendingFileCreate() {
	address, stop, err := startPendingFileServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	pendingFileRequest := projects.NewPendingFileCreateRequest("plan.md", []byte("# Plan\n"))

	pendingFileResponse, err := projects.PendingFileCreate(ctx, engine, pendingFileRequest)
	if err != nil {
		fmt.Printf("failed to create pending file: %s", err)
	} else {
		fmt.Printf("created pending file with reference %s\n", pendingFileResponse.Ref)
	}

	// Output: created pending file with reference tf_12345.md
}

// ExamplePendingFileUpload runs the two steps PendingFileCreate groups, which is
// only worth doing by hand to stream the contents or to let something else send
// them.
func ExamplePendingFileUpload() {
	address, stop, err := startPendingFileServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	contents := "# Plan\n"

	// The API reserves the space and says where the contents go.
	presignedResponse, err := projects.PendingFilePresignedURL(ctx, engine,
		projects.NewPendingFilePresignedURLRequest("plan.md", int64(len(contents))))
	if err != nil {
		fmt.Printf("failed to reserve pending file: %s", err)
		return
	}

	// The contents then go straight to storage, not through the API.
	_, err = projects.PendingFileUpload(ctx, engine, projects.NewPendingFileUploadRequest(
		presignedResponse.URL,
		strings.NewReader(contents),
		int64(len(contents)),
	))
	if err != nil {
		fmt.Printf("failed to upload pending file: %s", err)
	} else {
		fmt.Printf("uploaded pending file with reference %s\n", presignedResponse.Ref)
	}

	// Output: uploaded pending file with reference tf_12345.md
}

func startPendingFileServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /projects/api/v1/pendingfiles/presignedurl", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("fileName") != "plan.md" || r.URL.Query().Get("fileSize") != "7" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		// The real URL points at storage; this one comes back here so the example
		// runs offline. Its signed headers tell the SDK to repeat the canned ACL.
		uploadURL := fmt.Sprintf("http://%s/storage/tf_12345.md"+
			"?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-SignedHeaders=host%%3Bx-amz-acl&X-Amz-Signature=c0ffee"+
			"&X-Amz-Date=20260826T120000Z&X-Amz-Expires=600", r.Host)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"ref":"tf_12345.md","url":%q}`+"\n", uploadURL)
	})
	mux.HandleFunc("PUT /storage/tf_12345.md", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Amz-Acl") != "public-read" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		contents, err := io.ReadAll(r.Body)
		if err != nil || string(contents) != "# Plan\n" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// The upload is the one request without the Teamwork session.
			if !strings.HasPrefix(r.URL.Path, "/storage/") {
				if r.Header.Get("Authorization") != "Bearer your_token" {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			}
			r.URL.Path = strings.TrimSuffix(r.URL.Path, ".json")
			mux.ServeHTTP(w, r)
		}),
	}

	stop := make(chan struct{})
	go func() {
		_ = server.Serve(ln)
	}()
	go func() {
		<-stop
		_ = server.Shutdown(context.Background())
	}()

	return ln.Addr().String(), func() {
		close(stop)
	}, nil
}

// ExampleNewPendingFileUploadPlan reserves the space and then describes the
// upload rather than performing it, which is what a caller that never holds the
// bytes has to do: a server handing the URL to the client that has the file, or
// to a browser. The rules cannot be guessed from the URL, so they travel with
// it.
func ExampleNewPendingFileUploadPlan() {
	address, stop, err := startPendingFileServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	const size = 7

	presignedResponse, err := projects.PendingFilePresignedURL(ctx, engine,
		projects.NewPendingFilePresignedURLRequest("plan.md", size))
	if err != nil {
		fmt.Printf("failed to reserve pending file: %s", err)
		return
	}

	// The media type is derived from the name when it is not given, but the
	// host mime table decides that, so this names it outright.
	plan, err := projects.NewPendingFileUploadPlan(presignedResponse.URL, "plan.md", "text/markdown")
	if err != nil {
		fmt.Printf("failed to plan the upload: %s", err)
		return
	}

	// Whoever holds the contents sends exactly this, adding a Content-Length of
	// the size reserved above and no authorization of its own.
	fmt.Printf("%s with %s\n", plan.Method, plan.Headers.Get("Content-Type"))
	fmt.Printf("canned ACL: %q\n", plan.Headers.Get("X-Amz-Acl"))
	fmt.Printf("before %s\n", plan.ExpiresAt.Format(time.RFC3339))
	fmt.Printf("then attach reference %s\n", presignedResponse.Ref)

	// Output:
	// PUT with text/markdown
	// canned ACL: "public-read"
	// before 2026-08-26T12:10:00Z
	// then attach reference tf_12345.md
}
