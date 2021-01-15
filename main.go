package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"
)

type Config struct {
	Strategy    string
	BaseBranch  string
	IssueConfig string
}

func main() {
	err := run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}

func run(args []string) error {
	issuable := ""

	if len(args) == 2 {
		// 第一引数
		issuable = args[1]
	}

	isnum := 0
	i, err := strconv.Atoi(issuable)
	if err == nil {
		isnum = i
	} else {
		fmt.Println("issue num is not valid: %s", err)
	}

	debug := false
	if os.Getenv("DEBUG") != "" {
		debug = true
		fmt.Println("debug")
	}

	if debug {
		err := os.Chdir("/Users/ykpythemind/git/github.com/ykpythemind/sandbox")
		if err != nil {
			return err
		}
	}

	_, err = git.PlainOpen("./")
	if err != nil {
		return err
	}

	// w, err := r.Worktree()
	// if err != nil {
	// 	return err
	// }

	// opt := &git.PullOptions{Auth:}
	// opt.Validate()

	// err = w.Pull(opt)
	// if err != nil {
	// 	return err
	// }

	f, err := os.Open("/Users/ykpythemind/.git-brws-token")
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: string(bytes.TrimSpace(b))},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	is, res, err := client.Issues.Get(ctx, "coubic", "coubic-issues", isnum)
	if err != nil {
		return err
	}

	_ = res

	fmt.Println(*is.Title)
	fmt.Println(*is.Body)

	return nil
}

// strategy
//   - github flow
// main branch
//   - master
