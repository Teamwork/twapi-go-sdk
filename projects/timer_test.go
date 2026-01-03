package projects_test

import (
	"context"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestTimerCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name  string
		input projects.TimerCreateRequest
	}{{
		name:  "only required fields",
		input: projects.NewTimerCreateRequest(testResources.ProjectID),
	}, {
		name: "all fields",
		input: projects.TimerCreateRequest{
			Description:       twapi.Ptr("This is a test timer"),
			Billable:          twapi.Ptr(true),
			Running:           twapi.Ptr(true),
			Seconds:           twapi.Ptr(int64(3600)), // 1 hour in seconds
			StopRunningTimers: twapi.Ptr(true),
			ProjectID:         testResources.ProjectID,
			TaskID:            &testResources.TaskID,
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			timerResponse, err := projects.TimerCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx = context.Background() // t.Context is always canceled in cleanup
				_, err := projects.TimerDelete(ctx, engine, projects.NewTimerDeleteRequest(timerResponse.Timer.ID))
				if err != nil {
					t.Errorf("failed to delete timer after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if timerResponse.Timer.ID == 0 {
				t.Error("expected a valid timer ID but got 0")
			}
		})
	}
}

func TestTimerUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	timerID, timerCleanup, err := createTimer(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(timerCleanup)

	tests := []struct {
		name  string
		input projects.TimerUpdateRequest
	}{{
		name: "all fields",
		input: projects.TimerUpdateRequest{
			Path: projects.TimerUpdateRequestPath{
				ID: timerID,
			},
			Description: twapi.Ptr("Updated description"),
			Billable:    twapi.Ptr(true),
			Running:     twapi.Ptr(true),
			ProjectID:   &testResources.ProjectID,
			TaskID:      &testResources.TaskID,
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.TimerUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestTimerPause(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	timerID, timerCleanup, err := createTimer(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(timerCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err := projects.TimerPause(ctx, engine, projects.NewTimerPauseRequest(timerID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestTimerResume(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	timerID, timerCleanup, err := createTimer(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(timerCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err := projects.TimerPause(ctx, engine, projects.NewTimerPauseRequest(timerID)); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if _, err := projects.TimerResume(ctx, engine, projects.NewTimerResumeRequest(timerID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestTimerComplete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	timerID, timerCleanup, err := createTimer(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(timerCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err := projects.TimerPause(ctx, engine, projects.NewTimerPauseRequest(timerID)); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if _, err := projects.TimerComplete(ctx, engine, projects.NewTimerCompleteRequest(timerID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestTimerDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	timerID, _, err := createTimer(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.TimerDelete(ctx, engine, projects.NewTimerDeleteRequest(timerID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestTimerGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	timerID, timerCleanup, err := createTimer(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(timerCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.TimerGet(ctx, engine, projects.NewTimerGetRequest(timerID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestTimerList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	_, timerCleanup, err := createTimer(t, testResources.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(timerCleanup)

	tests := []struct {
		name          string
		input         projects.TimerListRequest
		expectedError bool
	}{{
		name: "all timers",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.TimerList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
