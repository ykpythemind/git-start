package main

import (
	"errors"
	"net/url"
	"strings"
)

type HostingService = string

var (
	GitHub HostingService = "GitHub"
)

type Repository struct {
	Hosting HostingService
	Owner   string
	Name    string
}

func GuessRepositoryFromRemoteURL(remoteURL *url.URL) (*Repository, error) {
	r := &Repository{}

	if strings.Contains(remoteURL.Host, "github.com") {
		r.Hosting = GitHub
	} else {
		return nil, errors.New("only github.com is supported now")
	}

	path := remoteURL.Path

	owner, reponame, err := extractRepositoryPath(path)
	if err != nil {
		return nil, err
	}

	r.Owner = owner
	r.Name = reponame

	return r, nil
}

func extractRepositoryPath(str string) (owner, repo string, err error) {
	// "/ykpythemind/git-start.git" => owner: ykpythemind, repo: git-start
	str = strings.TrimPrefix(str, "/")

	sp := strings.Split(str, "/")

	if len(sp) != 2 {
		return "", "", errors.New("fail to extract repository like string")
	}

	owner = sp[0]
	repo = strings.TrimSuffix(sp[1], ".git")

	return
}
