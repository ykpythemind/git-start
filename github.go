package main

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

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
