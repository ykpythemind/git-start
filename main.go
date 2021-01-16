package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/go-git/go-git/v5"

	"github.com/MakeNowJust/heredoc/v2"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"golang.org/x/exp/utf8string"
)

type StarterConfig struct {
	Strategy         string
	BaseBranch       string
	IssueConfig      string
	switchBranch     string
	pullRequestTitle string
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
	r, err := git.PlainOpen("./")
	if err != nil {
		return err
	}

	remotes, err := r.Remotes()
	if err != nil {
		return err
	}
	if len(remotes) < 1 {
		return errors.New("remote is not exist")
	}
	remote := remotes[0] // first remote

	remoteEndpoint, err := transport.NewEndpoint(remote.Config().URLs[0])
	if err != nil {
		return err
	}

	config := &StarterConfig{}

	issuable := ""

	if len(args) == 2 {
		// 第一引数
		issuable = args[1]
	}

	ghIssue, err := ParseGitHubIssuable(issuable)
	if err != nil {
		return err
	}

	// issue numだけ指定されてownerとrepoが分からなかった場合はgit remoteから推測しにいく
	if ghIssue.Owner == "" && ghIssue.Repo == "" {
		owner, repo, err := extractRepositoryPath(remoteEndpoint.Path)
		if err != nil {
			return err
		}
		ghIssue.Owner = owner
		ghIssue.Repo = repo
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

	ctx := context.Background()
	client, err := NewGitHubClient(ctx)
	if err != nil {
		return err
	}

	is, res, err := client.Issues.Get(ctx, ghIssue.Owner, ghIssue.Repo, ghIssue.Number)
	if err != nil {
		return err
	}

	_ = res

	template := fmt.Sprintf(heredoc.Doc(`
	branch:
	title: %s
	base issue: %s

	---

	%s
	`), *is.Title, *is.HTMLURL, *is.Body)

	editedTemplate, err := CaptureInputFromEditor(template)
	if err != nil {
		return err
	}

	err = parseTemplate(editedTemplate, config)
	if err != nil {
		return err
	}

	cmd, err := GitCommand("switch", "-c", config.switchBranch)
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	cmd, err = GitCommand("status")
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func parseTemplate(template string, config *StarterConfig) error {
	scanner := bufio.NewScanner(strings.NewReader(template))
	scanner.Split(bufio.ScanLines)

	title := ""
	titleFound := false
	branch := ""
	branchFound := false

	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if !titleFound && strings.HasPrefix(text, "title:") {
			title = strings.TrimSpace(text[6:])
			titleFound = true
		}

		if !branchFound && strings.HasPrefix(text, "branch:") {
			branch = strings.TrimSpace(text[7:])
			branchFound = true
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if branch == "" {
		return errors.New("branch is not specified")
	}

	// validate branch name
	utf8str := utf8string.NewString(branch)
	if !utf8str.IsASCII() {
		return fmt.Errorf("invalid branch name: %s. only ascii code is allowed", branch)
	}

	// todo: type Template struct
	config.pullRequestTitle = title
	config.switchBranch = branch

	return nil
}

func OpenFileInEditor(editor string, filename string) error {
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "vim"
	}

	executable, err := exec.LookPath(editor)
	if err != nil {
		return err
	}

	cmd := exec.Command(executable, filename)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func CaptureInputFromEditor(content string) (string, error) {
	file, err := ioutil.TempFile(os.TempDir(), "*")
	if err != nil {
		return "", err
	}

	if _, err := file.Write([]byte(content)); err != nil {
		return "", err
	}

	filename := file.Name()

	defer os.Remove(filename)

	if err = file.Close(); err != nil {
		return "", err
	}

	if err = OpenFileInEditor("", filename); err != nil {
		return "", err
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
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

func GitCommand(args ...string) (*exec.Cmd, error) {
	gitExe, err := exec.LookPath("git")
	if err != nil {
		return nil, err
	}

	return exec.Command(gitExe, args...), nil
}

// strategy
//   - github flow
// main branch
//   - master

// Issuable
//   issueっぽいやつ urlないしはissue number
//   ParseIssuable
// Template
//   作成ブランチ名, PRタイトルを決める
//   ParseTemplate
// StarterConfig
// GitConfig
//   localからIssueBaseとかをとってくる
// Starter
//   Configを元にgit startを実行
// PullRequestStarter
