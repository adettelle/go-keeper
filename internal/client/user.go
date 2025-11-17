// Package client provides functionality for managing card, file, password and user information through
// HTTP requests, including operations to add, retrieve, update, and delete.
package client

import (
	"context"
	"log"
	"net/http"

	"github.com/adettelle/go-keeper/cmd/settings"
	"github.com/adettelle/go-keeper/internal/localstorage"
	"github.com/carlmjohnson/requests"
)

// UserService is a service for managing user-related operations.
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
	Name           string `json:"name" validate:"required,min=1"`
	Login          string `json:"login" validate:"required,email"`
	MasterPassword string `json:"masterpassword" validate:"required,min=3"`
	Authentication bool   `json:"signin"`
}

// Register registers a new user with the provided credentials and optionally logs in the user.
func (us *UserService) Register(name, login, masterPassword string, signIn bool) error {
	customer := CustomerToReg{
		Name:           name,
		Login:          login,
		MasterPassword: masterPassword,
		Authentication: signIn,
	}

	err := validate.Struct(customer)
	if err != nil {
		log.Println("error in validating:", err)
		return err
	}

	err = requests.
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
	Login    string `json:"login" validate:"required,email"`
	Password string `json:"pwd" validate:"required,min=3"`
}

func (us *UserService) LogIn(login, password string) error {
	customer := CustomerToLogin{
		Login:    login,
		Password: password,
	}

	err := validate.Struct(customer)
	if err != nil {
		log.Println("error in validating:", err)
		return err
	}

	headers := http.Header{}

	err = requests.
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
