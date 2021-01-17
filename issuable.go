package gitstart

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
)

type GitHubIssuable struct {
	Owner  string
	Repo   string
	Number int
}

type Issue struct {
	Owner  string
	Repo   string
	Number int
	URL    string
	Body   string
	Title  string
}

func ParseGitHubIssuable(str string) (*GitHubIssuable, error) {
	str = strings.TrimSpace(str)

	// 番号だけだった場合
	if i, err := strconv.Atoi(str); err == nil {
		return &GitHubIssuable{Number: i}, nil
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

	return &GitHubIssuable{Owner: owner, Repo: repo, Number: isNum}, nil
}

func StarterTemplate(issue *Issue) string {
	t := fmt.Sprintf(heredoc.Doc(`
	branch: %d_
	title: %s

	// Edit branch name to checkout. title will be used for PR title.
	// Original issue URL is %s

	---

	%s
	`), issue.Number, issue.Title, issue.URL, issue.Body)

	return t
}
