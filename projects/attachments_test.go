package projects_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

// TestAttachmentsEncoding pins how attachments are encoded on the wire. The
// integration tests only assert that a write was accepted, and the API answers
// 2xx whether or not it understood the attachment, so a wrong key or a field
// nested in the wrong place would otherwise go unnoticed.
func TestAttachmentsEncoding(t *testing.T) {
	tests := []struct {
		name      string
		requester twapi.HTTPRequester
		want      map[string]any
	}{{
		// The v3 routes take attachments as a sibling of the entity, not as one
		// of its attributes.
		name: "task create with a pending file",
		requester: func() projects.TaskCreateRequest {
			req := projects.NewTaskCreateRequest(777, "example")
			req.Attachments = projects.TaskAttachments{
				PendingFiles: []projects.TaskAttachmentPendingFile{{Reference: "tf_A"}},
			}
			return req
		}(),
		want: map[string]any{
			"attachments": map[string]any{
				"pendingFiles": []any{map[string]any{"reference": "tf_A"}},
			},
		},
	}, {
		name: "task create with an existing file and a category",
		requester: func() projects.TaskCreateRequest {
			req := projects.NewTaskCreateRequest(777, "example")
			req.Attachments = projects.TaskAttachments{
				Files: []projects.TaskAttachmentFile{{ID: 12345, CategoryID: new(int64(67))}},
			}
			return req
		}(),
		want: map[string]any{
			"attachments": map[string]any{
				"files": []any{map[string]any{"id": float64(12345), "categoryId": float64(67)}},
			},
		},
	}, {
		name: "task update with a pending file",
		requester: func() projects.TaskUpdateRequest {
			req := projects.NewTaskUpdateRequest(12345)
			req.Attachments = projects.TaskAttachments{
				PendingFiles: []projects.TaskAttachmentPendingFile{{Reference: "tf_A"}, {Reference: "tf_B"}},
			}
			return req
		}(),
		want: map[string]any{
			"attachments": map[string]any{
				"pendingFiles": []any{
					map[string]any{"reference": "tf_A"},
					map[string]any{"reference": "tf_B"},
				},
			},
		},
	}, {
		// The legacy routes accept either a JSON array or a comma-separated
		// string. The array is used because it needs no encoding of its own.
		name: "comment create with two pending files",
		requester: func() projects.CommentCreateRequest {
			req := projects.NewCommentCreateRequestInTask(777, "example")
			req.PendingFileAttachments = []projects.PendingFileRef{"tf_A", "tf_B"}
			return req
		}(),
		want: map[string]any{
			"comment": map[string]any{"pendingFileAttachments": []any{"tf_A", "tf_B"}},
		},
	}, {
		name: "message create with pending and existing files",
		requester: func() projects.MessageCreateRequest {
			req := projects.NewMessageCreateRequest(777, "example", "body")
			req.PendingFileAttachments = []projects.PendingFileRef{"tf_A", "tf_B"}
			req.Attachments = projects.LegacyNumericList{1, 2}
			return req
		}(),
		want: map[string]any{
			"post": map[string]any{
				"pendingFileAttachments": []any{"tf_A", "tf_B"},
				"attachments":            "1,2",
			},
		},
	}, {
		name:      "file create carries the reference in the file envelope",
		requester: projects.NewFileCreateRequest(777, "tf_A"),
		want: map[string]any{
			"file": map[string]any{"pendingFileRef": "tf_A"},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := encodeRequestBody(t, tt.requester)
			for key, want := range tt.want {
				got, ok := body[key]
				if !ok {
					t.Fatalf("expected key %q in the request body, got %v", key, body)
				}
				assertSubset(t, key, want, got)
			}
		})
	}
}

// TestAttachmentsOmittedWhenNotRequested guards the requests that hoist
// attachments into the payload. An empty attachments object is not the same as
// no attachments object to a reader of the payload, and every existing caller
// of these requests sends none, so the key has to stay absent.
func TestAttachmentsOmittedWhenNotRequested(t *testing.T) {
	tests := []struct {
		name      string
		requester twapi.HTTPRequester
		keys      []string
	}{{
		name:      "task create",
		requester: projects.NewTaskCreateRequest(777, "example"),
		keys:      []string{"attachments", "attachmentOptions"},
	}, {
		name:      "task update",
		requester: projects.NewTaskUpdateRequest(12345),
		keys:      []string{"attachments", "attachmentOptions"},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := encodeRequestBody(t, tt.requester)
			for _, key := range tt.keys {
				if _, ok := body[key]; ok {
					t.Errorf("expected no %q key in the request body, got %v", key, body[key])
				}
			}
		})
	}
}

func TestFileCreateRequestRejectsMissingFields(t *testing.T) {
	tests := []struct {
		name    string
		request projects.FileCreateRequest
	}{{
		name:    "no project",
		request: projects.NewFileCreateRequest(0, "tf_A"),
	}, {
		name:    "no pending file reference",
		request: projects.NewFileCreateRequest(777, ""),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.request.HTTPRequest(context.Background(), "http://example.com"); err == nil {
				t.Error("expected an error, got none")
			}
		})
	}
}

// encodeRequestBody builds the request and decodes its JSON body.
func encodeRequestBody(t *testing.T, requester twapi.HTTPRequester) map[string]any {
	t.Helper()

	httpRequest, err := requester.HTTPRequest(context.Background(), "http://example.com")
	if err != nil {
		t.Fatalf("failed to build the request: %s", err)
	}
	if httpRequest.Body == nil {
		t.Fatalf("expected a request body, got none")
	}
	var body map[string]any
	if err := json.NewDecoder(httpRequest.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode the request body: %s", err)
	}
	return body
}

// assertSubset checks that every value in want appears in got, ignoring any
// extra keys got may carry, so that the tests only pin the attachment fields.
func assertSubset(t *testing.T, path string, want, got any) {
	t.Helper()

	switch want := want.(type) {
	case map[string]any:
		gotMap, ok := got.(map[string]any)
		if !ok {
			t.Fatalf("%s: expected an object, got %T (%v)", path, got, got)
		}
		for key, value := range want {
			gotValue, ok := gotMap[key]
			if !ok {
				t.Fatalf("%s: expected key %q, got %v", path, key, gotMap)
			}
			assertSubset(t, path+"."+key, value, gotValue)
		}
	case []any:
		gotSlice, ok := got.([]any)
		if !ok {
			t.Fatalf("%s: expected an array, got %T (%v)", path, got, got)
		}
		if len(gotSlice) != len(want) {
			t.Fatalf("%s: expected %d items, got %d (%v)", path, len(want), len(gotSlice), gotSlice)
		}
		for i, value := range want {
			assertSubset(t, fmt.Sprintf("%s[%d]", path, i), value, gotSlice[i])
		}
	default:
		if want != got {
			t.Errorf("%s: expected %v (%T), got %v (%T)", path, want, want, got, got)
		}
	}
}
