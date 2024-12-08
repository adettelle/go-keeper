package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/adettelle/go-keeper/internal/jwt"
	"github.com/adettelle/go-keeper/internal/repo"
	"github.com/adettelle/go-keeper/internal/server/config"
)

type ICustomerRepo interface {
	AddCustomer(ctx context.Context, name, login, masterPassword string) (int, error)
	GetCustomerByLogin(ctx context.Context, login string) (*repo.CustomerGetByLogin, error)
	VerifyUser(ctx context.Context, login string, pass string) (bool, int)
}

type IJwtRepo interface {
	AddJwtToken(ctx context.Context, custID int, token string) error // isvalid сначала всега true
}

type IMinioService interface {
	GetObject(fileCLoudID string) (io.Reader, error)
	Upload(fileCloudID string, reader io.Reader) error
}

type CustomerHandlers struct {
	CustomerRepo ICustomerRepo
	JwtRepo      IJwtRepo
	SignKey      []byte
	Config       *config.Config
}

func NewCustomerHandlers(
	customerRepo ICustomerRepo,
	jwtRepo IJwtRepo,
	signKey []byte, cfg *config.Config) *CustomerHandlers {
	return &CustomerHandlers{
		CustomerRepo: customerRepo,
		JwtRepo:      jwtRepo,
		SignKey:      signKey,
		Config:       cfg,
	}
}

type authRequestDTO struct {
	Login    string `json:"login" validate:"required,email"`
	Password string `json:"pwd" validate:"required,min=3"`
}

// Login происходит по логину и паролю (masterPassword)
func (ch *CustomerHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	var auth authRequestDTO

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Println("error in reading body:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &auth); err != nil {
		log.Println("error in unmarshalling json:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = validate.Struct(auth)
	if err != nil {
		log.Println("error in validating:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ok, custID := ch.CustomerRepo.VerifyUser(context.Background(), auth.Login, auth.Password)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized) // неверная пара логин/пароль
		return
	}

	token, err := jwt.GenerateJwtToken(ch.SignKey, auth.Login, custID)
	if err != nil {
		log.Println("error in generating token:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Authorization", "Bearer "+token)

	err = ch.JwtRepo.AddJwtToken(context.Background(), custID, token)
	if err != nil {
		log.Println("error in adding jwt token:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type CustomerRegistrationRequestDTO struct {
	Name           string `json:"name" validate:"required,min=1"`
	Login          string `json:"login" validate:"required,email"`
	MasterPassword string `json:"masterpassword" validate:"required,min=3"`
}

func (ch *CustomerHandlers) RegisterCustomer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var buf bytes.Buffer
	var customer CustomerRegistrationRequestDTO

	// читаем тело запроса
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Println("error in reading body:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &customer); err != nil {
		log.Println("error in unmarshalling json:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = validate.Struct(customer)
	if err != nil {
		log.Println("error in validating:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// получаем хеш пароля
	hashedPassword := sha256.Sum256([]byte(customer.MasterPassword))
	hashStringPassword := hex.EncodeToString(hashedPassword[:]) // дополнительно кодируем пароль

	_, err = ch.CustomerRepo.AddCustomer(
		context.Background(), customer.Name, customer.Login, hashStringPassword)
	if err != nil {
		if repo.IsCustomerExistsErr(err) {
			log.Printf("error %v in registering user %s", err, customer.Login)
			w.WriteHeader(http.StatusConflict)
			return
		}
		log.Println("error in adding user:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Println("customer: ", customer)

	w.WriteHeader(http.StatusOK)
}
