package projects_test

import (
	"context"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestMessageReplyCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name  string
		input projects.MessageReplyCreateRequest
	}{{
		name: "only required fields",
		input: projects.NewMessageReplyCreateRequest(
			testResources.MessageID,
			"This is a test reply",
		),
	}, {
		name: "all fields",
		input: projects.MessageReplyCreateRequest{
			Path: projects.MessageReplyCreateRequestPath{
				MessageID: testResources.MessageID,
			},
			Body:              "This is a test reply",
			NotifyCurrentUser: new(true),
			Notify:            projects.NewMessageNotifyAll(),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			message, err := projects.MessageReplyCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx = context.Background() // t.Context is always canceled in cleanup
				_, err := projects.MessageReplyDelete(ctx, engine, projects.NewMessageReplyDeleteRequest(int64(message.ID)))
				if err != nil {
					t.Errorf("failed to delete message after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if message.ID == 0 {
				t.Error("expected a valid message ID but got 0")
			}
		})
	}
}

func TestMessageReplyUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	messageID, messageCleanup, err := createMessageReply(t, testResources.MessageID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(messageCleanup)

	tests := []struct {
		name  string
		input projects.MessageReplyUpdateRequest
	}{{
		name: "all fields",
		input: projects.MessageReplyUpdateRequest{
			Path: projects.MessageReplyUpdateRequestPath{
				ID: messageID,
			},
			Body:              new("This is an updated test reply"),
			NotifyCurrentUser: new(true),
			Notify: projects.NewMessageNotifyGroup(projects.LegacyUserGroups{
				UserIDs:    []int64{testResources.UserID},
				CompanyIDs: []int64{testResources.CompanyID},
			}),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.MessageReplyUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestMessageReplyDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	messageID, _, err := createMessageReply(t, testResources.MessageID)
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.MessageReplyDelete(ctx, engine, projects.NewMessageReplyDeleteRequest(messageID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestMessageReplyGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	messageID, messageCleanup, err := createMessageReply(t, testResources.MessageID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(messageCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.MessageReplyGet(ctx, engine, projects.NewMessageReplyGetRequest(messageID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestMessageReplyList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	_, messageCleanup, err := createMessageReply(t, testResources.MessageID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(messageCleanup)

	tests := []struct {
		name  string
		input projects.MessageReplyListRequest
	}{{
		name: "all messages",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.MessageReplyList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
