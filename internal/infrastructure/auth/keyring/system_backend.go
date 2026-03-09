package keyring

import (
	"errors"

	keyringlib "github.com/zalando/go-keyring"
)

type systemBackend struct {
	service string
}

func (b *systemBackend) Get(key string) (string, error) {
	value, err := keyringlib.Get(b.service, key)
	if err != nil {
		if errors.Is(err, keyringlib.ErrNotFound) {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

func (b *systemBackend) Set(key, value string) error {
	return keyringlib.Set(b.service, key, value)
}

func (b *systemBackend) Delete(key string) error {
	err := keyringlib.Delete(b.service, key)
	if errors.Is(err, keyringlib.ErrNotFound) {
		return nil
	}
	return err
}
