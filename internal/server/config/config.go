package config

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/kelseyhightower/envconfig"
)

const (
// defaultAddress = "localhost:8080"
// defaultDBParams = "host=localhost port=5433 user=postgres password=password dbname=praktikum-fin sslmode=disable"
// dbHost     = "localhost"
// dbPort     = "5433"
// dbUser     = "postgres"
// dbPassword = "password"
// dbName     = "praktikum-fin"
// minioEndpoint = "localhost:9000"
// bucketName = "test"
)

type Config struct {
	Address string `envconfig:"ADDRESS" default:"localhost:8080"`
	// DBParams string `envconfig:"DATABASE_DSN"`

	DBHost     string `envconfig:"DATABASE_HOST" default:"localhost"`
	DBPort     string `envconfig:"DATABASE_PORT" default:"5433"`
	DBUser     string `envconfig:"DATABASE_USER" default:"postgres"`
	DBPassword string `envconfig:"DATABASE_PASSWORD" default:"password"`
	DBName     string `envconfig:"DATABASE_NAME" default:"praktikum-fin"`

	MinioEndPoint   string `envconfig:"MINIO_ENDPOINT" default:"localhost:9000"`
	JwtSignKey      string `envconfig:"JWT_SIGNKEY" required:"true"`
	AccessKeyID     string `envconfig:"ACCESS_KEYID" required:"true"`
	SecretAccessKey string `envconfig:"SECRET_ACCESSKEY" required:"true"`
	UseSSL          bool   `envconfig:"USE_SSL" default:"false"`
	BucketName      string `envconfig:"BUCKET_NAME" default:"test"`
}

func New() (*Config, error) {
	var cfg Config

	// flag.StringVar(&cfg.Address, "a", cfg.Address, "Net address localhost:port")
	// flag.StringVar(&cfg.DBParams, "d", cfg.DBParams, "db connection parameters")
	// flag.StringVar(&cfg.JwtSignKey, "j", cfg.JwtSignKey, "jwt sign key")

	// flag.StringVar(&cfg.MinioEndPoint, "e", cfg.MinioEndPoint, "minio endpoint")
	// flag.StringVar(&cfg.AccessKeyID, "k", cfg.AccessKeyID, "minio access key ID")
	// flag.StringVar(&cfg.SecretAccessKey, "s", cfg.SecretAccessKey, "minio secret access key")
	// flag.BoolVar(&cfg.UseSSL, "u", cfg.UseSSL, "UseSSL")

	// flag.Parse()

	if err := envconfig.Process("", &cfg); err != nil {
		log.Println("error in process:", err)
		return nil, err
	}

	// if cfg.Address == "" {
	// 	cfg.Address = defaultAddress
	// }

	ensureAddrFLagIsCorrect(cfg.Address)

	// if cfg.MinioEndPoint == "" {
	// 	cfg.MinioEndPoint = minioEndpoint
	// }

	// if cfg.JwtSignKey == "" {
	// 	log.Fatal("no jwt sign key")
	// 	cfg.JwtSignKey = jwtSignKey
	// }

	// if cfg.AccessKeyID == "" {
	// 	//
	// 	cfg.AccessKeyID = accessKeyID
	// }

	// if cfg.SecretAccessKey == "" {
	// 	//
	// 	cfg.SecretAccessKey = secretAccessKey
	// }

	// непонятно про bool, как может быть не задан ??????
	// if cfg.UseSSL == nil {
	// 	*cfg.UseSSL = false
	// }

	return &cfg, nil
}

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

func (cfg *Config) DBConnStr() string {
	dbParams := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	// if cfg.DBParams == "" {
	// 	cfg.DBParams = dbParams
	// }
	return dbParams
}
