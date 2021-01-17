package gitstart

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cli/cli/git"
	"github.com/pkg/browser"
)

type CLIMode int

const (
	Start CLIMode = iota
	PR
)

type CLI struct {
	configDir         string
	gitRemote         string
	currentBranch     string
	currentRepository Repository
	debug             bool
	stdin             io.Reader
	stdout            io.Writer
	stderr            io.Writer
	mode              CLIMode
}

// NewCLIはCLIの設定を初期化します.
func NewCLI(debug bool, stdin io.Reader, stdout io.Writer, stderr io.Writer, configDir string) (*CLI, error) {
	if configDir == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return nil, err
		}

		configDir = path.Join(dir, "git-start")
	}

	cli := &CLI{
		debug:     debug,
		stdin:     stdin,
		stdout:    stdout,
		stderr:    stderr,
		configDir: configDir,
	}

	// setup configdir
	if err := cli.setupConfigDir(); err != nil {
		return nil, err
	}

	remotes, err := git.Remotes()
	if err != nil {
		return nil, err
	}
	if len(remotes) == 0 {
		return nil, errors.New("no remotes found")
	}

	// first remote
	cli.gitRemote = remotes[0].Name
	remoteURL := remotes[0].FetchURL

	repo, err := GuessRepositoryFromRemoteURL(remoteURL)
	if err != nil {
		return nil, err
	}
	cli.currentRepository = *repo

	br, err := git.CurrentBranch()
	if err != nil {
		return nil, err
	}

	cli.currentBranch = br

	return cli, nil
}

func (cli *CLI) setupConfigDir() error {
	_, err := os.Stat(cli.configDir)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(cli.configDir, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

func (cli *CLI) RunStartCommand(issuable string) error {
	cli.mode = Start
	var template string

	// 引数に与えられたissue番号っぽいもの or issueのURL から実際にissueを取得する

	if cli.currentRepository.Hosting == GitHub {
		ghIssue, err := ParseGitHubIssuable(issuable)
		if err != nil {
			return err
		}

		// issue numだけ指定されてownerとrepoが分からなかった場合はgit remoteから推測しにいく
		if ghIssue.Owner == "" && ghIssue.Repo == "" {
			ghIssue.Owner = cli.currentRepository.Owner
			ghIssue.Repo = cli.currentRepository.Name
		}
		ctx := context.Background()

		foundIssue, err := FetchGitHubIssue(ctx, ghIssue)
		if err != nil {
			return err
		}

		template = StarterTemplate(foundIssue)
	} else {
		// unreachable
		panic("not github")
	}

	// 取得したissueをエディタで開きテンプレートに入力させる
	editedTemplate, err := CaptureInputFromEditor(template)
	if err != nil {
		return err
	}

	opt, err := NewStarterOptionFromTemplate(editedTemplate)
	if err != nil {
		return err
	}
	opt.BaseBranch = cli.currentBranch

	optStorage, err := cli.historyStorage()
	if err != nil {
		return err
	}

	cmd, err := git.GitCommand("switch", "-c", opt.SwitchBranch)
	if err != nil {
		return err
	}
	cmd.Stdout = cli.stdout
	cmd.Stderr = cli.stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	// branch switched
	cli.currentBranch = opt.SwitchBranch

	// save PR title and option for later use
	key := cli.starterOptionKey()
	if err := optStorage.Set(key, opt); err != nil {
		return err
	}

	return nil
}

func (cli *CLI) RunPRCommand() error {
	cli.mode = PR

	optStorage, err := cli.historyStorage()
	if err != nil {
		return err
	}

	key := cli.starterOptionKey()

	starterOption := optStorage.Get(key)
	if starterOption == nil {
		// not found. fallback?
		return errors.New("git-start history not found. did you exec git-start on this branch?")
	}

	// first, push
	cmd, err := git.GitCommand("push", "-u", "origin", cli.currentBranch)
	if err != nil {
		return err
	}
	cmd.Stdout = cli.stdout
	cmd.Stderr = cli.stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	url := cli.newPullRequestURL(starterOption)
	return browser.OpenURL(url)
}

func (c *CLI) historyStorage() (*HistoryStorage, error) {
	path := filepath.Join(c.configDir, "starter-option-storage.json")

	return NewHistoryStorage(path)
}

func (c *CLI) starterOptionKey() string {
	ky := []string{
		c.currentRepository.Hosting,
		c.currentRepository.Owner,
		c.currentRepository.Name,
		c.currentBranch,
	}

	return strings.Join(ky, "/")
}

func (cli *CLI) newPullRequestURL(opt *StarterOption) string {
	title := url.QueryEscape(opt.PullRequestTitle)

	url := fmt.Sprintf(
		"https://github.com/%s/%s/compare/%s...%s?quick_pull=1&title=%s",
		cli.currentRepository.Owner,
		cli.currentRepository.Name,
		opt.BaseBranch,
		cli.currentBranch,
		title,
	)

	return url
}

type StarterOption struct {
	SwitchBranch     string `json:"switchBranch"`
	PullRequestTitle string `json:"pullRequestTitle"`
	BaseBranch       string `json:"baseBranch"`
}
