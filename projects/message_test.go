package projects_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestMessageCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name  string
		input projects.MessageCreateRequest
	}{{
		name: "only required fields",
		input: projects.NewMessageCreateRequest(
			testResources.ProjectID,
			fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
			"This is a test message",
		),
	}, {
		name: "all fields",
		input: projects.MessageCreateRequest{
			Path: projects.MessageCreateRequestPath{
				ProjectID: testResources.ProjectID,
			},
			Title:             fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
			Body:              "This is a test message",
			NotifyCurrentUser: new(true),
			Notify:            projects.NewMessageNotifyAll(),
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			message, err := projects.MessageCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx = context.Background() // t.Context is always canceled in cleanup
				_, err := projects.MessageDelete(ctx, engine, projects.NewMessageDeleteRequest(int64(message.ID)))
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

func TestMessageUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	messageID, messageCleanup, err := createMessage(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(messageCleanup)

	tests := []struct {
		name  string
		input projects.MessageUpdateRequest
	}{{
		name: "all fields",
		input: projects.MessageUpdateRequest{
			Path: projects.MessageUpdateRequestPath{
				ID: messageID,
			},
			Title:             new(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
			Body:              new("This is an updated test message"),
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

			if _, err := projects.MessageUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestMessageDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	messageID, _, err := createMessage(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.MessageDelete(ctx, engine, projects.NewMessageDeleteRequest(messageID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestMessageGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	messageID, messageCleanup, err := createMessage(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(messageCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.MessageGet(ctx, engine, projects.NewMessageGetRequest(messageID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestMessageList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	_, messageCleanup, err := createMessage(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(messageCleanup)

	tests := []struct {
		name  string
		input projects.MessageListRequest
	}{{
		name: "all messages",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.MessageList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
