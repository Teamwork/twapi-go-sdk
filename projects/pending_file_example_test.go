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

	pendingFileRequest := projects.NewPendingFileCreateRequestFromBytes("plan.md", []byte("# Plan\n"))

	pendingFileResponse, err := projects.PendingFileCreate(ctx, engine, pendingFileRequest)
	if err != nil {
		fmt.Printf("failed to create pending file: %s", err)
	} else {
		fmt.Printf("created pending file with reference %s\n", pendingFileResponse.PendingFile.Ref)
	}

	// Output: created pending file with reference tf_12345
}

func startPendingFileServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /projects/api/v1/pendingfiles", func(w http.ResponseWriter, r *http.Request) {
		// Unlike every other endpoint in this package the body is multipart, and
		// the content type carries a generated boundary, so it cannot be compared
		// for equality.
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		defer func() {
			_ = file.Close()
		}()
		contents, err := io.ReadAll(file)
		if err != nil || header.Filename != "plan.md" || string(contents) != "# Plan\n" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"pendingFile":{"ref":"tf_12345"}}`)
	})

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer your_token" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
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
