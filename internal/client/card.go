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
	"github.com/go-playground/validator/v10"
	"github.com/jedib0t/go-pretty/v6/table"
)

// use a single instance of Validate, it caches struct info
var validate *validator.Validate = validator.New(validator.WithRequiredStructEnabled())

// CardService is a service for managing card-related operations.
type CardService struct {
	transport *http.Transport
	keyStore  localstorage.IKeyStorage
}

func NewCardService(transport *http.Transport, keyStore localstorage.IKeyStorage) *CardService {
	return &CardService{
		transport: transport,
		keyStore:  keyStore,
	}
}

type CardToGet struct {
	Title       string `json:"title"`
	Num         string `json:"num"`
	Description string `json:"description"`
}

// AllCards retrieves all cards associated with the user.
// It fetches card data from the server and displays it in a tabular format.
func (cs *CardService) AllCards() error {
	jwtToken, err := cs.keyStore.Get()
	if err != nil {
		return err
	}

	var cards []CardToGet

	err = requests.
		URL("/api/user/cards").
		Host(settings.ServerURL).
		Scheme("https").
		Transport(cs.transport).
		Header("Authorization", string(jwtToken)).
		ToJSON(&cards).
		Method(http.MethodGet).
		Fetch(context.Background())

	if err != nil {
		return err
	} else {
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Title", "Number", "Description"})

		for _, card := range cards {
			t.AppendRow([]interface{}{card.Title, card.Num, card.Description})
		}
		t.Render()
	}
	return nil
}

type CardToAdd struct {
	Num         string `json:"num" validate:"required,credit_card,len=16"`
	Expire      string `json:"expires_at" validate:"required,numeric,len=4"`
	Cvc         string `json:"cvc" validate:"required,numeric,len=3"`
	Title       string `json:"title" validate:"required,min=1"`
	Description string `json:"description"`
}

func (cs *CardService) AddCard(num, expire, cvc, title, description string) error {
	cardToAdd := CardToAdd{
		Num:         num,
		Expire:      expire,
		Cvc:         cvc,
		Title:       title,
		Description: description,
	}

	err := validate.Struct(cardToAdd)
	if err != nil {
		log.Println("error in validating:", err)
		return err
	}

	jwtToken, err := cs.keyStore.Get()
	if err != nil {
		return err
	}

	err = requests.
		URL("/api/user/card").
		Host(settings.ServerURL).
		Scheme("https").
		Header("Authorization", string(jwtToken)).
		Transport(cs.transport).
		BodyJSON(&cardToAdd).
		Method(http.MethodPut).
		Fetch(context.Background())

	if err != nil {
		return err
	} else {
		log.Println("Card is added.")
	}

	return nil
}

func (cs *CardService) DeleteCardByTitle(cardTitle string) error {
	jwtToken, err := cs.keyStore.Get()
	if err != nil {
		return err
	}

	err = requests.
		URL("/api/user/card/"+cardTitle).
		Host(settings.ServerURL).
		Scheme("https").
		Transport(cs.transport).
		Header("Authorization", string(jwtToken)).
		Method(http.MethodDelete).
		Fetch(context.Background())

	if err != nil {
		return err
	} else {
		log.Println("Card is deleted.")
	}
	return nil
}

type CardToGetByTitle struct {
	Num         string `json:"num"`
	Expire      string `json:"expires_at"`
	Cvc         string `json:"cvc"`
	Description string `json:"description"`
}

func (cs *CardService) GetCardByTitle(title string) error {
	jwtToken, err := cs.keyStore.Get()
	if err != nil {
		return err
	}

	var card CardToGetByTitle

	err = requests.
		URL("/api/user/card/"+title).
		Host(settings.ServerURL).
		Scheme("https").
		Transport(cs.transport).
		Header("Authorization", string(jwtToken)).
		Method(http.MethodGet).
		ToJSON(&card).
		Fetch(context.Background())

	if err != nil {
		return err
	}
	fmt.Printf("%s, %s, %s, %s\n", card.Num, card.Expire, card.Cvc, card.Description)
	return nil
}

type CardToUpdate struct {
	Num         string `json:"num,omitempty" validate:"omitnil,omitempty,credit_card,len=16"`
	Expire      string `json:"expires_at,omitempty"  validate:"omitempty,omitnil,numeric,len=4"`
	Cvc         string `json:"cvc,omitempty" validate:"omitempty,omitnil,numeric,len=3"`
	Description string `json:"description,omitempty"`
}

// UpdateCard updates card's number, date of expire, cvc and description by unique title.
// It updates only arguments which are provided.
func (ps *CardService) UpdateCard(title string, args ...string) error {
	jwtToken, err := ps.keyStore.Get()
	if err != nil {
		return err
	}

	card := CardToUpdate{
		Num:         args[0],
		Expire:      args[1],
		Cvc:         args[2],
		Description: args[3],
	}

	err = validate.Struct(card)
	if err != nil {
		log.Println("error in validating:", err)
		return err
	}

	err = requests.
		URL("/api/user/card/update/"+title).
		Host(settings.ServerURL).
		Scheme("https").
		Transport(ps.transport).
		Header("Authorization", string(jwtToken)).
		BodyJSON(&card).
		Method(http.MethodPost).
		Fetch(context.Background())

	if err != nil {
		return err
	} else {
		log.Println("Card info is updated.")
	}
	return nil
}
