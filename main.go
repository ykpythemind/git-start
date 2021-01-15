package main

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
)

type Config struct {
	Strategy    string
	BaseBranch  string
	IssueConfig string
}

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}

func run() error {
	debug := false
	if os.Getenv("DEBUG") != "" {
		debug = true
	}

	if debug {
		err := os.Chdir("/Users/ykpythemind/git/github.com/ykpythemind/sandbox")
		if err != nil {
			return err
		}
	}

	r, err := git.PlainOpen("./")
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	err = w.Pull(&git.PullOptions{})
	if err != nil {
		return err
	}

	return nil
}

// strategy
//   - github flow
// main branch
//   - master
