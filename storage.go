package gitstart

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/exp/utf8string"
)

type HistoryStorage struct {
	Content map[string]*StarterOption
	Path    string
}

func (s *HistoryStorage) Get(key string) *StarterOption {
	opt, ok := s.Content[key]
	if !ok {
		return nil
	}

	return opt
}

func (s *HistoryStorage) Set(key string, opt *StarterOption) error {
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

func NewHistoryStorage(storagePath string) (*HistoryStorage, error) {
	storage := &HistoryStorage{
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

func NewStarterOptionFromTemplate(template string) (*StarterOption, error) {
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
