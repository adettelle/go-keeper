package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	"github.com/adettelle/go-keeper/internal/jwt"
	"github.com/adettelle/go-keeper/internal/repo"
)

type CustomerHandlers struct {
	CustomerRepo ICustomerRepo
	PwdRepo      IPwdRepo
	JwtSignKey   []byte
}

func NewCustomerHandlers(customerRepo ICustomerRepo, pwdRepo IPwdRepo, jwtSignKey []byte) *CustomerHandlers {
	return &CustomerHandlers{
		CustomerRepo: customerRepo,
		PwdRepo:      pwdRepo,
		JwtSignKey:   jwtSignKey,
	}
}

type ICustomerRepo interface {
	AddCustomer(ctx context.Context, name, email, masterPassword string) error
	GetCustomerByLogin(ctx context.Context, login string) (*repo.Customer, error)
}

type IPwdRepo interface {
	GetAllPasswords(ctx context.Context, name string) ([]repo.Password, error)
	CreatePassword(ctx context.Context, password, title, description string, login string) error
	UpdatePassword(ctx context.Context, id int, password, title, description *string) error
	DeletePassword(ctx context.Context, title string, login string) error
	GetPasswordByTitle(ctx context.Context, title string, login string) (string, error)
}

type customerRegistrationRequestDTO struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	MasterPassword string `json:"masterpassword"`
}

type authRequestDTO struct {
	Login    string `json:"login"`
	Password string `json:"pwd"`
}

type pwdCreateRequestDTO struct {
	Password    string `json:"pwd"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type pwdUpdateRequestDTO struct {
	ID          int     `json:"id"`
	Password    *string `json:"pwd"`
	Title       *string `json:"title"`
	Description *string `json:"description"` // для ссылки значение по умолчанию - nil
}

func NewPwdDTO(pwd repo.Password) *pwdCreateRequestDTO {
	return &pwdCreateRequestDTO{
		Password:    pwd.Password,
		Title:       pwd.Title,
		Description: pwd.Description,
	}
}

func NewPwdListDTO(pwds []repo.Password) []*pwdCreateRequestDTO {
	res := []*pwdCreateRequestDTO{}
	for _, pwd := range pwds {
		res = append(res, NewPwdDTO(pwd))
	}

	return res
}

// Login происходит по логину (email) и паролю (masterPassword)
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

	if !VerifyUser(context.Background(), auth.Login, auth.Password, ch.CustomerRepo) {
		w.WriteHeader(http.StatusUnauthorized) // неверная пара логин/пароль
		return
	}

	token, err := jwt.GenerateJwtToken(ch.JwtSignKey, auth.Login)
	if err != nil {
		log.Println("error in generating token:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Authorization", "Bearer "+token)

	w.WriteHeader(http.StatusOK)
}

func (ch *CustomerHandlers) RegisterCustomer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var buf bytes.Buffer
	var customer customerRegistrationRequestDTO

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

	// получаем хеш пароля
	hashedPassword := sha256.Sum256([]byte(customer.MasterPassword))
	hashStringPassword := hex.EncodeToString(hashedPassword[:]) // дополнительно кодируем пароль

	err = ch.CustomerRepo.AddCustomer(
		context.Background(), customer.Name, customer.Email, hashStringPassword)
	if err != nil {
		if repo.IsCustomerExistsErr(err) { // !!!!!!!!!!!!!!!!!!!!!!!!!
			log.Printf("error %v in registering user %s", err, customer.Email)
			w.WriteHeader(http.StatusConflict)
			return
		}
		log.Println("error in adding user:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Println("customer: ", customer)

	token, err := jwt.GenerateJwtToken([]byte(ch.JwtSignKey), customer.Email)
	if err != nil {
		log.Println("error in generating token:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Authorization", "Bearer "+token)

	w.WriteHeader(http.StatusOK)
}

// VerifyUser — функция, которая выполняет аутентификацию и авторизацию пользователя
// login — email пользователя, pass — это masterpassword, permission — необходимая привилегия.
// если пользователь ввел правильные данные, и у него есть необходимая привилегия — возвращаем true, иначе — false
func VerifyUser(ctx context.Context, login string, pass string, customerRepo ICustomerRepo) bool {
	// получаем хеш пароля
	hashedPassword := sha256.Sum256([]byte(pass))
	hashStringPassword := hex.EncodeToString(hashedPassword[:]) // дополнительно кодируем пароль

	// проверяем введенные данные
	cust, err := customerRepo.GetCustomerByLogin(ctx, login)
	if err != nil {
		log.Printf("Error in authorization %s", cust.Email)
		return false
	}

	return cust.MasterPassword == hashStringPassword
}

// Хендлер доступен только авторизованному пользователю
func (ch *CustomerHandlers) AllPasswords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	userLogin := r.Header.Get("x-user")

	customer, err := ch.CustomerRepo.GetCustomerByLogin(context.Background(), userLogin)
	log.Println("user from get customer by login:", *customer)
	if err != nil {
		log.Println("error in getting user by login:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if customer == nil {
		log.Println("customer == nil")
		w.WriteHeader(http.StatusNotFound) // это значит, нет такого пользователя
		return
	}

	pwds, err := ch.PwdRepo.GetAllPasswords(context.Background(), customer.Name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Println("passwords: ", pwds)
	if len(pwds) == 0 {
		log.Println("len(passwords) == 0")
		w.WriteHeader(http.StatusNoContent) // нет данных для ответа
		return
	}

	resp, err := json.Marshal(NewPwdListDTO(pwds))
	if err != nil {
		log.Println("error in marshalling json:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(resp)
	if err != nil {
		log.Println("error in writing resp:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// надо ли проверять, что такой пароль уже есть у кого-то еще????????????????
// Хендлер доступен только авторизованному пользователю
func (ch *CustomerHandlers) PasswordCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userLogin := r.Header.Get("x-user")

	var buf bytes.Buffer
	var pwd pwdCreateRequestDTO

	// читаем тело запроса
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Println("error in reading body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &pwd); err != nil {
		log.Println("error in unmarshalling json:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = ch.PwdRepo.CreatePassword(
		context.Background(), pwd.Password, pwd.Title, pwd.Description, userLogin)
	if err != nil {
		log.Println("error in adding password:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted) // новый пароль принят
	return
}

// где использовать userLogin ?????????????????????????
// Хендлер доступен только авторизованному пользователю
func (ch *CustomerHandlers) PasswordUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	//userLogin := r.Header.Get("x-user")

	var buf bytes.Buffer
	var pwd pwdUpdateRequestDTO

	// читаем тело запроса
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Println("error in reading body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &pwd); err != nil {
		log.Println("error in unmarshalling json:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = ch.PwdRepo.UpdatePassword(
		context.Background(), pwd.ID, pwd.Password, pwd.Title, pwd.Description)
	if err != nil {
		log.Println("error in adding password:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted) // новый пароль принят
	return
}

// PasswordDelete удаляет
func (ch *CustomerHandlers) PasswordDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	pwdTitle := r.PathValue("title")
	userLogin := r.Header.Get("x-user")

	err := ch.PwdRepo.DeletePassword(context.Background(), pwdTitle, userLogin)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (ch *CustomerHandlers) PasswordByTitle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userLogin := r.Header.Get("x-user")
	pwdTitle := r.PathValue("title")

	pwd, err := ch.PwdRepo.GetPasswordByTitle(context.Background(), pwdTitle, userLogin)
	if err != nil {
		log.Println("error in getting password by title:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(pwd)
	if err != nil {
		log.Println("error in marshalling json:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(resp)
	if err != nil {
		log.Println("error in writing resp:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
