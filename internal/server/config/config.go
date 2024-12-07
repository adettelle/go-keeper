// Package config provides functionality for loading and managing application configuration
// from environment variables and defaults. It also includes helpers for constructing
// connection strings and validating configuration values.
package config

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Address string `envconfig:"ADDRESS" default:"localhost:8080"`

	DBHost     string `envconfig:"DATABASE_HOST" default:"localhost"`
	DBPort     string `envconfig:"DATABASE_PORT" default:"5433"`
	DBUser     string `envconfig:"DATABASE_USER" default:"postgres"`
	DBPassword string `envconfig:"DATABASE_PASSWORD" default:"password"`
	DBName     string `envconfig:"DATABASE_NAME" default:"postgres"`

	MinioEndPoint        string `envconfig:"MINIO_ENDPOINT" default:"localhost:9000"`
	JwtSignKey           string `envconfig:"JWT_SIGNKEY" required:"true"`
	MinioAccessKeyID     string `envconfig:"ACCESS_KEYID" required:"true"`
	MinioSecretAccessKey string `envconfig:"SECRET_ACCESSKEY" required:"true"`
	UseSSL               bool   `envconfig:"USE_SSL" default:"false"`
	BucketName           string `envconfig:"BUCKET_NAME" default:"test"`
}

func New() (*Config, error) {
	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		log.Println("error in process:", err)
		return nil, err
	}

	ensureAddrFLagIsCorrect(cfg.Address)

	return &cfg, nil
}

// ensureAddrFLagIsCorrect validates the format of the provided address,
// ensuring it contains a valid host and port.
func ensureAddrFLagIsCorrect(addr string) {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		log.Fatal(err)
	}

	_, err = strconv.Atoi(port)
	if err != nil {
		log.Fatal(fmt.Errorf("invalid port: '%s'", port))
	}
}

// DBConnStr constructs and returns the PostgreSQL database connection string
func (cfg *Config) DBConnStr() string {
	dbParams := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	return dbParams
}
