// Package localstorage provides functionality for securely storing and retrieving JWT tokens
// using the `keyring` library. It abstracts the keyring operations via an interface for flexibility.
package localstorage

import (
	"github.com/99designs/keyring"
)

const (
	service = "gokeeper" // service defines the name of the keyring service used for storing the token.
	user    = "default"  // user defines the default keyring item key under which the token is stored.
)

// IKeyStorage defines an interface for managing JWT token storage.
// It provides methods to set and get the token securely.
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

// Set securely stores the provided JWT token in the keyring.
//
// Parameters:
//   - jwtToken: The JWT token string to be stored.
//
// Returns:
//   - An error if the operation fails, otherwise nil.
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

// Get retrieves the JWT token securely stored in the keyring.
//
// Returns:
//   - The JWT token as a string.
//   - An error if the operation fails or the token is not found.
func (ks *KeyStore) Get() (string, error) {
	ring, err := keyring.Open(*ks.Config)
	if err != nil {
		return "", err
	}

	i, err := ring.Get(user)
	if err != nil {
		return "", err
	}

	return string(i.Data), nil
}
