package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"

	"github.com/MakeNowJust/heredoc/v2"
)

type Config struct {
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
	config := &Config{}

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

	template := fmt.Sprintf(heredoc.Doc(`
	branch: xxx
	title: %s

	---

	%s
	`), *is.Title, *is.Body)

	editedTemplate, err := CaptureInputFromEditor(template)
	if err != nil {
		return err
	}

	err = parseTemplate(editedTemplate, config)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", config)

	return nil
}

func parseTemplate(template string, config *Config) error {
	scanner := bufio.NewScanner(strings.NewReader(template))
	scanner.Split(bufio.ScanLines)

	title := ""
	titleFound := false
	branch := ""
	branchFound := false

	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if !titleFound && strings.HasPrefix(text, "title:") {
			title = text[6:]
			titleFound = true
		}

		if !branchFound && strings.HasPrefix(text, "branch:") {
			branch = text[7:]
			branchFound = true
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if branch == "" {
		return errors.New("branch is not specified")
	}

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

	// Defer removal of the temporary file in case any of the next steps fail.
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

// strategy
//   - github flow
// main branch
//   - master
