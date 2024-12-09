// Server allows users to store logins, passwords, binary data and other private information safely and securely.
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/adettelle/go-keeper/internal/database"
	"github.com/adettelle/go-keeper/internal/migrator"
	"github.com/adettelle/go-keeper/internal/repo"
	"github.com/adettelle/go-keeper/internal/server/api"
	"github.com/adettelle/go-keeper/internal/server/config"
	"github.com/adettelle/go-keeper/internal/service"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Println("error in config")
		log.Fatal(err)
	}

	srv, err := initializeServer(cfg)
	if err != nil {
		log.Fatal(err)
	}

	err = srv.ListenAndServeTLS("./keys/server_cert.pem", "./keys/server_privatekey.pem")
	if err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func initializeServer(cfg *config.Config) (*http.Server, error) {
	connStr := cfg.DBConnStr()

	migrator.MustApplyMigrations(connStr)

	db, err := database.Connect(connStr)
	if err != nil {
		return nil, err
	}
	//defer db.Close() // TODO outside the function/pass db in params

	customerRepo := repo.NewCustomerRepo(db)
	pwdRepo := repo.NewPasswordRepo(db)
	fileRepo := repo.NewFileRepo(db)
	cardRepo := repo.NewCardRepo(db)
	jwtRepo := repo.NewJwtRepo(db)

	// Initialize minio client object.
	minioClient, err := minio.New(cfg.MinioEndPoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKeyID, cfg.MinioSecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	minioService := service.NewMinioService(minioClient, cfg.BucketName, 3*time.Minute)
	err = minioService.CreateBucket()
	if err != nil {
		return nil, err
	}
	fmt.Println("Starting minio service")

	handlers := api.NewCustomerHandlers(customerRepo, jwtRepo, []byte(cfg.SignKey), cfg)

	cardHandlers := api.NewCardHandlers(cardRepo, []byte(cfg.SignKey), cfg)
	passHandlers := api.NewPassHandlers(pwdRepo, []byte(cfg.SignKey), cfg)
	fileHandlers := api.NewFileHandlers(fileRepo, minioService, []byte(cfg.SignKey), cfg)

	address := cfg.Address
	fmt.Println("Starting server at address:", address)

	r := api.NewRouter(handlers, cardHandlers, passHandlers, fileHandlers, jwtRepo)

	srv := &http.Server{
		Addr:    address,
		Handler: r,
	}
	return srv, nil
}
