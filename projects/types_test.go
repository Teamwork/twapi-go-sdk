package projects_test

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/teamwork/twapi-go-sdk/projects"
)

func TestLegacyUserGroups_MarshalJSON(t *testing.T) {
	tests := []struct {
		name  string
		input projects.LegacyUserGroups
		want  string
	}{{
		name:  "empty",
		input: projects.LegacyUserGroups{},
		want:  `""`,
	}, {
		name:  "users only",
		input: projects.LegacyUserGroups{UserIDs: []int64{1, 2}},
		want:  `"1,2"`,
	}, {
		name:  "job roles only",
		input: projects.LegacyUserGroups{JobRoleIDs: []int64{5}},
		want:  `"r5"`,
	}, {
		name: "all groups ordered users,teams,companies,jobRoles",
		input: projects.LegacyUserGroups{
			UserIDs:    []int64{1},
			TeamIDs:    []int64{2},
			CompanyIDs: []int64{3},
			JobRoleIDs: []int64{4},
		},
		want: `"1,t2,c3,r4"`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if string(got) != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestLegacyUserGroups_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    projects.LegacyUserGroups
		wantErr bool
	}{{
		name:  "empty",
		input: `""`,
		want:  projects.LegacyUserGroups{},
	}, {
		name:  "job roles only",
		input: `"r5"`,
		want:  projects.LegacyUserGroups{JobRoleIDs: []int64{5}},
	}, {
		name:  "all groups",
		input: `"1,t2,c3,r4"`,
		want: projects.LegacyUserGroups{
			UserIDs:    []int64{1},
			TeamIDs:    []int64{2},
			CompanyIDs: []int64{3},
			JobRoleIDs: []int64{4},
		},
	}, {
		name:    "invalid job role format",
		input:   `"r"`,
		wantErr: true,
	}, {
		name:    "invalid job role id",
		input:   `"rabc"`,
		wantErr: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got projects.LegacyUserGroups
			err := json.Unmarshal([]byte(tt.input), &got)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if !slices.Equal(got.UserIDs, tt.want.UserIDs) ||
				!slices.Equal(got.TeamIDs, tt.want.TeamIDs) ||
				!slices.Equal(got.CompanyIDs, tt.want.CompanyIDs) ||
				!slices.Equal(got.JobRoleIDs, tt.want.JobRoleIDs) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestLegacyUserGroups_RoundTrip(t *testing.T) {
	want := projects.LegacyUserGroups{
		UserIDs:    []int64{1, 2},
		TeamIDs:    []int64{3},
		CompanyIDs: []int64{4},
		JobRoleIDs: []int64{5, 6},
	}

	encoded, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("unexpected marshal error: %s", err)
	}

	var got projects.LegacyUserGroups
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatalf("unexpected unmarshal error: %s", err)
	}

	if !slices.Equal(got.UserIDs, want.UserIDs) ||
		!slices.Equal(got.TeamIDs, want.TeamIDs) ||
		!slices.Equal(got.CompanyIDs, want.CompanyIDs) ||
		!slices.Equal(got.JobRoleIDs, want.JobRoleIDs) {
		t.Errorf("round trip mismatch: got %+v, want %+v", got, want)
	}
}

func TestLegacyUserGroups_IsEmpty(t *testing.T) {
	tests := []struct {
		name  string
		input projects.LegacyUserGroups
		want  bool
	}{{
		name:  "empty",
		input: projects.LegacyUserGroups{},
		want:  true,
	}, {
		name:  "users only",
		input: projects.LegacyUserGroups{UserIDs: []int64{1}},
		want:  false,
	}, {
		name:  "job roles only",
		input: projects.LegacyUserGroups{JobRoleIDs: []int64{5}},
		want:  false,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.input.IsEmpty(); got != tt.want {
				t.Errorf("got %t, want %t", got, tt.want)
			}
		})
	}
}
