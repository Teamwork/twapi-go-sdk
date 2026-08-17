package projects

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strconv"
	"strings"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var (
	_ twapi.HTTPRequester = (*PendingFilePresignedURLRequest)(nil)
	_ twapi.HTTPResponser = (*PendingFilePresignedURLResponse)(nil)
)

// PendingFileRef is the opaque handle the API issues for an uploaded file that
// is not attached to anything yet. It is consumed the first time it is attached,
// so each attachment needs its own.
//
// Produced by PendingFileCreate and PendingFilePresignedURL, consumed by the
// attachment fields on tasks, comments and messages.
type PendingFileRef string

// PendingFilePresignedURLRequest represents the request for the first step of an
// upload: reserving space and asking where to send the contents. Use
// PendingFileCreate to run both steps at once.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/file-uploading/put-projects-api-v1-pendingfiles-presignedurl-json
type PendingFilePresignedURLRequest struct {
	// FileName is the name of the file, including its extension. It becomes the
	// name of the attached file.
	FileName string

	// FileSize is the size of the file in bytes. It must be greater than zero.
	FileSize int64
}

// NewPendingFilePresignedURLRequest creates a new PendingFilePresignedURLRequest
// with the provided required fields.
func NewPendingFilePresignedURLRequest(fileName string, fileSize int64) PendingFilePresignedURLRequest {
	return PendingFilePresignedURLRequest{
		FileName: fileName,
		FileSize: fileSize,
	}
}

// HTTPRequest creates an HTTP request for the PendingFilePresignedURLRequest.
func (p PendingFilePresignedURLRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	switch {
	case p.FileName == "":
		return nil, fmt.Errorf("pending file requires a file name")
	case p.FileSize <= 0:
		return nil, fmt.Errorf("pending file requires a size greater than zero")
	}

	// Unlike the other v1 routes here, this one lives under /projects/api/v1/.
	uri := server + "/projects/api/v1/pendingfiles/presignedurl.json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Set("fileName", p.FileName)
	query.Set("fileSize", strconv.FormatInt(p.FileSize, 10))
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// PendingFilePresignedURLResponse represents the response for reserving space
// for a file.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/file-uploading/put-projects-api-v1-pendingfiles-presignedurl-json
type PendingFilePresignedURLResponse struct {
	// Ref identifies the file, but only resolves once the contents are uploaded.
	Ref PendingFileRef `json:"ref"`

	// URL is where the contents must be PUT. It expires in ten minutes and carries
	// its own credentials, so the request to it must not be authenticated.
	URL string `json:"url"`
}

// HandleHTTPResponse handles the HTTP response for the
// PendingFilePresignedURLResponse. If some unexpected HTTP status code is
// returned by the API, a twapi.HTTPError is returned.
func (p *PendingFilePresignedURLResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to reserve pending file")
	}
	if err := json.NewDecoder(resp.Body).Decode(p); err != nil {
		return fmt.Errorf("failed to decode reserve pending file response: %w", err)
	}
	switch {
	case p.Ref == "":
		return fmt.Errorf("reserve pending file response does not contain a valid identifier")
	case p.URL == "":
		return fmt.Errorf("reserve pending file response does not contain an upload URL")
	}
	return nil
}

// PendingFilePresignedURL reserves space for a file and returns its reference
// along with the URL the contents must be sent to. First of the two steps
// PendingFileCreate performs.
func PendingFilePresignedURL(
	ctx context.Context,
	engine *twapi.Engine,
	req PendingFilePresignedURLRequest,
) (*PendingFilePresignedURLResponse, error) {
	return twapi.Execute[PendingFilePresignedURLRequest, *PendingFilePresignedURLResponse](ctx, engine, req)
}

// PendingFileUploadRequest represents the request for the second step of an
// upload: sending the contents to the pre-signed URL.
//
// It addresses the storage service, not the API, and the URL authenticates it,
// so it is not a twapi.HTTPRequester: it is sent with twapi.Engine.Do.
type PendingFileUploadRequest struct {
	// URL is the pre-signed URL from PendingFilePresignedURLResponse.
	URL string

	// Contents is the file body. It is read once, so the request cannot be sent
	// twice.
	Contents io.Reader

	// Size is the number of bytes in Contents. Required: the storage service needs
	// the length up front.
	Size int64

	// ContentType is the media type stored with the file. Defaults to the type of
	// the extension in URL.
	ContentType string
}

// NewPendingFileUploadRequest creates a new PendingFileUploadRequest with the
// provided required fields.
func NewPendingFileUploadRequest(uploadURL string, contents io.Reader, size int64) PendingFileUploadRequest {
	return PendingFileUploadRequest{
		URL:      uploadURL,
		Contents: contents,
		Size:     size,
	}
}

// PendingFileUploadResponse represents the response for sending the contents of
// a file to a pre-signed URL. The storage service returns nothing of interest;
// the reference to attach came from PendingFilePresignedURL.
type PendingFileUploadResponse struct{}

// PendingFileUpload sends the contents of a file to the pre-signed URL returned
// by PendingFilePresignedURL. Second of the two steps PendingFileCreate
// performs, exposed for callers that stream the contents or got the URL
// elsewhere.
func PendingFileUpload(
	ctx context.Context,
	engine *twapi.Engine,
	req PendingFileUploadRequest,
) (*PendingFileUploadResponse, error) {
	switch {
	case req.URL == "":
		return nil, fmt.Errorf("pending file upload requires a pre-signed URL")
	case req.Contents == nil:
		return nil, fmt.Errorf("pending file upload requires the file contents")
	case req.Size <= 0:
		return nil, fmt.Errorf("pending file upload requires a size greater than zero")
	}

	uploadURL, err := url.Parse(req.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the pre-signed URL: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, req.URL, req.Contents)
	if err != nil {
		return nil, err
	}
	// An unmeasured reader would be sent chunked, which the storage service rejects.
	httpReq.ContentLength = req.Size

	contentType := req.ContentType
	if contentType == "" {
		// The reference in the URL keeps the original extension.
		contentType = contentTypeForFileName(uploadURL.Path)
	}
	httpReq.Header.Set("Content-Type", contentType)

	// The signature covers the headers it lists: one missing, or an unsigned
	// x-amz-* one added, and the upload fails. Whether the ACL is signed depends on
	// the installation's bucket, so read it from the URL instead of guessing.
	signedHeaders := strings.Split(uploadURL.Query().Get("X-Amz-SignedHeaders"), ";")
	if slices.Contains(signedHeaders, "x-amz-acl") {
		httpReq.Header.Set("X-Amz-Acl", "public-read")
	}

	resp, err := engine.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// The storage service picks its own success status, so accept any 2xx.
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, twapi.NewHTTPError(resp, "failed to upload pending file")
	}

	return &PendingFileUploadResponse{}, nil
}

// PendingFileCreateRequest represents the request for uploading a file. The file
// is stored on its own, and a later request attaches it to a task, a comment or
// a message.
type PendingFileCreateRequest struct {
	// FileName is the name of the file, including its extension. It becomes the
	// name of the attached file.
	FileName string

	// Contents is the file body. It is read once, so the request cannot be sent
	// twice.
	Contents io.Reader

	// Size is the number of bytes in Contents. Both constructors fill it in.
	Size int64

	// ContentType is the media type stored with the file. Defaults to the type of
	// FileName's extension, as known to the standard library.
	ContentType string
}

// NewPendingFileCreateRequest creates a new PendingFileCreateRequest for
// contents held in memory.
func NewPendingFileCreateRequest(fileName string, contents []byte) PendingFileCreateRequest {
	return PendingFileCreateRequest{
		FileName: fileName,
		Contents: bytes.NewReader(contents),
		Size:     int64(len(contents)),
	}
}

// NewPendingFileCreateRequestFromReader creates a new PendingFileCreateRequest
// for contents read as they are sent, keeping a large file out of memory. The
// size must be known in advance; for a file on disk it comes from os.File.Stat.
func NewPendingFileCreateRequestFromReader(
	fileName string,
	contents io.Reader,
	size int64,
) PendingFileCreateRequest {
	return PendingFileCreateRequest{
		FileName: fileName,
		Contents: contents,
		Size:     size,
	}
}

// PendingFileCreateResponse represents the response for uploading a file.
type PendingFileCreateResponse struct {
	// Ref identifies the uploaded file until it is attached to something.
	Ref PendingFileRef
}

// PendingFileCreate uploads a file and returns the reference identifying it,
// which a later request attaches to a task, a comment or a message.
//
// It performs both steps of the upload: PendingFilePresignedURL reserves the
// space, then PendingFileUpload sends the contents straight to storage rather
// than through the API.
func PendingFileCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req PendingFileCreateRequest,
) (*PendingFileCreateResponse, error) {
	switch {
	case req.FileName == "":
		return nil, fmt.Errorf("pending file requires a file name")
	case req.Contents == nil:
		return nil, fmt.Errorf("pending file requires the file contents")
	case req.Size <= 0:
		return nil, fmt.Errorf("pending file requires a size greater than zero")
	}

	presigned, err := PendingFilePresignedURL(ctx, engine,
		NewPendingFilePresignedURLRequest(req.FileName, req.Size))
	if err != nil {
		return nil, err
	}

	upload := NewPendingFileUploadRequest(presigned.URL, req.Contents, req.Size)
	// The file name is known here; the upload only sees the reference in the URL.
	upload.ContentType = req.ContentType
	if upload.ContentType == "" {
		upload.ContentType = contentTypeForFileName(req.FileName)
	}
	if _, err := PendingFileUpload(ctx, engine, upload); err != nil {
		return nil, err
	}

	return &PendingFileCreateResponse{Ref: presigned.Ref}, nil
}

// contentTypeForFileName derives a media type from a file extension, falling
// back to the generic binary type.
func contentTypeForFileName(fileName string) string {
	if contentType := mime.TypeByExtension(path.Ext(fileName)); contentType != "" {
		return contentType
	}
	return "application/octet-stream"
}
