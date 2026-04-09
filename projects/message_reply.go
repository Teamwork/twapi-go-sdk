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

// MessageReplyMeta contains additional information about a message reply, such
// as whether the reply has been read and the permissions of the current user
// regarding the reply (e.g., whether they can edit it). This metadata helps
// provide context for how the reply can be interacted with and its current
// state in the conversation.
type MessageReplyMeta struct {
	// IsRead indicates whether the message reply has been read by the current
	// user. If true, it means the user has already seen the reply; if false, it
	// indicates that the reply is new or unread, which can help users identify
	// which parts of the conversation they may need to catch up on.
	IsRead bool `json:"isRead"`

	// Permissions contains the permissions of the current user regarding the
	// message reply, such as whether they can edit it. This information is
	// crucial for determining what actions the user can take on the reply,
	// ensuring that they have the appropriate access to modify or interact with
	// the content based on their role and permissions within the project.
	Permissions struct {
		// CanEdit indicates whether the current user has permission to edit the
		// message reply. If true, the user can modify the content of the reply; if
		// false, they do not have editing rights, which helps maintain the
		// integrity of the conversation and ensures that only authorized users can
		// make changes to the replies.
		CanEdit bool `json:"canEdit"`
	} `json:"permissions"`
}

// MessageReplyStatus defines the message reply status.
type MessageReplyStatus string

const (
	// MessageReplyStatusActive represents an active message reply.
	MessageReplyStatusActive MessageReplyStatus = "active"

	// MessageReplyStatusArchived represents an archived message reply.
	MessageReplyStatusDeleted MessageReplyStatus = "deleted"
)

// MessageReply is a response within a project message thread that allows team
// members to contribute to the discussion, ask questions, or provide updates
// while keeping all communication organized under the original message. Replies
// maintain context by staying linked to the main topic, include the author and
// timestamp, and help create a clear, ongoing conversation that is easy for
// everyone involved to follow and reference.
type MessageReply struct {
	// ID is the unique identifier of the message reply.
	ID int64 `json:"id"`

	// Body is the content of the message reply, which contains the actual text of
	// the reply, allowing team members to share their thoughts, updates, or
	// questions in response to the original message, while keeping the
	// conversation organized and easy to follow within the project context.
	Body string `json:"body"`

	// Author is the relationship to the author of the message reply, which
	// provides information about who contributed the reply, helping to attribute
	// comments and updates to specific team members, and fostering accountability
	// and clear communication within the project.
	Author twapi.Relationship `json:"author"`

	// Message is the relationship to the message that this reply belongs to,
	// which maintains the connection between the reply and the original message,
	// ensuring that all contributions to the discussion are properly linked and
	// organized under the main topic for easy reference and context within the
	// project.
	Message twapi.Relationship `json:"message"`

	// Meta contains additional information about the message reply, such as
	// whether the reply has been read and the permissions of the current user
	// regarding the reply (e.g., whether they can edit it). This metadata helps
	// provide context for how the reply can be interacted with and its current
	// state in the conversation.
	Meta *MessageReplyMeta `json:"meta,omitempty"`

	// CreatedAt is the timestamp of when the message reply was created, which
	// provides a chronological context for the discussion, allowing team members
	// to understand the sequence of contributions and track the evolution of the
	// conversation over time within the project.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt is the timestamp of when the message reply was last updated,
	// which indicates if and when the content of the reply has been modified
	// after its initial creation, helping team members stay informed about any
	// changes to the discussion and ensuring that they are aware of the most
	// current information within the project conversation.
	UpdatedAt *time.Time `json:"updatedAt"`

	// Status is the current state of the message reply, which can be active or
	// deleted, indicating whether the reply is currently relevant and visible to
	// the team, or if it has been set aside for reference or removed from view.
	Status MessageReplyStatus `json:"status"`
}

// MessageReplyUpdateRequestPath contains the path parameters for creating a
// message.
type MessageReplyCreateRequestPath struct {
	// MessageID is the unique identifier of the message that will contain the
	// reply.
	MessageID int64
}

// MessageReplyCreateRequest represents the request body for creating a new
// message reply.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/message-replies/post-messages-id-message-replies-json
type MessageReplyCreateRequest struct {
	// Path contains the path parameters for the request.
	Path MessageReplyCreateRequestPath `json:"-"`

	// Body is the content of the message reply, which contains the actual text of
	// the reply, allowing team members to share their thoughts, updates, or
	// questions in a structured format, while keeping the conversation organized
	// and easy to follow within the project context.
	Body string `json:"body"`

	// NotifyCurrentUser indicates whether the user creating the message reply
	// should be notified about the new message reply. If not provided, it
	// defaults to false.
	NotifyCurrentUser *bool `json:"notify-current-user,omitempty"`

	// Notify is the message notifier that specifies who should be notified about
	// the new message reply. It can be a MessageReplyNotifyAll, or
	// MessageReplyNotifyGroup. If not provided, no notifications will be sent.
	Notify messageNotifier `json:"notify,omitempty"`
}

// NewMessageReplyCreateRequest creates a new MessageReplyCreateRequest with the provided
// required fields.
func NewMessageReplyCreateRequest(messageID int64, body string) MessageReplyCreateRequest {
	return MessageReplyCreateRequest{
		Path: MessageReplyCreateRequestPath{
			MessageID: messageID,
		},
		Body: body,
	}
}

// HTTPRequest creates an HTTP request for the MessageReplyCreateRequest.
func (m MessageReplyCreateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := fmt.Sprintf("%s/messages/%d/replies.json", server, m.Path.MessageID)

	payload := struct {
		MessageReply MessageReplyCreateRequest `json:"messagereply"`
	}{MessageReply: m}

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

// MessageReplyCreateResponse represents the response body for creating a new
// message reply.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/message-replies/post-messages-id-message-replies-json
type MessageReplyCreateResponse struct {
	// ID is the unique identifier of the created message reply.
	ID LegacyNumber `json:"postId"`
}

// HandleHTTPResponse handles the HTTP response for the
// MessageReplyCreateResponse. If some unexpected HTTP status code is returned
// by the API, a twapi.HTTPError is returned.
func (m *MessageReplyCreateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusCreated {
		return twapi.NewHTTPError(resp, "failed to create message reply")
	}
	if err := json.NewDecoder(resp.Body).Decode(m); err != nil {
		return fmt.Errorf("failed to decode create message reply response: %w", err)
	}
	if m.ID == 0 {
		return fmt.Errorf("create message reply response does not contain a valid identifier")
	}
	return nil
}

// MessageReplyCreate creates a new message reply using the provided request and
// returns the response.
func MessageReplyCreate(
	ctx context.Context,
	engine *twapi.Engine,
	req MessageReplyCreateRequest,
) (*MessageReplyCreateResponse, error) {
	return twapi.Execute[MessageReplyCreateRequest, *MessageReplyCreateResponse](ctx, engine, req)
}

// MessageReplyUpdateRequestPath contains the path parameters for updating a
// message reply.
type MessageReplyUpdateRequestPath struct {
	// ID is the unique identifier of the message reply to be updated.
	ID int64
}

// MessageReplyUpdateRequest represents the request body for updating a message
// reply. Besides the identifier, all other fields are optional. When a field is
// not provided, it will not be modified.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/message-replies/put-message-replies-id-json
type MessageReplyUpdateRequest struct {
	// Path contains the path parameters for the request.
	Path MessageReplyUpdateRequestPath `json:"-"`

	// Body is the content of the message reply, which contains the actual text of
	// the message reply, allowing team members to share their thoughts, updates,
	// or questions in a structured format, while keeping the conversation
	// organized and easy to follow within the project context.
	Body *string `json:"body,omitempty"`

	// NotifyCurrentUser indicates whether the user creating the message reply
	// should be notified about the new message reply. If not provided, it defaults to
	// false.
	NotifyCurrentUser *bool `json:"notify-current-user,omitempty"`

	// Notify is the message notifier that specifies who should be notified about
	// the new message reply. It can be a MessageReplyNotifyAll, or
	// MessageReplyNotifyGroup. If not provided, no notifications will be sent.
	Notify messageNotifier `json:"notify,omitempty"`
}

// NewMessageReplyUpdateRequest creates a new MessageReplyUpdateRequest with the
// provided message reply ID. The ID is required to update a message reply.
func NewMessageReplyUpdateRequest(messageID int64) MessageReplyUpdateRequest {
	return MessageReplyUpdateRequest{
		Path: MessageReplyUpdateRequestPath{
			ID: messageID,
		},
	}
}

// HTTPRequest creates an HTTP request for the MessageReplyUpdateRequest.
func (m MessageReplyUpdateRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/messageReplies/" + strconv.FormatInt(m.Path.ID, 10) + ".json"

	payload := struct {
		MessageReply MessageReplyUpdateRequest `json:"messagereply"`
	}{MessageReply: m}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, fmt.Errorf("failed to encode update message reply request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// MessageReplyUpdateResponse represents the response body for updating a
// message reply.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/message-replies/put-message-replies-id-json
type MessageReplyUpdateResponse struct{}

// HandleHTTPResponse handles the HTTP response for the MessageReplyUpdateResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (m *MessageReplyUpdateResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to update message reply")
	}
	if err := json.NewDecoder(resp.Body).Decode(m); err != nil {
		return fmt.Errorf("failed to decode update message reply response: %w", err)
	}
	return nil
}

// MessageReplyUpdate updates a message reply using the provided request and
// returns the response.
func MessageReplyUpdate(
	ctx context.Context,
	engine *twapi.Engine,
	req MessageReplyUpdateRequest,
) (*MessageReplyUpdateResponse, error) {
	return twapi.Execute[MessageReplyUpdateRequest, *MessageReplyUpdateResponse](ctx, engine, req)
}

// MessageReplyDeleteRequestPath contains the path parameters for deleting a message reply.
type MessageReplyDeleteRequestPath struct {
	// ID is the unique identifier of the message reply to be deleted.
	ID int64
}

// MessageReplyDeleteRequest represents the request body for deleting a message
// reply.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/message-replies/delete-message-replies-id-json
type MessageReplyDeleteRequest struct {
	// Path contains the path parameters for the request.
	Path MessageReplyDeleteRequestPath
}

// NewMessageReplyDeleteRequest creates a new MessageReplyDeleteRequest with the
// provided message reply ID.
func NewMessageReplyDeleteRequest(messageID int64) MessageReplyDeleteRequest {
	return MessageReplyDeleteRequest{
		Path: MessageReplyDeleteRequestPath{
			ID: messageID,
		},
	}
}

// HTTPRequest creates an HTTP request for the MessageReplyDeleteRequest.
func (m MessageReplyDeleteRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/messageReplies/" + strconv.FormatInt(m.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// MessageReplyDeleteResponse represents the response body for deleting a message reply.
//
// https://apidocs.teamwork.com/docs/teamwork/v1/message-replies/delete-message-replies-id-json
type MessageReplyDeleteResponse struct{}

// HandleHTTPResponse handles the HTTP response for the MessageReplyDeleteResponse.
// If some unexpected HTTP status code is returned by the API, a twapi.HTTPError
// is returned.
func (m *MessageReplyDeleteResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to delete message reply")
	}
	if err := json.NewDecoder(resp.Body).Decode(m); err != nil {
		return fmt.Errorf("failed to decode delete message reply response: %w", err)
	}
	return nil
}

// MessageReplyDelete deletes a message reply using the provided request and
// returns the response.
func MessageReplyDelete(
	ctx context.Context,
	engine *twapi.Engine,
	req MessageReplyDeleteRequest,
) (*MessageReplyDeleteResponse, error) {
	return twapi.Execute[MessageReplyDeleteRequest, *MessageReplyDeleteResponse](ctx, engine, req)
}

// MessageReplyGetRequestPath contains the path parameters for loading a single
// message reply.
type MessageReplyGetRequestPath struct {
	// ID is the unique identifier of the message reply to be retrieved.
	ID int64 `json:"id"`
}

// MessageReplyGetRequest represents the request body for loading a single
// message reply.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/message-replies/get-projects-api-v3-message-replies-message-id-json
type MessageReplyGetRequest struct {
	// Path contains the path parameters for the request.
	Path MessageReplyGetRequestPath
}

// NewMessageReplyGetRequest creates a new MessageReplyGetRequest with the provided
// message reply ID. The ID is required to load a message reply.
func NewMessageReplyGetRequest(messageReplyID int64) MessageReplyGetRequest {
	return MessageReplyGetRequest{
		Path: MessageReplyGetRequestPath{
			ID: messageReplyID,
		},
	}
}

// HTTPRequest creates an HTTP request for the MessageReplyGetRequest.
func (m MessageReplyGetRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	uri := server + "/projects/api/v3/messagereplies/" + strconv.FormatInt(m.Path.ID, 10) + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// MessageReplyGetResponse contains all the information related to a message
// reply.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/message-replies/get-projects-api-v3-message-replies-message-id-json
type MessageReplyGetResponse struct {
	MessageReply MessageReply `json:"messageReply"`
}

// HandleHTTPResponse handles the HTTP response for the MessageReplyGetResponse. If
// some unexpected HTTP status code is returned by the API, a twapi.HTTPError is
// returned.
func (m *MessageReplyGetResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to retrieve message reply")
	}

	if err := json.NewDecoder(resp.Body).Decode(m); err != nil {
		return fmt.Errorf("failed to decode retrieve message reply response: %w", err)
	}
	return nil
}

// MessageReplyGet retrieves a single message reply using the provided request
// and returns the response.
func MessageReplyGet(
	ctx context.Context,
	engine *twapi.Engine,
	req MessageReplyGetRequest,
) (*MessageReplyGetResponse, error) {
	return twapi.Execute[MessageReplyGetRequest, *MessageReplyGetResponse](ctx, engine, req)
}

// MessageReplyListRequestPath contains the path parameters for loading multiple
// message replies.
type MessageReplyListRequestPath struct {
	// MessageID is the unique identifier of the message whose replies are to be
	// retrieved.
	MessageID int64
}

// MessageReplyListRequestFilters contains the filters for loading multiple
// message replies.
type MessageReplyListRequestFilters struct {
	// MessageIDs is an optional list of message IDs to filter message replies by
	// their parent messages.
	MessageIDs []int64

	// ProjectIDs is an optional list of project IDs to filter message replies by
	// their parent projects.
	ProjectIDs []int64

	// Page is the page number to retrieve. Defaults to 1.
	Page int64

	// PageSize is the number of messages to retrieve per page. Defaults to 50.
	PageSize int64
}

// MessageReplyListRequest represents the request body for loading multiple
// message replies.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/message-replies/get-projects-api-v3-message-replies-json
// https://apidocs.teamwork.com/docs/teamwork/v3/message-replies/get-projects-api-v3-messages-message-id-replies-json
type MessageReplyListRequest struct {
	// Path contains the path parameters for the request.
	Path MessageReplyListRequestPath

	// Filters contains the filters for loading multiple message replies.
	Filters MessageReplyListRequestFilters
}

// NewMessageReplyListRequest creates a new MessageReplyListRequest with default values.
func NewMessageReplyListRequest() MessageReplyListRequest {
	return MessageReplyListRequest{
		Filters: MessageReplyListRequestFilters{
			Page:     1,
			PageSize: 50,
		},
	}
}

// HTTPRequest creates an HTTP request for the MessageReplyListRequest.
func (m MessageReplyListRequest) HTTPRequest(ctx context.Context, server string) (*http.Request, error) {
	var uri string
	switch {
	case m.Path.MessageID > 0:
		uri = fmt.Sprintf("%s/messages/%d/replies.json", server, m.Path.MessageID)
	default:
		uri = server + "/projects/api/v3/messagereplies.json"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	query := req.URL.Query()
	if len(m.Filters.MessageIDs) > 0 {
		messageIDs := make([]string, len(m.Filters.MessageIDs))
		for i, id := range m.Filters.MessageIDs {
			messageIDs[i] = strconv.FormatInt(id, 10)
		}
		query.Set("messageIds", strings.Join(messageIDs, ","))
	}
	if len(m.Filters.ProjectIDs) > 0 {
		projectIDs := make([]string, len(m.Filters.ProjectIDs))
		for i, id := range m.Filters.ProjectIDs {
			projectIDs[i] = strconv.FormatInt(id, 10)
		}
		query.Set("projectIds", strings.Join(projectIDs, ","))
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

// MessageReplyListResponse contains information by multiple message replies
// matching the request filters.
//
// https://apidocs.teamwork.com/docs/teamwork/v3/message-replies/get-projects-api-v3-message-replies-json
type MessageReplyListResponse struct {
	request MessageReplyListRequest

	Meta struct {
		Page struct {
			HasMore bool `json:"hasMore"`
		} `json:"page"`
	} `json:"meta"`
	MessageReplies []MessageReply `json:"messageReplies"`
}

// HandleHTTPResponse handles the HTTP response for the
// MessageReplyListResponse. If some unexpected HTTP status code is returned by
// the API, a twapi.HTTPError is returned.
func (m *MessageReplyListResponse) HandleHTTPResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return twapi.NewHTTPError(resp, "failed to list message replies")
	}

	if err := json.NewDecoder(resp.Body).Decode(m); err != nil {
		return fmt.Errorf("failed to decode list message replies response: %w", err)
	}
	return nil
}

// SetRequest sets the request used to load this response. This is used for
// pagination purposes, so the Iterate method can return the next page.
func (m *MessageReplyListResponse) SetRequest(req MessageReplyListRequest) {
	m.request = req
}

// Iterate returns the request set to the next page, if available. If there are
// no more pages, a nil request is returned.
func (m *MessageReplyListResponse) Iterate() *MessageReplyListRequest {
	if !m.Meta.Page.HasMore {
		return nil
	}
	req := m.request
	req.Filters.Page++
	return &req
}

// MessageReplyList retrieves multiple message replies using the provided
// request and returns the response.
func MessageReplyList(
	ctx context.Context,
	engine *twapi.Engine,
	req MessageReplyListRequest,
) (*MessageReplyListResponse, error) {
	return twapi.Execute[MessageReplyListRequest, *MessageReplyListResponse](ctx, engine, req)
}
