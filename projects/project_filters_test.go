package projects_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	twapi "github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
)

// listQuery builds the HTTP request for req and returns its parsed query string.
func listQuery(t *testing.T, req twapi.HTTPRequester) url.Values {
	t.Helper()

	httpReq, err := req.HTTPRequest(context.Background(), "https://test.com")
	if err != nil {
		t.Fatalf("unexpected error creating HTTP request: %s", err)
	}
	query, err := url.ParseQuery(httpReq.URL.RawQuery)
	if err != nil {
		t.Fatalf("failed to parse query string: %s", err)
	}
	return query
}

// TestProjectListFiltersApplied pins the whole query string the project list
// builds when every filter is populated, against the parameter names the
// endpoint documents:
//
// https://apidocs.teamwork.com/docs/teamwork/v3/projects/get-projects-api-v3-projects-json
//
// Comparing the complete map rather than a subset is deliberate. An unrecognised
// query key is silently ignored by the API, so a misspelled parameter looks
// exactly like a working one from the caller's side — only an exact comparison
// catches it, and only an exact comparison catches a filter that stopped
// reaching the wire.
func TestProjectListFiltersApplied(t *testing.T) {
	updatedAfter := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
	notCompletedBefore := twapi.Date(time.Date(2026, 3, 4, 0, 0, 0, 0, time.UTC))

	req := projects.ProjectListRequest{
		Filters: projects.ProjectListRequestFilters{
			ProjectRequestFilters: projects.ProjectRequestFilters{
				Include: []projects.ProjectRequestSideload{
					projects.ProjectRequestSideloadProjectCategories,
					projects.ProjectRequestSideloadWorkflowStages,
				},
			},

			SearchTerm:      "acme",
			SearchCompanies: new(true),

			ProjectIDs:        []int64{777, 888},
			ExcludeProjectIDs: []int64{999},

			ProjectType: projects.ProjectTypeNormal,
			ProjectStatuses: []projects.ProjectListStatus{
				projects.ProjectListStatusCurrent,
				projects.ProjectListStatusLate,
			},
			ProjectHealths: []projects.ProjectHealth{
				projects.ProjectHealthNotSet,
				projects.ProjectHealthGood,
			},

			ProjectCategoryIDs:   []int64{12345},
			IncludeSubCategories: new(true),
			ProjectCompanyIDs:    []int64{23456},
			ProjectOwnerIDs:      []int64{34567},

			UserIDs: []int64{45678},
			TeamIDs: []int64{78901},

			TagIDs:               []int64{111, 222},
			MatchAllTags:         new(true),
			ExcludeTagIDs:        []int64{333},
			MatchAllExcludedTags: new(false),

			UpdatedAfter:       &updatedAfter,
			NotCompletedBefore: &notCompletedBefore,

			OnlyStarredProjects:         new(true),
			OnlyProjectsWithAdminAccess: new(true),
			HideObservedProjects:        new(true),

			IncludeArchivedProjects:  new(true),
			OnlyArchivedProjects:     new(false),
			IncludeTentativeProjects: new(true),

			IncludeCustomFields:   new(true),
			IncludeCustomFieldIDs: []int64{444, 555},
			UseFormulaFields:      new(true),

			IncludeProjectStats:         new(true),
			IncludeOverallStats:         new(true),
			IncludeProjectDates:         new(true),
			IncludeProjectUserInfo:      new(true),
			IncludeProjectProfitability: new(true),
			TimeMode:                    projects.ProjectTimeModeEstimated,

			OrderBy:              projects.ProjectOrderByCustomField,
			OrderMode:            twapi.OrderModeDescending,
			OrderByCustomFieldID: 666,

			Page:      2,
			PageSize:  25,
			CountMode: twapi.ListCountModeExact,

			Fields: projects.ProjectListFields{
				Projects: []projects.ProjectField{projects.ProjectFieldName},
			},
		},
	}

	want := map[string]string{
		"include": "projectCategories,workflows.stages",

		"searchTerm":      "acme",
		"searchCompanies": "true",

		"projectIds":        "777,888",
		"excludeProjectIds": "999",

		"projectType":     "normal",
		"projectStatuses": "current,late",
		"projectHealths":  "0,3",

		"projectCategoryIds":   "12345",
		"includeSubCategories": "true",
		"projectCompanyIds":    "23456",
		"projectOwnerIds":      "34567",

		"usersWithExplicitMembershipIds": "45678",
		"teamIds":                        "78901",

		"projectTagIds":        "111,222",
		"matchAllProjectTags":  "true",
		"excludeTagIds":        "333",
		"matchAllExcludedTags": "false",

		"updatedAfter":       "2026-02-03T04:05:06Z",
		"notCompletedBefore": "2026-03-04",

		"onlyStarredProjects":         "true",
		"onlyProjectsWithAdminAccess": "true",
		"hideObservedProjects":        "true",

		"includeArchivedProjects":  "true",
		"onlyArchivedProjects":     "false",
		"includeTentativeProjects": "true",

		"includeCustomFields":   "true",
		"includeCustomFieldIds": "444,555",
		"useFormulaFields":      "true",

		"includeCounts":               "true",
		"includeStats":                "true",
		"includeProjectDates":         "true",
		"includeProjectUserInfo":      "true",
		"includeProjectProfitability": "true",
		"timeMode":                    "estimated",

		"orderBy":              "customfield",
		"orderMode":            "desc",
		"orderByCustomFieldId": "666",

		"page":       "2",
		"pageSize":   "25",
		"skipCounts": "false",

		"fields[projects]": "name",
	}

	query := listQuery(t, req)

	for key, expected := range want {
		if got := query.Get(key); got != expected {
			t.Errorf("expected %s=%q but got %q", key, expected, got)
		}
	}
	for key := range query {
		if _, ok := want[key]; !ok {
			t.Errorf("unexpected query parameter %s=%q", key, query.Get(key))
		}
	}
}

// TestProjectListFiltersUnset checks the zero-value filters send nothing, so the
// endpoint applies its own defaults rather than receiving a wall of false.
func TestProjectListFiltersUnset(t *testing.T) {
	query := listQuery(t, projects.ProjectListRequest{})
	if len(query) != 0 {
		t.Errorf("expected no query parameters but got %v", query)
	}
}
