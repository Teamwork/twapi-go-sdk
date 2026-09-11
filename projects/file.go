package projects

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var (
	_ twapi.HTTPRequester = (*FileCreateRequest)(nil)
	_ twapi.HTTPResponser = (*FileCreateResponse)(nil)
	_ twapi.HTTPRequester = (*FileDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*FileDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*FileGetRequest)(nil)
	_ twapi.HTTPResponser = (*FileGetResponse)(nil)
	_ twapi.HTTPRequester = (*FileListRequest)(nil)
	_ twapi.HTTPResponser = (*FileListResponse)(nil)
	_ twapi.HTTPRequester = (*FileDownloadRequest)(nil)
)

// FileStatus is the storage state of a file.
type FileStatus string

// Supported file statuses.
const (
	FileStatusActive  FileStatus = "active"
	FileStatusDeleted FileStatus = "deleted"
)

// FileRequestSideload contains the possible sideload options when loading
// files.
type FileRequestSideload string

// List of possible sideload options for FileRequestSideload.
const (
	FileRequestSideloadUsers    FileRequestSideload = "users"
	FileRequestSideloadProjects FileRequestSideload = "projects"
	FileRequestSideloadTags     FileRequestSideload = "tags"
	FileRequestSideloadTasks    FileRequestSideload = "tasks"
)

// FileOrderBy identifies the attributes a file list can be ordered by.
type FileOrderBy string

// Supported file order-by values.
const (
	FileOrderByName         FileOrderBy = "name"
	FileOrderByProjectName  FileOrderBy = "projectname"
	FileOrderByCategoryName FileOrderBy = "categoryname"
	FileOrderByDateUploaded FileOrderBy = "dateuploaded"
	FileOrderBySize         FileOrderBy = "size"
	FileOrderByID           FileOrderBy = "id"
)

// FileVersion is one upload of a file. A file keeps every version that was
// uploaded under its name, and the file's own attributes describe the latest
// one unless a request selects another.
type FileVersion struct {
	// ID is the unique identifier of the version.
	ID int64 `json:"fileVersionId"`

	// File is the file this version belongs to.
	File twapi.Relationship `json:"file"`

	// VersionNo is the position of this version in the file's history, starting
	// at 1.
	VersionNo int64 `json:"versionNo"`

	// Name is the name the version was uploaded with.
	Name string `json:"name"`

	// OriginalName is the name of the uploaded file before any rename.
	OriginalName string `json:"originalName"`

	// DisplayName is the name shown for the version.
	DisplayName string `json:"displayName"`

	// Description is the description entered for the version.
	Description string `json:"description"`

	// Size is the size of the version in bytes.
	Size int64 `json:"size"`

	// Status is the storage state of the version.
	Status FileStatus `json:"status"`

	// Project is the project the file belongs to.
	Project twapi.Relationship `json:"project"`

	// UploadedBy is the unique identifier of the user who uploaded the version.
	UploadedBy int64 `json:"uploadedBy"`

	// UploadedAt is when the version was uploaded.
	UploadedAt time.Time `json:"uploadedAt"`
}

// FileRelatedItems lists the items a file is attached to.
type FileRelatedItems struct {
	// Tasks is the list of task IDs the file is attached to.
	Tasks []int64 `json:"tasks"`

	// Messages is the list of message IDs the file is attached to.
	Messages []int64 `json:"messages"`

	// Comments is the list of comment IDs the file is attached to.
	Comments []int64 `json:"comments"`
}

// File is a document stored in a project's files area. Files are uploaded
// directly or arrive as attachments on tasks, comments and messages, and every
// upload under the same name is kept as a version.
//
// More information can be found at:
// https://support.teamwork.com/projects/files
//
// sparsefields:gen
type File struct {
	// ID is the unique identifier of the file.
	ID int64 `json:"id"`

	// DisplayName is the name shown for the file.
	DisplayName string `json:"displayName"`

	// OriginalName is the name of the uploaded file before any rename.
	OriginalName string `json:"originalName"`

	// Description is the description entered for the file.
	Description string `json:"description"`

	// Size is the size of the selected version in bytes.
	Size int64 `json:"size"`

	// Status is the storage state of the file.
	Status FileStatus `json:"status"`

	// IsPrivate is 1 when the file is hidden from client users, 0 otherwise.
	IsPrivate int64 `json:"isPrivate"`

	// IsLocked indicates whether the file is locked against new versions.
	IsLocked bool `json:"isLocked"`

	// LockedBy is the unique identifier of the user holding the lock, 0 when the
	// file is not locked.
	LockedBy int64 `json:"lockedBy"`

	// LockedAt is when the lock was taken, if the file is locked.
	LockedAt *time.Time `json:"lockedAt"`

	// Version is the selected version.
	Version twapi.Relationship `json:"version"`

	// LatestFileVersionNo is the number of versions the file has.
	LatestFileVersionNo int64 `json:"latestFileVersionNo"`

	// Versions is the file's version history. It is only populated when the
	// request sets IncludeVersions.
	Versions []FileVersion `json:"versions,omitempty"`

	// Project is the project the file belongs to.
	Project twapi.Relationship `json:"project"`

	// Category is the file category, if any.
	Category *twapi.Relationship `json:"category"`

	// Tags is the list of tags associated with the file.
	Tags []twapi.Relationship `json:"tags,omitempty"`

	// FileSource identifies where the file is stored: "teamworkpm" for uploads,
	// or the name of the linked cloud storage provider.
	FileSource string `json:"fileSource"`

	// UploadedBy is the unique identifier of the user who uploaded the selected
	// version.
	UploadedBy int64 `json:"uploadedBy"`

	// UploadedAt is when the selected version was uploaded.
	UploadedAt time.Time `json:"uploadedAt"`

	// UpdatedAt is when the file was last changed.
	UpdatedAt *time.Time `json:"updatedAt"`

	// DeletedAt is when the file was deleted, if it was.
	DeletedAt *time.Time `json:"deletedAt"`

	// DeletedBy is the unique identifier of the user who deleted the file, if it
	// was deleted.
	DeletedBy *int64 `json:"deletedBy"`

	// DownloadURL is the address the selected version can be downloaded from,
	// authenticated like any other request to the site. It is nil unless both
	// the file and the version are active. FileDownload fetches it.
	DownloadURL *string `json:"downloadURL"`

	// PreviewURL is the address of the document viewer for the selected version.
	// It is nil unless both the file and the version are active.
	PreviewURL *string `json:"previewURL"`

	// ThumbURL is the address of a thumbnail, for files that have one.
	ThumbURL string `json:"thumbURL,omitempty"`

	// RelatedItems lists the items the file is attached to. It is only populated
	// when a single file is retrieved.
	RelatedItems *FileRelatedItems `json:"relatedItems,omitempty"`

	// CommentsCount is the number of comments on the file. It is only populated
	// when a single file is retrieved.
	CommentsCount *int64 `json:"commentsCount,omitempty"`

	// Shareable indicates whether a public sharing link can be created for the
	// file.
	Shareable *bool `json:"shareable,omitempty"`
}

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
	switch {
	case f.Path.ProjectID <= 0:
		return nil, fmt.Errorf("file requires a project")
	case f.PendingFileRef == "":
		return nil, fmt.Errorf("file requires a pending file reference")
	}

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
	ID LegacyNumber `json:"id"`
}

// HandleHTTPResponse handles the HTTP response for the FileCreateResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (f *FileCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create file")
	}
	// "id" is the documented field and the one every other v1 create response in
	// this package uses, but this endpoint also returns it as "fileId"; accept
	// that as a fallback so a payload carrying only one of them still resolves.
	var body struct {
		ID     LegacyNumber `json:"id"`
		FileID LegacyNumber `json:"fileId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return fmt.Errorf("failed to decode create file response: %w", err)
	}
	if f.ID = body.ID; f.ID == 0 {
		f.ID = body.FileID
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

// FileGetRequestPath contains the path parameters for loading a single file.
type FileGetRequestPath struct {
	// ID is the unique identifier of the file to be retrieved.
	ID int64
}

// FileGetRequest represents the request body for loading a single file.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/files/get-projects-api-v3-files-file-id-json
type FileGetRequest struct {
	// Path contains the path parameters for the request.
	Path FileGetRequestPath

	// Version selects the version whose attributes, size and download address
	// are reported. Zero selects the latest version.
	Version int64

	// IncludeVersions populates File.Versions with the whole version history.
	IncludeVersions bool

	// Include is the list of related entities to sideload alongside the file.
	Include []FileRequestSideload

	// Fields restricts the attributes returned for the file and each of its
	// sideloads. Each slot of FileGetFields is a separate `fields[entity]=…`
	// selection; populated slots restrict the response, empty slots return the
	// API default. Use the generated FileField constants to ensure values match
	// real attributes.
	Fields FileGetFields
}

// NewFileGetRequest creates a new FileGetRequest with the provided file ID. The
// ID is required to load a file.
func NewFileGetRequest(fileID int64) FileGetRequest {
	return FileGetRequest{
		Path: FileGetRequestPath{
			ID: fileID,
		},
	}
}

// HTTPRequest creates an HTTP request for the FileGetRequest.
func (f FileGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/files/" + strconv.FormatInt(f.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	querySetInt64(query, "version", f.Version)
	if f.IncludeVersions {
		query.Set("getVersions", "true")
	}
	querySetStrings(query, "include", f.Include)
	f.Fields.apply(query)
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// FileGetResponse contains all the information related to a file.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/files/get-projects-api-v3-files-file-id-json
//
// sparsefields:get
type FileGetResponse struct {
	File File `json:"file"`

	Included struct {
		// Users contains the users referenced by the file, such as the uploader.
		//
		// The key is the string representation of the user ID.
		Users map[string]User `json:"users,omitempty"`
		// Projects contains the project the file belongs to.
		//
		// The key is the string representation of the project ID.
		Projects map[string]Project `json:"projects,omitempty"`
		// Tags contains the tags associated with the file.
		//
		// The key is the string representation of the tag ID.
		Tags map[string]Tag `json:"tags,omitempty"`
		// Tasks contains the tasks the file is attached to.
		//
		// The key is the string representation of the task ID.
		Tasks map[string]Task `json:"tasks,omitempty"`
	} `json:"included"`
}

// HandleHTTPResponse handles the HTTP response for the FileGetResponse. If some
// unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (f *FileGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve file")
	}

	if err := json.NewDecoder(resp.Body).Decode(f); err != nil {
		return fmt.Errorf("failed to decode retrieve file response: %w", err)
	}
	return nil
}

// FileGet retrieves a single file using the provided request and returns the
// response.
func FileGet(
	ctx context.Context,
	engine *twapi.Engine,
	req FileGetRequest,
) (*FileGetResponse, error) {
	return twapi.Execute[FileGetRequest, *FileGetResponse](ctx, engine, req)
}

// FileListRequestPath contains the path parameters for loading multiple files.
type FileListRequestPath struct {
	// ProjectID scopes the list to one project's files area. Zero lists the
	// files of every project the user can access.
	ProjectID int64
}

// FileListRequestFilters contains the filters for loading multiple files.
type FileListRequestFilters struct {
	// IDs restricts the list to the given file IDs.
	IDs []int64

	// TaskID restricts the list to the files attached to the given task.
	TaskID int64

	// CategoryID restricts the list to the files in the given file category.
	CategoryID int64

	// TagIDs restricts the list to files carrying any of the given tags.
	TagIDs []int64

	// UserIDs restricts the list to files uploaded by the given users.
	UserIDs []int64

	// SearchTerm restricts the list to files whose name contains the term.
	SearchTerm string

	// SearchAllFields widens SearchTerm to the file extension, the file
	// category, the original file name and the name of the latest uploader.
	SearchAllFields bool

	// UploadedStartDate restricts the list to files whose selected version was
	// uploaded on or after the start of the given day.
	UploadedStartDate *twapi.Date

	// UploadedEndDate restricts the list to files whose selected version was
	// uploaded no later than the start of the given day. The bound is the first
	// instant of the day, so a file uploaded later that day does not match.
	UploadedEndDate *twapi.Date

	// UpdatedAfter restricts the list to files changed strictly after the given
	// instant, on either the file or its selected version.
	UpdatedAfter *time.Time

	// ShowDeleted also lists deleted files.
	ShowDeleted bool

	// SkipExternalFiles leaves out files that live in a linked cloud storage
	// provider rather than in the project's own storage, so only uploads are
	// listed. File.FileSource is what distinguishes the two on a row.
	SkipExternalFiles bool

	// IncludeVersions populates File.Versions of every row with the whole
	// version history.
	IncludeVersions bool

	// Include is the list of related entities to sideload alongside the files.
	Include []FileRequestSideload

	// OrderBy is the field to sort the results by. Use the FileOrderBy
	// constants. The endpoint defaults to name.
	OrderBy FileOrderBy

	// OrderMode is the direction to sort the results in. See twapi.OrderMode for
	// the supported values. The endpoint defaults to ascending.
	OrderMode twapi.OrderMode

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of files to retrieve per page. Defaults to 50.
	PageSize int64

	// CountMode selects whether the API computes the exact number of files
	// matching the filters, reported in Meta.Page.Count. Defaults to
	// twapi.ListCountModeDefault, which leaves the decision to the API.
	CountMode twapi.ListCountMode

	// Fields restricts the attributes returned for the file and each of its
	// sideloads. Each slot of FileListFields is a separate `fields[entity]=…`
	// selection; populated slots restrict the response, empty slots return the
	// API default. Use the generated FileField constants to ensure values match
	// real attributes.
	Fields FileListFields
}

func (f FileListRequestFilters) apply(req *http.Request) {
	query := req.URL.Query()
	querySetInt64s(query, "ids", f.IDs)
	querySetInt64(query, "taskId", f.TaskID)
	querySetInt64(query, "categoryId", f.CategoryID)
	querySetInt64s(query, "tagIds", f.TagIDs)
	querySetInt64s(query, "userIds", f.UserIDs)
	querySetString(query, "searchTerm", f.SearchTerm)
	if f.SearchAllFields {
		query.Set("searchAllFields", "true")
	}
	querySetDate(query, "uploadedStartDate", f.UploadedStartDate)
	querySetDate(query, "uploadedEndDate", f.UploadedEndDate)
	querySetTimestamp(query, "updatedAfter", f.UpdatedAfter)
	if f.ShowDeleted {
		query.Set("showDeleted", "true")
	}
	if f.SkipExternalFiles {
		query.Set("skipExternalFiles", "true")
	}
	if f.IncludeVersions {
		query.Set("getVersions", "true")
	}
	querySetStrings(query, "include", f.Include)
	querySetString(query, "orderBy", f.OrderBy)
	querySetString(query, "orderMode", f.OrderMode)
	querySetInt64(query, "page", f.Page)
	querySetInt64(query, "pageSize", f.PageSize)
	f.Fields.apply(query)
	f.CountMode.Apply(query)
	req.URL.RawQuery = query.Encode()
}

// FileListRequest represents the request body for loading multiple files.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/files/get-projects-api-v3-files-json
type FileListRequest struct {
	// Path contains the path parameters for the request.
	Path FileListRequestPath

	// Filters contains the filters for loading multiple files.
	Filters FileListRequestFilters
}

// NewFileListRequest creates a new FileListRequest with default values.
func NewFileListRequest() FileListRequest {
	return FileListRequest{
		Filters: FileListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the FileListRequest.
func (f FileListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/files.json"
	if f.Path.ProjectID > 0 {
		uri = fmt.Sprintf("%s/projects/api/v3/projects/%d/files.json", server, f.Path.ProjectID)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	f.Filters.apply(req)

	return req, nil
}

// FileListResponse contains information by multiple files matching the request
// filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/files/get-projects-api-v3-files-json
//
// sparsefields:list
type FileListResponse struct {
	request FileListRequest

	Meta  twapi.ListMeta `json:"meta"`
	Files []File         `json:"files"`

	Included struct {
		// Users contains the users referenced by the files, such as uploaders.
		//
		// The key is the string representation of the user ID.
		Users map[string]User `json:"users,omitempty"`
		// Projects contains the projects the files belong to.
		//
		// The key is the string representation of the project ID.
		Projects map[string]Project `json:"projects,omitempty"`
		// Tags contains the tags associated with the files.
		//
		// The key is the string representation of the tag ID.
		Tags map[string]Tag `json:"tags,omitempty"`
		// Tasks contains the tasks the files are attached to.
		//
		// The key is the string representation of the task ID.
		Tasks map[string]Task `json:"tasks,omitempty"`
	} `json:"included"`
}

// HandleHTTPResponse handles the HTTP response for the FileListResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (f *FileListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list files")
	}

	if err := json.NewDecoder(resp.Body).Decode(f); err != nil {
		return fmt.Errorf("failed to decode list files response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (f *FileListResponse) SetRequest(req FileListRequest) {
	f.request = req
	f.Meta.ResolveCount(req.Filters.CountMode)
}

// Iterate returns the request set to the next page, if available. If there are
// no more pages, a nil request is returned.
func (f *FileListResponse) Iterate() *FileListRequest {
	if !f.Meta.Page.HasMore {
		return nil
	}
	req := f.request
	req.Filters.Page++
	return &req
}

// FileList retrieves multiple files using the provided request and returns the
// response.
func FileList(
	ctx context.Context,
	engine *twapi.Engine,
	req FileListRequest,
) (*FileListResponse, error) {
	return twapi.Execute[FileListRequest, *FileListResponse](ctx, engine, req)
}

// FileDownloadRequestPath contains the path parameters for downloading a file.
type FileDownloadRequestPath struct {
	// ID is the unique identifier of the file to be downloaded.
	ID int64
}

// FileDownloadRequest represents the request for downloading the content of a
// file. It resolves to the address File.DownloadURL reports, built from the
// session's server so the session can authenticate it.
//
// The route answers with a redirect to the storage service, which the HTTP
// client follows. The redirect target is signed, so the session's credentials
// are not needed there, and Go's client drops them on the change of host.
type FileDownloadRequest struct {
	// Path contains the path parameters for the request.
	Path FileDownloadRequestPath

	// Version is the number of the version to download. Zero downloads the
	// latest version.
	Version int64
}

// NewFileDownloadRequest creates a new FileDownloadRequest with the provided
// file ID. The ID is required to download a file.
func NewFileDownloadRequest(fileID int64) FileDownloadRequest {
	return FileDownloadRequest{
		Path: FileDownloadRequestPath{
			ID: fileID,
		},
	}
}

// HTTPRequest creates an HTTP request for the FileDownloadRequest.
func (f FileDownloadRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	if f.Path.ID <= 0 {
		return nil, fmt.Errorf("file download requires a file identifier")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server+"/", nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Set("action", "viewFile")
	query.Set("fileId", strconv.FormatInt(f.Path.ID, 10))
	querySetInt64(query, "v", f.Version)
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// FileDownloadResponse carries the content of a downloaded file. It is produced
// by FileDownload only, and deliberately does not implement
// twapi.HTTPResponser: twapi.Execute closes the response body before
// returning, which would leave Body unreadable.
type FileDownloadResponse struct {
	// Name is the file name the server suggests, taken from the
	// Content-Disposition header. It is empty when the header was not sent.
	Name string

	// ContentType is the media type of the content.
	ContentType string

	// Size is the number of bytes in Body, or -1 when the server did not report
	// it.
	Size int64

	// Body is the content of the file. The caller must close it.
	Body io.ReadCloser
}

// FileDownload downloads the content of a file using the provided request. The
// caller must close the returned Body.
func FileDownload(
	ctx context.Context,
	engine *twapi.Engine,
	req FileDownloadRequest,
) (*FileDownloadResponse, error) {
	resp, err := twapi.ExecuteRaw(ctx, engine, req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		defer func() { _ = resp.Body.Close() }()
		return nil, twapi.NewHTTPError(resp, "failed to download file")
	}

	var name string
	if _, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition")); err == nil {
		name = params["filename"]
	}

	return &FileDownloadResponse{
		Name:        name,
		ContentType: resp.Header.Get("Content-Type"),
		Size:        resp.ContentLength,
		Body:        resp.Body,
	}, nil
}
