package gitstart

import (
	"testing"

	"github.com/MakeNowJust/heredoc/v2"

	"github.com/google/go-cmp/cmp"
)

func TestNewStarterOptionFromTemplate(t *testing.T) {
	type testcase struct {
		template      string
		starterOption *StarterOption
		wantErr       bool
	}

	tc := []testcase{
		{
			template: heredoc.Doc(`
			branch: hoge
			title: PR title

			---
			`),
			starterOption: &StarterOption{SwitchBranch: "hoge", PullRequestTitle: "PR title"},
			wantErr:       false,
		},
		{
			template: heredoc.Doc(`
			branch:
			title: PR title
			`),
			wantErr: true,
		},
		{
			template: heredoc.Doc(`
			branch: hoge
			title:
			`),
			starterOption: &StarterOption{SwitchBranch: "hoge", PullRequestTitle: ""},
			wantErr:       false,
		},
		{
			template: heredoc.Doc(`
			branch: hoge
			title: branch
			branch: piyo
			`),
			wantErr:       false,
			starterOption: &StarterOption{SwitchBranch: "hoge", PullRequestTitle: "branch"},
		},
	}

	for _, tt := range tc {
		tt := tt

		got, err := NewStarterOptionFromTemplate(tt.template)
		hasErr := err != nil
		if hasErr != tt.wantErr {
			t.Errorf("hasErr: %v but wantErr: %v", hasErr, tt.wantErr)
			return
		}

		if diff := cmp.Diff(tt.starterOption, got); diff != "" {
			t.Errorf("diff:\n%s", diff)
		}
	}
}
