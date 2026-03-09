package keyring

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type fileBackend struct {
	path string
	mu   sync.Mutex
}

func newFileBackend(path string) *fileBackend {
	return &fileBackend{path: path}
}

func (b *fileBackend) Get(key string) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	data, err := b.read()
	if err != nil {
		return "", err
	}
	return data[key], nil
}

func (b *fileBackend) Set(key, value string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	data, err := b.read()
	if err != nil {
		return err
	}
	data[key] = value
	return b.write(data)
}

func (b *fileBackend) Delete(key string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	data, err := b.read()
	if err != nil {
		return err
	}
	delete(data, key)
	return b.write(data)
}

func (b *fileBackend) read() (map[string]string, error) {
	payload, err := os.ReadFile(b.path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}

	var data map[string]string
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, err
	}
	if data == nil {
		data = map[string]string{}
	}
	return data, nil
}

func (b *fileBackend) write(data map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(b.path), 0o755); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(b.path, payload, 0o600)
}
