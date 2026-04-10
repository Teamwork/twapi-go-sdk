package projects

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	twapi "github.com/teamwork/twapi-go-sdk"
)

var (
	_ twapi.HTTPRequester = (*MessageCreateRequest)(nil)
	_ twapi.HTTPResponser = (*MessageCreateResponse)(nil)
	_ twapi.HTTPRequester = (*MessageUpdateRequest)(nil)
	_ twapi.HTTPResponser = (*MessageUpdateResponse)(nil)
	_ twapi.HTTPRequester = (*MessageDeleteRequest)(nil)
	_ twapi.HTTPResponser = (*MessageDeleteResponse)(nil)
	_ twapi.HTTPRequester = (*MessageGetRequest)(nil)
	_ twapi.HTTPResponser = (*MessageGetResponse)(nil)
	_ twapi.HTTPRequester = (*MessageListRequest)(nil)
	_ twapi.HTTPResponser = (*MessageListResponse)(nil)
)

// MessageStatus defines the message status.
type MessageStatus string

const (
	// MessageStatusActive represents an active message.
	MessageStatusActive MessageStatus = "active"
	// MessageStatusArchived represents an archived message.
	MessageStatusArchived MessageStatus = "archived"
	// MessageStatusDeleted represents a deleted message.
	MessageStatusDeleted MessageStatus = "deleted"
)

// Message is a structured communication post within a project that allows team
// members to share updates, discuss topics, and document decisions in a
// centralized, threaded format. It includes a title, a detailed message body,
// and replies from collaborators, all tied to the project for clear context and
// visibility, making it ideal for important discussions that need to be
// organized and easily referenced over time.
//
// More information can be found at:
// https://support.teamwork.com/projects/getting-started/messages-overview
type Message struct {
	// ID is the unique identifier of the message.
	ID int64 `json:"id"`

	// Title is the title of the message, which provides a brief summary of the
	// content and purpose of the message, allowing team members to quickly
	// understand the topic at a glance.
	Title string `json:"title"`

	// Project is the project associated with the message.
	Project twapi.Relationship `json:"project"`

	// Tags is the list of tags associated with the message.
	Tags []twapi.Relationship `json:"tags"`

	// MessageReply is the embedded message reply, which contains the most recent
	// reply to the message.
	MessageReply

	// LastReply is the most recent reply to the message, providing a quick
	// snapshot of the latest discussion and updates related to the message,
	// helping team members stay informed about the current state of the
	// conversation without needing to read through all replies.
	LastReply *twapi.Relationship `json:"lastReply"`

	// LastReplyStatus indicates the status of the last reply, which can be
	// active, or deleted, providing insight into whether the most recent
	// contribution to the discussion is currently relevant and visible to the
	// team, or if it has been set aside for reference or removed from view.
	LastReplyStatus MessageReplyStatus `json:"replyStatus"`

	// Status is the current state of the message, which can be active, archived,
	// or deleted, indicating whether the message is currently relevant and
	// visible to the team, or if it has been set aside for reference or removed
	// from view.
	Status MessageStatus `json:"status"`
}

// MessageUpdateRequestPath contains the path parameters for creating a
// message.
type MessageCreateRequestPath struct {
	// ProjectID is the unique identifier of the project that will contain the
	// message.
	ProjectID int64
}

// messageNotifier is an interface that represents the different options for
// notifying users when a message is created or updated. It is implemented by
// MessageNotifyAll and MessageNotifyGroup.
type messageNotifier interface {
	messageNotifier()
}

// MessageNotifyAll is a message notifier that notifies all users from the
// project.
type MessageNotifyAll struct{}

// NewMessageNotifyAll creates a new MessageNotifyAll notifier.
func NewMessageNotifyAll() MessageNotifyAll {
	return MessageNotifyAll{}
}

// messageNotifier is a marker method to indicate that MessageNotifyAll
// implements the messageNotifier interface.
func (MessageNotifyAll) messageNotifier() {}

// MarshalJSON encodes the MessageNotifyAll as a string "ALL", which is the
// value expected by the API to notify all users.
func (MessageNotifyAll) MarshalJSON() ([]byte, error) {
	return []byte(`"ALL"`), nil
}

// MessageNotifyGroup is a message notifier that notifies a specific group of
// users, teams and/or companies.
type MessageNotifyGroup LegacyUserGroups

// NewMessageNotifyGroup creates a new MessageNotifyGroup notifier with the
// provided group of users, teams and/or companies.
func NewMessageNotifyGroup(groups LegacyUserGroups) MessageNotifyGroup {
	return MessageNotifyGroup(groups)
}

// messageNotifier is a marker method to indicate that MessageNotifyGroup
// implements the messageNotifier interface.
func (MessageNotifyGroup) messageNotifier() {}

// MarshalJSON encodes the MessageNotifyGroup as a string containing the encoded
// group of users, teams and/or companies, which is the value expected by the
// API to notify a specific group.
func (c MessageNotifyGroup) MarshalJSON() ([]byte, error) {
	return LegacyUserGroups(c).MarshalJSON()
}

// MessageCreateRequest represents the request body for creating a new
// message.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/messages/post-projects-id-posts-json
type MessageCreateRequest struct {
	// Path contains the path parameters for the request.
	Path MessageCreateRequestPath `json:"-"`

	// Title is the title of the message, which provides a brief summary of the
	// content and purpose of the message, allowing team members to quickly
	// understand the topic at a glance.
	Title string `json:"title"`

	// Body is the content of the message, which contains the actual text of the
	// message, allowing team members to share their thoughts, updates, or
	// questions in a structured format, while keeping the conversation organized
	// and easy to follow within the project context.
	Body string `json:"body"`

	// NotifyCurrentUser indicates whether the user creating the message should be
	// notified about the new message. If not provided, it defaults to false.
	NotifyCurrentUser *bool `json:"notify-current-user,omitempty"`

	// Notify is the message notifier that specifies who should be notified about
	// the new message. It can be a MessageNotifyAll, or MessageNotifyGroup. If
	// not provided, no notifications will be sent.
	Notify messageNotifier `json:"notify,omitempty"`
}

// NewMessageCreateRequest creates a new MessageCreateRequest with the provided
// required fields.
func NewMessageCreateRequest(
	projectID int64,
	title string,
	body string,
) MessageCreateRequest {
	return MessageCreateRequest{
		Path: MessageCreateRequestPath{
			ProjectID: projectID,
		},
		Title: title,
		Body:  body,
	}
}

// HTTPRequest creates an HTTP request for the MessageCreateRequest.
func (m MessageCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/projects/%d/messages.json", server, m.Path.ProjectID)

	payload := struct {
		Message MessageCreateRequest `json:"post"`
	}{Message: m}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode create message request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// MessageCreateResponse represents the response body for creating a new
// message.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/messages/post-projects-id-posts-json
type MessageCreateResponse struct {
	// ID is the unique identifier of the created message.
	ID LegacyNumber `json:"messageId"`
}

// HandleHTTPResponse handles the HTTP response for the MessageCreateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (m *MessageCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create message")
	}
	if err := json.NewDecoder(resp.Body).Decode(m); err != nil {
		return fmt.Errorf("failed to decode create message response: %w", err)
	}
	if m.ID == 0 {
		return fmt.Errorf("create message response does not contain a valid identifier")
	}
	return nil
}

// MessageCreate creates a new message using the provided request and returns
// the response.
func MessageCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req MessageCreateRequest,
) (*MessageCreateResponse, error) {
	return twapi.Execute[MessageCreateRequest, *MessageCreateResponse](ctx, engine, req)
}

// MessageUpdateRequestPath contains the path parameters for updating a message.
type MessageUpdateRequestPath struct {
	// ID is the unique identifier of the message to be updated.
	ID int64
}

// MessageUpdateRequest represents the request body for updating a message.
// Besides the identifier, all other fields are optional. When a field is not
// provided, it will not be modified.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/messages/put-posts-id-json
type MessageUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path MessageUpdateRequestPath `json:"-"`

	// Title is the title of the message, which provides a brief summary of the
	// content and purpose of the message, allowing team members to quickly
	// understand the topic at a glance.
	Title *string `json:"title,omitempty"`

	// Body is the content of the message, which contains the actual text of the
	// message, allowing team members to share their thoughts, updates, or
	// questions in a structured format, while keeping the conversation organized
	// and easy to follow within the project context.
	Body *string `json:"body,omitempty"`

	// NotifyCurrentUser indicates whether the user creating the message should be
	// notified about the new message. If not provided, it defaults to false.
	NotifyCurrentUser *bool `json:"notify-current-user,omitempty"`

	// Notify is the message notifier that specifies who should be notified about
	// the new message. It can be a MessageNotifyAll, or MessageNotifyGroup. If
	// not provided, no notifications will be sent.
	Notify messageNotifier `json:"notify,omitempty"`
}

// NewMessageUpdateRequest creates a new MessageUpdateRequest with the provided
// message ID. The ID is required to update a message.
func NewMessageUpdateRequest(messageID int64) MessageUpdateRequest {
	return MessageUpdateRequest{
		Path: MessageUpdateRequestPath{
			ID: messageID,
		},
	}
}

// HTTPRequest creates an HTTP request for the MessageUpdateRequest.
func (m MessageUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/messages/" + strconv.FormatInt(m.Path.ID, 10) + ".json"

	payload := struct {
		Message MessageUpdateRequest `json:"post"`
	}{Message: m}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update message request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// MessageUpdateResponse represents the response body for updating a message.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/messages/put-posts-id-json
type MessageUpdateResponse struct{}

// HandleHTTPResponse handles the HTTP response for the MessageUpdateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (m *MessageUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update message")
	}
	if err := json.NewDecoder(resp.Body).Decode(m); err != nil {
		return fmt.Errorf("failed to decode update message response: %w", err)
	}
	return nil
}

// MessageUpdate updates a message using the provided request and returns the
// response.
func MessageUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req MessageUpdateRequest,
) (*MessageUpdateResponse, error) {
	return twapi.Execute[MessageUpdateRequest, *MessageUpdateResponse](ctx, engine, req)
}

// MessageDeleteRequestPath contains the path parameters for deleting a message.
type MessageDeleteRequestPath struct {
	// ID is the unique identifier of the message to be deleted.
	ID int64
}

// MessageDeleteRequest represents the request body for deleting a message.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/messages/delete-posts-id-json
type MessageDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path MessageDeleteRequestPath
}

// NewMessageDeleteRequest creates a new MessageDeleteRequest with the
// provided message ID.
func NewMessageDeleteRequest(messageID int64) MessageDeleteRequest {
	return MessageDeleteRequest{
		Path: MessageDeleteRequestPath{
			ID: messageID,
		},
	}
}

// HTTPRequest creates an HTTP request for the MessageDeleteRequest.
func (m MessageDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/messages/" + strconv.FormatInt(m.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// MessageDeleteResponse represents the response body for deleting a message.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/messages/delete-posts-id-json
type MessageDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the MessageDeleteResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (m *MessageDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to delete message")
	}
	if err := json.NewDecoder(resp.Body).Decode(m); err != nil {
		return fmt.Errorf("failed to decode delete message response: %w", err)
	}
	return nil
}

// MessageDelete deletes a message using the provided request and returns the
// response.
func MessageDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req MessageDeleteRequest,
) (*MessageDeleteResponse, error) {
	return twapi.Execute[MessageDeleteRequest, *MessageDeleteResponse](ctx, engine, req)
}

// MessageGetRequestPath contains the path parameters for loading a single
// message.
type MessageGetRequestPath struct {
	// ID is the unique identifier of the message to be retrieved.
	ID int64 `json:"id"`
}

// MessageGetRequest represents the request body for loading a single message.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/messages/get-projects-api-v3-messages-message-id-json
type MessageGetRequest struct {
	// Path contains the path parameters for the request.
	Path MessageGetRequestPath
}

// NewMessageGetRequest creates a new MessageGetRequest with the provided
// message ID. The ID is required to load a message.
func NewMessageGetRequest(messageID int64) MessageGetRequest {
	return MessageGetRequest{
		Path: MessageGetRequestPath{
			ID: messageID,
		},
	}
}

// HTTPRequest creates an HTTP request for the MessageGetRequest.
func (m MessageGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/messages/" + strconv.FormatInt(m.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// MessageGetResponse contains all the information related to a message.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/messages/get-projects-api-v3-messages-message-id-json
type MessageGetResponse struct {
	Message Message `json:"message"`
}

// HandleHTTPResponse handles the HTTP response for the MessageGetResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (m *MessageGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve message")
	}

	if err := json.NewDecoder(resp.Body).Decode(m); err != nil {
		return fmt.Errorf("failed to decode retrieve message response: %w", err)
	}
	return nil
}

// MessageGet retrieves a single message using the provided request and returns
// the response.
func MessageGet(
	ctx context.Context,
	engine *twapi.Engine,
	req MessageGetRequest,
) (*MessageGetResponse, error) {
	return twapi.Execute[MessageGetRequest, *MessageGetResponse](ctx, engine, req)
}

// MessageListRequestFilters contains the filters for loading multiple
// messages.
type MessageListRequestFilters struct {
	// SearchTerm is an optional search term to filter messages by body content or
	// title.
	SearchTerm string

	// ProjectIDs is an optional list of project IDs to filter messages by
	// projects.
	ProjectIDs []int64

	// TagIDs is an optional list of tag IDs to filter messages by tags.
	TagIDs []int64

	// MatchAllTags is an optional flag to indicate if all tags must match. If set
	// to true, only messages matching all specified tags will be returned.
	MatchAllTags *bool

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of messages to retrieve per page. Defaults to 50.
	PageSize int64
}

// MessageListRequest represents the request body for loading multiple messages.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/messages/get-projects-api-v3-messages-json
type MessageListRequest struct {
	// Filters contains the filters for loading multiple messages.
	Filters MessageListRequestFilters
}

// NewMessageListRequest creates a new MessageListRequest with default values.
func NewMessageListRequest() MessageListRequest {
	return MessageListRequest{
		Filters: MessageListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the MessageListRequest.
func (m MessageListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/messages.json"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	query := req.URL.Query()
	if m.Filters.SearchTerm != "" {
		query.Set("searchTerm", m.Filters.SearchTerm)
	}
	if len(m.Filters.ProjectIDs) > 0 {
		projectIDs := make([]string, len(m.Filters.ProjectIDs))
		for i, id := range m.Filters.ProjectIDs {
			projectIDs[i] = strconv.FormatInt(id, 10)
		}
		query.Set("projectIds", strings.Join(projectIDs, ","))
	}
	if len(m.Filters.TagIDs) > 0 {
		tagIDs := make([]string, len(m.Filters.TagIDs))
		for i, id := range m.Filters.TagIDs {
			tagIDs[i] = strconv.FormatInt(id, 10)
		}
		query.Set("tagIds", strings.Join(tagIDs, ","))
	}
	if m.Filters.MatchAllTags != nil {
		query.Set("matchAllTags", strconv.FormatBool(*m.Filters.MatchAllTags))
	}
	if m.Filters.Page > 0 {
		query.Set("page", strconv.FormatInt(m.Filters.Page, 10))
	}
	if m.Filters.PageSize > 0 {
		query.Set("pageSize", strconv.FormatInt(m.Filters.PageSize, 10))
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}

// MessageListResponse contains information by multiple messages matching the
// request filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/messages/get-projects-api-v3-messages-json
type MessageListResponse struct {
	request MessageListRequest

	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`
	Messages []Message `json:"messages"`
}

// HandleHTTPResponse handles the HTTP response for the MessageListResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (m *MessageListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list messages")
	}

	if err := json.NewDecoder(resp.Body).Decode(m); err != nil {
		return fmt.Errorf("failed to decode list messages response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (m *MessageListResponse) SetRequest(req MessageListRequest) {
	m.request = req
}

// Iterate returns the request set to the next page, if available. If there are
// no more pages, a nil request is returned.
func (m *MessageListResponse) Iterate() *MessageListRequest {
	if !m.Meta.Page.HasMore {
		return nil
	}
	req := m.request
	req.Filters.Page++
	return &req
}

// MessageList retrieves multiple messages using the provided request and
// returns the response.
func MessageList(
	ctx context.Context,
	engine *twapi.Engine,
	req MessageListRequest,
) (*MessageListResponse, error) {
	return twapi.Execute[MessageListRequest, *MessageListResponse](ctx, engine, req)
}
