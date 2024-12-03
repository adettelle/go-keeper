package client

import (
	"context"
	"net/http"

	"github.com/adettelle/go-keeper/cmd/settings"
	"github.com/adettelle/go-keeper/internal/localstorage"
	"github.com/carlmjohnson/requests"
)

type UserService struct {
	transport *http.Transport
	keyStore  localstorage.IKeyStorage
}

func NewUserService(transport *http.Transport, keyStore localstorage.IKeyStorage) *UserService {
	return &UserService{
		transport: transport,
		keyStore:  keyStore,
	}
}

type CustomerToReg struct {
	Name           string `json:"name"`
	Login          string `json:"login"`
	MasterPassword string `json:"masterpassword"`
	Authentication bool   `json:"signin"`
}

func (us *UserService) Register(name, login, masterPassword string, signIn bool) error {
	customer := CustomerToReg{
		Name:           name,
		Login:          login,
		MasterPassword: masterPassword,
		Authentication: signIn,
	}

	err := requests.
		URL("/api/user/register").
		CheckStatus(http.StatusOK).
		Host(settings.ServerURL).
		Scheme("https").
		Transport(us.transport).
		Method(http.MethodPost).
		BodyJSON(&customer).
		Fetch(context.Background())

	if err != nil {
		return err
	}

	if customer.Authentication {
		err := us.LogIn(customer.Login, customer.MasterPassword)
		if err != nil {
			return err
		}
	}
	return nil
}

type CustomerToLogin struct {
	Login    string `json:"login"`
	Password string `json:"pwd"`
}

func (us *UserService) LogIn(login, password string) error {
	customer := CustomerToLogin{
		Login:    login,
		Password: password,
	}

	headers := http.Header{}

	err := requests.
		URL("/api/user/login").
		CopyHeaders(headers).
		CheckStatus(http.StatusOK).
		Host(settings.ServerURL).
		Scheme("https").
		Transport(us.transport).
		BodyJSON(&customer).
		Fetch(context.Background())

	if err != nil {
		return err
	}

	jwtToken := headers.Get("Authorization")
	err = us.keyStore.Set(jwtToken)
	if err != nil {
		return err
	}
	return nil
}
