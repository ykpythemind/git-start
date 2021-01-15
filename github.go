package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// todo: make interface
func FetchGitHubIssue(ctx context.Context, ghIssue *GitHubIssue) (*GitHubIssue, error) {
	client, err := NewGitHubClient(ctx)
	if err != nil {
		return nil, err
	}

	is, res, err := client.Issues.Get(ctx, ghIssue.Owner, ghIssue.Repo, ghIssue.Number)
	if err != nil {
		if res.StatusCode == 404 {
			return nil, fmt.Errorf("issue %d is not found", ghIssue.Number)
		}
		return nil, err
	}

	return &GitHubIssue{
		Owner:  ghIssue.Owner,
		Repo:   ghIssue.Repo,
		Number: ghIssue.Number,
		URL:    is.GetHTMLURL(),
		Body:   is.GetBody(),
		Title:  is.GetTitle(),
	}, nil
}

func NewGitHubClient(ctx context.Context) (*github.Client, error) {
	ghToken, err := fetchGitHubToken()
	if err != nil {
		return nil, err
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ghToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc), nil
}

func fetchGitHubToken() (string, error) {
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t, nil
	}
	if t := os.Getenv("GIT_START_GITHUB_TOKEN"); t != "" {
		return t, nil
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	f, err := os.Open(path.Join(homedir, ".git-start-token"))
	if err != nil {
		return "", err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(b)), nil
}
