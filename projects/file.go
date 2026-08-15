package projects

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var (
	_ twapi.HTTPRequester = (*FileCreateRequest)(nil)
	_ twapi.HTTPResponser = (*FileCreateResponse)(nil)
	_ twapi.HTTPRequester = (*FileDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*FileDeleteResponse)(nil)
)

// FileCreateRequestPath contains the path parameters for creating a file.
type FileCreateRequestPath struct {
	// ProjectID is the unique identifier of the project that will contain the
	// file.
	ProjectID int64
}

// FileCreateRequest represents the request body for adding a file to a
// project's files area.
//
// Attaching a pending file to a task, a comment or a message already stores it
// in the project's files area, so this endpoint is not a prerequisite for those.
// It is the way to store a file on its own, and the only way to obtain a file
// identifier that can be attached to more than one task, since a pending file
// reference is consumed the first time it is used.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/files/post-projects-id-files-json
type FileCreateRequest struct {
	// Path contains the path parameters for the request.
	Path FileCreateRequestPath `json:"-"`

	// PendingFileRef is the reference of a file uploaded with PendingFileCreate.
	// Unlike the attachment fields on tasks, comments and messages, this endpoint
	// accepts a single reference, so one request stores one file.
	PendingFileRef PendingFileRef `json:"pendingFileRef"`

	// Name overrides the name the file was uploaded with.
	Name *string `json:"name,omitempty"`

	// Description is an optional description of the file.
	Description *string `json:"description,omitempty"`

	// Private hides the file from client users.
	Private *bool `json:"private,omitempty"`

	// CategoryID files the upload under an existing file category.
	CategoryID *int64 `json:"category-id,omitempty"`

	// CategoryName creates a file category with this name, or reuses it when the
	// project already has one. It is ignored when CategoryID is provided.
	CategoryName *string `json:"category-name,omitempty"`

	// TagIDs is the list of tag IDs associated with this file.
	TagIDs LegacyNumericList `json:"tagIds,omitempty"`

	// AutoNewVersion stores the upload as a new version of an existing file with
	// the same name in the project, instead of creating a separate file.
	AutoNewVersion *bool `json:"autoNewVersion,omitempty"`

	// NotifyCurrentUser indicates whether the user adding the file should be
	// notified about it. If not provided, it defaults to false.
	NotifyCurrentUser *bool `json:"notifyCurrentUser,omitempty"`
}

// NewFileCreateRequest creates a new FileCreateRequest with the provided
// required fields.
func NewFileCreateRequest(projectID int64, pendingFileRef PendingFileRef) FileCreateRequest {
	return FileCreateRequest{
		Path: FileCreateRequestPath{
			ProjectID: projectID,
		},
		PendingFileRef: pendingFileRef,
	}
}

// HTTPRequest creates an HTTP request for the FileCreateRequest.
func (f FileCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/%d/files.json", server, f.Path.ProjectID)

	payload := struct {
		File FileCreateRequest `json:"file"`
	}{File: f}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create file request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// FileCreateResponse represents the response body for adding a file to a
// project's files area.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/files/post-projects-id-files-json
type FileCreateResponse struct {
	// ID is the unique identifier of the created file.
	ID LegacyNumber `json:"fileId"`
}

// HandleHTTPResponse handles the HTTP response for the FileCreateResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (f *FileCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create file")
	}
	if err := json.NewDecoder(resp.Body).Decode(f); err != nil {
		return fmt.Errorf("failed to decode create file response: %w", err)
	}
	if f.ID == 0 {
		return fmt.Errorf("create file response does not contain a valid identifier")
	}
	return nil
}

// FileCreate adds a file to a project's files area using the provided request
// and returns the response.
func FileCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req FileCreateRequest,
) (*FileCreateResponse, error) {
	return twapi.Execute[FileCreateRequest, *FileCreateResponse](ctx, engine, req)
}

// FileDeleteRequestPath contains the path parameters for deleting a file.
type FileDeleteRequestPath struct {
	// ID is the unique identifier of the file to be deleted.
	ID int64
}

// FileDeleteRequest represents the request body for deleting a file.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/files/delete-files-id-json
type FileDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path FileDeleteRequestPath
}

// NewFileDeleteRequest creates a new FileDeleteRequest with the provided file
// ID. The ID is required to delete a file.
func NewFileDeleteRequest(fileID int64) FileDeleteRequest {
	return FileDeleteRequest{
		Path: FileDeleteRequestPath{
			ID: fileID,
		},
	}
}

// HTTPRequest creates an HTTP request for the FileDeleteRequest.
func (f FileDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/files/" + strconv.FormatInt(f.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// FileDeleteResponse represents the response body for deleting a file.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/files/delete-files-id-json
type FileDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the FileDeleteResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (f *FileDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to delete file")
	}
	return nil
}

// FileDelete deletes a file using the provided request and returns the
// response.
func FileDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req FileDeleteRequest,
) (*FileDeleteResponse, error) {
	return twapi.Execute[FileDeleteRequest, *FileDeleteResponse](ctx, engine, req)
}
