package gitstart

import (
	"encoding/json"
	"os"
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
