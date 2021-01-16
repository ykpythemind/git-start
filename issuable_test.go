package main

import (
	"reflect"
	"testing"
)

func TestParseGitHubIssuable(t *testing.T) {
	type arg string

	tests := []struct {
		name    string
		args    arg
		want    *GitHubIssue
		wantErr bool
	}{
		{
			name: "only issue number",
			args: "1000",
			want: &GitHubIssue{Number: 1000},
		},
		{
			name: "GitHub repo",
			args: "https://github.com/ykpythemind/piyo/issues/1234",
			want: &GitHubIssue{Number: 1234, Owner: "ykpythemind", Repo: "piyo"},
		},
		{
			name: "GitHub repo with blank",
			args: "  https://github.com/ykpythemind/piyo/issues/1234\n  ",
			want: &GitHubIssue{Number: 1234, Owner: "ykpythemind", Repo: "piyo"},
		},
		{
			name:    "GitHub repo (issue num is not present)",
			args:    "https://github.com/ykpythemind/piyo/issues",
			wantErr: true,
		},
		{
			name:    "GitHub repo (pull request)",
			args:    "https://github.com/ykpythemind/piyo/pulls/1234",
			wantErr: true,
		},
		{
			name:    "invalid",
			args:    "hoge",
			wantErr: true,
		},
		{
			name:    "gitlab",
			args:    "https://gitlab.com/ykpythemind/piyo/issues/1234",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseGitHubIssuable(string(tt.args))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseGitHubIssuable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseGitHubIssuable() = %v, want %v", got, tt.want)
			}
		})
	}
}
