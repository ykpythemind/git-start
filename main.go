package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/go-git/go-git/v5"

	"github.com/MakeNowJust/heredoc/v2"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"golang.org/x/exp/utf8string"

	"github.com/pkg/browser"
)

type StarterConfig struct {
	switchBranch     string
	pullRequestTitle string
}

func (s *StarterConfig) Save() error {
	configDir := s.configDir()

	_, err := os.Stat(configDir)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(s.savePath(), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err := f.Write([]byte(s.pullRequestTitle)); err != nil {
		return err
	}

	return nil
}

func (s *StarterConfig) configDir() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return path.Join(homedir, ".config", "git-start")
}

func (s *StarterConfig) savePath() string {
	// todo: branch名に/とか入ると死ぬ？
	return path.Join(s.configDir(), s.switchBranch)
}

func newStarterConfigFromFile(branch string) (*StarterConfig, error) {
	conf := &StarterConfig{switchBranch: branch}

	f, err := os.Open(conf.savePath())
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	// 今はprタイトルしか保存しない. 構造化して保存しておいたほうがいいかも
	conf.pullRequestTitle = string(b)

	return conf, nil
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
	if len(args) == 1 {
		return errors.New("arg is invalid. issue is required")
	}

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

	if args[1] == "pr" || args[1] == "pull-request" {
		owner, repo, err := extractRepositoryPath(remoteEndpoint.Path)
		if err != nil {
			return err
		}
		return runPR(r, owner, repo)
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
		if res.StatusCode == 404 {
			return fmt.Errorf("issue %d is not found", ghIssue.Number)
		}
		return err
	}

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

	// save PR title for later use
	if err := config.Save(); err != nil {
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

func runPR(r *git.Repository, owner string, repo string) error {
	ref, err := r.Head()
	if err != nil {
		return err
	}

	if !ref.Name().IsBranch() {
		return errors.New("current head is not branch")
	}

	branch := ref.Name().String()
	branch = strings.Replace(branch, "refs/heads/", "", 1)

	// first, push
	cmd, err := GitCommand("push", "-u", "origin")
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	conf, err := newStarterConfigFromFile(branch)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// todo: fallback
		}

		return err
	}

	title := url.QueryEscape(conf.pullRequestTitle)

	// fixme: base branchを記録する必要あり
	url := fmt.Sprintf("https://github.com/%s/%s/compare/main...%s?quick_pull=1&title=%s", owner, repo, branch, title)

	return browser.OpenURL(url)
}
