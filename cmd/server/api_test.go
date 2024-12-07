package main

import (
	"context"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/adettelle/go-keeper/internal/client"
	"github.com/adettelle/go-keeper/internal/database"
	"github.com/adettelle/go-keeper/internal/server/config"
	"github.com/carlmjohnson/requests"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"
)

func Test(t *testing.T) {
	cfg := &config.Config{
		JwtSignKey:           "my_secret_key",
		MinioAccessKeyID:     "AAAAAAAAAAAAAAAAAAAA",
		MinioSecretAccessKey: "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
		DBPort:               "5433",
		DBHost:               "localhost",
		DBUser:               "postgres",
		DBPassword:           "password",
		DBName:               "postgres",
		Address:              "localhost:8088",
		MinioEndPoint:        "localhost:9000",
		BucketName:           "test",
	}

	connStr := cfg.DBConnStr()

	db, err := database.Connect(connStr)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec("drop database if exists e2e_test;")
	require.NoError(t, err)

	_, err = db.Exec("create database e2e_test;")
	require.NoError(t, err)

	cfg.DBName = "e2e_test"

	// defer func() {
	// 	_, err = db.Exec("drop database if exists e2e_test;")
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }()

	srv, err := initializeServer(cfg)
	require.NoError(t, err)

	go func() {
		_ = srv.ListenAndServe()
	}()
	time.Sleep(1 * time.Second)
	defer func() {
		log.Println("Closing server")
		srv.Close()
	}()

	name := uuid.NewString()
	login := name + "@google.com"

	customer := client.CustomerToReg{
		Name:           name,
		Login:          login,
		MasterPassword: uuid.NewString(),
		Authentication: false,
	}

	// Test registration
	err = requests.
		URL("/api/user/register").
		Scheme("http").
		CheckStatus(http.StatusOK).
		Host(cfg.Address).
		Method(http.MethodPost).
		BodyJSON(&customer).
		Fetch(context.Background())
	require.NoError(t, err)

	// Test login
	customerToLogin := client.CustomerToLogin{
		Login:    login,
		Password: customer.MasterPassword,
	}
	headers := http.Header{}
	err = requests.
		URL("/api/user/login").
		Scheme("http").
		CopyHeaders(headers).
		CheckStatus(http.StatusOK).
		Host(cfg.Address).
		BodyJSON(&customerToLogin).
		Fetch(context.Background())
	require.NoError(t, err)

	jwtToken := headers.Get("Authorization")
	require.NotEmpty(t, jwtToken)

	// ----------------------------------- password -----------------------------------
	// Test get all passwords, should be empty
	var pwds []client.PasswordToGet

	getAllPass(t, cfg, []byte(jwtToken), pwds, 0)

	// Test add password
	pwdToAdd := client.PwdToAdd{
		Password:    "password1",
		Title:       uuid.NewString(),
		Description: "pass 1",
	}
	pwdToAdd2 := client.PwdToAdd{
		Password:    "password1",
		Title:       uuid.NewString(),
		Description: "pass 1",
	}

	addPass(t, cfg, []byte(jwtToken), &pwdToAdd)
	addPass(t, cfg, []byte(jwtToken), &pwdToAdd2)

	// Test get all passwords, user should have two passwords
	getAllPass(t, cfg, []byte(jwtToken), pwds, 2)

	// Test get password by title
	var pwdToGet string

	getPass(t, cfg, []byte(jwtToken), pwdToGet, pwdToAdd.Title, "password1")

	// Test update password
	pwdToUpdate := client.PasswordToUpdate{
		Password: "newPassword1",
	}

	updatePass(t, cfg, []byte(jwtToken), pwdToAdd.Title, pwdToUpdate)

	// Test get password after update
	getPass(t, cfg, []byte(jwtToken), pwdToGet, pwdToAdd.Title, "newPassword1")

	// Test delete passwords
	err = requests.
		URL("/api/user/password/"+pwdToAdd.Title).
		Scheme("http").
		Host(cfg.Address).
		Header("Authorization", string(jwtToken)).
		Method(http.MethodDelete).
		Fetch(context.Background())

	// Test all passwords, user should have one password
	getAllPass(t, cfg, []byte(jwtToken), pwds, 1)

	// ----------------------------------- card -----------------------------------
	// Test get all cards, should be empty
	var cards []client.CardToGet

	getAllCards(t, cfg, []byte(jwtToken), cards, 0)

	// Test add card

	cardToAdd := client.CardToAdd{
		Num:         "3530111333300000",
		Expire:      "0131",
		Cvc:         "111",
		Title:       RandStringBytes(10),
		Description: "pass 1",
	}
	cardToAdd2 := client.CardToAdd{
		Num:         "3566002020360505",
		Expire:      "0232",
		Cvc:         "222",
		Title:       RandStringBytes(10),
		Description: "pass 2",
	}

	addCard(t, cfg, []byte(jwtToken), &cardToAdd)
	addCard(t, cfg, []byte(jwtToken), &cardToAdd2)

	// Test get all cards, user should have two cards
	getAllCards(t, cfg, []byte(jwtToken), cards, 2)

	// Test get card by title
	expectedCard := client.CardToGetByTitle{
		Num:         cardToAdd.Num,
		Expire:      cardToAdd.Expire,
		Cvc:         cardToAdd.Cvc,
		Description: cardToAdd.Description,
	}

	getCard(t, cfg, []byte(jwtToken), cardToAdd.Title, expectedCard)

	// Test update card
	cardToUpdate := client.CardToUpdate{
		Expire:      "0130",
		Description: "new description",
	}

	expectedCard2 := client.CardToGetByTitle{
		Num:         cardToAdd.Num,
		Expire:      cardToAdd.Expire,
		Cvc:         cardToAdd.Cvc,
		Description: cardToUpdate.Description,
	}
	updateCard(t, cfg, []byte(jwtToken), cardToAdd.Title, cardToUpdate)

	// Test get card after update
	getCard(t, cfg, []byte(jwtToken), cardToAdd.Title, expectedCard2)

	// Test delete card
	err = requests.
		URL("/api/user/card/"+cardToAdd.Title).
		Scheme("http").
		Host(cfg.Address).
		Header("Authorization", string(jwtToken)).
		Method(http.MethodDelete).
		Fetch(context.Background())

	// Test all cards, user should have one card
	getAllCards(t, cfg, []byte(jwtToken), cards, 1)

}

// ----------------------------------- password -----------------------------------

func getAllPass(t *testing.T, cfg *config.Config, jwtToken []byte,
	pwds []client.PasswordToGet, len int) {

	err := requests.
		URL("/api/user/passwords").
		Scheme("http").
		Host(cfg.Address).
		Header("Authorization", string(jwtToken)).
		ToJSON(&pwds).
		Method(http.MethodGet).
		Fetch(context.Background())
	require.NoError(t, err)
	require.Len(t, pwds, len)
}

func addPass(t *testing.T, cfg *config.Config, jwtToken []byte, pwd *client.PwdToAdd) {
	err := requests.
		URL("/api/user/password").
		Scheme("http").
		Host(cfg.Address).
		Header("Authorization", string(jwtToken)).
		BodyJSON(pwd).
		Method(http.MethodPut).
		Fetch(context.Background())
	require.NoError(t, err)
}

func getPass(t *testing.T, cfg *config.Config, jwtToken []byte,
	pwdToGet string, title string, expectedPass string) {

	err := requests.
		URL("/api/user/password/"+title).
		Scheme("http").
		Host(cfg.Address).
		Header("Authorization", string(jwtToken)).
		Method(http.MethodGet).
		ToString(&pwdToGet).
		Fetch(context.Background())
	require.NoError(t, err)
	require.Equal(t, expectedPass, pwdToGet)
}

func updatePass(t *testing.T, cfg *config.Config, jwtToken []byte,
	title string, pwdToUpdate client.PasswordToUpdate) {

	err := requests.
		URL("/api/user/password/update/"+title).
		Scheme("http").
		Host(cfg.Address).
		Header("Authorization", string(jwtToken)).
		BodyJSON(&pwdToUpdate).
		Method(http.MethodPost).
		Fetch(context.Background())
	require.NoError(t, err)
}

// ----------------------------------- card -----------------------------------
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func getAllCards(t *testing.T, cfg *config.Config, jwtToken []byte,
	cards []client.CardToGet, len int) {

	err := requests.
		URL("/api/user/cards").
		Scheme("http").
		Host(cfg.Address).
		Header("Authorization", string(jwtToken)).
		ToJSON(&cards).
		Method(http.MethodGet).
		Fetch(context.Background())
	require.NoError(t, err)
	require.Len(t, cards, len)
}

func addCard(t *testing.T, cfg *config.Config, jwtToken []byte, card *client.CardToAdd) {
	err := requests.
		URL("/api/user/card").
		Scheme("http").
		Host(cfg.Address).
		Header("Authorization", string(jwtToken)).
		BodyJSON(card).
		Method(http.MethodPut).
		Fetch(context.Background())
	require.NoError(t, err)
}

func getCard(t *testing.T, cfg *config.Config, jwtToken []byte,
	title string, expectedCard client.CardToGetByTitle) {

	var cardToGet client.CardToGetByTitle

	err := requests.
		URL("/api/user/card/"+title).
		Scheme("http").
		Host(cfg.Address).
		Header("Authorization", string(jwtToken)).
		Method(http.MethodGet).
		ToJSON(&cardToGet).
		Fetch(context.Background())
	require.NoError(t, err)

	log.Println(cardToGet)

	require.Equal(t, expectedCard, cardToGet)
}

func updateCard(t *testing.T, cfg *config.Config, jwtToken []byte,
	title string, cardToUpdate client.CardToUpdate) {

	err := requests.
		URL("/api/user/card/update/"+title).
		Scheme("http").
		Host(cfg.Address).
		Header("Authorization", string(jwtToken)).
		BodyJSON(&cardToUpdate).
		Method(http.MethodPost).
		Fetch(context.Background())
	require.NoError(t, err)
	log.Println("!!!!!!!!!", cardToUpdate)
}
