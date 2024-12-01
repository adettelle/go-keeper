package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/adettelle/go-keeper/internal/database"
	"github.com/adettelle/go-keeper/internal/migrator"
	"github.com/adettelle/go-keeper/internal/repo"
	"github.com/adettelle/go-keeper/internal/server/api"
	"github.com/adettelle/go-keeper/internal/server/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Println("error in config")
		log.Fatal(err)
	}
	log.Println("config:", cfg)

	connStr := cfg.DBConnStr()

	migrator.MustApplyMigrations(connStr) // dbParams)

	db, err := database.Connect(connStr) // dbParams)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	customerRepo := repo.NewCustomerRepo(db)
	pwdRepo := repo.NewPasswordRepo(db)
	fileRepo := repo.NewFileRepo(db)
	cardRepo := repo.NewCardRepo(db)
	jwtRepo := repo.NewJwtRepo(db)

	// Initialize minio client object.
	minioClient, err := minio.New(cfg.MinioEndPoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	handlers := api.NewCustomerHandlers(
		customerRepo, pwdRepo, fileRepo, cardRepo, jwtRepo, minioClient, []byte(cfg.JwtSignKey), cfg)

	address := cfg.Address // "localhost:8080"
	fmt.Println("Starting server at address:", address)

	r := api.NewRouter(handlers, jwtRepo)

	srv := &http.Server{
		Addr:    address,
		Handler: r,
	}

	err = srv.ListenAndServeTLS("./keys/server_cert.pem", "./keys/server_privatekey.pem")
	if err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
