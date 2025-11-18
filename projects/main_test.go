package projects_test

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
	"github.com/teamwork/twapi-go-sdk/session"
)

var engine *twapi.Engine

var testResources struct {
	CompanyID         int64
	ProjectID         int64
	ProjectCategoryID int64
	TasklistID        int64
	TaskID            int64
	UserID            int64
	MilestoneID       int64
	TagID             int64
}

func TestMain(m *testing.M) {
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	if engine = startEngine(); engine == nil {
		logger.Info("Missing setup environment variables, skipping tests")
		return
	}

	// initialize some resources that can be shared across tests
	testEngine := newMainTestEngine(context.Background(), logger)

	companyID, companyCleanup, err := createCompany(testEngine)
	if err != nil {
		logger.Error("Failed to create company for tests",
			slog.String("error", err.Error()),
		)
		exitCode = 1
		return
	}
	defer companyCleanup()
	testResources.CompanyID = companyID

	projectID, projectCleanup, err := createProject(testEngine)
	if err != nil {
		logger.Error("Failed to create project for tests",
			slog.String("error", err.Error()),
		)
		exitCode = 1
		return
	}
	defer projectCleanup()
	testResources.ProjectID = projectID

	projectCategoryID, projectCategoryCleanup, err := createProjectCategory(testEngine)
	if err != nil {
		logger.Error("Failed to create project category for tests",
			slog.String("error", err.Error()),
		)
		exitCode = 1
		return
	}
	defer projectCategoryCleanup()
	testResources.ProjectCategoryID = projectCategoryID

	tasklistID, tasklistCleanup, err := createTasklist(testEngine, projectID)
	if err != nil {
		logger.Error("Failed to create tasklist for tests",
			slog.String("error", err.Error()),
		)
		exitCode = 1
		return
	}
	defer tasklistCleanup()
	testResources.TasklistID = tasklistID

	taskID, taskCleanup, err := createTask(testEngine, tasklistID)
	if err != nil {
		logger.Error("Failed to create task for tests",
			slog.String("error", err.Error()),
		)
		exitCode = 1
		return
	}
	defer taskCleanup()
	testResources.TaskID = taskID

	userID, userCleanup, err := createUser(testEngine)
	if err != nil {
		logger.Error("Failed to create user for tests",
			slog.String("error", err.Error()),
		)
		exitCode = 1
		return
	}
	defer userCleanup()
	testResources.UserID = userID

	if err := addProjectMember(testEngine, projectID, userID); err != nil {
		logger.Error("Failed to add user to project for tests",
			slog.String("error", err.Error()),
		)
		exitCode = 1
		return
	}

	milestoneID, milestoneCleanup, err := createMilestone(testEngine, projectID, projects.LegacyUserGroups{
		UserIDs: []int64{testResources.UserID},
	})
	if err != nil {
		logger.Error("Failed to create milestone for tests",
			slog.String("error", err.Error()),
		)
		exitCode = 1
		return
	}
	defer milestoneCleanup()
	testResources.MilestoneID = milestoneID

	tagID, tagCleanup, err := createTag(testEngine)
	if err != nil {
		logger.Error("Failed to create tag for tests",
			slog.String("error", err.Error()),
		)
		exitCode = 1
		return
	}
	defer tagCleanup()
	testResources.TagID = tagID

	exitCode = m.Run()
}

func startEngine() *twapi.Engine {
	server, token := strings.TrimSuffix(os.Getenv("TWAPI_SERVER"), "/"), os.Getenv("TWAPI_TOKEN")
	if server == "" || token == "" {
		return nil
	}
	return twapi.NewEngine(session.NewBearerToken(token, server))
}

func createProject(t testEngine) (int64, func(), error) {
	project, err := projects.ProjectCreate(t.Context(), engine, projects.ProjectCreateRequest{
		Name: fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
	})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create project for test: %w", err)
	}
	id := int64(project.ID)
	return id, func() {
		ctx := context.Background() // t.Context is always canceled in cleanup
		_, err := projects.ProjectDelete(ctx, engine, projects.NewProjectDeleteRequest(id))
		if err != nil {
			t.Errorf("failed to delete project after test: %s", err)
		}
	}, nil
}

func createProjectCategory(t testEngine) (int64, func(), error) {
	projectCategory, err := projects.ProjectCategoryCreate(t.Context(), engine, projects.ProjectCategoryCreateRequest{
		Name:  fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
		Color: twapi.Ptr("#00ff00"),
	})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create project category for test: %w", err)
	}
	id := int64(projectCategory.ID)
	return id, func() {
		ctx := context.Background() // t.Context is always canceled in cleanup
		_, err := projects.ProjectCategoryDelete(ctx, engine, projects.NewProjectCategoryDeleteRequest(id))
		if err != nil {
			t.Errorf("failed to delete project category after test: %s", err)
		}
	}, nil
}

func createTasklist(t testEngine, projectID int64) (int64, func(), error) {
	tasklist, err := projects.TasklistCreate(t.Context(), engine, projects.TasklistCreateRequest{
		Path: projects.TasklistCreateRequestPath{
			ProjectID: projectID,
		},
		Name: fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
	})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create tasklist for test: %w", err)
	}
	id := int64(tasklist.ID)
	return id, func() {
		ctx := context.Background() // t.Context is always canceled in cleanup
		_, err := projects.TasklistDelete(ctx, engine, projects.NewTasklistDeleteRequest(id))
		if err != nil {
			t.Errorf("failed to delete tasklist after test: %s", err)
		}
	}, nil
}

func createTask(t testEngine, tasklistID int64) (int64, func(), error) {
	taskResponse, err := projects.TaskCreate(t.Context(), engine, projects.TaskCreateRequest{
		Path: projects.TaskCreateRequestPath{
			TasklistID: tasklistID,
		},
		Name: fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
	})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create task for test: %w", err)
	}
	id := taskResponse.Task.ID
	return id, func() {
		ctx := context.Background() // t.Context is always canceled in cleanup
		_, err := projects.TaskDelete(ctx, engine, projects.NewTaskDeleteRequest(id))
		if err != nil {
			t.Errorf("failed to delete task after test: %s", err)
		}
	}, nil
}

func createUser(t testEngine) (int64, func(), error) {
	user, err := projects.UserCreate(t.Context(), engine, projects.NewUserCreateRequest(
		fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
		fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
		fmt.Sprintf("testuser%d%d@example.com", time.Now().UnixNano(), rand.Intn(100)),
	))
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create user for test: %w", err)
	}
	id := int64(user.ID)
	return id, func() {
		ctx := context.Background() // t.Context is always canceled in cleanup
		_, err := projects.UserDelete(ctx, engine, projects.NewUserDeleteRequest(id))
		if err != nil {
			t.Errorf("failed to delete user after test: %s", err)
		}
	}, nil
}

func addProjectMember(t testEngine, projectID, userID int64) error {
	_, err := projects.ProjectMemberAdd(t.Context(), engine, projects.NewProjectMemberAddRequest(projectID, userID))
	if err != nil {
		return fmt.Errorf("failed to add user %d to project %d: %w", userID, projectID, err)
	}
	return nil
}

func createMilestone(t testEngine, projectID int64, assignees projects.LegacyUserGroups) (int64, func(), error) {
	milestone, err := projects.MilestoneCreate(t.Context(), engine, projects.MilestoneCreateRequest{
		Path: projects.MilestoneCreateRequestPath{
			ProjectID: projectID,
		},
		Name:      fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
		DueAt:     projects.NewLegacyDate(time.Now().Add(24 * time.Hour)), // Due tomorrow
		Assignees: assignees,
	})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create milestone for test: %w", err)
	}
	id := int64(milestone.ID)
	return id, func() {
		ctx := context.Background() // t.Context is always canceled in cleanup
		_, err := projects.MilestoneDelete(ctx, engine, projects.NewMilestoneDeleteRequest(id))
		if err != nil {
			t.Errorf("failed to delete milestone after test: %s", err)
		}
	}, nil
}

func createCompany(t testEngine) (int64, func(), error) {
	companyResponse, err := projects.CompanyCreate(t.Context(), engine, projects.CompanyCreateRequest{
		Name: fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
	})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create company for test: %w", err)
	}
	id := companyResponse.Company.ID
	return id, func() {
		ctx := context.Background() // t.Context is always canceled in cleanup
		_, err := projects.CompanyDelete(ctx, engine, projects.NewCompanyDeleteRequest(id))
		if err != nil {
			t.Errorf("failed to delete company after test: %s", err)
		}
	}, nil
}

func createTag(t testEngine) (int64, func(), error) {
	tagResponse, err := projects.TagCreate(t.Context(), engine, projects.TagCreateRequest{
		Name: fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
	})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create tag for test: %w", err)
	}
	id := tagResponse.Tag.ID
	return id, func() {
		ctx := context.Background() // t.Context is always canceled in cleanup
		_, err := projects.TagDelete(ctx, engine, projects.NewTagDeleteRequest(id))
		if err != nil {
			t.Errorf("failed to delete tag after test: %s", err)
		}
	}, nil
}

func createTeam(t testEngine) (int64, func(), error) {
	team, err := projects.TeamCreate(t.Context(), engine, projects.NewTeamCreateRequest(
		fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
	))
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create team for test: %w", err)
	}
	id := int64(team.ID)
	return id, func() {
		ctx := context.Background() // t.Context is always canceled in cleanup
		_, err := projects.TeamDelete(ctx, engine, projects.NewTeamDeleteRequest(id))
		if err != nil {
			t.Errorf("failed to delete team after test: %s", err)
		}
	}, nil
}

func createCommentInTask(t testEngine, taskID int64) (int64, func(), error) {
	comment, err := projects.CommentCreate(t.Context(), engine, projects.CommentCreateRequest{
		Path: projects.CommentCreateRequestPath{
			TaskID: taskID,
		},
		Body:        "<h1>This is a test comment</h1>",
		ContentType: twapi.Ptr("HTML"),
	})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create comment for test: %w", err)
	}
	id := int64(comment.ID)
	return id, func() {
		ctx := context.Background() // t.Context is always canceled in cleanup
		_, err := projects.CommentDelete(ctx, engine, projects.NewCommentDeleteRequest(id))
		if err != nil {
			t.Errorf("failed to delete comment after test: %s", err)
		}
	}, nil
}

func createCommentInMilestone(t testEngine, milestoneID int64) (int64, func(), error) {
	comment, err := projects.CommentCreate(t.Context(), engine, projects.CommentCreateRequest{
		Path: projects.CommentCreateRequestPath{
			MilestoneID: milestoneID,
		},
		Body:        "<h1>This is a test comment</h1>",
		ContentType: twapi.Ptr("HTML"),
	})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create comment for test: %w", err)
	}
	id := int64(comment.ID)
	return id, func() {
		ctx := context.Background() // t.Context is always canceled in cleanup
		_, err := projects.CommentDelete(ctx, engine, projects.NewCommentDeleteRequest(id))
		if err != nil {
			t.Errorf("failed to delete comment after test: %s", err)
		}
	}, nil
}

func createTimelogInTask(t testEngine, taskID int64) (int64, func(), error) {
	timelogResponse, err := projects.TimelogCreate(t.Context(), engine, projects.TimelogCreateRequest{
		Path: projects.TimelogCreateRequestPath{
			TaskID: taskID,
		},
		Date:    twapi.Date(time.Now().UTC()),
		Time:    twapi.Time(time.Now().UTC()),
		IsUTC:   true,
		Minutes: 30,
	})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create timelog for test: %w", err)
	}
	id := timelogResponse.Timelog.ID
	return id, func() {
		ctx := context.Background() // t.Context is always canceled in cleanup
		_, err := projects.TimelogDelete(ctx, engine, projects.NewTimelogDeleteRequest(id))
		if err != nil {
			t.Errorf("failed to delete timelog after test: %s", err)
		}
	}, nil
}

func createTimelogInProject(t testEngine, projectID int64) (int64, func(), error) {
	timelogResponse, err := projects.TimelogCreate(t.Context(), engine, projects.TimelogCreateRequest{
		Path: projects.TimelogCreateRequestPath{
			ProjectID: projectID,
		},
		Date:    twapi.Date(time.Now().UTC()),
		Time:    twapi.Time(time.Now().UTC()),
		IsUTC:   true,
		Minutes: 30,
	})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create timelog for test: %w", err)
	}
	id := timelogResponse.Timelog.ID
	return id, func() {
		ctx := context.Background() // t.Context is always canceled in cleanup
		_, err := projects.TimelogDelete(ctx, engine, projects.NewTimelogDeleteRequest(id))
		if err != nil {
			t.Errorf("failed to delete timelog after test: %s", err)
		}
	}, nil
}

func createTimer(t testEngine, projectID int64) (int64, func(), error) {
	timerResponse, err := projects.TimerCreate(t.Context(), engine, projects.NewTimerCreateRequest(projectID))
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create timer for test: %w", err)
	}
	id := timerResponse.Timer.ID
	return id, func() {
		ctx := context.Background() // t.Context is always canceled in cleanup
		_, err := projects.TimerDelete(ctx, engine, projects.NewTimerDeleteRequest(id))
		if err != nil {
			t.Errorf("failed to delete timer after test: %s", err)
		}
	}, nil
}

func createNotebook(t testEngine, projectID int64) (int64, func(), error) {
	notebookResponse, err := projects.NotebookCreate(t.Context(), engine, projects.NewNotebookCreateRequest(
		projectID,
		fmt.Sprintf("test%d%d", time.Now().UnixNano(), rand.Intn(100)),
		"An amazing content",
		projects.NotebookTypeMarkdown,
	))
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create notebook for test: %w", err)
	}
	id := notebookResponse.Notebook.ID
	return id, func() {
		ctx := context.Background() // t.Context is always canceled in cleanup
		_, err := projects.NotebookDelete(ctx, engine, projects.NewNotebookDeleteRequest(id))
		if err != nil {
			t.Errorf("failed to delete notebook after test: %s", err)
		}
	}, nil
}

type testEngine interface {
	Context() context.Context
	Errorf(string, ...any)
}

type mainTestEngine struct {
	ctx    context.Context
	logger *slog.Logger
}

func newMainTestEngine(ctx context.Context, logger *slog.Logger) *mainTestEngine {
	return &mainTestEngine{
		ctx:    ctx,
		logger: logger,
	}
}

func (m *mainTestEngine) Context() context.Context {
	return m.ctx
}

func (m *mainTestEngine) Errorf(format string, args ...any) {
	m.logger.Error(fmt.Sprintf(format, args...))
}
