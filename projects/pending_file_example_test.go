package projects_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

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

// ExamplePendingFileUpload runs the two steps of an upload separately, which is
// what PendingFileCreate does in one call. Doing it by hand is only worthwhile
// when something other than this process sends the contents, or when they are
// streamed from somewhere the size is already known.
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

	// First the API reserves the space and says where the contents should go.
	presignedResponse, err := projects.PendingFilePresignedURL(ctx, engine,
		projects.NewPendingFilePresignedURLRequest("plan.md", int64(len(contents))))
	if err != nil {
		fmt.Printf("failed to reserve pending file: %s", err)
		return
	}

	// Then the contents travel straight to the storage service, without passing
	// through the API.
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
		// The real URL points at the storage service and carries an AWS signature;
		// this one points back at this server so that the example can run offline.
		// The signature lists the headers it covers, which is what tells the SDK
		// whether the upload has to repeat the canned ACL.
		uploadURL := fmt.Sprintf("http://%s/storage/tf_12345.md"+
			"?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-SignedHeaders=host%%3Bx-amz-acl&X-Amz-Signature=c0ffee", r.Host)
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
			// The upload is the one request that is not authenticated with the
			// Teamwork session: the pre-signed URL carries its own credentials.
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
