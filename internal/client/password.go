// Package client provides functionality for managing card, file, password and user information through
// HTTP requests, including operations to add, retrieve, update, and delete.
package client

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/adettelle/go-keeper/cmd/settings"
	"github.com/adettelle/go-keeper/internal/localstorage"
	"github.com/carlmjohnson/requests"
	"github.com/jedib0t/go-pretty/v6/table"
)

// PasswordService is a service for managing password-related operations.
type PasswordService struct {
	transport *http.Transport
	keyStore  localstorage.IKeyStorage
}

func NewPasswordService(transport *http.Transport, keyStore localstorage.IKeyStorage) *PasswordService {
	return &PasswordService{
		transport: transport,
		keyStore:  keyStore,
	}
}

func (ps *PasswordService) GetPasswordByTitle(title string) error {
	jwtToken, err := ps.keyStore.Get()
	if err != nil {
		return err
	}

	var pwd string

	err = requests.
		URL("/api/user/password/"+title).
		Host(settings.ServerURL).
		Scheme("https").
		Transport(ps.transport).
		Header("Authorization", string(jwtToken)).
		Method(http.MethodGet).
		ToString(&pwd).
		Fetch(context.Background())

	if err != nil {
		return err
	}
	fmt.Printf("%s\n", pwd)
	return nil
}

type PasswordToGet struct {
	// ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// AllPass retrieves all passwords' info (only title and description) associated with the user.
// It fetches password data from the server and displays it in a tabular format.
func (ps *PasswordService) AllPass() error {
	jwtToken, err := ps.keyStore.Get()
	if err != nil {
		return err
	}

	var pwds []PasswordToGet

	err = requests.
		URL("/api/user/passwords").
		Host(settings.ServerURL).
		Scheme("https").
		Transport(ps.transport).
		Header("Authorization", string(jwtToken)).
		ToJSON(&pwds).
		Method(http.MethodGet).
		Fetch(context.Background())

	if err != nil {
		return err
	} else {
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Title", "Description"})

		for _, pwd := range pwds {
			t.AppendRow([]interface{}{pwd.Title, pwd.Description})
		}
		t.Render()
	}
	return nil
}

type PasswordToUpdate struct {
	Password    string `json:"pwd,omitempty"`
	Description string `json:"description,omitempty"`
}

// UpdateCard updates password itself and description by unique title.
// It updates only arguments which are provided.
func (ps *PasswordService) UpdatePassword(title string, args ...string) error {
	jwtToken, err := ps.keyStore.Get()
	if err != nil {
		return err
	}

	pwd := PasswordToUpdate{
		Password:    args[0],
		Description: args[1],
	}

	err = requests.
		URL("/api/user/password/update/"+title).
		Host(settings.ServerURL).
		Scheme("https").
		Transport(ps.transport).
		Header("Authorization", string(jwtToken)).
		BodyJSON(&pwd).
		Method(http.MethodPost).
		Fetch(context.Background())

	if err != nil {
		return err
	} else {
		log.Println("Password info is updated.")
	}
	return nil
}

func (ps *PasswordService) DeletePasswordByTitle(title string) error {
	jwtToken, err := ps.keyStore.Get()
	if err != nil {
		return err
	}

	err = requests.
		URL("/api/user/password/"+title).
		Host(settings.ServerURL).
		Scheme("https").
		Transport(ps.transport).
		Header("Authorization", string(jwtToken)).
		Method(http.MethodDelete).
		Fetch(context.Background())

	if err != nil {
		return err
	} else {
		log.Println("Password is deleted.")
	}
	return nil
}

type PwdToAdd struct {
	Password    string `json:"pwd" validate:"required,min=1"`
	Title       string `json:"title" validate:"required,min=1"`
	Description string `json:"description"`
}

func (ps *PasswordService) AddPassword(password, title, description string) error {
	pwdToAdd := PwdToAdd{
		Password:    password,
		Title:       title,
		Description: description,
	}

	err := validate.Struct(pwdToAdd)
	if err != nil {
		log.Println("error in validating:", err)
		return err
	}

	jwtToken, err := ps.keyStore.Get()
	if err != nil {
		return err
	}

	err = requests.
		URL("/api/user/password").
		Host(settings.ServerURL).
		Scheme("https").
		Header("Authorization", string(jwtToken)).
		Transport(ps.transport).
		BodyJSON(&pwdToAdd).
		Method(http.MethodPut).
		Fetch(context.Background())

	if err != nil {
		return err
	} else {
		log.Println("Password is added.")
	}
	return nil
}
