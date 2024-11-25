package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/adettelle/go-keeper/internal/database"
	"github.com/adettelle/go-keeper/internal/migrator"
	"github.com/adettelle/go-keeper/internal/repo"
	"github.com/adettelle/go-keeper/internal/server/api"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	dbParams := "host=localhost port=5433 user=postgres password=password dbname=praktikum-fin sslmode=disable"

	migrator.MustApplyMigrations(dbParams)

	db, err := database.Connect(dbParams)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// TODO унести все параметры в config
	endpoint := "localhost:9000"
	accessKeyID := "RPClJMVJmUJyRF2PgZSK"
	secretAccessKey := "qJa8Bl0VHixgkDoymsJC7yEgb88nPTUQsZNLPUBM"
	useSSL := false // true

	customerRepo := repo.NewCustomerRepo(db)
	pwdRepo := repo.NewPasswordRepo(db)
	fileRepo := repo.NewFileRepo(db)

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	handlers := api.NewCustomerHandlers(customerRepo, pwdRepo, fileRepo, minioClient, []byte("my_secret_key"))
	// убрать в config TODO []byte("my_secret_key")

	address := "localhost:8080"
	fmt.Println("Starting server at address:", address)

	r := api.NewRouter(handlers)

	err = http.ListenAndServe(address, r)
	if err != nil {
		log.Fatal(err)
	}
}
