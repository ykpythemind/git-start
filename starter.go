package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cli/cli/git"
	"golang.org/x/exp/utf8string"
)

type StarterOption struct {
	SwitchBranch     string `json:"switchBranch"`
	PullRequestTitle string `json:"pullRequestTitle"`
	BaseBranch       string `json:"baseBranch"`
}

type StarterOptionStorage struct {
	Content map[string]*StarterOption
	Path    string
}

func (s *StarterOptionStorage) fetch(key string) *StarterOption {
	opt, ok := s.Content[key]
	if !ok {
		return nil
	}

	return opt
}

func (s *StarterOptionStorage) write(key string, opt *StarterOption) error {
	s.Content[key] = opt

	f, err := os.OpenFile(s.Path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(s.Content); err != nil {
		return err
	}

	return nil
}

func NewStarterOptionStorage(storagePath string) (*StarterOptionStorage, error) {
	storage := &StarterOptionStorage{
		Path:    storagePath,
		Content: make(map[string]*StarterOption),
	}

	var storageFile *os.File
	newFile := false

	_, err := os.Stat(storage.Path)

	if os.IsNotExist(err) {
		newFile = true

		f, err := os.Create(storage.Path)
		if err != nil {
			return nil, err
		}
		storageFile = f

	} else {
		f, err := os.Open(storage.Path)
		if err != nil {
			return nil, err
		}
		storageFile = f
	}

	defer storageFile.Close()

	if !newFile {
		if err := json.NewDecoder(storageFile).Decode(&storage.Content); err != nil {
			// dirty: 強制初期化
			_, err = storageFile.Write([]byte("{}"))
			if err != nil {
				return nil, err
			}
		}
	}

	return storage, nil
}

func parseStarterTemplate(template string) (*StarterOption, error) {
	starterOption := &StarterOption{}

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
		return nil, err
	}

	if branch == "" {
		return nil, errors.New("branch is not specified")
	}

	// validate branch name
	utf8str := utf8string.NewString(branch)
	if !utf8str.IsASCII() {
		return nil, fmt.Errorf("invalid branch name: %s. only ascii code is allowed", branch)
	}

	starterOption.PullRequestTitle = title
	starterOption.SwitchBranch = branch

	return starterOption, nil
}

func Start(config *Config, opt *StarterOption) error {
	optStorage, err := NewStarterOptionStorage(config.StarterOptionStoragePath())
	if err != nil {
		return err
	}

	cmd, err := git.GitCommand("switch", "-c", opt.SwitchBranch)
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	// save PR title and option for later use
	key := config.StarterOptionKey(opt.SwitchBranch)
	if err := optStorage.write(key, opt); err != nil {
		return err
	}

	return nil
}
