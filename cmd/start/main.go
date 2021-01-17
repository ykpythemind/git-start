package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/cli/cli/git"

	"github.com/pkg/browser"

	"github.com/ykpythemind/git-start"
)

func main() {
	err := run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}

func run(args []string) error {
	debug := false
	if os.Getenv("DEBUG") != "" {
		debug = true
		fmt.Println("debug")
	}

	_ = debug

	if len(args) == 1 {
		// todo: showhelp
		return errors.New("arg is invalid. issue is required")
	}

	config, err := gitstart.NewConfig()
	if err != nil {
		return err
	}
	if err := gitstart.SetupConfigDir(config); err != nil {
		return err
	}

	if args[1] == "pr" || args[1] == "pull-request" {
		return runPRCommand(config)
	}

	issuable := ""

	if len(args) == 2 {
		issuable = args[1]
	}

	var template string

	if config.CurrentRepository.Hosting == gitstart.GitHub {

		ghIssue, err := gitstart.ParseGitHubIssuable(issuable)
		if err != nil {
			return err
		}

		// issue numだけ指定されてownerとrepoが分からなかった場合はgit remoteから推測しにいく
		if ghIssue.Owner == "" && ghIssue.Repo == "" {
			ghIssue.Owner = config.CurrentRepository.Owner
			ghIssue.Repo = config.CurrentRepository.Name
		}
		ctx := context.Background()

		foundIssue, err := gitstart.FetchGitHubIssue(ctx, ghIssue)
		if err != nil {
			return err
		}

		template = gitstart.StarterTemplate(foundIssue)
	} else {
		// unreachable
		panic("not github")
	}

	editedTemplate, err := gitstart.CaptureInputFromEditor(template)
	if err != nil {
		return err
	}

	opt, err := gitstart.ParseStarterTemplate(editedTemplate)
	if err != nil {
		return err
	}
	opt.BaseBranch = config.CurrentBranch

	if err := gitstart.Start(config, opt); err != nil {
		return err
	}

	return nil
}

func runPRCommand(config *gitstart.Config) error {
	currentBranch := config.CurrentBranch

	optStorage, err := gitstart.NewStarterOptionStorage(config.StarterOptionStoragePath())
	if err != nil {
		return err
	}

	key := config.StarterOptionKey(currentBranch)

	starterOption := optStorage.Fetch(key)
	if starterOption == nil {
		// not found. fallback?
		return errors.New("git-start history not found. did you exec git-start on this branch?")
	}

	// first, push
	cmd, err := git.GitCommand("push", "-u", "origin", currentBranch)
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	title := url.QueryEscape(starterOption.PullRequestTitle)

	url := fmt.Sprintf(
		"https://github.com/%s/%s/compare/%s...%s?quick_pull=1&title=%s",
		config.CurrentRepository.Owner,
		config.CurrentRepository.Name,
		starterOption.BaseBranch,
		config.CurrentBranch,
		title,
	)

	return browser.OpenURL(url)
}
