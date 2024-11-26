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
	"path/filepath"
	"time"

	"github.com/adettelle/go-keeper/internal/jwt"
	"github.com/adettelle/go-keeper/internal/repo"
	"github.com/adettelle/go-keeper/internal/server/config"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type ICustomerRepo interface {
	AddCustomer(ctx context.Context, name, login, masterPassword string) error
	GetCustomerByLogin(ctx context.Context, login string) (*repo.Customer, error)
}

type IPwdRepo interface {
	GetAllPasswords(ctx context.Context, name string) ([]repo.Password, error)
	CreatePassword(ctx context.Context, password, title, description string, login string) error
	UpdatePassword(ctx context.Context, id int, password, title, description *string) error
	DeletePassword(ctx context.Context, title string, login string) error
	GetPasswordByTitle(ctx context.Context, title string, login string) (string, error)
}

type IFileRepo interface {
	AddFile(ctx context.Context, fileName, title, description, cloudID string, login string) error
	GetFileCoudIDByID(ctx context.Context, fileID, login string) (string, error)
	GetAllFiles(ctx context.Context, name string) ([]repo.FileToGet, error)
}

type ICardRepo interface {
	AddCard(ctx context.Context, cardNum, expire, description, cvc string, login string) error
	GetCardByID(ctx context.Context, cardID, login string) (repo.CardGetByID, error)
	GetAllCards(ctx context.Context, login string) ([]repo.CardToGet, error)
}

// use a single instance of Validate, it caches struct info
var validate *validator.Validate = validator.New(validator.WithRequiredStructEnabled())

type CustomerHandlers struct {
	CustomerRepo ICustomerRepo
	PwdRepo      IPwdRepo
	FileRepo     IFileRepo
	CardRepo     ICardRepo
	MinioClient  *minio.Client
	JwtSignKey   []byte
	Config       *config.Config
}

func NewCustomerHandlers(
	customerRepo ICustomerRepo,
	pwdRepo IPwdRepo,
	fileRepo IFileRepo,
	cardRepo ICardRepo,
	minioClient *minio.Client,
	jwtSignKey []byte, cfg *config.Config) *CustomerHandlers {
	return &CustomerHandlers{
		CustomerRepo: customerRepo,
		PwdRepo:      pwdRepo,
		FileRepo:     fileRepo,
		CardRepo:     cardRepo,
		MinioClient:  minioClient,
		JwtSignKey:   jwtSignKey,
		Config:       cfg,
	}
}

type customerRegistrationRequestDTO struct {
	Name           string `json:"name"`
	Login          string `json:"login" validate:"required,email"`
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

func NewPwdDTO(pwd repo.Password) *pwdCreateRequestDTO {
	return &pwdCreateRequestDTO{
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

type pwdUpdateRequestDTO struct {
	ID          int     `json:"id"`
	Password    *string `json:"pwd"`
	Title       *string `json:"title"`
	Description *string `json:"description"` // для ссылки значение по умолчанию - nil
}

type fileCreateRequestDTO struct {
	FileName    string `json:"fname"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type fileGetRequestDTO struct {
	ID          string `json:"id"`
	FileName    string `json:"file_name"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func NewFileDTO(file repo.FileToGet) *fileGetRequestDTO {
	return &fileGetRequestDTO{
		ID:          file.ID,
		FileName:    file.FileName,
		Title:       file.Title,
		Description: file.Description,
	}
}

func NewFileListDTO(files []repo.FileToGet) []*fileGetRequestDTO {
	res := []*fileGetRequestDTO{}
	for _, file := range files {
		res = append(res, NewFileDTO(file))
	}
	return res
}

type cardCreateRequestDTO struct {
	Num         string `json:"num" validate:"required,credit_card"`
	Expire      string `json:"expire" validate:"required,numeric,len=4"`
	Description string `json:"description"`
	Cvc         string `json:"cvc" validate:"required,numeric,len=3"`
}

type cardGetRequestDTO struct {
	ID          string `json:"id"`
	Num         string `json:"num"`
	Expire      string `json:"expire"`
	Description string `json:"description"`
	Cvc         string `json:"cvc"`
}

func NewCardDTO(card repo.CardToGet) *cardGetRequestDTO {
	return &cardGetRequestDTO{
		ID:          card.ID,
		Num:         card.Num,
		Expire:      card.Expire,
		Description: card.Description,
		Cvc:         card.Cvc,
	}
}

func NewCardListDTO(cards []repo.CardToGet) []*cardGetRequestDTO {
	res := []*cardGetRequestDTO{}
	for _, card := range cards {
		res = append(res, NewCardDTO(card))
	}
	return res
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

	if !VerifyUser(context.Background(), auth.Login, auth.Password, ch.CustomerRepo) {
		log.Println("authlogin: ", auth.Login, auth.Password)
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
		context.Background(), customer.Name, customer.Login, hashStringPassword)
	if err != nil {
		if repo.IsCustomerExistsErr(err) { // !!!!!!!!!!!!!!!!!!!!!!!!!
			log.Printf("error %v in registering user %s", err, customer.Login)
			w.WriteHeader(http.StatusConflict)
			return
		}
		log.Println("error in adding user:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Println("customer: ", customer)

	token, err := jwt.GenerateJwtToken([]byte(ch.JwtSignKey), customer.Login)
	if err != nil {
		log.Println("error in generating token:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Authorization", "Bearer "+token)

	w.WriteHeader(http.StatusOK)
}

// VerifyUser — функция, которая выполняет аутентификацию и авторизацию пользователя
// login — это email пользователя, pass — это masterpassword, permission — необходимая привилегия.
// если пользователь ввел правильные данные, и у него есть необходимая привилегия — возвращаем true, иначе — false
func VerifyUser(ctx context.Context, login string, pass string, customerRepo ICustomerRepo) bool {
	if login == "" || pass == "" {
		return false
	}
	// получаем хеш пароля
	hashedPassword := sha256.Sum256([]byte(pass))
	hashStringPassword := hex.EncodeToString(hashedPassword[:]) // дополнительно кодируем пароль

	// проверяем введенные данные
	cust, err := customerRepo.GetCustomerByLogin(ctx, login)
	if err != nil {
		log.Printf("Error in authorization %s", cust.Login)
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
		log.Println("error in getting passwords: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Println("passwords: ", pwds)

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

// Хендлер доступен только авторизованному пользователю
func (ch *CustomerHandlers) FileAdd(w http.ResponseWriter, r *http.Request) {
	log.Println("In FileAdd")
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userLogin := r.Header.Get("x-user")

	var buf bytes.Buffer
	var file fileCreateRequestDTO

	// читаем тело запроса
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Println("error in reading body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &file); err != nil {
		log.Println("error in unmarshalling json:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cloudID := uuid.NewString() // генерируем случайную строку опр-го формата

	// убрем из filename полный путь, оставив только название файла
	fileNameWithoutPath := filepath.Base(file.FileName)

	err = ch.FileRepo.AddFile(
		context.Background(), fileNameWithoutPath, file.Title, file.Description, cloudID, userLogin)
	if err != nil {
		log.Println("error in adding password:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO сколько надо?
	url, err := ch.MinioClient.PresignedPutObject(context.Background(), "test", cloudID, 3*time.Minute)
	if err != nil {
		log.Println("error in generating upload url:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	type addFileResponseDTO struct {
		URL string
	}

	res := addFileResponseDTO{
		URL: url.String(),
	}
	resBody, err := json.Marshal(res)
	if err != nil {
		log.Println("error in marshalling json:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resBody)

	return
}

func (ch *CustomerHandlers) FileGetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userLogin := r.Header.Get("x-user")
	fileID := r.PathValue("id")

	fileCLoudID, err := ch.FileRepo.GetFileCoudIDByID(context.Background(), fileID, userLogin)
	if err != nil {
		log.Println("error in getting file by id:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO config test bucket name
	obj, err := ch.MinioClient.GetObject(context.Background(), ch.Config.BucketName, fileCLoudID, minio.GetObjectOptions{}) //  "test"
	if err != nil {
		log.Println("error in getting minio:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/binary")
	w.WriteHeader(http.StatusOK)

	_, err = io.Copy(w, obj)
	if err != nil {
		log.Println("error in copy response:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// Хендлер доступен только авторизованному пользователю
func (ch *CustomerHandlers) AllFiles(w http.ResponseWriter, r *http.Request) {
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

	files, err := ch.FileRepo.GetAllFiles(context.Background(), customer.Name)
	if err != nil {
		log.Println("error in getting files: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Println("files: ", files)
	if len(files) == 0 {
		log.Println("len(files) == 0")
		w.WriteHeader(http.StatusNoContent) // нет данных для ответа
		return
	}

	resp, err := json.Marshal(NewFileListDTO(files))
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

// Хендлер доступен только авторизованному пользователю
func (ch *CustomerHandlers) CardAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userLogin := r.Header.Get("x-user")

	var buf bytes.Buffer
	var card cardCreateRequestDTO

	// читаем тело запроса
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Println("error in reading body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &card); err != nil {
		log.Println("error in unmarshalling json:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = validate.Struct(card)
	if err != nil {
		log.Println("error in validating:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// if !luhn.CheckLuhn(card.Num) {
	// 	w.WriteHeader(http.StatusUnprocessableEntity) // неверный номер карты 422
	// 	return
	// }

	err = ch.CardRepo.AddCard(
		context.Background(), card.Num, card.Expire, card.Description, card.Cvc, userLogin)
	if err != nil {
		log.Println("error in adding card:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// // TODO сколько надо?
	// url, err := ch.MinioClient.PresignedPutObject(context.Background(), "test", cloudID, 3*time.Minute)
	// if err != nil {
	// 	log.Println("error in generating upload url:", err)
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }

	// type addFileResponseDTO struct {
	// 	URL string
	// }

	// res := addFileResponseDTO{
	// 	URL: url.String(),
	// }
	// resBody, err := json.Marshal(res)
	// if err != nil {
	// 	log.Println("error in marshalling json:", err)
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }
	// w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	return
}

// Хендлер доступен только авторизованному пользователю
func (ch *CustomerHandlers) AllCards(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	userLogin := r.Header.Get("x-user") // w.Header().Get("x-user")
	log.Println("!!!!!!!!!!!!!!!", userLogin)
	cards, err := ch.CardRepo.GetAllCards(context.Background(), userLogin)
	if err != nil {
		log.Println("error in getting card by login:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Println("cards in handler:", cards)
	if len(cards) == 0 {
		log.Println("len(cards) == 0")
		w.WriteHeader(http.StatusNoContent) // нет данных для ответа
		return
	}

	resp, err := json.Marshal(NewCardListDTO(cards))
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
