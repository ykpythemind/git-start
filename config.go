package main

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cli/cli/git"
)

type Config struct {
	Dir               string
	Remote            string
	CurrentBranch     string
	CurrentRepository Repository
}

// SetupConfigはCLIの設定を初期化します.
func NewConfig() (*Config, error) {
	confdir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	conf := &Config{}
	conf.Dir = path.Join(confdir, "git-start")

	remotes, err := git.Remotes()
	if err != nil {
		return nil, err
	}
	if len(remotes) == 0 {
		return nil, errors.New("no remotes found")
	}

	// first remote
	conf.Remote = remotes[0].Name
	remoteURL := remotes[0].FetchURL

	repo, err := GuessRepositoryFromRemoteURL(remoteURL)
	if err != nil {
		return nil, err
	}
	conf.CurrentRepository = *repo

	b, err := git.CurrentBranch()
	if err != nil {
		return nil, err
	}

	conf.CurrentBranch = b

	return conf, nil
}

func SetupConfigDir(c *Config) error {
	_, err := os.Stat(c.Dir)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(c.Dir, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) StarterOptionStoragePath() string {
	return filepath.Join(c.Dir, "starter-option-storage.json")
}

func (c *Config) StarterOptionKey(branch string) string {
	ky := []string{
		c.CurrentRepository.Hosting,
		c.CurrentRepository.Owner,
		c.CurrentRepository.Name,
		branch,
	}

	return strings.Join(ky, "/")
}
