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
		// fmt.Println("Could not connect to localhost:8080/api/user/cards", err)
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
	Num         string `json:"num"`
	Expire      string `json:"expires_at"`
	Cvc         string `json:"cvc"`
	Title       string `json:"title"`
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
		// fmt.Println("Could not connect to localhost:8080/api/user/addcard: ", err)
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
		// fmt.Println("Could not connect to localhost:8080/api/user/delete/"+cardID, err)
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

	// var cardNum string
	var card CardToGetByTitle

	err = requests.
		URL("/api/user/card/"+title).
		Host(settings.ServerURL).
		Scheme("https").
		Transport(cs.transport).
		Header("Authorization", string(jwtToken)).
		Method(http.MethodGet).
		// ToJSON(&cardNum).
		ToJSON(&card).
		Fetch(context.Background())

	if err != nil {
		return err
	}
	fmt.Printf("%s, %s, %s, %s\n", card.Num, card.Expire, card.Cvc, card.Description)
	return nil
}

type CardToUpdate struct {
	Num         string `json:"num,omitempty"`
	Expire      string `json:"expires_at,omitempty"`
	Cvc         string `json:"cvc,omitempty"`
	Description string `json:"description,omitempty"`
}

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
		// fmt.Println("could not connect to localhost:8080/api/user/card/update", err)
		return err
	} else {
		log.Println("Card info is updated.")
	}
	return nil
}
