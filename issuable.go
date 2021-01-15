package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
)

type GitHubIssue struct {
	Owner  string
	Repo   string
	Number int
	URL    string
	Body   string
	Title  string
}

func ParseGitHubIssuable(str string) (*GitHubIssue, error) {
	str = strings.TrimSpace(str)

	// 番号だけだった場合
	if i, err := strconv.Atoi(str); err == nil {
		return &GitHubIssue{Number: i}, nil
	}

	str = strings.TrimPrefix(str, "https://")

	if !strings.HasPrefix(str, "github.com") {
		return nil, errors.New("can't parse non GitHub url")
	}

	splited := strings.Split(str, "/")
	// "github.com", "ykpythemind", "repoName", "issues", "1201"

	if len(splited) < 5 {
		return nil, errors.New("invalid url")
	}

	if splited[3] != "issues" {
		return nil, errors.New("not issue?")
	}

	owner := splited[1]
	repo := splited[2]
	is := splited[4]

	isNum, err := strconv.Atoi(is)
	if err != nil {
		return nil, errors.New("can't parse issue number")
	}

	return &GitHubIssue{Owner: owner, Repo: repo, Number: isNum}, nil
}

func (i *GitHubIssue) StarterTemplate() string {
	t := fmt.Sprintf(heredoc.Doc(`
	branch: %d_
	title: %s
	issue URL: %s

	---

	%s
	`), i.Number, i.Title, i.URL, i.Body)

	return t
}
