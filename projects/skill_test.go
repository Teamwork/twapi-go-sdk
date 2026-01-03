package projects_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestSkillCreate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	tests := []struct {
		name  string
		input projects.SkillCreateRequest
	}{{
		name: "only required fields",
		input: projects.NewSkillCreateRequest(
			fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
		),
	}, {
		name: "all fields",
		input: projects.SkillCreateRequest{
			Name:    fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
			UserIDs: []int64{testResources.UserID},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			skillResponse, err := projects.SkillCreate(ctx, engine, tt.input)
			t.Cleanup(func() {
				if err != nil {
					return
				}
				ctx = context.Background() // t.Context is always canceled in cleanup
				_, err := projects.SkillDelete(ctx, engine, projects.NewSkillDeleteRequest(skillResponse.Skill.ID))
				if err != nil {
					t.Errorf("failed to delete skill after test: %s", err)
				}
			})
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if skillResponse.Skill.ID == 0 {
				t.Error("expected a valid skill ID but got 0")
			}
		})
	}
}

func TestSkillUpdate(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	skillID, skillCleanup, err := createSkill(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(skillCleanup)

	tests := []struct {
		name  string
		input projects.SkillUpdateRequest
	}{{
		name: "all fields",
		input: projects.SkillUpdateRequest{
			Path: projects.SkillUpdateRequestPath{
				ID: skillID,
			},
			Name:    twapi.Ptr(fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100))),
			UserIDs: []int64{testResources.UserID},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.SkillUpdate(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}

func TestSkillDelete(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	skillID, _, err := createSkill(t)
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.SkillDelete(ctx, engine, projects.NewSkillDeleteRequest(skillID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestSkillGet(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	skillID, skillCleanup, err := createSkill(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(skillCleanup)

	ctx := t.Context()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)

	if _, err = projects.SkillGet(ctx, engine, projects.NewSkillGetRequest(skillID)); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestSkillList(t *testing.T) {
	if engine == nil {
		t.Skip("Skipping test because the engine is not initialized")
	}

	_, skillCleanup, err := createSkill(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(skillCleanup)

	tests := []struct {
		name          string
		input         projects.SkillListRequest
		expectedError bool
	}{{
		name: "all skills",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			t.Cleanup(cancel)

			if _, err := projects.SkillList(ctx, engine, tt.input); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
		})
	}
}
