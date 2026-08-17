package projects

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
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

// PendingFileRef is the handle the API issues for a file that has been uploaded
// but is not yet attached to anything. Treat it as opaque: it is a token to pass
// back, not a value to parse.
//
// A reference is scoped to the installation, does not expire, and is consumed
// the first time it is attached to something. Attaching the same reference twice
// fails with "Temporary file reference not found".
//
// References are produced by PendingFileCreate, or by PendingFilePresignedURL
// for a caller running the upload steps itself, and are consumed by the
// attachment fields on tasks, comments and messages.
type PendingFileRef string

// PendingFilePresignedURLRequest represents the request for the first step of an
// upload: reserving space for a file and asking where to send it.
//
// The file itself is not part of this request. The API answers with a reference
// and a pre-signed URL, and the contents then travel straight to the storage
// service, so they never pass through the API. PendingFileCreate performs both
// steps; use this request only when the contents are sent by something else, for
// example a browser that is handed the URL.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/file-uploading/put-projects-api-v1-pendingfiles-presignedurl-json
type PendingFilePresignedURLRequest struct {
	// FileName is the name of the file, including its extension. The name is
	// stored with the reservation and becomes the name of the attached file, and
	// its extension decides how the file is presented, so a name without one is
	// harder for the recipient to open.
	FileName string

	// FileSize is the size of the file in bytes, which the API requires to be
	// greater than zero. It reserves the space, so it should be the real size of
	// the contents that follow.
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

	// Unlike the other v1 routes in this package, which hang off the bare server
	// root, this endpoint genuinely lives under /projects/api/v1/.
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
	// Ref identifies the reserved file until it is attached to something. It only
	// resolves to a file once the contents have been sent to URL; attaching it
	// before that fails.
	Ref PendingFileRef `json:"ref"`

	// URL is where the contents must be sent with a PUT request. It carries its
	// own credentials, so a request to it must not also carry the Teamwork
	// session, and it is short-lived, so it is not worth storing. PendingFileUpload
	// sends the contents with the headers the URL was signed for.
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

// PendingFilePresignedURL reserves space for a file and returns the reference
// that will identify it along with the URL its contents must be sent to. It is
// the first of the two steps PendingFileCreate performs.
func PendingFilePresignedURL(
	ctx context.Context,
	engine *twapi.Engine,
	req PendingFilePresignedURLRequest,
) (*PendingFilePresignedURLResponse, error) {
	return twapi.Execute[PendingFilePresignedURLRequest, *PendingFilePresignedURLResponse](ctx, engine, req)
}

// PendingFileUploadRequest represents the request for the second step of an
// upload: sending the contents to the URL obtained from
// PendingFilePresignedURL.
//
// This request does not address the Teamwork API but the storage service the
// pre-signed URL points at, which authenticates it with the credentials in the
// URL. That is also why it is not a twapi.HTTPRequester: the URL is absolute and
// the Teamwork session must be left out of it.
type PendingFileUploadRequest struct {
	// URL is the pre-signed URL from PendingFilePresignedURLResponse.
	URL string

	// Contents is the file body. It is read once while the request is sent, so a
	// request value holding a reader cannot be sent twice.
	Contents io.Reader

	// Size is the number of bytes in Contents. The storage service needs the
	// length up front, and it is the size the API reserved, so it cannot be
	// discovered while sending.
	Size int64

	// ContentType is the media type recorded with the stored file, which decides
	// whether a browser later displays it or downloads it. When empty it is
	// derived from the extension of the file name embedded in URL.
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
// a file to a pre-signed URL. The storage service answers with no content of
// interest; the reference to attach is the one that came back from
// PendingFilePresignedURL.
type PendingFileUploadResponse struct{}

// PendingFileUpload sends the contents of a file to the pre-signed URL returned
// by PendingFilePresignedURL. It is the second of the two steps
// PendingFileCreate performs, and is exposed for callers that hold the contents
// as a stream, or that obtained the URL elsewhere.
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
	// A reader the standard library cannot measure leaves the length unset, and
	// the request would be sent chunked, which the storage service rejects.
	httpReq.ContentLength = req.Size

	contentType := req.ContentType
	if contentType == "" {
		// The reference in the URL keeps the extension of the original file name,
		// which is all there is to go on when the caller did not say.
		contentType = contentTypeForFileName(uploadURL.Path)
	}
	httpReq.Header.Set("Content-Type", contentType)

	// The signature in the URL covers a set of headers, and the URL itself lists
	// which ones. A signed header missing from the request, or an x-amz-* header
	// present that was not signed, and the upload is rejected. Whether the API
	// signs a canned ACL depends on the installation's bucket, so the URL decides
	// this rather than a guess at the environment.
	signedHeaders := strings.Split(uploadURL.Query().Get("X-Amz-SignedHeaders"), ";")
	if slices.Contains(signedHeaders, "x-amz-acl") {
		httpReq.Header.Set("X-Amz-Acl", "public-read")
	}

	resp, err := engine.HTTPClient().Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			engine.Logger().Error("failed to close response body",
				slog.String("error", err.Error()),
			)
		}
	}()

	// The response comes from a storage service rather than the Teamwork API, and
	// which success status it answers with is its own business, so anything in the
	// 2xx range counts.
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
	// name of the attached file, and its extension decides how the file is
	// presented, so a name without one is harder for the recipient to open.
	FileName string

	// Contents is the file body. It is read once while the file is sent, so a
	// request value built from a reader cannot be sent twice.
	Contents io.Reader

	// Size is the number of bytes in Contents. The API reserves the space before
	// the contents are sent, so the size has to be known in advance. Both
	// constructors fill it in.
	Size int64

	// ContentType is the media type recorded with the stored file, which decides
	// whether a browser later displays it or downloads it. When empty it is
	// derived from the extension of FileName, using the same table as the standard
	// library, so set it explicitly for a name without a useful extension or for a
	// type the host does not know about.
	ContentType string
}

// NewPendingFileCreateRequest creates a new PendingFileCreateRequest for
// contents already held in memory.
func NewPendingFileCreateRequest(fileName string, contents []byte) PendingFileCreateRequest {
	return PendingFileCreateRequest{
		FileName: fileName,
		Contents: bytes.NewReader(contents),
		Size:     int64(len(contents)),
	}
}

// NewPendingFileCreateRequestFromReader creates a new PendingFileCreateRequest
// for contents read as they are sent, which keeps a large file out of memory.
// The size has to be known in advance, as it is what the API reserves; for a
// file on disk it comes from os.File.Stat.
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

// PendingFileCreate uploads a file and returns the reference that identifies it.
// The reference is not attached to anything yet; attaching it is a separate
// operation on whichever entity should own the file.
//
// The upload takes two requests, which this function performs in order: the API
// reserves space for the file and answers with a pre-signed URL
// (PendingFilePresignedURL), and the contents are then sent to that URL
// (PendingFileUpload). The contents travel straight to the storage service
// rather than through the API, which is what keeps large files viable.
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
	// The upload request can only derive the media type from the reference in the
	// URL, while the original file name is known here.
	upload.ContentType = req.ContentType
	if upload.ContentType == "" {
		upload.ContentType = contentTypeForFileName(req.FileName)
	}
	if _, err := PendingFileUpload(ctx, engine, upload); err != nil {
		return nil, err
	}

	return &PendingFileCreateResponse{Ref: presigned.Ref}, nil
}

// contentTypeForFileName derives the media type of a file from its extension,
// falling back to the generic binary type, which is what an unrecognised
// extension would end up stored as anyway.
func contentTypeForFileName(fileName string) string {
	if contentType := mime.TypeByExtension(path.Ext(fileName)); contentType != "" {
		return contentType
	}
	return "application/octet-stream"
}
