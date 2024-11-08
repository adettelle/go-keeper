package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/adettelle/go-keeper/internal/database"
	"github.com/adettelle/go-keeper/internal/migrator"
	"github.com/adettelle/go-keeper/internal/server/api"
)

func main() {
	dbParams := "host=localhost port=5433 user=postgres password=password dbname=praktikum-fin sslmode=disable"

	migrator.MustApplyMigrations(dbParams)

	db, err := database.Connect(dbParams)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	address := "localhost:8080"
	fmt.Println("Starting server at address:", address)

	r := api.NewRouter()

	err = http.ListenAndServe(address, r)
	if err != nil {
		log.Fatal(err)
	}
}
