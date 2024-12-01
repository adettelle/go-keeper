package localstorage

import (
	// "github.com/zalando/go-keyring"
	"fmt"

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

// func Set(jwtToken string) error {
// 	err := keyring.Set(service, user, jwtToken)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	return nil
// }

// func Get() (string, error) {
// 	jwtToken, err := keyring.Get(service, user)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	return jwtToken, nil
// }

func Set(jwtToken string) error {
	ring, err := keyring.Open(keyring.Config{
		ServiceName:      service,
		AllowedBackends:  nil, // []keyring.BackendType{keyring.FileBackend}
		FilePasswordFunc: keyring.TerminalPrompt,
		// AllowedBackends: []keyring.BackendType{keyring.KWalletBackend},
		FileDir: "~/",
	})
	if err != nil {
		return err
	}

	fmt.Println("!!!!!!!", ring)
	err = ring.Set(keyring.Item{
		Key:  user,
		Data: []byte(jwtToken),
	})
	if err != nil {
		return err
	}
	return nil
}

func Get() (string, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: service,
		// TODO на крайний случай
		AllowedBackends:  nil, // []keyring.BackendType{keyring.FileBackend}
		FilePasswordFunc: keyring.TerminalPrompt,
		FileDir:          "~/",
	})
	if err != nil {
		return "", err
	}

	i, err := ring.Get(user)
	if err != nil {
		return "", err
	}

	fmt.Printf("%s", i.Data)
	return string(i.Data), nil
}
