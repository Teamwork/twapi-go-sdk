package projects

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var (
	_ twapi.HTTPRequester = (*CommentCreateRequest)(nil)
	_ twapi.HTTPResponser = (*CommentCreateResponse)(nil)
	_ twapi.HTTPRequester = (*CommentUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*CommentUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*CommentDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*CommentDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*CommentGetRequest)(nil)
	_ twapi.HTTPResponser = (*CommentGetResponse)(nil)
	_ twapi.HTTPRequester = (*CommentListRequest)(nil)
	_ twapi.HTTPResponser = (*CommentListResponse)(nil)
)

// Comment is a way for users to communicate and collaborate directly within
// tasks, milestones, files, or other project items. Comments allow team members
// to provide updates, ask questions, give feedback, or share relevant
// information in a centralized and contextual manner. They support rich text
// formatting, file attachments, and @mentions to notify specific users or
// teams, helping keep discussions organized and easily accessible within the
// project. Comments are visible to all users with access to the item, promoting
// transparency and keeping everyone aligned.
//
// More information can be found at:
// https://support.teamwork.com/projects/getting-started/comments-overview
type Comment struct {
	// ID is the unique identifier of the comment.
	ID int64 `json:"id"`

	// Body is the body of the comment.
	Body string `json:"body"`

	// HTMLBody is the HTML representation of the comment body.
	HTMLBody string `json:"htmlBody"`

	// ContentType is the content type of the comment body. It can be "TEXT" or
	// "HTML".
	ContentType string `json:"contentType"`

	// Object is the relationship to the object (task, milestone, project) that
	// this comment is associated with.
	Object *twapi.Relationship `json:"object"`

	// Project is the relationship to the project that this comment belongs to.
	Project twapi.Relationship `json:"project"`

	// PostedBy is the ID of the user who posted the comment.
	PostedBy *int64 `json:"postedBy"`

	// PostedAt is the date and time when the comment was posted.
	PostedAt *time.Time `json:"postedDateTime"`

	// LastEditedBy is the ID of the user who last edited the comment, if it was
	// edited.
	LastEditedBy *int64 `json:"lastEditedBy"`

	// EditedAt is the date and time when the comment was last edited, if it was
	// edited.
	EditedAt *time.Time `json:"dateLastEdited"`

	// Deleted indicates whether the comment has been deleted.
	Deleted bool `json:"deleted"`

	// DeletedBy is the ID of the user who deleted the comment, if it was deleted.
	DeletedBy *int64 `json:"deletedBy"`

	// DeletedAt is the date and time when the comment was deleted, if it was
	// deleted.
	DeletedAt *time.Time `json:"dateDeleted"`
}

// CommentUpdateRequestPath contains the path parameters for creating a
// comment.
type CommentCreateRequestPath struct {
	// FileVersionID is the unique identifier of the file version where the
	// comment will be created. Each file can have multiple versions, and the
	// comments are associated with a specific version.
	FileVersionID int64

	// MilestoneID is the unique identifier of the milestone where the comment
	// will be created.
	MilestoneID int64

	// NotebookID is the unique identifier of the notebook where the comment will
	// be created.
	NotebookID int64

	// TaskID is the unique identifier of the task where the comment will be
	// created.
	TaskID int64

	// LinkID is the unique identifier of the link where the comment will be
	// created.
	LinkID int64
}

// CommentCreateRequest represents the request body for creating a new
// comment.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/comments/post-resource-resource-id-comments-json
type CommentCreateRequest struct {
	// Path contains the path parameters for the request.
	Path CommentCreateRequestPath `json:"-"`

	// Body is the body of the comment. It can contain plain text or HTML.
	Body string `json:"body"`

	// ContentType is the content type of the comment body. It can be "TEXT" or
	// "HTML". If not provided, it defaults to "TEXT".
	ContentType *string `json:"contentType,omitempty"`
}

// NewCommentCreateRequestInFileVersion creates a new CommentCreateRequest with
// the provided file version ID.
func NewCommentCreateRequestInFileVersion(fileVersionID int64, body string) CommentCreateRequest {
	return CommentCreateRequest{
		Path: CommentCreateRequestPath{
			FileVersionID: fileVersionID,
		},
		Body: body,
	}
}

// NewCommentCreateRequestInMilestone creates a new CommentCreateRequest with
// the provided milestone ID.
func NewCommentCreateRequestInMilestone(milestoneID int64, body string) CommentCreateRequest {
	return CommentCreateRequest{
		Path: CommentCreateRequestPath{
			MilestoneID: milestoneID,
		},
		Body: body,
	}
}

// NewCommentCreateRequestInNotebook creates a new CommentCreateRequest with
// the provided notebook ID.
func NewCommentCreateRequestInNotebook(notebookID int64, body string) CommentCreateRequest {
	return CommentCreateRequest{
		Path: CommentCreateRequestPath{
			NotebookID: notebookID,
		},
		Body: body,
	}
}

// NewCommentCreateRequestInTask creates a new CommentCreateRequest with the
// provided task ID.
func NewCommentCreateRequestInTask(taskID int64, body string) CommentCreateRequest {
	return CommentCreateRequest{
		Path: CommentCreateRequestPath{
			TaskID: taskID,
		},
		Body: body,
	}
}

// NewCommentCreateRequestInLink creates a new CommentCreateRequest with the
// provided link ID.
func NewCommentCreateRequestInLink(linkID int64, body string) CommentCreateRequest {
	return CommentCreateRequest{
		Path: CommentCreateRequestPath{
			LinkID: linkID,
		},
		Body: body,
	}
}

// HTTPRequest creates an HTTP request for the CommentCreateRequest.
func (t CommentCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	var uri string
	switch {
	case t.Path.FileVersionID > 0:
		uri = fmt.Sprintf("%s/fileversions/%d/comments.json", server, t.Path.FileVersionID)
	case t.Path.MilestoneID > 0:
		uri = fmt.Sprintf("%s/milestones/%d/comments.json", server, t.Path.MilestoneID)
	case t.Path.NotebookID > 0:
		uri = fmt.Sprintf("%s/notebooks/%d/comments.json", server, t.Path.NotebookID)
	case t.Path.TaskID > 0:
		uri = fmt.Sprintf("%s/tasks/%d/comments.json", server, t.Path.TaskID)
	case t.Path.LinkID > 0:
		uri = fmt.Sprintf("%s/links/%d/comments.json", server, t.Path.LinkID)
	default:
		return nil, fmt.Errorf("no valid path provided for creating comment")
	}

	payload := struct {
		Comment CommentCreateRequest `json:"comment"`
	}{Comment: t}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create comment request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// CommentCreateResponse represents the response body for creating a new
// comment.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/comments/post-resource-resource-id-comments-json
type CommentCreateResponse struct {
	// ID is the unique identifier of the created comment.
	ID LegacyNumber `json:"id"`
}

// HandleHTTPResponse handles the HTTP response for the CommentCreateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (t *CommentCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create comment")
	}
	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode create comment response: %w", err)
	}
	if t.ID == 0 {
		return fmt.Errorf("create comment response does not contain a valid identifier")
	}
	return nil
}

// CommentCreate creates a new comment using the provided request and returns
// the response.
func CommentCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req CommentCreateRequest,
) (*CommentCreateResponse, error) {
	return twapi.Execute[CommentCreateRequest, *CommentCreateResponse](ctx, engine, req)
}

// CommentUpdateRequestPath contains the path parameters for updating a comment.
type CommentUpdateRequestPath struct {
	// ID is the unique identifier of the comment to be updated.
	ID int64
}

// CommentUpdateRequest represents the request body for updating a comment.
// Besides the identifier, all other fields are optional. When a field is not
// provided, it will not be modified.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/comments/put-comments-id-json
type CommentUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path CommentUpdateRequestPath `json:"-"`

	// Body is the body of the comment. It can contain plain text or HTML.
	Body string `json:"body"`

	// ContentType is the content type of the comment body. It can be "TEXT" or
	// "HTML". If not provided, it defaults to "TEXT".
	ContentType *string `json:"contentType,omitempty"`
}

// NewCommentUpdateRequest creates a new CommentUpdateRequest with the
// provided comment ID. The ID is required to update a comment.
func NewCommentUpdateRequest(commentID int64) CommentUpdateRequest {
	return CommentUpdateRequest{
		Path: CommentUpdateRequestPath{
			ID: commentID,
		},
	}
}

// HTTPRequest creates an HTTP request for the CommentUpdateRequest.
func (t CommentUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/comments/" + strconv.FormatInt(t.Path.ID, 10) + ".json"

	payload := struct {
		Comment CommentUpdateRequest `json:"comment"`
	}{Comment: t}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update comment request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// CommentUpdateResponse represents the response body for updating a comment.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/comments/put-comments-id-json
type CommentUpdateResponse struct{}

// HandleHTTPResponse handles the HTTP response for the CommentUpdateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (t *CommentUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update comment")
	}
	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode update comment response: %w", err)
	}
	return nil
}

// CommentUpdate updates a comment using the provided request and returns the
// response.
func CommentUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req CommentUpdateRequest,
) (*CommentUpdateResponse, error) {
	return twapi.Execute[CommentUpdateRequest, *CommentUpdateResponse](ctx, engine, req)
}

// CommentDeleteRequestPath contains the path parameters for deleting a comment.
type CommentDeleteRequestPath struct {
	// ID is the unique identifier of the comment to be deleted.
	ID int64
}

// CommentDeleteRequest represents the request body for deleting a comment.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/comments/delete-comments-id-json
type CommentDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path CommentDeleteRequestPath
}

// NewCommentDeleteRequest creates a new CommentDeleteRequest with the
// provided comment ID.
func NewCommentDeleteRequest(commentID int64) CommentDeleteRequest {
	return CommentDeleteRequest{
		Path: CommentDeleteRequestPath{
			ID: commentID,
		},
	}
}

// HTTPRequest creates an HTTP request for the CommentDeleteRequest.
func (t CommentDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/comments/" + strconv.FormatInt(t.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// CommentDeleteResponse represents the response body for deleting a comment.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/comments/delete-comments-id-json
type CommentDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the CommentDeleteResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (t *CommentDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to delete comment")
	}
	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode delete comment response: %w", err)
	}
	return nil
}

// CommentDelete deletes a comment using the provided request and returns the
// response.
func CommentDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req CommentDeleteRequest,
) (*CommentDeleteResponse, error) {
	return twapi.Execute[CommentDeleteRequest, *CommentDeleteResponse](ctx, engine, req)
}

// CommentGetRequestPath contains the path parameters for loading a single
// comment.
type CommentGetRequestPath struct {
	// ID is the unique identifier of the comment to be retrieved.
	ID int64 `json:"id"`
}

// CommentGetRequest represents the request body for loading a single comment.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/comments/get-projects-api-v3-comments-id-json
type CommentGetRequest struct {
	// Path contains the path parameters for the request.
	Path CommentGetRequestPath
}

// NewCommentGetRequest creates a new CommentGetRequest with the provided
// comment ID. The ID is required to load a comment.
func NewCommentGetRequest(commentID int64) CommentGetRequest {
	return CommentGetRequest{
		Path: CommentGetRequestPath{
			ID: commentID,
		},
	}
}

// HTTPRequest creates an HTTP request for the CommentGetRequest.
func (t CommentGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/comments/" + strconv.FormatInt(t.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// CommentGetResponse contains all the information related to a comment.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/comments/get-projects-api-v3-comments-id-json
type CommentGetResponse struct {
	Comment Comment `json:"comments"`
}

// HandleHTTPResponse handles the HTTP response for the CommentGetResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (t *CommentGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve comment")
	}

	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode retrieve comment response: %w", err)
	}
	return nil
}

// CommentGet retrieves a single comment using the provided request and returns
// the response.
func CommentGet(
	ctx context.Context,
	engine *twapi.Engine,
	req CommentGetRequest,
) (*CommentGetResponse, error) {
	return twapi.Execute[CommentGetRequest, *CommentGetResponse](ctx, engine, req)
}

// CommentListRequestPath contains the path parameters for loading multiple
// comments.
type CommentListRequestPath struct {
	// FileVersionID is the unique identifier of the file version whose comments
	// are to be retrieved. Each file can have multiple versions, and the comments
	// are associated with a specific version.
	FileVersionID int64

	// MilestoneID is the unique identifier of the milestone whose comments are to
	// be retrieved.
	MilestoneID int64

	// NotebookID is the unique identifier of the notebook whose comments are to
	// be retrieved.
	NotebookID int64

	// TaskID is the unique identifier of the task whose comments are to be
	// retrieved.
	TaskID int64
}

// CommentListRequestFilters contains the filters for loading multiple comments.
type CommentListRequestFilters struct {
	// SearchTerm is an optional search term to filter comments by name, description
	// or commentlist's name.
	SearchTerm string

	// UserIDs is an optional list of user IDs to filter comments by users.
	UserIDs []int64

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of comments to retrieve per page. Defaults to 50.
	PageSize int64
}

// CommentListRequest represents the request body for loading multiple comments.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/comments/get-projects-api-v3-comments-json
// https://apidocs.teamwork.com/docs/teamwork/v3/file-version-comments/get-projects-api-v3-fileversions-id-comments-json
// https://apidocs.teamwork.com/docs/teamwork/v3/milestone-comments/get-projects-api-v3-milestones-milestone-id-comments-json
// https://apidocs.teamwork.com/docs/teamwork/v3/notebook-comments/get-projects-api-v3-notebooks-notebook-id-comments-json
// https://apidocs.teamwork.com/docs/teamwork/v3/task-comments/get-projects-api-v3-tasks-task-id-comments-json
//
//nolint:lll
type CommentListRequest struct {
	// Path contains the path parameters for the request.
	Path CommentListRequestPath

	// Filters contains the filters for loading multiple comments.
	Filters CommentListRequestFilters
}

// NewCommentListRequest creates a new CommentListRequest with default values.
func NewCommentListRequest() CommentListRequest {
	return CommentListRequest{
		Filters: CommentListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the CommentListRequest.
func (t CommentListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	var uri string
	switch {
	case t.Path.FileVersionID > 0:
		uri = fmt.Sprintf("%s/projects/api/v3/fileversions/%d/comments.json", server, t.Path.FileVersionID)
	case t.Path.MilestoneID > 0:
		uri = fmt.Sprintf("%s/projects/api/v3/milestones/%d/comments.json", server, t.Path.MilestoneID)
	case t.Path.NotebookID > 0:
		uri = fmt.Sprintf("%s/projects/api/v3/notebooks/%d/comments.json", server, t.Path.NotebookID)
	case t.Path.TaskID > 0:
		uri = fmt.Sprintf("%s/projects/api/v3/tasks/%d/comments.json", server, t.Path.TaskID)
	default:
		uri = server + "/projects/api/v3/comments.json"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	if t.Filters.SearchTerm != "" {
		query.Set("searchTerm", t.Filters.SearchTerm)
	}
	if len(t.Filters.UserIDs) > 0 {
		tagIDs := make([]string, len(t.Filters.UserIDs))
		for i, id := range t.Filters.UserIDs {
			tagIDs[i] = strconv.FormatInt(id, 10)
		}
		query.Set("userIds", strings.Join(tagIDs, ","))
	}
	if t.Filters.Page > 0 {
		query.Set("page", strconv.FormatInt(t.Filters.Page, 10))
	}
	if t.Filters.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(t.Filters.PageSize, 10))
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// CommentListResponse contains information by multiple comments matching the request
// filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/file-version-comments/get-projects-api-v3-fileversions-id-comments-json
// https://apidocs.teamwork.com/docs/teamwork/v3/milestone-comments/get-projects-api-v3-milestones-milestone-id-comments-json
// https://apidocs.teamwork.com/docs/teamwork/v3/notebook-comments/get-projects-api-v3-notebooks-notebook-id-comments-json
// https://apidocs.teamwork.com/docs/teamwork/v3/task-comments/get-projects-api-v3-tasks-task-id-comments-json
//
//nolint:lll
type CommentListResponse struct {
	request CommentListRequest

	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`
	Comments []Comment `json:"comments"`
}

// HandleHTTPResponse handles the HTTP response for the CommentListResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (t *CommentListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list comments")
	}

	if err := json.NewDecoder(resp.Body).Decode(t); err != nil {
		return fmt.Errorf("failed to decode list comments response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (t *CommentListResponse) SetRequest(req CommentListRequest) {
	t.request = req
}

// Iterate returns the request set to the next page, if available. If there
// are no more pages, a nil request is returned.
func (t *CommentListResponse) Iterate() *CommentListRequest {
	if !t.Meta.Page.HasMore {
		return nil
	}
	req := t.request
	req.Filters.Page++
	return &req
}

// CommentList retrieves multiple comments using the provided request and
// returns the response.
func CommentList(
	ctx context.Context,
	engine *twapi.Engine,
	req CommentListRequest,
) (*CommentListResponse, error) {
	return twapi.Execute[CommentListRequest, *CommentListResponse](ctx, engine, req)
}
