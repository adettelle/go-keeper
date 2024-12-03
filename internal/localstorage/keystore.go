package localstorage

import (
	"github.com/99designs/keyring"
)

const (
	service = "gokeeper"
	user    = "default"
)

type IKeyStorage interface {
	Set(jwtToken string) error
	Get() (string, error)
}

type KeyStore struct {
	Config *keyring.Config
}

func NewKeyStore(config *keyring.Config) *KeyStore {
	return &KeyStore{
		Config: config,
	}
}

func (ks *KeyStore) Set(jwtToken string) error {
	ring, err := keyring.Open(*ks.Config)
	if err != nil {
		return err
	}

	err = ring.Set(keyring.Item{
		Key:  user,
		Data: []byte(jwtToken),
	})
	if err != nil {
		return err
	}
	return nil
}

func (ks *KeyStore) Get() (string, error) {
	ring, err := keyring.Open(*ks.Config)
	if err != nil {
		return "", err
	}

	i, err := ring.Get(user)
	if err != nil {
		return "", err
	}

	// log.Println("----", string(i.Data))
	return string(i.Data), nil
}
