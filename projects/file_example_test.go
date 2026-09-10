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

func ExampleFileCreate() {
	address, stop, err := startFileServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	fileRequest := projects.NewFileCreateRequest(777, "tf_12345")
	fileRequest.Description = new("Notes from the kickoff call")

	fileResponse, err := projects.FileCreate(ctx, engine, fileRequest)
	if err != nil {
		fmt.Printf("failed to create file: %s", err)
	} else {
		fmt.Printf("created file with identifier %d\n", fileResponse.ID)
	}

	// Output: created file with identifier 12345
}

func ExampleFileDelete() {
	address, stop, err := startFileServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	_, err = projects.FileDelete(ctx, engine, projects.NewFileDeleteRequest(12345))
	if err != nil {
		fmt.Printf("failed to delete file: %s", err)
	} else {
		fmt.Println("file deleted!")
	}

	// Output: file deleted!
}

func ExampleFileGet() {
	address, stop, err := startFileServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	fileResponse, err := projects.FileGet(ctx, engine, projects.NewFileGetRequest(12345))
	if err != nil {
		fmt.Printf("failed to retrieve file: %s", err)
	} else {
		fmt.Printf("retrieved file %q with identifier %d\n", fileResponse.File.DisplayName, fileResponse.File.ID)
	}

	// Output: retrieved file "kickoff-notes.txt" with identifier 12345
}

func ExampleFileList() {
	address, stop, err := startFileServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	fileRequest := projects.NewFileListRequest()
	fileRequest.Path.ProjectID = 777

	filesResponse, err := projects.FileList(ctx, engine, fileRequest)
	if err != nil {
		fmt.Printf("failed to list files: %s", err)
	} else {
		for _, file := range filesResponse.Files {
			fmt.Printf("retrieved file with identifier %d\n", file.ID)
		}
	}

	// Output:
	// retrieved file with identifier 12345
	// retrieved file with identifier 12346
}

func ExampleFileDownload() {
	address, stop, err := startFileServer() // mock server for demonstration purposes
	if err != nil {
		fmt.Printf("failed to start server: %s", err)
		return
	}
	defer stop()

	ctx := context.Background()
	engine := twapi.NewEngine(session.NewBearerToken("your_token", fmt.Sprintf("http://%s", address)))

	download, err := projects.FileDownload(ctx, engine, projects.NewFileDownloadRequest(12345))
	if err != nil {
		fmt.Printf("failed to download file: %s", err)
		return
	}
	defer func() { _ = download.Body.Close() }()

	content, err := io.ReadAll(download.Body)
	if err != nil {
		fmt.Printf("failed to read file: %s", err)
	} else {
		fmt.Printf("downloaded %s (%s): %s\n", download.Name, download.ContentType, content)
	}

	// Output: downloaded kickoff-notes.txt (text/plain): Notes from the kickoff call
}

func startFileServer() (string, func(), error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /projects/{id}/files", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
			return
		}
		if r.PathValue("id") != "777" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK","id":"12345","fileId":"12345","fileIds":["12345"]}`)
	})
	mux.HandleFunc("DELETE /files/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"STATUS":"OK"}`)
	})

	mux.HandleFunc("GET /projects/api/v3/files/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"file":{"id":12345,"displayName":"kickoff-notes.txt"}}`)
	})
	mux.HandleFunc("GET /projects/api/v3/projects/{id}/files", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "777" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"files":[{"id":12345},{"id":12346}]}`)
	})
	// The download route redirects to storage, as the real one does.
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("action") != "viewFile" || r.URL.Query().Get("fileId") != "12345" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		http.Redirect(w, r, "/storage/kickoff-notes.txt", http.StatusFound)
	})
	mux.HandleFunc("GET /storage/{name}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Disposition", `attachment; filename="`+r.PathValue("name")+`"`)
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "Notes from the kickoff call")
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
